package testhelper

import (
	tflightsail "github.com/deyoungtech/terraform-provider-awslightsail/internal/lightsail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// initialized once, so it can be shared by each acceptance test
var provider = tflightsail.Provider()

// GetProvider returns the lightsail provider
func GetProvider() *schema.Provider {
	return provider
}

func GetProviders() map[string]*schema.Provider {
	return map[string]*schema.Provider{
		"awslightsail": GetProvider(),
	}
}

// func testAccPreCheck(t *testing.T) {
// 	conn := provider.Meta(*lightsail.Client)

// 	input := &lightsail.GetInstancesInput{}

// 	_, err := conn.GetInstances(input)

// 	if err != nil {
// 		t.Fatalf("unexpected PreCheck error: %s", err)
// 	}
// }
