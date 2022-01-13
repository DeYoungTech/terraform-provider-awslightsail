package lightsail_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/smithy-go"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/conns"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/testhelper"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)


func TestAccLoadBalancer_basic(t *testing.T) {
	rName := "awslightsail_instance.instance"
	lName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy:  testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfigBasic(lName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerExists(rName),
					resource.TestCheckResourceAttr(rName, "health_check_path", "/"),
					resource.TestCheckResourceAttr(rName, "instance_port", "80"),
					resource.TestCheckResourceAttrSet(rName, "dns_name"),
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

func TestAccLoadBalancer_Name(t *testing.T) {
	rName := "awslightsail_instance.instance"
	lName := acctest.RandomWithPrefix("tf-acc-test")
	lightsailNameWithSpaces := fmt.Sprint(rName, "string with spaces")
	lightsailNameWithStartingDigit := fmt.Sprintf("01-%s", rName)
	lightsailNameWithUnderscore := fmt.Sprintf("%s_123456", rName)

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config:      testAccLoadBalancerConfigBasic(lightsailNameWithSpaces),
				ExpectError: regexp.MustCompile(`must contain only alphanumeric characters, underscores, hyphens, and dots`),
			},
			{
				Config:      testAccLoadBalancerConfigBasic(lightsailNameWithStartingDigit),
				ExpectError: regexp.MustCompile(`must begin with an alphabetic character`),
			},
			{
				Config: testAccLoadBalancerConfigBasic(lName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(rName),
					resource.TestCheckResourceAttrSet(rName, "health_check_path"),
					resource.TestCheckResourceAttrSet(rName, "instance_port"),
				),
			},
			{
				Config: testAccLoadBalancerConfigBasic(lightsailNameWithUnderscore),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(rName),
					resource.TestCheckResourceAttrSet(rName, "health_check_path"),
					resource.TestCheckResourceAttrSet(rName, "instance_port"),
				),
			},
		},
	})
}

func TestAccLoadBalancer_HealthCheckPath(t *testing.T) {
	rName := "awslightsail_instance.instance"
	lName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfigHealthCheckPath(lName, "/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerExists(rName),
					resource.TestCheckResourceAttr(rName, "health_check_path", "/"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLoadBalancerConfigHealthCheckPath(lName, "/healthcheck"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerExists(rName),
					resource.TestCheckResourceAttr(rName, "health_check_path", "/healthcheck"),
				),
			},
		},
	})
}

func TestAccLoadBalancer_Tags(t *testing.T) {
	rName := "awslightsail_instance.instance"
	lName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfigTags1(lName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerExists(rName),
					resource.TestCheckResourceAttr(rName, "tags.%", "1"),
					resource.TestCheckResourceAttr(rName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLoadBalancerConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerExists(rName),
					resource.TestCheckResourceAttr(rName, "tags.%", "2"),
					resource.TestCheckResourceAttr(rName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(rName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccLoadBalancerConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerExists(rName),
					resource.TestCheckResourceAttr(rName, "tags.%", "1"),
					resource.TestCheckResourceAttr(rName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckLoadBalancerExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No LightsailLoadBalancer ID is set")
		}

		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn

		respLoadBalancer, err := conn.GetLoadBalancer(context.TODO(), &lightsail.GetLoadBalancerInput{
			LoadBalancerName: aws.String(rs.Primary.Attributes["name"]),
		})

		if err != nil {
			return err
		}

		if respLoadBalancer == nil || respLoadBalancer.LoadBalancer == nil {
			return fmt.Errorf("Load Balancer (%s) not found", rs.Primary.Attributes["name"])
		}

		return nil
	}
}

func TestAccLoadBalancer_disappears(t *testing.T) {
	rName := "awslightsail_instance.instance"
	lName := acctest.RandomWithPrefix("tf-acc-test")

	testDestroy := func(*terraform.State) error {
		// reach out and DELETE the Load Balancer
		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn
		_, err := conn.DeleteLoadBalancer(context.TODO(), &lightsail.DeleteLoadBalancerInput{
			LoadBalancerName: aws.String(lName),
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
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfigBasic(lName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerExists(rName),
					testDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckLoadBalancerDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "awslightsail_load_balancer" {
			continue
		}

		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn

		resp, err := conn.GetLoadBalancer(context.TODO(), &lightsail.GetLoadBalancerInput{
			LoadBalancerName: aws.String(rs.Primary.ID),
		})

		if err == nil {
			if resp.LoadBalancer != nil {
				return fmt.Errorf("Lightsail Load Balancer %q still exists", rs.Primary.ID)
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

func testAccLoadBalancerConfigBasic(rName string) string {
	return fmt.Sprintf(`
resource "awslightsail_load_balancer" "test" {
  name              = %[1]q
  health_check_path = "/"
  instance_port     = "80"
}
`, rName)
}

func testAccLoadBalancerConfigHealthCheckPath(rName string, rPath string) string {
	return fmt.Sprintf(`
resource "awslightsail_load_balancer" "test" {
  name              = %[1]q
  health_check_path = %[2]q
  instance_port     = "80"
}
`, rName, rPath)
}

func testAccLoadBalancerConfigTags1(rName string, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "awslightsail_load_balancer" "test" {
  name              = %[1]q
  health_check_path = "/"
  instance_port     = "80"
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccLoadBalancerConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "awslightsail_load_balancer" "test" {
  name              = %[1]q
  health_check_path = "/"
  instance_port     = "80"
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}