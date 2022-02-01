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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// AWS Accounts are limited to 5 Load Balancers per account.
func TestAccBucket_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Bucket": {
			"basic":             testAccBucket_basic,
			"Name":              testAccBucket_Name,
			"VersioningEnabled": testAccBucket_VersioningEnabled,
			"BundleId":          testAccBucket_BundleId,
			"Tags":              testAccBucket_Tags,
			"disappears":        testAccBucket_disappears,
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

func testAccBucket_basic(t *testing.T) {
	rName := "awslightsail_bucket.test"
	lName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfigBasic(lName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(rName),
					resource.TestCheckResourceAttr(rName, "bundle_id", "small_1_0"),
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

func testAccBucket_Name(t *testing.T) {
	rName := "awslightsail_bucket.test"
	lName := acctest.RandomWithPrefix("tf-acc-test")
	lName2 := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		Providers: testhelper.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfigBasic(lName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(rName),
					resource.TestCheckResourceAttrSet(rName, "bundle_id"),
					resource.TestCheckResourceAttrSet(rName, "arn"),
				),
			},
			{
				Config: testAccBucketConfigBasic(lName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(rName),
					resource.TestCheckResourceAttrSet(rName, "bundle_id"),
					resource.TestCheckResourceAttrSet(rName, "arn"),
				),
			},
		},
	})
}

func testAccBucket_VersioningEnabled(t *testing.T) {
	rName := "awslightsail_bucket.test"
	lName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		Providers: testhelper.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfigVersinoingEnabled(lName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(rName),
					resource.TestCheckResourceAttr(rName, string("versioning_enabled"), "true"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketConfigVersinoingEnabled(lName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(rName),
					resource.TestCheckResourceAttr(rName, string("versioning_enabled"), "false"),
				),
			},
		},
	})
}

func testAccBucket_BundleId(t *testing.T) {
	rName := "awslightsail_bucket.test"
	lName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		Providers: testhelper.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfigBundleId(lName, "small_1_0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(rName),
					resource.TestCheckResourceAttr(rName, "bundle_id", "small_1_0"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketConfigBundleId(lName, "medium_1_0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(rName),
					resource.TestCheckResourceAttr(rName, "bundle_id", "medium_1_0"),
				),
			},
		},
	})
}

func testAccBucket_Tags(t *testing.T) {
	rName := "awslightsail_bucket.test"
	lName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		Providers: testhelper.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfigTags1(lName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(rName),
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
				Config: testAccBucketConfigTags2(lName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(rName),
					resource.TestCheckResourceAttr(rName, "tags.%", "2"),
					resource.TestCheckResourceAttr(rName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(rName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccBucketConfigTags1(lName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(rName),
					resource.TestCheckResourceAttr(rName, "tags.%", "1"),
					resource.TestCheckResourceAttr(rName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccBucket_disappears(t *testing.T) {
	rName := "awslightsail_bucket.test"
	lName := acctest.RandomWithPrefix("tf-acc-test")

	testDestroy := func(*terraform.State) error {
		// reach out and DELETE the Load Balancer
		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn
		resp, err := conn.DeleteBucket(context.TODO(), &lightsail.DeleteBucketInput{
			BucketName: aws.String(lName),
		})

		if err != nil {
			return fmt.Errorf("error deleting Lightsail Instance in disappear test")
		}

		op := resp.Operations[0]

		err = testhelper.WaitLightsailOperation(conn, op.Id)
		if err != nil {
			return fmt.Errorf("Error waiting for load balancer attatchment (%s) to become ready: %s", lName, err)
		}

		return nil
	}

	resource.Test(t, resource.TestCase{
		Providers: testhelper.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfigBasic(lName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(rName),
					testDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckBucketExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No LightsailBucket ID is set")
		}

		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn

		resp, err := conn.GetBuckets(context.TODO(), &lightsail.GetBucketsInput{
			BucketName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if resp == nil || resp.Buckets == nil {
			return fmt.Errorf("Bucket (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBucketDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "awslightsail_bucket" {
			continue
		}

		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn

		resp, err := conn.GetBuckets(context.TODO(), &lightsail.GetBucketsInput{
			BucketName: aws.String(rs.Primary.ID),
		})

		if err == nil {
			if resp.Buckets != nil {
				return fmt.Errorf("Bucket %q still exists", rs.Primary.ID)
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

func testAccBucketConfigBasic(rName string) string {
	return fmt.Sprintf(`
resource "awslightsail_bucket" "test" {
  name      = %[1]q
  bundle_id = "small_1_0"
}
`, rName)
}

func testAccBucketConfigVersinoingEnabled(rName string, rVersioningEnabled bool) string {
	return fmt.Sprintf(`
resource "awslightsail_bucket" "test" {
  name               = %[1]q
  bundle_id          = "small_1_0"
  versioning_enabled = %[2]t
}
`, rName, rVersioningEnabled)
}

func testAccBucketConfigBundleId(rName string, rBundleId string) string {
	return fmt.Sprintf(`
resource "awslightsail_bucket" "test" {
  name      = %[1]q
  bundle_id = %[2]q
}
`, rName, rBundleId)
}

func testAccBucketConfigTags1(rName string, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "awslightsail_bucket" "test" {
  name      = %[1]q
  bundle_id = "small_1_0"
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccBucketConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "awslightsail_bucket" "test" {
  name      = %[1]q
  bundle_id = "small_1_0"
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
