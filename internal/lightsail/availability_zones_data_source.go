package lightsail

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceAvailabilityZones() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAvailabilityZonesRead,

		Schema: map[string]*schema.Schema{
			"names": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"database_names": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAvailabilityZonesRead(d *schema.ResourceData, meta interface{}) error {
	conn := Client()

	resp, err := conn.GetRegions(context.TODO(), &lightsail.GetRegionsInput{
		IncludeAvailabilityZones:                   aws.Bool(true),
		IncludeRelationalDatabaseAvailabilityZones: aws.Bool(true),
	})

	if err != nil {
		return fmt.Errorf("Error fetching Availability Zones: %w", err)
	}

	names := []string{}
	databasenames := []string{}

	for _, v := range resp.Regions {

		for _, v := range v.AvailabilityZones {
			name := aws.ToString(v.ZoneName)
			names = append(names, name)
		}

		for _, v := range v.RelationalDatabaseAvailabilityZones {
			name := aws.ToString(v.ZoneName)
			databasenames = append(names, name)
		}
	}

	d.SetId(string(resp.Regions[0].Name))
	d.Set("names", names)
	d.Set("database_names", databasenames)

	return nil
}
