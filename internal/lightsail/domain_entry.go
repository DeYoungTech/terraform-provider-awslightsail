package lightsail

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/aws-sdk-go-v2/service/lightsail/types"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/conns"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceDomainEntry() *schema.Resource {
	return &schema.Resource{
		Create: resourceDomainEntryCreate,
		Read:   resourceDomainEntryRead,
		Delete: resourceDomainEntryDelete,

		Schema: map[string]*schema.Schema{
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"is_alias": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"target": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"A",
					"CNAME",
					"MX",
					"NS",
					"SOA",
					"SRV",
					"TXT",
				}, false),
				ForceNew: true,
			},
		},
	}
}

func resourceDomainEntryCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn
	req := &lightsail.CreateDomainEntryInput{
		DomainName: aws.String(d.Get("domain_name").(string)),

		DomainEntry: &types.DomainEntry{
			IsAlias: aws.Bool(d.Get("is_alias").(bool)),
			Name:    aws.String(fmt.Sprintf("%s.%s", d.Get("name").(string), d.Get("domain_name").(string))),
			Target:  aws.String(d.Get("target").(string)),
			Type:    aws.String(d.Get("type").(string)),
		},
	}

	resp, err := conn.CreateDomainEntry(context.TODO(), req)

	if err != nil {
		return err
	}

	op := resp.Operation

	err = waitLightsailOperation(conn, op.Id)
	if err != nil {
		return fmt.Errorf("Error waiting for Domain Entry (%s) to become ready: %s", d.Id(), err)
	}

	// Generate an ID
	vars := []string{
		d.Get("name").(string),
		d.Get("domain_name").(string),
		d.Get("type").(string),
		d.Get("target").(string),
	}

	d.SetId(strings.Join(vars, "_"))

	return resourceDomainEntryRead(d, meta)
}

func resourceDomainEntryRead(d *schema.ResourceData, meta interface{}) error {

	id_parts := strings.SplitN(d.Id(), "_", -1)
	if len(id_parts) != 4 {
		return nil
	}
	name := fmt.Sprintf("%s.%s", id_parts[0], id_parts[1])
	domainname := id_parts[1]
	recordType := id_parts[2]
	recordTarget := id_parts[3]

	conn := meta.(*conns.AWSClient).LightsailConn
	resp, err := conn.GetDomain(context.TODO(), &lightsail.GetDomainInput{
		DomainName: aws.String(domainname),
	})

	if err != nil {
		return err
	}

	var entry types.DomainEntry
	entryExists := false

	for _, n := range resp.Domain.DomainEntries {
		if name == *n.Name && recordType == *n.Type && recordTarget == *n.Target {
			entry = n
			entryExists = true
			break
		}
	}

	if !entryExists {
		d.SetId("")
		return nil
	}

	d.Set("name", strings.TrimSuffix(aws.ToString(entry.Name), fmt.Sprintf(".%s", domainname)))
	d.Set("domain_name", domainname)
	d.Set("type", entry.Type)
	d.Set("is_alias", entry.IsAlias)
	d.Set("target", entry.Target)

	return nil
}

func resourceDomainEntryDelete(d *schema.ResourceData, meta interface{}) error {

	id_parts := strings.SplitN(d.Id(), "_", -1)
	name := fmt.Sprintf("%s.%s", id_parts[0], id_parts[1])
	domainname := id_parts[1]
	recordType := id_parts[2]
	recordTarget := id_parts[3]

	conn := meta.(*conns.AWSClient).LightsailConn
	_, err := conn.DeleteDomainEntry(context.TODO(), &lightsail.DeleteDomainEntryInput{
		DomainName: aws.String(domainname),
		DomainEntry: &types.DomainEntry{
			Name:    aws.String(name),
			Type:    aws.String(recordType),
			Target:  aws.String(recordTarget),
			IsAlias: aws.Bool(d.Get("is_alias").(bool)),
		},
	})

	return err
}
