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

func TestAccDisk_basic(t *testing.T) {
	rName := "awslightsail_disk.test"
	lName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckDiskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDiskConfigBasic(lName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskExists(rName),
					resource.TestCheckResourceAttr(rName, "size_in_gb", "8"),
					resource.TestCheckResourceAttr(rName, "name", lName),
					resource.TestCheckResourceAttrSet(rName, "arn"),
					resource.TestCheckResourceAttrSet(rName, "created_at"),
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

func TestAccDisk_Name(t *testing.T) {
	rName := "awslightsail_disk.test"
	lName := acctest.RandomWithPrefix("tf-acc-test")
	lightsailNameWithSpaces := fmt.Sprint(rName, "string with spaces")
	lightsailNameWithStartingDigit := fmt.Sprintf("01-%s", rName)
	lightsailNameWithUnderscore := fmt.Sprintf("%s_123456", rName)

	resource.Test(t, resource.TestCase{
		Providers: testhelper.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config:      testAccDiskConfigBasic(lightsailNameWithSpaces),
				ExpectError: regexp.MustCompile(`must contain only alphanumeric characters, underscores, hyphens, and dots`),
			},
			{
				Config:      testAccDiskConfigBasic(lightsailNameWithStartingDigit),
				ExpectError: regexp.MustCompile(`must begin with an alphabetic character`),
			},
			{
				Config: testAccDiskConfigBasic(lName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDiskExists(rName),
					resource.TestCheckResourceAttrSet(rName, "size_in_gb"),
					resource.TestCheckResourceAttrSet(rName, "name"),
				),
			},
			{
				Config: testAccDiskConfigBasic(lightsailNameWithUnderscore),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDiskExists(rName),
					resource.TestCheckResourceAttrSet(rName, "size_in_gb"),
					resource.TestCheckResourceAttrSet(rName, "name"),
				),
			},
		},
	})
}

func TestAccDisk_Tags(t *testing.T) {
	rName := "awslightsail_disk.test"
	lName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		Providers: testhelper.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: testAccDiskConfigTags1(lName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskExists(rName),
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
				Config: testAccDiskConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskExists(rName),
					resource.TestCheckResourceAttr(rName, "tags.%", "2"),
					resource.TestCheckResourceAttr(rName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(rName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDiskConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskExists(rName),
					resource.TestCheckResourceAttr(rName, "tags.%", "1"),
					resource.TestCheckResourceAttr(rName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccDisk_disappears(t *testing.T) {
	rName := "awslightsail_disk.test"
	lName := acctest.RandomWithPrefix("tf-acc-test")

	testDestroy := func(*terraform.State) error {
		// reach out and DELETE the Disk
		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn
		_, err := conn.DeleteDisk(context.TODO(), &lightsail.DeleteDiskInput{
			DiskName: aws.String(lName),
		})

		if err != nil {
			return fmt.Errorf("error deleting Lightsail Instance in disappear test")
		}

		// sleep 7 seconds to give it time, so we don't have to poll
		time.Sleep(7 * time.Second)

		return nil
	}

	resource.Test(t, resource.TestCase{
		Providers: testhelper.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: testAccDiskConfigBasic(lName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskExists(rName),
					testDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDiskExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No LightsailDisk ID is set")
		}

		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn

		respDisk, err := conn.GetDisk(context.TODO(), &lightsail.GetDiskInput{
			DiskName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if respDisk == nil || respDisk.Disk == nil {
			return fmt.Errorf("Disk (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDiskDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "awslightsail_disk" {
			continue
		}

		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn

		resp, err := conn.GetDisk(context.TODO(), &lightsail.GetDiskInput{
			DiskName: aws.String(rs.Primary.ID),
		})

		if err == nil {
			if resp.Disk != nil {
				return fmt.Errorf("Lightsail Disk %q still exists", rs.Primary.ID)
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

func testAccDiskConfigBasic(rName string) string {
	return fmt.Sprintf(`
data "awslightsail_availability_zones" "all" {}

resource "awslightsail_disk" "test" {
  name              = %[1]q
  size_in_gb        = 8
  availability_zone = data.awslightsail_availability_zones.all.names[0]
}
`, rName)
}

func testAccDiskConfigTags1(rName string, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
data "awslightsail_availability_zones" "all" {}

resource "awslightsail_disk" "test" {
  name              = %[1]q
  size_in_gb        = 8
  availability_zone = data.awslightsail_availability_zones.all.names[0]
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccDiskConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
data "awslightsail_availability_zones" "all" {}

resource "awslightsail_disk" "test" {
  name              = %[1]q
  size_in_gb        = 8
  availability_zone = data.awslightsail_availability_zones.all.names[0]
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
