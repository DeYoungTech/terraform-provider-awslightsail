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
	resourceName := "awslightsail_instance.instance"
	lightsailName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_basic(lightsailName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttrSet(resourceName, "blueprint_id"),
					resource.TestCheckResourceAttrSet(resourceName, "bundle_id"),
					resource.TestMatchResourceAttr(resourceName, "ipv6_address", regexp.MustCompile(`([a-f0-9]{1,4}:){7}[a-f0-9]{1,4}`)),
					resource.TestCheckResourceAttr(resourceName, "ipv6_addresses.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "key_pair_name"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestMatchResourceAttr(resourceName, "ram_size", regexp.MustCompile(`\d+(.\d+)?`)),
				),
			},
		},
	})
}

func TestAccInstance_name(t *testing.T) {

	resourceName := "awslightsail_instance.instance"
	lightsailName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())
	lightsailNameWithSpaces := fmt.Sprint(lightsailName, "string with spaces")
	lightsailNameWithStartingDigit := fmt.Sprintf("01-%s", lightsailName)
	lightsailNameWithUnderscore := fmt.Sprintf("%s_123456", lightsailName)

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccInstanceConfig_basic(lightsailNameWithSpaces),
				ExpectError: regexp.MustCompile(`must contain only alphanumeric characters, underscores, hyphens, and dots`),
			},
			{
				Config:      testAccInstanceConfig_basic(lightsailNameWithStartingDigit),
				ExpectError: regexp.MustCompile(`must begin with an alphabetic character`),
			},
			{
				Config: testAccInstanceConfig_basic(lightsailName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttrSet(resourceName, "blueprint_id"),
					resource.TestCheckResourceAttrSet(resourceName, "bundle_id"),
					resource.TestCheckResourceAttrSet(resourceName, "key_pair_name"),
				),
			},
			{
				Config: testAccInstanceConfig_basic(lightsailNameWithUnderscore),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttrSet(resourceName, "blueprint_id"),
					resource.TestCheckResourceAttrSet(resourceName, "bundle_id"),
					resource.TestCheckResourceAttrSet(resourceName, "key_pair_name"),
				),
			},
		},
	})
}

func TestAccInstance_tags(t *testing.T) {
	resourceName := "awslightsail_instance.instance"
	lightsailName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_tags1(lightsailName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttrSet(resourceName, "blueprint_id"),
					resource.TestCheckResourceAttrSet(resourceName, "bundle_id"),
					resource.TestCheckResourceAttrSet(resourceName, "key_pair_name"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
				),
			},
			{
				Config: testAccInstanceConfig_tags2(lightsailName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttrSet(resourceName, "blueprint_id"),
					resource.TestCheckResourceAttrSet(resourceName, "bundle_id"),
					resource.TestCheckResourceAttrSet(resourceName, "key_pair_name"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
				),
			},
		},
	})
}

func TestAccInstance_disappears(t *testing.T) {
	resourceName := "awslightsail_instance.instance"
	lightsailName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())

	testDestroy := func(*terraform.State) error {
		// reach out and DELETE the Instance
		conn := testhelper.GetProvider().Meta().(*lightsail.Client)
		_, err := conn.DeleteInstance(context.TODO(), &lightsail.DeleteInstanceInput{
			InstanceName: aws.String(lightsailName),
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
				Config: testAccInstanceConfig_basic(lightsailName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName),
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

func testAccInstanceConfig_basic(lightsailName string) string {
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
`, lightsailName)
}

func testAccInstanceConfig_tags1(lightsailName string) string {
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
`, lightsailName)
}

func testAccInstanceConfig_tags2(lightsailName string) string {
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
`, lightsailName)
}
