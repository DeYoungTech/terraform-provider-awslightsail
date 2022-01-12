package lightsail

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/smithy-go"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/conns"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/tftags"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/verify"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceDomainCreate,
		Read:   resourceDomainRead,
		Update: resourceDomainUpdate,
		Delete: resourceDomainDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     TagsSchema(),
			"tags_all": TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDomainCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	req := lightsail.CreateDomainInput{
		DomainName: aws.String(d.Get("domain_name").(string)),
	}

	if len(tags) > 0 {
		req.Tags = Tags(tags.IgnoreAWS())
	}

	_, err := conn.CreateDomain(context.TODO(), &req)

	if err != nil {
		return err
	}

	d.SetId(d.Get("domain_name").(string))

	return resourceDomainRead(d, meta)
}

func resourceDomainRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	resp, err := conn.GetDomain(context.TODO(), &lightsail.GetDomainInput{
		DomainName: aws.String(d.Id()),
	})

	if err != nil {
		var oe *smithy.OperationError
		if errors.As(err, &oe) {
			log.Printf("failed to call service: %s, operation: %s, error: %v", oe.Service(), oe.Operation(), oe.Unwrap())
		}
		d.SetId("")
		return nil
	}

	domain := resp.Domain

	d.Set("arn", domain.Arn)
	d.Set("domain_name", domain.Name)
	d.SetId(d.Get("domain_name").(string))

	tags := KeyValueTags(domain.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}
	return nil
}

func resourceDomainUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating Lightsail Instance (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceDomainRead(d, meta)
}

func resourceDomainDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn
	_, err := conn.DeleteDomain(context.TODO(), &lightsail.DeleteDomainInput{
		DomainName: aws.String(d.Id()),
	})

	return err
}
