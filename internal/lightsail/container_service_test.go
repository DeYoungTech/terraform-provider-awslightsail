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
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/conns"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/testhelper"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// Lightsail limits to 5 pending containers at a time
func TestAccContainerService_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"ContainerService": {
			"basic":       testAccContainerService_basic,
			"disappears":  testAccContainerService_disappears,
			"name":        testAccContainerService_Name,
			"is_disabled": testAccContainerService_IsDisabled,
			"power":       testAccContainerService_Power,
			"scale":       testAccContainerService_Scale,
			"Tags":        testAccContainerService_Tags,
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

func testAccContainerService_basic(t *testing.T) {
	lName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rName := "awslightsail_container_service.test"

	resource.Test(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckContainerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerServiceConfigBasic(lName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(rName),
					resource.TestCheckResourceAttr(rName, "name", lName),
					resource.TestCheckResourceAttr(rName, "power", "nano"),
					resource.TestCheckResourceAttr(rName, "scale", "1"),
					resource.TestCheckResourceAttr(rName, "is_disabled", "false"),
					resource.TestCheckResourceAttrSet(rName, "arn"),
					resource.TestCheckResourceAttrSet(rName, "availability_zone"),
					resource.TestCheckResourceAttrSet(rName, "power_id"),
					resource.TestCheckResourceAttrSet(rName, "principal_arn"),
					resource.TestCheckResourceAttrSet(rName, "private_domain_name"),
					resource.TestCheckResourceAttr(rName, "resource_type", "ContainerService"),
					resource.TestCheckResourceAttr(rName, "state", "READY"),
					resource.TestCheckResourceAttrSet(rName, "url"),
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

func testAccContainerService_disappears(t *testing.T) {
	lName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rName := "awslightsail_container_service.test"

	testDestroy := func(*terraform.State) error {
		// reach out and DELETE the Container Service
		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn

		_, err := conn.DeleteContainerService(context.TODO(), &lightsail.DeleteContainerServiceInput{
			ServiceName: aws.String(lName),
		})

		if err != nil {
			return fmt.Errorf("error deleting Lightsail Container Service in disappear test")
		}

		// sleep 7 seconds to give it time, so we don't have to poll
		time.Sleep(7 * time.Second)

		return nil
	}

	resource.Test(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckContainerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerServiceConfigBasic(lName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(rName),
					testDestroy),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccContainerService_Name(t *testing.T) {
	lName1 := sdkacctest.RandomWithPrefix("tf-acc-test")
	lName2 := sdkacctest.RandomWithPrefix("tf-acc-test")
	rName := "awslightsail_container_service.test"

	resource.Test(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckContainerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerServiceConfigBasic(lName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(rName),
					resource.TestCheckResourceAttr(rName, "name", lName1),
				),
			},
			{
				Config: testAccContainerServiceConfigBasic(lName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(rName),
					resource.TestCheckResourceAttr(rName, "name", lName2),
				),
			},
		},
	})
}

func testAccContainerService_IsDisabled(t *testing.T) {
	lName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rName := "awslightsail_container_service.test"

	resource.Test(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckContainerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerServiceConfigBasic(lName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(rName),
					resource.TestCheckResourceAttr(rName, "is_disabled", "false"),
				),
			},
			{
				Config: testAccContainerServiceConfigIsDisabled(lName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(rName),
					resource.TestCheckResourceAttr(rName, "is_disabled", "true"),
				),
			},
		},
	})
}

func testAccContainerService_Power(t *testing.T) {
	lName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rName := "awslightsail_container_service.test"

	resource.Test(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckContainerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerServiceConfigBasic(lName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(rName),
					resource.TestCheckResourceAttr(rName, "power", "nano"),
				),
			},
			{
				Config: testAccContainerServiceConfigPower(lName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(rName),
					resource.TestCheckResourceAttr(rName, "power", "micro"),
				),
			},
		},
	})
}

func testAccContainerService_Scale(t *testing.T) {
	lName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rName := "awslightsail_container_service.test"

	resource.Test(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckContainerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerServiceConfigBasic(lName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(rName),
					resource.TestCheckResourceAttr(rName, "scale", "1"),
				),
			},
			{
				Config: testAccContainerServiceConfigScale(lName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(rName),
					resource.TestCheckResourceAttr(rName, "scale", "2"),
				),
			},
		},
	})

}

func testAccContainerService_Tags(t *testing.T) {
	lName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rName := "awslightsail_container_service.test"

	resource.Test(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckContainerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerServiceConfigTag1(lName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckContainerServiceExists(rName),
					resource.TestCheckResourceAttr(rName, "tags.%", "2"),
					resource.TestCheckResourceAttr(rName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(rName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccContainerServiceConfigTag2(lName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckContainerServiceExists(rName),
					resource.TestCheckResourceAttr(rName, "tags.%", "1"),
					resource.TestCheckResourceAttr(rName, "tags.key1", "new_value1"),
				),
			},
		},
	})
}

func testAccCheckContainerServiceDestroy(s *terraform.State) error {
	conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn

	for _, r := range s.RootModule().Resources {
		if r.Type != "awslightsail_container_service" {
			continue
		}

		input := lightsail.GetContainerServicesInput{
			ServiceName: aws.String(r.Primary.ID),
		}

		_, err := conn.GetContainerServices(context.TODO(), &input)
		if err == nil {
			return fmt.Errorf("container service still exists: %s", r.Primary.ID)
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

func testAccCheckContainerServiceExists(rName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rName]
		if !ok {
			return fmt.Errorf("not finding Lightsail Container Service (%s)", rName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Lightsail Container Service ID is set")
		}

		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn
		_, err := conn.GetContainerServices(context.TODO(), &lightsail.GetContainerServicesInput{
			ServiceName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccContainerServiceConfigBasic(lName string) string {
	return fmt.Sprintf(`
resource "awslightsail_container_service" "test" {
  name  = %q
  power = "nano"
  scale = 1
}
`, lName)
}

func testAccContainerServiceConfigIsDisabled(lName string) string {
	return fmt.Sprintf(`
resource "awslightsail_container_service" "test" {
  name        = %q
  power       = "nano"
  scale       = 1
  is_disabled = true
}
`, lName)
}

func testAccContainerServiceConfigPower(lName string) string {
	return fmt.Sprintf(`
resource "awslightsail_container_service" "test" {
  name  = %q
  power = "micro"
  scale = 1
}
`, lName)
}

func testAccContainerServiceConfigScale(lName string) string {
	return fmt.Sprintf(`
resource "awslightsail_container_service" "test" {
  name  = %q
  power = "nano"
  scale = 2
}
`, lName)
}

func testAccContainerServiceConfigTag1(lName string) string {
	return fmt.Sprintf(`
resource "awslightsail_container_service" "test" {
  name  = %q
  power = "nano"
  scale = 1

  tags = {
    key1 = "value1"
    key2 = "value2"
  }
}
`, lName)
}

func testAccContainerServiceConfigTag2(lName string) string {
	return fmt.Sprintf(`
resource "awslightsail_container_service" "test" {
  name  = %q
  power = "nano"
  scale = 1
  tags = {
    key1 = "new_value1"
  }
}
`, lName)
}
