package lightsail

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/aws-sdk-go-v2/service/lightsail/types"
	"github.com/aws/smithy-go"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/conns"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/tftags"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/verify"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceContainerService() *schema.Resource {
	return &schema.Resource{
		Create: resourceContainerServiceCreate,
		Read:   resourceContainerServiceRead,
		Update: resourceContainerServiceUpdate,
		Delete: resourceContainerServiceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			// required fields
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"power": {
				Type:     schema.TypeString,
				Required: true,
			},
			"scale": {
				Type:     schema.TypeInt,
				Required: true,
			},
			// optional
			"is_disabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			// additional fields returned by lightsail container API
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"power_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"principal_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},

			// tags
			"tags":     TagsSchema(),
			"tags_all": TagsSchemaComputed(),
		},
	}
}

func resourceContainerServiceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	serviceName := aws.String(d.Get("name").(string))

	containerServiceInput := lightsail.CreateContainerServiceInput{
		ServiceName: serviceName,
		Power:       types.ContainerServicePowerName(d.Get("power").(string)),
		Scale:       aws.Int32(int32(d.Get("scale").(int))),
	}

	if len(tags) > 0 {
		containerServiceInput.Tags = Tags(tags.IgnoreAWS())
	}

	if _, err := conn.CreateContainerService(context.TODO(), &containerServiceInput); err != nil {
		log.Printf("[ERROR] Lightsail Container Service (%s) create failed: %s", aws.ToString(serviceName), err)
		return err
	}

	d.SetId(d.Get("name").(string))
	log.Printf("[INFO] Lightsail Container Service (%s) CreateContainerService call successful, now waiting for ContainerServiceState change", d.Id())

	err := waitContainerService(conn, serviceName)
	if err != nil {
		log.Printf("[ERROR] Container Service (%s) failed to become ready: %s", d.Id(), err)
		return err
	}

	log.Printf("[INFO] Lightsail Container Service (%s) create successful", d.Id())

	// once container service creation and/or deployment successful (now enabled by default), disable it if "is_disabled" is true
	if isDisabled := d.Get("is_disabled"); isDisabled.(bool) {
		updateContainerServiceInput := lightsail.UpdateContainerServiceInput{
			ServiceName: serviceName,
			IsDisabled:  aws.Bool(true),
		}

		if _, err := conn.UpdateContainerService(context.TODO(), &updateContainerServiceInput); err != nil {
			log.Printf("[ERROR] Lightsail Container Service (%s) create and/or deployment successful, but disabling it failed: %s", d.Id(), err)
			return err
		}

		err := waitContainerService(conn, serviceName)
		if err != nil {
			log.Printf("[ERROR] Container Service (%s) failed to become ready: %s", d.Id(), err)
			return err
		}
	}

	return resourceContainerServiceRead(d, meta)
}

func resourceContainerServiceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	resp, err := conn.GetContainerServices(context.TODO(),
		&lightsail.GetContainerServicesInput{
			ServiceName: aws.String(d.Id()),
		},
	)

	if err != nil {
		var oe *smithy.OperationError
		if errors.As(err, &oe) {
			log.Printf("failed to call service: %s, operation: %s, error: %v", oe.Service(), oe.Operation(), oe.Unwrap())
		}
		d.SetId("")
		return nil
	}
	// just look at index 0 because we only looked up 1 container service
	cs := resp.ContainerServices[0]

	d.Set("name", cs.ContainerServiceName)
	d.Set("power", cs.Power)
	d.Set("scale", cs.Scale)
	d.Set("is_disabled", cs.IsDisabled)
	d.Set("arn", cs.Arn)
	d.Set("availability_zone", cs.Location.AvailabilityZone)
	d.Set("power_id", cs.PowerId)
	d.Set("principal_arn", cs.PrincipalArn)
	d.Set("private_domain_name", cs.PrivateDomainName)
	d.Set("resource_type", cs.ResourceType)
	d.Set("state", cs.State)
	d.Set("url", cs.Url)

	tags := KeyValueTags(cs.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceContainerServiceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn
	serviceName := aws.String(d.Id())
	requestUpdate := false

	req := lightsail.UpdateContainerServiceInput{
		ServiceName: serviceName,
	}

	if d.HasChange("is_disabled") {
		req.IsDisabled = aws.Bool(d.Get("is_disabled").(bool))
		requestUpdate = true
	}

	if d.HasChange("power") {
		req.Power = types.ContainerServicePowerName(d.Get("power").(string))
		requestUpdate = true
	}

	if d.HasChange("scale") {
		req.Scale = aws.Int32(int32(d.Get("scale").(int)))
		requestUpdate = true
	}

	if requestUpdate {
		_, err := conn.UpdateContainerService(context.TODO(), &req)

		if err != nil {
			return err
		}

		err = waitContainerService(conn, serviceName)
		if err != nil {
			log.Printf("[ERROR] Container Service (%s) failed to become ready: %s", d.Id(), err)
			return err
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating Lightsail Container Service (%s) tags: %s", d.Id(), err)
		}
	}

	log.Printf("[INFO] Lightsail Container Service (%s) update successful", d.Id())
	return resourceContainerServiceRead(d, meta)
}

func resourceContainerServiceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn
	serviceName := aws.String(d.Id())

	req := lightsail.DeleteContainerServiceInput{
		ServiceName: serviceName,
	}

	if _, err := conn.DeleteContainerService(context.TODO(), &req); err != nil {
		log.Printf("[ERROR] Lightsail Container Service (%s) delete failed: %s", d.Id(), err)
		return err
	}

	return nil
}
