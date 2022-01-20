package lightsail

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/conns"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceLoadBalancerAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceLoadBalancerAttachmentCreate,
		Read:   resourceLoadBalancerAttachmentRead,
		Delete: resourceLoadBalancerAttachmentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"load_balancer_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(2, 255),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z]`), "must begin with an alphabetic character"),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_\-.]+[^._\-]$`), "must contain only alphanumeric characters, underscores, hyphens, and dots"),
				),
			},
			"instance_name": {
				Type:     schema.TypeString,
				Set:      schema.HashString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceLoadBalancerAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn
	instanceNames := expandInstanceNames(d.Get("instance_name").(string))

	req := lightsail.AttachInstancesToLoadBalancerInput{
		LoadBalancerName: aws.String(d.Get("load_balancer_name").(string)),
		InstanceNames:    instanceNames,
	}

	resp, err := conn.AttachInstancesToLoadBalancer(context.TODO(), &req)
	if err != nil {
		return err
	}

	if len(resp.Operations) == 0 {
		return fmt.Errorf("No operations found for AttachInstancesToLoadBalancer request")
	}

	op := resp.Operations[0]

	err = waitLightsailOperation(conn, op.Id)
	if err != nil {
		return fmt.Errorf("Error waiting for load balancer attatchment (%s) to become ready: %s", d.Id(), err)
	}

	// Generate an ID
	vars := []string{
		d.Get("load_balancer_name").(string),
		d.Get("instance_name").(string),
	}

	d.SetId(strings.Join(vars, "_"))

	return resourceLoadBalancerAttachmentRead(d, meta)
}

func resourceLoadBalancerAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn

	id_parts := strings.SplitN(d.Id(), "_", -1)
	if len(id_parts) != 2 {
		return nil
	}

	lbname := id_parts[0]
	iname := id_parts[1]

	resp, err := conn.GetLoadBalancer(context.TODO(), &lightsail.GetLoadBalancerInput{
		LoadBalancerName: aws.String(lbname),
	})

	if err != nil {
		return err
	}

	var entry string
	entryExists := false

	for _, n := range resp.LoadBalancer.InstanceHealthSummary {
		if iname == *n.InstanceName {
			entry = *n.InstanceName
			entryExists = true
			break
		}
	}

	if !entryExists {
		d.SetId("")
		return nil
	}

	d.Set("load_balancer_name", lbname)
	d.Set("instance_name", entry)

	return nil
}

func resourceLoadBalancerAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn

	id_parts := strings.SplitN(d.Id(), "_", -1)
	if len(id_parts) != 2 {
		return nil
	}

	lbname := id_parts[0]
	iname := id_parts[1]

	req := lightsail.DetachInstancesFromLoadBalancerInput{
		LoadBalancerName: aws.String(lbname),
		InstanceNames:    expandInstanceNames(iname),
	}

	resp, err := conn.DetachInstancesFromLoadBalancer(context.TODO(), &req)

	if err != nil {
		return err
	}

	if len(resp.Operations) == 0 {
		return fmt.Errorf("No operations found for DetachInstancesFromLoadBalancer request")
	}

	op := resp.Operations[0]

	err = waitLightsailOperation(conn, op.Id)
	if err != nil {
		return fmt.Errorf("Error waiting for load balancer attatchment (%s) to become detached: %s", d.Id(), err)
	}

	return err
}

func expandInstanceNames(iname string) []string {
	instanceNames := make([]string, 1)
	instanceNames[0] = iname
	return instanceNames
}
