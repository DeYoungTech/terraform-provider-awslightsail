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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceStaticIP() *schema.Resource {
	return &schema.Resource{
		Create: resourceStaticIPCreate,
		Read:   resourceStaticIPRead,
		Delete: resourceStaticIPDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"support_code": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceStaticIPCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn

	name := d.Get("name").(string)
	log.Printf("[INFO] Allocating Lightsail Static IP: %q", name)
	resp, err := conn.AllocateStaticIp(context.TODO(), &lightsail.AllocateStaticIpInput{
		StaticIpName: aws.String(name),
	})
	if err != nil {
		var oe *smithy.OperationError
		if errors.As(err, &oe) {
			log.Printf("failed to call service: %s, operation: %s, error: %v", oe.Service(), oe.Operation(), oe.Unwrap())
		}
		return err
	}

	op := resp.Operations[0]

	err = waitLightsailOperation(conn, op.Id)
	if err != nil {
		return fmt.Errorf("Error waiting for Static IP (%s) to become ready: %s", d.Id(), err)
	}

	d.SetId(name)

	return resourceStaticIPRead(d, meta)
}

func resourceStaticIPRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn

	log.Printf("[INFO] Reading Lightsail Static IP: %q", d.Id())
	resp, err := conn.GetStaticIp(context.TODO(), &lightsail.GetStaticIpInput{
		StaticIpName: aws.String(d.Id()),
	})

	if err != nil {
		var oe *smithy.OperationError
		if errors.As(err, &oe) {
			log.Printf("failed to call service: %s, operation: %s, error: %v", oe.Service(), oe.Operation(), oe.Unwrap())
		}
		d.SetId("")
		return nil
	}

	d.Set("arn", resp.StaticIp.Arn)
	d.Set("ip_address", resp.StaticIp.IpAddress)
	d.Set("support_code", resp.StaticIp.SupportCode)
    d.Set("name", resp.StaticIp.Name)

	return nil
}

func resourceStaticIPDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn

	name := d.Get("name").(string)
	log.Printf("[INFO] Deleting Lightsail Static IP: %q", name)
	resp, err := conn.ReleaseStaticIp(context.TODO(), &lightsail.ReleaseStaticIpInput{
		StaticIpName: aws.String(name),
	})
	if err != nil {
		return err
	}

	op := resp.Operations[0]

	err = waitLightsailOperation(conn, op.Id)
	if err != nil {
		return fmt.Errorf("Error waiting for Static IP (%s) to become ready: %s", d.Id(), err)
	}

	return nil
}
