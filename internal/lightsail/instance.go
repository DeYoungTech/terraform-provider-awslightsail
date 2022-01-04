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
	"github.com/aws/smithy-go"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/tftags"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceInstanceCreate,
		Read:   resourceInstanceRead,
		Update: resourceInstanceUpdate,
		Delete: resourceInstanceDelete,
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
			"availability_zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"blueprint_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"bundle_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			// Optional attributes
			"key_pair_name": {
				// Not compatible with aws_key_pair (yet)
				// We'll need a new awslightsail_key_pair resource
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == "LightsailDefaultKeyPair" && new == "" {
						return true
					}
					return false
				},
			},

			// cannot be retrieved from the API
			"user_data": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			// additional info returned from the API
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cpu_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"ram_size": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"ipv6_address": {
				Type:       schema.TypeString,
				Computed:   true,
				Deprecated: "use `ipv6_addresses` attribute instead",
			},
			"ipv6_addresses": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"is_static_ip": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"private_ip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_ip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"username": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": TagsSchema(),
		},
	}
}

func resourceInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*lightsail.Client)
	iName := d.Get("name").(string)
	// tags := d.Get("tags")
	tags := tftags.New(d.Get("tags").(map[string]interface{}))

	req := lightsail.CreateInstancesInput{
		AvailabilityZone: aws.String(d.Get("availability_zone").(string)),
		BlueprintId:      aws.String(d.Get("blueprint_id").(string)),
		BundleId:         aws.String(d.Get("bundle_id").(string)),
		InstanceNames:    []string{iName},
	}

	if len(tags) > 0 {
		req.Tags = Tags(tags.IgnoreAWS())
	}

	if v, ok := d.GetOk("key_pair_name"); ok {
		req.KeyPairName = aws.String(v.(string))
	}
	if v, ok := d.GetOk("user_data"); ok {
		req.UserData = aws.String(v.(string))
	}

	resp, err := conn.CreateInstances(context.TODO(), &req)
	if err != nil {
		return err
	}

	if len(resp.Operations) == 0 {
		return fmt.Errorf("No operations found for CreateInstance request")
	}

	op := resp.Operations[0]
	d.SetId(d.Get("name").(string))

	err = waitLightsailOperation(conn, op.Id)
	if err != nil {
		return fmt.Errorf("Error waiting for Relational Database (%s) to become ready: %s", d.Id(), err)
	}

	return resourceInstanceRead(d, meta)
}

func resourceInstanceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*lightsail.Client)

	resp, err := conn.GetInstance(context.TODO(), &lightsail.GetInstanceInput{
		InstanceName: aws.String(d.Id()),
	})

	if err != nil {
		var oe *smithy.OperationError
		if errors.As(err, &oe) {
			log.Printf("failed to call service: %s, operation: %s, error: %v", oe.Service(), oe.Operation(), oe.Unwrap())
		}
		d.SetId("")
		return nil
	}

	if resp == nil {
		log.Printf("[WARN] Lightsail Instance (%s) not found, nil response from server, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	i := resp.Instance

	d.Set("availability_zone", i.Location.AvailabilityZone)
	d.Set("blueprint_id", i.BlueprintId)
	d.Set("bundle_id", i.BundleId)
	d.Set("key_pair_name", i.SshKeyName)
	d.Set("name", i.Name)

	// additional attributes
	d.Set("arn", i.Arn)
	d.Set("username", i.Username)
	d.Set("created_at", i.CreatedAt.Format(time.RFC3339))
	d.Set("cpu_count", i.Hardware.CpuCount)
	d.Set("ram_size", i.Hardware.RamSizeInGb)

	// Deprecated: AWS Go SDK v1.36.25 removed Ipv6Address field
	if len(i.Ipv6Addresses) > 0 {
		d.Set("ipv6_address", i.Ipv6Addresses[0])
	}

	d.Set("ipv6_addresses", aws.StringSlice(i.Ipv6Addresses))
	d.Set("is_static_ip", i.IsStaticIp)
	d.Set("private_ip_address", i.PrivateIpAddress)
	d.Set("public_ip_address", i.PublicIpAddress)

	tags := KeyValueTags(i.Tags).IgnoreAWS()

	//lintignore:AWSR002
	if err := d.Set("tags", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func resourceInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*lightsail.Client)

	resp, err := conn.DeleteInstance(context.TODO(), &lightsail.DeleteInstanceInput{
		InstanceName: aws.String(d.Id()),
	})

	if err != nil {
		return err
	}

	op := resp.Operations[0]

	err = waitLightsailOperation(conn, op.Id)
	if err != nil {
		return fmt.Errorf("Error waiting for Relational Database (%s) to become ready: %s", d.Id(), err)
	}

	return nil
}

func resourceInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*lightsail.Client)

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating Lightsail Instance (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceInstanceRead(d, meta)
}
