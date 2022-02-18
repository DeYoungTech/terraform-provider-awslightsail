package lightsail

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/smithy-go"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/conns"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceDiskAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceDiskAttachmentCreate,
		Read:   resourceDiskAttachmentRead,
		Delete: resourceDiskAttachmentDelete,

		Schema: map[string]*schema.Schema{
			"disk_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"instance_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"disk_path": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceDiskAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn

	resp, err := conn.AttachDisk(context.TODO(), &lightsail.AttachDiskInput{
		DiskName:     aws.String(d.Get("disk_name").(string)),
		InstanceName: aws.String(d.Get("instance_name").(string)),
		DiskPath:     aws.String(d.Get("disk_path").(string)),
	})

	if err != nil {
		return err
	}

	op := resp.Operations[0]

	err = waitLightsailOperation(conn, op.Id)
	if err != nil {
		return fmt.Errorf("Error waiting for disk Attachent (%s) to become ready: %s", d.Id(), err)
	}

	// Generate an ID
	vars := []string{
		d.Get("disk_name").(string),
		d.Get("instance_name").(string),
	}

	d.SetId(strings.Join(vars, "_"))

	return resourceDiskAttachmentRead(d, meta)
}

func resourceDiskAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn

	id_parts := strings.SplitN(d.Id(), "_", -1)
	if len(id_parts) != 2 {
		return nil
	}

	dname := id_parts[0]
	iname := id_parts[1]

	resp, err := conn.GetDisk(context.TODO(), &lightsail.GetDiskInput{
		DiskName: aws.String(dname),
	})

	if err != nil {
		var oe *smithy.OperationError
		if errors.As(err, &oe) {
			log.Printf("failed to call service: %s, operation: %s, error: %v", oe.Service(), oe.Operation(), oe.Unwrap())
		}
		d.SetId("")
		return nil
	}

	disk := resp.Disk

	if !*disk.IsAttached {
		log.Printf("[WARN] Lightsail Disk (%s) is not attached, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if *disk.AttachedTo != iname {
		log.Printf("[WARN] Lightsail Disk (%s) is not attached, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("instance_name", disk.AttachedTo)
	d.Set("disk_name", disk.Name)
	d.Set("disk_path", disk.Path)

	return nil
}

func resourceDiskAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn

	id_parts := strings.SplitN(d.Id(), "_", -1)
	if len(id_parts) != 2 {
		return nil
	}

	dname := id_parts[0]
	iname := id_parts[1]

	// You must first Stop an instance in order to detach a disk.
	stopresp, stoperr := conn.StopInstance(context.TODO(), &lightsail.StopInstanceInput{
		InstanceName: aws.String(iname),
	})

	if stoperr != nil {
		return stoperr
	}

	stopop := stopresp.Operations[0]

	stoperr = waitLightsailOperation(conn, stopop.Id)
	if stoperr != nil {
		return fmt.Errorf("Error waiting for Instance (%s) to Stop: %s", iname, stoperr)
	}

	resp, err := conn.DetachDisk(context.TODO(), &lightsail.DetachDiskInput{
		DiskName: aws.String(dname),
	})

	if err != nil {
		return err
	}

	op := resp.Operations[0]

	err = waitLightsailOperation(conn, op.Id)
	if err != nil {
		return fmt.Errorf("Error waiting for disk Attachent (%s) to become ready: %s", d.Id(), err)
	}

	startresp, starterr := conn.StartInstance(context.TODO(), &lightsail.StartInstanceInput{
		InstanceName: aws.String(iname),
	})

	if starterr != nil {
		return err
	}

	startop := startresp.Operations[0]

	starterr = waitLightsailOperation(conn, startop.Id)
	if starterr != nil {
		return fmt.Errorf("Error waiting for Instance to Start (%s) to become ready: %s", iname, starterr)
	}

	return nil
}
