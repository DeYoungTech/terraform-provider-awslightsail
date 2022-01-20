package lightsail_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/smithy-go"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/conns"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/testhelper"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// AWS Accounts are limited to 5 Load Balancers per account.
func TestAccLoadBalancerAttachment_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"LoadBalancerAttachment": {
			"basic":      testAccLoadBalancerAttachment_basic,
			"disappears": testAccLoadBalancerAttachment_disappears,
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

func testAccLoadBalancerAttachment_basic(t *testing.T) {
	rName := "awslightsail_lb_attachment.test"
	lbName := acctest.RandomWithPrefix("tf-acc-test")
	liName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckLoadBalancerAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerAttachmentConfigBasic(lbName, liName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerAttachmentExists(rName),
				),
			},
		},
	})
}

func testAccLoadBalancerAttachment_disappears(t *testing.T) {
	rName := "awslightsail_lb_attachment.test"
	lbName := acctest.RandomWithPrefix("tf-acc-test")
	liName := acctest.RandomWithPrefix("tf-acc-test")

	instanceNames := make([]string, 1)
	instanceNames[0] = liName

	testDestroy := func(*terraform.State) error {
		// reach out and detach Instance from the Load Balancer
		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn
		resp, err := conn.DetachInstancesFromLoadBalancer(context.TODO(), &lightsail.DetachInstancesFromLoadBalancerInput{
			LoadBalancerName: aws.String(lbName),
			InstanceNames:    instanceNames,
		})

		if err != nil {
			return fmt.Errorf("error detaching Instance from Load Balancer")
		}

		op := resp.Operations[0]

		err = testhelper.WaitLightsailOperation(conn, op.Id)
		if err != nil {
			return fmt.Errorf("Error waiting for load balancer attatchment (%s) to become ready: %s", lbName, err)
		}

		return nil
	}

	resource.Test(t, resource.TestCase{
		Providers: testhelper.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerAttachmentConfigBasic(lbName, liName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerAttachmentExists(rName),
					testDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckLoadBalancerAttachmentExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No LightsailLoadBalancerAttachment ID is set")
		}

		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn

		id_parts := strings.SplitN(rs.Primary.ID, "_", -1)

		if len(id_parts) != 2 {
			return nil
		}

		lbname := id_parts[0]

		resp, err := conn.GetLoadBalancer(context.TODO(), &lightsail.GetLoadBalancerInput{
			LoadBalancerName: aws.String(lbname),
		})

		if err != nil {
			return err
		}

		if resp == nil || resp.LoadBalancer == nil {
			return fmt.Errorf("Load Balancer (%s) not found", rs.Primary.Attributes["lb_name"])
		}

		entryExists := false
		for _, n := range resp.LoadBalancer.InstanceHealthSummary {
			if rs.Primary.Attributes["instance_name"] == *n.InstanceName {
				entryExists = true
				break
			}
		}

		if !entryExists {
			return fmt.Errorf("Attachment (%s) not found", rs.Primary.Attributes["instance_name"])
		}

		return nil
	}
}

func testAccCheckLoadBalancerAttachmentDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "awslightsail_lb_attachment" {
			continue
		}

		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn

		id_parts := strings.SplitN(rs.Primary.ID, "_", -1)
		if len(id_parts) != 2 {
			return nil
		}

		lbname := id_parts[0]

		resp, err := conn.GetLoadBalancer(context.TODO(), &lightsail.GetLoadBalancerInput{
			LoadBalancerName: aws.String(lbname),
		})

		// Verify the error
		if err != nil {
			var oe *smithy.OperationError
			if errors.As(err, &oe) {
				log.Printf("failed to call service: %s, operation: %s, error: %v", oe.Service(), oe.Operation(), oe.Unwrap())
			}
			return nil
		}

		entryExists := false
		for _, n := range resp.LoadBalancer.InstanceHealthSummary {
			if rs.Primary.Attributes["instance_name"] == *n.InstanceName {
				entryExists = true
				break
			}
		}

		if entryExists {
			return fmt.Errorf("Attachment (%s) Still Exists", rs.Primary.Attributes["instance_name"])
		}

		return err
	}

	return nil
}

func testAccLoadBalancerAttachmentConfigBasic(lbName string, liName string) string {
	return fmt.Sprintf(`
data "awslightsail_availability_zones" "all" {}

resource "awslightsail_lb" "test" {
  name              = %[1]q
  health_check_path = "/"
  instance_port     = "80"
}

resource "awslightsail_instance" "test" {
  name              = %[2]q
  availability_zone = data.awslightsail_availability_zones.all.names[0]
  blueprint_id      = "amazon_linux"
  bundle_id         = "nano_1_0"
}

resource "awslightsail_lb_attachment" "test" {
  load_balancer_name = awslightsail_lb.test.name
  instance_name      = awslightsail_instance.test.name
}
`, lbName, liName)
}
