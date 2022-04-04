package lightsail

import (
	"context"
	"errors"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/smithy-go"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/conns"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceContainerPublicDomainNames() *schema.Resource {
	return &schema.Resource{
		Create: resourceContainerPublicDomainNamesCreate,
		Read:   resourceContainerPublicDomainNamesRead,
		Update: resourceContainerPublicDomainNamesUpdate,
		Delete: resourceContainerPublicDomainNamesDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		// CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			// required fields
			"container_service_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"public_domain_names": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"certificate": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"certificate_name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"domain_names": {
										Type:     schema.TypeList,
										Required: true,
										MinItems: 1,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
func resourceContainerPublicDomainNamesCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn
	serviceName := aws.String(d.Get("container_service_name").(string))

	UpdateContainerServiceInput := lightsail.UpdateContainerServiceInput{
		ServiceName:       serviceName,
		PublicDomainNames: expandLightsailContainerServicePublicDomainNames(d.Get("public_domain_names")),
	}

	if _, err := conn.UpdateContainerService(context.TODO(), &UpdateContainerServiceInput); err != nil {
		log.Printf("[ERROR] Lightsail Container Service (%s) failed to update Public Domain Names: %s", aws.ToString(serviceName), err)
		return err
	}

	d.SetId(d.Get("container_service_name").(string))
	log.Printf("[INFO] Lightsail Container Service (%s) added Public Domain Names", d.Id())

	err := waitContainerService(conn, serviceName)
	if err != nil {
		log.Printf("[ERROR] Container Service (%s) failed to become ready: %s", d.Id(), err)
		return err
	}

	log.Printf("[INFO] Lightsail Container Service (%s) Public Domain Names set Successfully", d.Id())

	return resourceContainerPublicDomainNamesRead(d, meta)
}

func resourceContainerPublicDomainNamesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn

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

	cs := resp.ContainerServices[0]

	d.Set("container_service_name", cs.ContainerServiceName)
	if err := d.Set("public_domain_names", flattenLightsailContainerServicePublicDomainNames(cs.PublicDomainNames)); err != nil {
		log.Printf("[ERROR] setting public_domain_names for Lightsail Container Service (%s): %s", d.Id(), err)
		return err
	}

	return nil
}

func resourceContainerPublicDomainNamesUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn
	serviceName := aws.String(d.Id())
	requestUpdate := false

	req := lightsail.UpdateContainerServiceInput{
		ServiceName: serviceName,
	}

	if d.HasChange("public_domain_names") {
		requestUpdate = true
	}

	if requestUpdate {
		req.PublicDomainNames = expandLightsailContainerServicePublicDomainNames(d.Get("public_domain_names"))

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

	log.Printf("[INFO] Lightsail Container Service: (%s) failed to update Public Domain Names", d.Id())
	return resourceContainerPublicDomainNamesRead(d, meta)
}

func resourceContainerPublicDomainNamesDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn
	serviceName := aws.String(d.Id())

	req := lightsail.UpdateContainerServiceInput{
		ServiceName:       serviceName,
		PublicDomainNames: make(map[string][]string),
	}

	_, err := conn.UpdateContainerService(context.TODO(), &req)

	if err != nil {
		return err
	}

	err = waitContainerService(conn, serviceName)
	if err != nil {
		log.Printf("[ERROR] Container Service (%s) failed to remove Public Domain Names: %s", d.Id(), err)
		return err
	}

	return nil
}

func flattenLightsailContainerServicePublicDomainNames(domainNames map[string][]string) []interface{} {
	if domainNames == nil {
		return []interface{}{}
	}

	var rawCertificates []interface{}

	for certName, domains := range domainNames {
		rawCertificate := map[string]interface{}{
			"certificate_name": certName,
			"domain_names":     domains,
		}

		rawCertificates = append(rawCertificates, rawCertificate)
	}

	return []interface{}{
		map[string]interface{}{
			"certificate": rawCertificates,
		},
	}
}

func expandLightsailContainerServicePublicDomainNames(rawPublicDomainNames interface{}) map[string][]string {
	if len(rawPublicDomainNames.([]interface{})) == 0 {
		return nil
	}

	resultMap := make(map[string][]string)

	for _, rpdn := range rawPublicDomainNames.([]interface{}) {
		rpdnMap := rpdn.(map[string]interface{})

		rawCertificates := rpdnMap["certificate"].(*schema.Set).List()

		for _, rc := range rawCertificates {
			rcMap := rc.(map[string]interface{})

			var domainNames []string
			for _, rawDomainName := range rcMap["domain_names"].([]interface{}) {
				domainNames = append(domainNames, rawDomainName.(string))
			}

			certificateName := rcMap["certificate_name"].(string)

			resultMap[certificateName] = domainNames
		}
	}

	return resultMap
}
