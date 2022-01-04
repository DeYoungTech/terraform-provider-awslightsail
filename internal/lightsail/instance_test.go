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
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/testhelper"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccInstance_basic(t *testing.T) {
	rName := "awslightsail_instance.instance"
	lName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_basic(lName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(rName),
					resource.TestCheckResourceAttrSet(rName, "availability_zone"),
					resource.TestCheckResourceAttrSet(rName, "blueprint_id"),
					resource.TestCheckResourceAttrSet(rName, "bundle_id"),
					resource.TestMatchResourceAttr(rName, "ipv6_address", regexp.MustCompile(`([a-f0-9]{1,4}:){7}[a-f0-9]{1,4}`)),
					resource.TestCheckResourceAttr(rName, "ipv6_addresses.#", "1"),
					resource.TestCheckResourceAttrSet(rName, "key_pair_name"),
					resource.TestCheckResourceAttr(rName, "tags.%", "0"),
					resource.TestMatchResourceAttr(rName, "ram_size", regexp.MustCompile(`\d+(.\d+)?`)),
				),
			},
		},
	})
}

func TestAccInstance_name(t *testing.T) {

	rName := "awslightsail_instance.instance"
	lName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())
	lNameWithSpaces := fmt.Sprint(lName, "string with spaces")
	lNameWithStartingDigit := fmt.Sprintf("01-%s", lName)
	lNameWithUnderscore := fmt.Sprintf("%s_123456", lName)

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccInstanceConfig_basic(lNameWithSpaces),
				ExpectError: regexp.MustCompile(`must contain only alphanumeric characters, underscores, hyphens, and dots`),
			},
			{
				Config:      testAccInstanceConfig_basic(lNameWithStartingDigit),
				ExpectError: regexp.MustCompile(`must begin with an alphabetic character`),
			},
			{
				Config: testAccInstanceConfig_basic(lName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(rName),
					resource.TestCheckResourceAttrSet(rName, "availability_zone"),
					resource.TestCheckResourceAttrSet(rName, "blueprint_id"),
					resource.TestCheckResourceAttrSet(rName, "bundle_id"),
					resource.TestCheckResourceAttrSet(rName, "key_pair_name"),
				),
			},
			{
				Config: testAccInstanceConfig_basic(lNameWithUnderscore),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(rName),
					resource.TestCheckResourceAttrSet(rName, "availability_zone"),
					resource.TestCheckResourceAttrSet(rName, "blueprint_id"),
					resource.TestCheckResourceAttrSet(rName, "bundle_id"),
					resource.TestCheckResourceAttrSet(rName, "key_pair_name"),
				),
			},
		},
	})
}

func TestAccInstance_tags(t *testing.T) {
	rName := "awslightsail_instance.instance"
	lName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_tags1(lName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(rName),
					resource.TestCheckResourceAttrSet(rName, "availability_zone"),
					resource.TestCheckResourceAttrSet(rName, "blueprint_id"),
					resource.TestCheckResourceAttrSet(rName, "bundle_id"),
					resource.TestCheckResourceAttrSet(rName, "key_pair_name"),
					resource.TestCheckResourceAttr(rName, "tags.%", "2"),
				),
			},
			{
				Config: testAccInstanceConfig_tags2(lName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(rName),
					resource.TestCheckResourceAttrSet(rName, "availability_zone"),
					resource.TestCheckResourceAttrSet(rName, "blueprint_id"),
					resource.TestCheckResourceAttrSet(rName, "bundle_id"),
					resource.TestCheckResourceAttrSet(rName, "key_pair_name"),
					resource.TestCheckResourceAttr(rName, "tags.%", "3"),
				),
			},
		},
	})
}

func TestAccInstance_disappears(t *testing.T) {
	rName := "awslightsail_instance.instance"
	lName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())

	testDestroy := func(*terraform.State) error {
		// reach out and DELETE the Instance
		conn := testhelper.GetProvider().Meta().(*lightsail.Client)
		_, err := conn.DeleteInstance(context.TODO(), &lightsail.DeleteInstanceInput{
			InstanceName: aws.String(lName),
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
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_basic(lName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(rName),
					testDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckInstanceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Instance ID is set")
		}

		conn := testhelper.GetProvider().Meta().(*lightsail.Client)

		respInstance, err := conn.GetInstance(context.TODO(), &lightsail.GetInstanceInput{
			InstanceName: aws.String(rs.Primary.Attributes["name"]),
		})

		if err != nil {
			return err
		}

		if respInstance == nil || respInstance.Instance == nil {
			return fmt.Errorf("Instance (%s) not found", rs.Primary.Attributes["name"])
		}

		return nil
	}
}

func testAccCheckInstanceDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "awslightsail_instance" {
			continue
		}

		conn := testhelper.GetProvider().Meta().(*lightsail.Client)

		respInstance, err := conn.GetInstance(context.TODO(), &lightsail.GetInstanceInput{
			InstanceName: aws.String(rs.Primary.Attributes["name"]),
		})

		if err == nil {
			if respInstance.Instance != nil {
				return fmt.Errorf("Instance %q still exists", rs.Primary.ID)
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

func testAccInstanceConfig_basic(lName string) string {
	return fmt.Sprintf(`
data "awslightsail_availability_zones" "all" {}

resource "awslightsail_instance" "instance" {
  name              = "%s"
  availability_zone = data.awslightsail_availability_zones.all.names[0]
  blueprint_id      = "amazon_linux"
  bundle_id         = "nano_1_0"
  tags ={
  }
}
`, lName)
}

func testAccInstanceConfig_tags1(lName string) string {
	return fmt.Sprintf(`
data "awslightsail_availability_zones" "all" {}

resource "awslightsail_instance" "instance" {
  name              = "%s"
  availability_zone = data.awslightsail_availability_zones.all.names[0]
  blueprint_id      = "amazon_linux"
  bundle_id         = "nano_1_0"

  tags = {
    Name       = "tf-test"
    KeyOnlyTag = ""
  }
}
`, lName)
}

func testAccInstanceConfig_tags2(lName string) string {
	return fmt.Sprintf(`
data "awslightsail_availability_zones" "all" {}

resource "awslightsail_instance" "instance" {
  name              = "%s"
  availability_zone = data.awslightsail_availability_zones.all.names[0]
  blueprint_id      = "amazon_linux"
  bundle_id         = "nano_1_0"

  tags = {
    Name       = "tf-test",
    KeyOnlyTag = ""
    ExtraName  = "tf-test"
  }
}
`, lName)
}
