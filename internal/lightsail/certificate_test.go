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

func TestAccCertificate_basic(t *testing.T) {
	rName := "awslightsail_certificate.test"
	lName := fmt.Sprintf("tf-test-lightsail-%s.com", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_basic(lName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(rName),
					resource.TestCheckResourceAttr(rName, "name", lName),
					resource.TestCheckResourceAttr(rName, "domain_name", lName),
					resource.TestCheckResourceAttr(rName, "subject_alternative_names.#", "0"),
					resource.TestCheckTypeSetElemNestedAttrs(rName, "domain_validation_options.*", map[string]string{
						"domain_name":          lName,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttrSet(rName, "arn"),
					resource.TestCheckResourceAttrSet(rName, "created_at"),
					resource.TestCheckResourceAttr(rName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccCertificate_SubjectAlternativeNames(t *testing.T) {
	rName := "awslightsail_certificate.test"
	lName := fmt.Sprintf("tf-test-lightsail-%s.com", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_SubjectAlternativeNames(lName, "www.test.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(rName),
					resource.TestCheckResourceAttr(rName, "name", lName),
					resource.TestCheckResourceAttr(rName, "domain_name", lName),
					resource.TestCheckResourceAttr(rName, "subject_alternative_names.#", "1"),
					resource.TestCheckTypeSetElemAttr(rName, "subject_alternative_names.*", "www.test.com"),
					resource.TestCheckTypeSetElemNestedAttrs(rName, "domain_validation_options.*", map[string]string{
						"domain_name":          lName,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttrSet(rName, "arn"),
					resource.TestCheckResourceAttrSet(rName, "created_at"),
					resource.TestCheckResourceAttr(rName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccCertificate_Tags(t *testing.T) {
	rName := "awslightsail_certificate.test"
	lName := fmt.Sprintf("tf-test-lightsail-%s.com", sdkacctest.RandString(5))

	resource.Test(t, resource.TestCase{
		Providers: testhelper.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfigTags1(lName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateExists(rName),
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
				Config: testAccCertificateConfigTags2(lName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateExists(rName),
					resource.TestCheckResourceAttr(rName, "tags.%", "2"),
					resource.TestCheckResourceAttr(rName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(rName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccCertificateConfigTags1(lName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateExists(rName),
					resource.TestCheckResourceAttr(rName, "tags.%", "1"),
					resource.TestCheckResourceAttr(rName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccCertificate_disappears(t *testing.T) {
	rName := "awslightsail_certificate.test"
	lName := fmt.Sprintf("tf-test-lightsail-%s.com", sdkacctest.RandString(5))

	testDestroy := func(*terraform.State) error {
		// reach out and DELETE the Certificate
		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn
		_, err := conn.DeleteCertificate(context.TODO(), &lightsail.DeleteCertificateInput{
			CertificateName: aws.String(lName),
		})

		if err != nil {
			return fmt.Errorf("error deleting Lightsail Certificate in disappear test")
		}

		// sleep 7 seconds to give it time, so we don't have to poll
		time.Sleep(7 * time.Second)

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_basic(lName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateExists(rName),
					testDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCertificateDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "awslightsail_certificate" {
			continue
		}

		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn

		respCertificate, err := conn.GetCertificates(context.TODO(), &lightsail.GetCertificatesInput{
			CertificateName: aws.String(rs.Primary.ID),
		})

		if err == nil {
			if len(respCertificate.Certificates) > 0 {
				return fmt.Errorf("Certificate %q still exists", rs.Primary.ID)
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

func testAccCheckCertificateExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Certificate ID is set")
		}

		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn

		respCertificate, err := conn.GetCertificates(context.TODO(), &lightsail.GetCertificatesInput{
			CertificateName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if respCertificate == nil || respCertificate.Certificates == nil {
			return fmt.Errorf("Certificate (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCertificateConfig_basic(lName string) string {
	return fmt.Sprintf(`
resource "awslightsail_certificate" "test" {
  name        = %[1]q
  domain_name = %[1]q
}
`, lName)
}

func testAccCertificateConfig_SubjectAlternativeNames(lName string, san string) string {
	return fmt.Sprintf(`
resource "awslightsail_certificate" "test" {
  name                      = %[1]q
  domain_name               = %[1]q
  subject_alternative_names = [%[2]q]
}
`, lName, san)
}

func testAccCertificateConfigTags1(rName string, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "awslightsail_certificate" "test" {
  name        = %[1]q
  domain_name = %[1]q
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccCertificateConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "awslightsail_certificate" "test" {
  name        = %[1]q
  domain_name = %[1]q
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
