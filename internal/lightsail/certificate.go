package lightsail

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/aws-sdk-go-v2/service/lightsail/types"
	"github.com/aws/smithy-go"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/conns"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/create"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/tftags"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/verify"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceCertificate() *schema.Resource {
	return &schema.Resource{
		Create: resourceCertificateCreate,
		Read:   resourceCertificateRead,
		Update: resourceCertificateUpdate,
		Delete: resourceCertificateDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(2, 255),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z]`), "must begin with an alphabetic character"),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_\-.]+[^._\-]$`), "must contain only alphanumeric characters, underscores, hyphens, and dots"),
				),
			},
			"domain_name": {
				// AWS Provider 3.0.0 aws_route53_zone references no longer contain a
				// trailing period, no longer requiring a custom StateFunc
				// to prevent ACM API error
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringDoesNotMatch(regexp.MustCompile(`\.$`), "cannot end with a period"),
			},
			"subject_alternative_names": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 253),
						validation.StringDoesNotMatch(regexp.MustCompile(`\.$`), "cannot end with a period"),
					),
				},
				Set: schema.HashString,
			},
			"tags":     TagsSchema(),
			"tags_all": TagsSchemaComputed(),
			// additional info returned from the API
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_validation_options": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"resource_record_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"resource_record_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"resource_record_value": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Set: domainValidationOptionsHash,
			},
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceCertificateCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	req := lightsail.CreateCertificateInput{
		CertificateName: aws.String(d.Get("name").(string)),
		DomainName:      aws.String(d.Get("domain_name").(string)),
	}

	if v, ok := d.GetOk("subject_alternative_names"); ok {
		req.SubjectAlternativeNames = expandSubjectAlternativeNames(v)
	}

	if len(tags) > 0 {
		req.Tags = Tags(tags.IgnoreAWS())
	}

	resp, err := conn.CreateCertificate(context.TODO(), &req)
	if err != nil {
		return err
	}

	if len(resp.Operations) == 0 {
		return fmt.Errorf("No operations found for CreateCertificate request")
	}

	op := resp.Operations[0]
	d.SetId(d.Get("name").(string))

	err = waitLightsailOperation(conn, op.Id)
	if err != nil {
		return fmt.Errorf("Error waiting for Certificate (%s) to become ready: %s", d.Id(), err)
	}

	return resourceCertificateRead(d, meta)
}

func resourceCertificateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	resp, err := conn.GetCertificates(context.TODO(), &lightsail.GetCertificatesInput{
		CertificateName: aws.String(d.Id()),
	})

	if err != nil {
		var oe *smithy.OperationError
		if errors.As(err, &oe) {
			log.Printf("failed to call service: %s, operation: %s, error: %v", oe.Service(), oe.Operation(), oe.Unwrap())
		}
		d.SetId("")
		return nil
	}

	if len(resp.Certificates) == 0 {
		log.Printf("[WARN] Lightsail Certificate (%s) not found removing from state", d.Id())
		d.SetId("")
		return nil
	}

	i := resp.Certificates[0]

	d.Set("name", i.CertificateName)
	d.Set("domain_name", i.DomainName)
	d.Set("subject_alternative_names", flattenSubjectAlternativeNames(i.CertificateDetail))

	// additional attributes
	d.Set("arn", i.CertificateArn)
	d.Set("created_at", i.CertificateDetail.CreatedAt.Format(time.RFC3339))

	d.Set("domain_validation_options", flattenDomainValidationRecords(i.CertificateDetail.DomainValidationRecords))

	tags := KeyValueTags(i.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceCertificateDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn

	resp, err := conn.DeleteCertificate(context.TODO(), &lightsail.DeleteCertificateInput{
		CertificateName: aws.String(d.Id()),
	})

	if err != nil {
		return err
	}

	op := resp.Operations[0]

	err = waitLightsailOperation(conn, op.Id)
	if err != nil {
		return fmt.Errorf("Error waiting for Certificate (%s) to become ready: %s", d.Id(), err)
	}

	return nil
}

func resourceCertificateUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating Lightsail Certificate (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceCertificateRead(d, meta)
}

func domainValidationOptionsHash(v interface{}) int {
	m, ok := v.(map[string]interface{})

	if !ok {
		return 0
	}

	if v, ok := m["domain_name"].(string); ok {
		return create.StringHashcode(v)
	}

	return 0
}

func flattenSubjectAlternativeNames(cert *types.Certificate) []string {
	sans := cert.SubjectAlternativeNames
	vs := make([]string, 0)
	for _, v := range sans {
		if v != aws.ToString(cert.DomainName) {
			vs = append(vs, v)
		}
	}
	return vs
}

func flattenDomainValidationRecords(domainValidationRecords []types.DomainValidationRecord) []map[string]interface{} {
	var domainValidationResult []map[string]interface{}

	for _, o := range domainValidationRecords {
		if o.ResourceRecord != nil {
			validationOption := map[string]interface{}{
				"domain_name":           aws.ToString(o.DomainName),
				"resource_record_name":  aws.ToString(o.ResourceRecord.Name),
				"resource_record_type":  aws.ToString(o.ResourceRecord.Type),
				"resource_record_value": aws.ToString(o.ResourceRecord.Value),
			}
			domainValidationResult = append(domainValidationResult, validationOption)
		}
	}

	return domainValidationResult
}

func expandSubjectAlternativeNames(sans interface{}) []string {
	subjectAlternativeNames := make([]string, len(sans.(*schema.Set).List()))
	for i, sanRaw := range sans.(*schema.Set).List() {
		subjectAlternativeNames[i] = sanRaw.(string)
	}

	return subjectAlternativeNames
}
