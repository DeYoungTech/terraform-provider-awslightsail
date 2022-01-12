package conns

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/tftags"
)

const (
	Lightsail = "lightsail"
)

type ServiceDatum struct {
	AWSClientName     string
	AWSServiceID      string
	ProviderNameUpper string
	HCLKeys           []string
}

var serviceData map[string]*ServiceDatum

func init() {
	serviceData = make(map[string]*ServiceDatum)
	serviceData[Lightsail] = &ServiceDatum{AWSClientName: "Lightsail", AWSServiceID: lightsail.ServiceID, ProviderNameUpper: "Lightsail", HCLKeys: []string{"lightsail"}}
}

type Config struct {
	Region            string
	DefaultTagsConfig *tftags.DefaultConfig
	IgnoreTagsConfig  *tftags.IgnoreConfig
	TerraformVersion  string
}

type AWSClient struct {
	DefaultTagsConfig *tftags.DefaultConfig
	IgnoreTagsConfig  *tftags.IgnoreConfig
	LightsailConn     *lightsail.Client
	TerraformVersion  string
}

// Client configures and returns a fully initialized AWSClient
func (c *Config) Client() (interface{}, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(c.Region))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	client := &AWSClient{
		IgnoreTagsConfig:  c.IgnoreTagsConfig,
		LightsailConn:     lightsail.NewFromConfig(cfg),
		DefaultTagsConfig: c.DefaultTagsConfig,
		TerraformVersion:  c.TerraformVersion,
	}

	return client, nil
}
