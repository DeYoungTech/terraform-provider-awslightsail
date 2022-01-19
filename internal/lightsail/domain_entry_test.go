package lightsail_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/aws-sdk-go-v2/service/lightsail/types"
	"github.com/aws/smithy-go"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/conns"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/testhelper"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDomainEntry_basic(t *testing.T) {
	rName := "awslightsail_domain_entry.test"
	lightsailDomainName := fmt.Sprintf("tf-test-lightsail-%s.com", acctest.RandString(5))
	lightsailDomainEntryName := fmt.Sprintf("test-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckDomainEntryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainEntryConfig_basic(lightsailDomainName, lightsailDomainEntryName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainEntryExists(rName),
				),
			},
		},
	})
}

func TestAccDomainEntry_disappears(t *testing.T) {
	rName := "awslightsail_domain_entry.test"
	lightsailDomainName := fmt.Sprintf("tf-test-lightsail-%s.com", acctest.RandString(5))
	lightsailDomainEntryName := fmt.Sprintf("test-%s", acctest.RandString(5))

	domainEntryDestroy := func(*terraform.State) error {

		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn
		_, err := conn.DeleteDomainEntry(context.TODO(), &lightsail.DeleteDomainEntryInput{
			DomainName: aws.String(lightsailDomainName),
			DomainEntry: &types.DomainEntry{
				Name:   aws.String(fmt.Sprintf("%s.%s", lightsailDomainEntryName, lightsailDomainName)),
				Type:   aws.String("A"),
				Target: aws.String("127.0.0.1"),
			},
		})

		if err != nil {
			return fmt.Errorf("Error deleting Lightsail Domain Entry in disapear test")
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckDomainEntryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainEntryConfig_basic(lightsailDomainName, lightsailDomainEntryName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainEntryExists(rName),
					domainEntryDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDomainEntryExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Lightsail Domain Entry ID is set")
		}

		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn

		resp, err := conn.GetDomain(context.TODO(), &lightsail.GetDomainInput{
			DomainName: aws.String(rs.Primary.Attributes["domain_name"]),
		})

		if err != nil {
			return err
		}

		if resp == nil || resp.Domain == nil {
			return fmt.Errorf("Domain (%s) not found", rs.Primary.Attributes["domain_name"])
		}

		entryExists := false
		for _, n := range resp.Domain.DomainEntries {
			if fmt.Sprintf("%s.%s", rs.Primary.Attributes["name"], rs.Primary.Attributes["domain_name"]) == *n.Name && rs.Primary.Attributes["type"] == *n.Type && rs.Primary.Attributes["target"] == *n.Target {
				entryExists = true
				break
			}
		}

		if !entryExists {
			return fmt.Errorf("Domain entry (%s) not found", rs.Primary.Attributes["name"])
		}

		return nil
	}
}

func testAccCheckDomainEntryDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lightsail_domain_entry" {
			continue
		}

		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn

		resp, err := conn.GetDomain(context.TODO(), &lightsail.GetDomainInput{
			DomainName: aws.String(rs.Primary.ID),
		})

		if err == nil {
			if resp.Domain != nil {
				return fmt.Errorf("Lightsail Domain Entry %q still exists", rs.Primary.ID)
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

func testAccDomainEntryConfig_basic(lightsailDomainName string, lightsailDomainEntryName string) string {
	return fmt.Sprintf(`
resource "awslightsail_domain" "test" {
  domain_name = %[1]q
}
resource "awslightsail_domain_entry" "test" {
  domain_name = awslightsail_domain.test.domain_name
  name        = %[2]q
  type        = "A"
  target      = "127.0.0.1"
  is_alias    = false
}
`, lightsailDomainName, lightsailDomainEntryName)
}
