package lightsail_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/aws-sdk-go-v2/service/lightsail/types"
	"github.com/aws/smithy-go"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/conns"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/testhelper"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccContactMethod_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"ContactMethod": {
			"basic":      testAccContactMethod_basic,
			"disappears": testAccContactMethod_disappears,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
					time.Sleep(3 * time.Second)
				})
			}
		})
	}
}

func testAccContactMethod_basic(t *testing.T) {
	rName := "awslightsail_contact_method.test"
	lEndpoint := testhelper.DefaultEmailAddress

	resource.Test(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckContactMethodDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContactMethodConfig_basic(lEndpoint),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckContactMethodExists(rName),
				),
			},
		},
	})
}

func testAccContactMethod_disappears(t *testing.T) {
	rName := "awslightsail_contact_method.test"
	lEndpoint := testhelper.DefaultEmailAddress

	contactMethodDestroy := func(*terraform.State) error {

		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn
		_, err := conn.DeleteContactMethod(context.TODO(), &lightsail.DeleteContactMethodInput{
			Protocol: types.ContactProtocol("Email"),
		})

		if err != nil {
			return fmt.Errorf("Error deleting Lightsail Contact Method in disapear test")
		}

		return nil
	}

	resource.Test(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckContactMethodDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContactMethodConfig_basic(lEndpoint),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckContactMethodExists(rName),
					contactMethodDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckContactMethodExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Lightsail Domain Entry ID is set")
		}

		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn

		protocols := make([]types.ContactProtocol, 0, 1)

		resp, err := conn.GetContactMethods(context.TODO(), &lightsail.GetContactMethodsInput{
			Protocols: append(protocols, types.ContactProtocol(rs.Primary.ID)),
		})

		if err != nil {
			return err
		}

		if resp == nil || resp.ContactMethods == nil {
			return fmt.Errorf("Contact Method (%s) not found", rs.Primary.ID)
		}

		entryExists := false
		for _, n := range resp.ContactMethods {
			if rs.Primary.Attributes["endpoint"] == *n.ContactEndpoint {
				entryExists = true
				break
			}
		}

		if !entryExists {
			return fmt.Errorf("Contact Method (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckContactMethodDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lightsail_contact_method" {
			continue
		}

		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn
		protocols := make([]types.ContactProtocol, 0, 1)

		resp, err := conn.GetContactMethods(context.TODO(), &lightsail.GetContactMethodsInput{
			Protocols: append(protocols, types.ContactProtocol(rs.Primary.ID)),
		})

		if err == nil {
			if resp.ContactMethods != nil {
				return fmt.Errorf("Lightsail Contact Entry %q still exists", rs.Primary.ID)
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

func testAccContactMethodConfig_basic(lEndpoint string) string {
	return fmt.Sprintf(`
resource "awslightsail_contact_method" "test" {
  protocol = "Email"
  endpoint = "%s"

}
`, lEndpoint)
}
