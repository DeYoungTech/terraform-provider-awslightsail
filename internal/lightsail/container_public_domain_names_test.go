package lightsail_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/deyoungtech/terraform-provider-awslightsail/internal/testhelper"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Lightsail limits to 5 pending containers at a time
func TestAccContainerPublicDomainNames_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"ContainerPublicDomainNames": {
			"basic": testAccContainerPublicDomainNames_basic,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}

func testAccContainerPublicDomainNames_basic(t *testing.T) {
	lName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		Providers: testhelper.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config:      testAccContainerPublicDomainNamesConfigBasic(lName),
				ExpectError: regexp.MustCompile(fmt.Sprintf(`InvalidInputException: Certificate "ExsitingCertificate" provided for service "%s" has PENDING_VALIDATION status, required status: ISSUED.`, lName)),
			},
		},
	})
}

func testAccContainerPublicDomainNamesConfigBasic(lName string) string {
	return fmt.Sprintf(`
resource "awslightsail_container_service" "test" {
  name  = %[1]q
  power = "nano"
  scale = 1
}

resource "awslightsail_certificate" "test" {
  name        = "ExsitingCertificate"
  domain_name = "existingdomain1.com"
  subject_alternative_names = [
    "www.existingdomain1.com"
  ]
}

resource "awslightsail_container_public_domain_names" "test" {
  container_service_name = awslightsail_container_service.test.id
  public_domain_names {
    certificate {
      certificate_name = awslightsail_certificate.test.name
      domain_names = [
        "existingdomain1.com",
        "www.existingdomain1.com",
      ]
    }
  }
}
`, lName)
}
