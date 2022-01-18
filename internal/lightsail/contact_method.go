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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceContactMethod() *schema.Resource {
	return &schema.Resource{
		Create: resourceContactMethodCreate,
		Read:   resourceContactMethodRead,
		Delete: resourceContactMethodDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"endpoint": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The destination of the contact method, such as an email address or a mobile phone number. Use the E.164 format when specifying a mobile phone number.",
			},
			"protocol": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The protocol of the contact method, such as Email or SMS (text messaging).",
				ValidateFunc: validation.StringInSlice([]string{
					"SMS",
					"Email",
				}, false),
			},
		},
	}
}

func resourceContactMethodCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn

	resp, err := conn.CreateContactMethod(context.TODO(), &lightsail.CreateContactMethodInput{
		ContactEndpoint: aws.String(d.Get("endpoint").(string)),
		Protocol:        types.ContactProtocol(d.Get("protocol").(string)),
	})

	if err != nil {
		return err
	}

	op := resp.Operations[0]

	err = waitLightsailOperation(conn, op.Id)
	if err != nil {
		return fmt.Errorf("Error waiting for Contact Method (%s) to become ready: %s", d.Id(), err)
	}

	d.SetId(d.Get("protocol").(string))

	return resourceContactMethodRead(d, meta)
}

func resourceContactMethodRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn

	protocols := make([]types.ContactProtocol, 0, 1)

	resp, err := conn.GetContactMethods(context.TODO(), &lightsail.GetContactMethodsInput{
		Protocols: append(protocols, types.ContactProtocol(d.Id())),
	})

	if err != nil {
		var oe *smithy.OperationError
		if errors.As(err, &oe) {
			log.Printf("failed to call service: %s, operation: %s, error: %v", oe.Service(), oe.Operation(), oe.Unwrap())
		}
		d.SetId("")
		return nil
	}

	var entry types.ContactMethod
	entryExists := false

	for _, n := range resp.ContactMethods {
		if d.Id() == string(n.Protocol) {
			entry = n
			entryExists = true
			break
		}
	}

	if !entryExists {
		d.SetId("")
		return nil
	}
	d.Set("protocol", entry.Protocol)
	d.Set("endpoint", string(aws.ToString(entry.ContactEndpoint)))

	return nil
}

func resourceContactMethodDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn

	resp, err := conn.DeleteContactMethod(context.TODO(), &lightsail.DeleteContactMethodInput{
		Protocol: types.ContactProtocol(d.Id()),
	})

	if err != nil {
		return err
	}

	op := resp.Operations[0]

	err = waitLightsailOperation(conn, op.Id)
	if err != nil {
		return fmt.Errorf("Error waiting for Contact Method (%s) to be deleted: %s", d.Id(), err)
	}

	return nil
}
