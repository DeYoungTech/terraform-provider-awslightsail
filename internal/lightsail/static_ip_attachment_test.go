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

func TestAccStaticIPAttachment_basic(t *testing.T) {
	rName := "awslightsail_static_ip_attachment.test"
	staticIpName := fmt.Sprintf("tf-test-lightsail-%s", sdkacctest.RandString(5))
	instanceName := fmt.Sprintf("tf-test-lightsail-%s", sdkacctest.RandString(5))
	keypairName := fmt.Sprintf("tf-test-lightsail-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckStaticIPAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStaticIPAttachmentConfig_basic(staticIpName, instanceName, keypairName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStaticIPAttachmentExists(rName),
					resource.TestCheckResourceAttrSet(rName, "ip_address"),
				),
			},
		},
	})
}

func TestAccStaticIPAttachment_disappears(t *testing.T) {
	rName := "awslightsail_static_ip_attachment.test"
	staticIpName := fmt.Sprintf("tf-test-lightsail-%s", sdkacctest.RandString(5))
	instanceName := fmt.Sprintf("tf-test-lightsail-%s", sdkacctest.RandString(5))
	keypairName := fmt.Sprintf("tf-test-lightsail-%s", sdkacctest.RandString(5))

	staticIpDestroy := func(*terraform.State) error {
		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn
		_, err := conn.DetachStaticIp(context.TODO(), &lightsail.DetachStaticIpInput{
			StaticIpName: aws.String(staticIpName),
		})

		if err != nil {
			return fmt.Errorf("Error deleting Lightsail Static IP in disappear test")
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckStaticIPAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStaticIPAttachmentConfig_basic(staticIpName, instanceName, keypairName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStaticIPAttachmentExists(rName),
					staticIpDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckStaticIPAttachmentExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Lightsail Static IP Attachment ID is set")
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

		if !*resp.StaticIp.IsAttached {
			return fmt.Errorf("Static IP (%s) not attached", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckStaticIPAttachmentDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "awslightsail_static_ip_attachment" {
			continue
		}

		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn

		resp, err := conn.GetStaticIp(context.TODO(), &lightsail.GetStaticIpInput{
			StaticIpName: aws.String(rs.Primary.ID),
		})

		if err == nil {
			if *resp.StaticIp.IsAttached {
				return fmt.Errorf("Lightsail Static IP %q is still attached (to %q)", rs.Primary.ID, *resp.StaticIp.AttachedTo)
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

func testAccStaticIPAttachmentConfig_basic(staticIpName, instanceName, keypairName string) string {
	return fmt.Sprintf(`
data "awslightsail_availability_zones" "all" {}

resource "awslightsail_static_ip_attachment" "test" {
  static_ip_name = awslightsail_static_ip.test.name
  instance_name  = awslightsail_instance.test.name
}
resource "awslightsail_static_ip" "test" {
  name = "%s"
}
resource "awslightsail_instance" "test" {
  name              = "%s"
  availability_zone = data.awslightsail_availability_zones.all.names[0]
  blueprint_id      = "amazon_linux"
  bundle_id         = "micro_1_0"
  key_pair_name     = awslightsail_key_pair.test.name
}
resource "awslightsail_key_pair" "test" {
  name = "%s"
}
`, staticIpName, instanceName, keypairName)
}
