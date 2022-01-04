package lightsail

import (
	"context"
	"errors"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/smithy-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceDomainCreate,
		Read:   resourceDomainRead,
		Delete: resourceDomainDelete,

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
		},
	}
}

func resourceDomainCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*lightsail.Client)
	_, err := conn.CreateDomain(context.TODO(), &lightsail.CreateDomainInput{
		DomainName: aws.String(d.Get("domain_name").(string)),
	})

	if err != nil {
		return err
	}

	d.SetId(d.Get("domain_name").(string))

	return resourceDomainRead(d, meta)
}

func resourceDomainRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*lightsail.Client)
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

	d.Set("arn", resp.Domain.Arn)
	return nil
}

func resourceDomainDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*lightsail.Client)
	_, err := conn.DeleteDomain(context.TODO(), &lightsail.DeleteDomainInput{
		DomainName: aws.String(d.Id()),
	})

	return err
}
