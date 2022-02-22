package lightsail_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/smithy-go"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/conns"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/testhelper"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccStaticIP_basic(t *testing.T) {
	rName := "awslightsail_static_ip.test"
	lName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckStaticIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStaticIPConfig_basic(lName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStaticIPExists(rName),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccStaticIP_disappears(t *testing.T) {
	rName := "awslightsail_static_ip.test"
	lName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())

	staticIpDestroy := func(*terraform.State) error {
		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn
		_, err := conn.ReleaseStaticIp(context.TODO(), &lightsail.ReleaseStaticIpInput{
			StaticIpName: aws.String(lName),
		})

		if err != nil {
			return fmt.Errorf("Error deleting Lightsail Static IP in disapear test")
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckStaticIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStaticIPConfig_basic(lName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStaticIPExists(rName),
					staticIpDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckStaticIPExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Lightsail Static IP ID is set")
		}

		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn

		resp, err := conn.GetStaticIp(context.TODO(), &lightsail.GetStaticIpInput{
			StaticIpName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if resp == nil || resp.StaticIp == nil {
			return fmt.Errorf("Static IP (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckStaticIPDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "awslightsail_static_ip" {
			continue
		}

		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn

		resp, err := conn.GetStaticIp(context.TODO(), &lightsail.GetStaticIpInput{
			StaticIpName: aws.String(rs.Primary.ID),
		})

		if err == nil {
			if resp.StaticIp != nil {
				return fmt.Errorf("Lightsail Static IP %q still exists", rs.Primary.ID)
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

func testAccStaticIPConfig_basic(staticIpName string) string {
	return fmt.Sprintf(`
resource "awslightsail_static_ip" "test" {
  name = "%s"
}
`, staticIpName)
}
