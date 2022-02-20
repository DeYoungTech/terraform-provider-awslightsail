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

func ResourceStaticIPAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceStaticIPAttachmentCreate,
		Read:   resourceStaticIPAttachmentRead,
		Delete: resourceStaticIPAttachmentDelete,
        Importer: &schema.ResourceImporter{
                State: schema.ImportStatePassthrough,
        },

		Schema: map[string]*schema.Schema{
			"static_ip_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"instance_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceStaticIPAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn

	staticIpName := d.Get("static_ip_name").(string)
	log.Printf("[INFO] Attaching Lightsail Static IP: %q", staticIpName)
	resp, err := conn.AttachStaticIp(context.TODO(), &lightsail.AttachStaticIpInput{
		StaticIpName: aws.String(staticIpName),
		InstanceName: aws.String(d.Get("instance_name").(string)),
	})

	if err != nil {
		return err
	}

	op := resp.Operations[0]

	err = waitLightsailOperation(conn, op.Id)
	if err != nil {
		return fmt.Errorf("Error waiting for Static IP Attachent (%s) to become ready: %s", d.Id(), err)
	}

	d.SetId(staticIpName)

	return resourceStaticIPAttachmentRead(d, meta)
}

func resourceStaticIPAttachmentRead(d *schema.ResourceData, meta interface{}) error {
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

	if !*resp.StaticIp.IsAttached {
		log.Printf("[WARN] Lightsail Static IP (%s) is not attached, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("instance_name", resp.StaticIp.AttachedTo)
	d.Set("ip_address", resp.StaticIp.IpAddress)
	d.Set("static_ip_name", d.Id())

	return nil
}

func resourceStaticIPAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn

	resp, err := conn.DetachStaticIp(context.TODO(), &lightsail.DetachStaticIpInput{
		StaticIpName: aws.String(d.Id()),
	})

	if err != nil {
		return err
	}

	op := resp.Operations[0]

	err = waitLightsailOperation(conn, op.Id)
	if err != nil {
		return fmt.Errorf("Error waiting for Static IP Attachent (%s) to become ready: %s", d.Id(), err)
	}

	return nil
}
