package lightsail

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider returns a *schema.Provider.
func Provider() *schema.Provider {
	// TODO: Move the validation to this, requires conditional schemas
	// TODO: Move the configuration to this, requires validation

	// The actual provider
	p := &schema.Provider{
		DataSourcesMap: map[string]*schema.Resource{
			"awslightsail_availability_zones": DataSourceAvailabilityZones(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"awslightsail_key_pair":             ResourceKeyPair(),
			"awslightsail_database":             ResourceDatabase(),
			"awslightsail_domain":               ResourceDomain(),
			"awslightsail_instance":             ResourceInstance(),
			"awslightsail_static_ip_attachment": ResourceStaticIPAttachment(),
			"awslightsail_static_ip":            ResourceStaticIP(),
		},
	}

	p.ConfigureFunc = providerConfigure(p)

	return p
}

func providerConfigure(p *schema.Provider) schema.ConfigureFunc {
	return func(d *schema.ResourceData) (interface{}, error) {
		terraformVersion := p.TerraformVersion
		if terraformVersion == "" {
			// Terraform 0.12 introduced this field to the protocol
			// We can therefore assume that if it's missing it's 0.10 or 0.11
			terraformVersion = "0.11+compatible"
		}
		// Using the SDK's default configuration, loading additional config
		// and credentials values from the environment variables, shared
		// credentials, and shared configuration files
		cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
		if err != nil {
			log.Fatalf("unable to load SDK config, %v", err)
		}

		// Create an Amazon Lightsail Service client
		client := lightsail.NewFromConfig(cfg)

		return client, err
	}
}
