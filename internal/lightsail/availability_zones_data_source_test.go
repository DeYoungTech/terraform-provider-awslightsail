package lightsail_test

import (
	"fmt"
	"testing"

	"github.com/deyoungtech/terraform-provider-awslightsail/internal/testhelper"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAvailabilityZonesDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		Providers: testhelper.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLightsailAvailabilityZonesConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAvailabilityZonesMeta("data.awslightsail_availability_zones.all"),
				),
			},
		},
	})
}

func testAccCheckAvailabilityZonesMeta(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find AZ resource: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("AZ resource ID not set.")
		}
		return nil
	}
}

const testAccCheckLightsailAvailabilityZonesConfig = ` 
data "awslightsail_availability_zones" "all" {}
`
