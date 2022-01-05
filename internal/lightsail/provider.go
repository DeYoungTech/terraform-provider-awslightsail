package lightsail

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/meta"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider returns a *schema.Provider.
func Provider() *schema.Provider {
	// TODO: Move the validation to this, requires conditional schemas
	// TODO: Move the configuration to this, requires validation

	// The actual provider
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"default_tags": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Configuration block with settings to default resource tags across all resources.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"tags": {
							Type:        schema.TypeMap,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Resource tags to default across all resources",
						},
					},
				},
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"awslightsail_default_tags":            meta.DataSourceDefaultTags(),
			"awslightsail_availability_zones": DataSourceAvailabilityZones(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"awslightsail_key_pair":             ResourceKeyPair(),
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

// func expandProviderDefaultTags(l []interface{}) *tftags.DefaultConfig {
// 	if len(l) == 0 || l[0] == nil {
// 		return nil
// 	}

// 	defaultConfig := &tftags.DefaultConfig{}
// 	m := l[0].(map[string]interface{})

// 	if v, ok := m["tags"].(map[string]interface{}); ok {
// 		defaultConfig.Tags = tftags.New(v)
// 	}
// 	return defaultConfig
// }