package lightsail

import (
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/conns"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/tftags"
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
			"ignore_tags": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Configuration block with settings to ignore resource tags across all resources.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"keys": {
							Type:        schema.TypeSet,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Description: "Resource tag keys to ignore across all resources.",
						},
						"key_prefixes": {
							Type:        schema.TypeSet,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Description: "Resource tag key prefixes to ignore across all resources.",
						},
					},
				},
			},
			"region": {
				Type:     schema.TypeString,
				Required: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"AWS_REGION",
					"AWS_DEFAULT_REGION",
				}, nil),
				Description:  descriptions["region"],
				InputDefault: "us-east-1", // lintignore:AWSAT003
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"awslightsail_availability_zones": DataSourceAvailabilityZones(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"awslightsail_contact_method":       ResourceContactMethod(),
			"awslightsail_database":             ResourceDatabase(),
			"awslightsail_domain":               ResourceDomain(),
			"awslightsail_domain_entry":         ResourceDomainEntry(),
			"awslightsail_instance":             ResourceInstance(),
			"awslightsail_key_pair":             ResourceKeyPair(),
			"awslightsail_lb":                   ResourceLoadBalancer(),
			"awslightsail_lb_attachment":        ResourceLoadBalancerAttachment(),
			"awslightsail_static_ip_attachment": ResourceStaticIPAttachment(),
			"awslightsail_static_ip":            ResourceStaticIP(),
		},
	}

	// p.ConfigureFunc = providerConfigure(p)
	p.ConfigureFunc = func(d *schema.ResourceData) (interface{}, error) {
		terraformVersion := p.TerraformVersion
		if terraformVersion == "" {
			// Terraform 0.12 introduced this field to the protocol
			// We can therefore assume that if it's missing it's 0.10 or 0.11
			terraformVersion = "0.11+compatible"
		}
		return providerConfigure(d, terraformVersion)
	}
	return p
}

func providerConfigure(d *schema.ResourceData, terraformVersion string) (interface{}, error) {

	config := conns.Config{
		Region:            d.Get("region").(string),
		DefaultTagsConfig: expandProviderDefaultTags(d.Get("default_tags").([]interface{})),
		IgnoreTagsConfig:  expandProviderIgnoreTags(d.Get("ignore_tags").([]interface{})),
		TerraformVersion:  terraformVersion,
	}

	// return func(d *schema.ResourceData) (interface{}, error) {
	// 	terraformVersion := p.TerraformVersion
	// 	if terraformVersion == "" {
	// 		// Terraform 0.12 introduced this field to the protocol
	// 		// We can therefore assume that if it's missing it's 0.10 or 0.11
	// 		terraformVersion = "0.11+compatible"
	// 	}
	// 	// Using the SDK's default configuration, loading additional config
	// 	// and credentials values from the environment variables, shared
	// 	// credentials, and shared configuration files
	// 	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	// 	if err != nil {
	// 		log.Fatalf("unable to load SDK config, %v", err)
	// 	}

	// 	// Create an Amazon Lightsail Service client
	// 	client := lightsail.NewFromConfig(cfg)

	// 	return config.Client()
	// }
	return config.Client()
}

var descriptions map[string]string

func init() {
	descriptions = map[string]string{
		"region": "The region where AWS operations will take place. Examples\n" +
			"are us-east-1, us-west-2, etc.", // lintignore:AWSAT003
	}
}

func expandProviderDefaultTags(l []interface{}) *tftags.DefaultConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	defaultConfig := &tftags.DefaultConfig{}
	m := l[0].(map[string]interface{})

	if v, ok := m["tags"].(map[string]interface{}); ok {
		defaultConfig.Tags = tftags.New(v)
	}
	return defaultConfig
}

func expandProviderIgnoreTags(l []interface{}) *tftags.IgnoreConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	ignoreConfig := &tftags.IgnoreConfig{}
	m := l[0].(map[string]interface{})

	if v, ok := m["keys"].(*schema.Set); ok {
		ignoreConfig.Keys = tftags.New(v.List())
	}

	if v, ok := m["key_prefixes"].(*schema.Set); ok {
		ignoreConfig.KeyPrefixes = tftags.New(v.List())
	}

	return ignoreConfig
}
