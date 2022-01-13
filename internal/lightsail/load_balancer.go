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
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/conns"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/tftags"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/verify"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceLoadBalancer() *schema.Resource {
	return &schema.Resource{
		Create: resourceLoadBalancerCreate,
		Read:   resourceLoadBalancerRead,
		Update: resourceLoadBalancerUpdate,
		Delete: resourceLoadBalancerDelete,
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
			"health_check_path": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "/",
			},
			"instance_port": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(0, 65535),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"protocol": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_ports": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},
			"tags": TagsSchema(),
			"tags_all": TagsSchemaComputed(),
		},
	CustomizeDiff: verify.SetTagsDiff,
  	}
}

func resourceLoadBalancerCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	req := lightsail.CreateLoadBalancerInput{
		HealthCheckPath:  aws.String(d.Get("health_check_path").(string)),
		InstancePort:     *aws.Int32(int32(d.Get("instance_port").(int))),
		LoadBalancerName: aws.String(d.Get("name").(string)),
	}

	if len(tags) > 0 {
		req.Tags = Tags(tags.IgnoreAWS())
	}

	resp, err := conn.CreateLoadBalancer(context.TODO(), &req)
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

	return resourceLoadBalancerRead(d, meta)
}

func resourceLoadBalancerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	resp, err := conn.GetLoadBalancer(context.TODO(), &lightsail.GetLoadBalancerInput{
		LoadBalancerName: aws.String(d.Id()),
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
		log.Printf("[WARN] Lightsail Load Balancer (%s) not found, nil response from server, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	lb := resp.LoadBalancer


	d.Set("arn", lb.Arn)
	d.Set("created_at", lb.CreatedAt.Format(time.RFC3339))
	d.Set("health_check_path", lb.HealthCheckPath)
	d.Set("instance_port", lb.InstancePort)
	d.Set("name", lb.Name)
	d.Set("protocol", lb.Protocol)
	d.Set("public_ports", lb.PublicPorts)
	d.Set("dns_name", lb.DnsName)

	tags := KeyValueTags(lb.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceLoadBalancerDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn
	resp, err := conn.DeleteLoadBalancer(context.TODO(), &lightsail.DeleteLoadBalancerInput{
		LoadBalancerName: aws.String(d.Id()),
	})

	op := resp.Operations[0]

	if err != nil {
		return err
	}

	err = waitLightsailOperation(conn, op.Id)
	if err != nil {
		return fmt.Errorf("Error waiting for Relational Database (%s) to become ready: %s", d.Id(), err)
	}

	return err
}

func resourceLoadBalancerUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn

	if d.HasChange("health_check_path") {
		_, err := conn.UpdateLoadBalancerAttribute(context.TODO(), &lightsail.UpdateLoadBalancerAttributeInput{
			AttributeName:    "HealthCheckPath",
			AttributeValue:   aws.String(d.Get("health_check_path").(string)),
			LoadBalancerName: aws.String(d.Get("name").(string)),
		})
		d.Set("health_check_path", d.Get("health_check_path").(string))
		if err != nil {
			return err
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating Lightsail Instance (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceLoadBalancerRead(d, meta)
}