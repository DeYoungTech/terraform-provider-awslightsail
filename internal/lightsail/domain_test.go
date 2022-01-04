package lightsail_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/smithy-go"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/testhelper"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDomain_basic(t *testing.T) {
	lName := fmt.Sprintf("tf-test-lightsail-%s.com", sdkacctest.RandString(5))
	rName := "awslightsail_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(lName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(rName),
				),
			},
		},
	})
}

func TestAccDomain_disappears(t *testing.T) {
	lName := fmt.Sprintf("tf-test-lightsail-%s.com", sdkacctest.RandString(5))
	rName := "awslightsail_domain.test"

	testDestroy := func(*terraform.State) error {
		// reach out and DELETE the Instance
		conn := testhelper.GetProvider().Meta().(*lightsail.Client)
		_, err := conn.DeleteDomain(context.TODO(), &lightsail.DeleteDomainInput{
			DomainName: aws.String(lName),
		})

		if err != nil {
			return fmt.Errorf("error deleting Lightsail Instance in disappear test")
		}

		// sleep 7 seconds to give it time, so we don't have to poll
		time.Sleep(7 * time.Second)

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(lName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(rName),
					testDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDomainExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Lightsail Domain ID is set")
		}

		conn := testhelper.GetProvider().Meta().(*lightsail.Client)

		resp, err := conn.GetDomain(context.TODO(), &lightsail.GetDomainInput{
			DomainName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if resp == nil || resp.Domain == nil {
			return fmt.Errorf("Domain (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDomainDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "awslightsail_domain" {
			continue
		}

		conn := testhelper.GetProvider().Meta().(*lightsail.Client)

		resp, err := conn.GetDomain(context.TODO(), &lightsail.GetDomainInput{
			DomainName: aws.String(rs.Primary.ID),
		})

		if err == nil {
			if resp.Domain != nil {
				return fmt.Errorf("Lightsail Domain %q still exists", rs.Primary.ID)
			}
		}

		// Verify the error
		if err != nil {
			var oe *smithy.OperationError
			if errors.As(err, &oe) {
				log.Printf("failed to call service: %s, operation: %s, error: %v", oe.Service(), oe.Operation(), oe.Unwrap())
			}
			return nil
		}
		return err
	}

	return nil
}

func testAccDomainConfig_basic(lName string) string {
	return fmt.Sprintf(`
resource "awslightsail_domain" "test" {
  domain_name = "%s"
}
`, lName)
}
