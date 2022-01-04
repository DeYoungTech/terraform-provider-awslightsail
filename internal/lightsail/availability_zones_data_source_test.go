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

// func testAccCheckAvailabilityZonesAllAvailabilityZonesConfig() string {
// 	return `
// data "aws_availability_zones" "test" {
//   all_availability_zones = true
// }
// `
// }

// func testAccCheckAvailabilityZonesFilterConfig() string {
// 	return `
// data "aws_availability_zones" "test" {
//   filter {
//     name   = "state"
//     values = ["available"]
//   }
// }
// `
// }

// func testAccCheckAvailabilityZonesExcludeNamesConfig() string {
// 	return `
// data "aws_availability_zones" "all" {}
// data "aws_availability_zones" "test" {
//   exclude_names = [data.aws_availability_zones.all.names[0]]
// }
// `
// }

// func testAccCheckAvailabilityZonesExcludeZoneIDsConfig() string {
// 	return `
// data "aws_availability_zones" "all" {}
// data "aws_availability_zones" "test" {
//   exclude_zone_ids = [data.aws_availability_zones.all.zone_ids[0]]
// }
// `
// }

// const testAccCheckAWSAvailabilityZonesStateConfig = `
// data "aws_availability_zones" "state_filter" {
//   state = "available"
// }
// `
