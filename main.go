package main

import (
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/lightsail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	opts := &plugin.ServeOpts{ProviderFunc: lightsail.Provider}
	plugin.Serve(opts)
}
