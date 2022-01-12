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
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccKeyPair_basic(t *testing.T) {
	rName := "awslightsail_key_pair.test"
	lName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckKeyPairDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPairConfig_basic(lName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeyPairExists(rName),
					resource.TestCheckResourceAttrSet(rName, "arn"),
					resource.TestCheckResourceAttrSet(rName, "fingerprint"),
					resource.TestCheckResourceAttrSet(rName, "public_key"),
					resource.TestCheckResourceAttrSet(rName, "private_key"),
				),
			},
		},
	})
}

func TestAccKeyPair_publicKey(t *testing.T) {
	rName := "awslightsail_key_pair.test"
	lName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())

	publicKey, _, err := sdkacctest.RandSSHKeyPair(testhelper.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckKeyPairDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPairConfig_imported(lName, publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeyPairExists(rName),
					resource.TestCheckResourceAttrSet(rName, "arn"),
					resource.TestCheckResourceAttrSet(rName, "fingerprint"),
					resource.TestCheckResourceAttrSet(rName, "public_key"),
					resource.TestCheckNoResourceAttr(rName, "encrypted_fingerprint"),
					resource.TestCheckNoResourceAttr(rName, "encrypted_private_key"),
					resource.TestCheckNoResourceAttr(rName, "private_key"),
				),
			},
		},
	})
}

func TestAccKeyPair_encrypted(t *testing.T) {
	rName := "awslightsail_key_pair.test"
	lName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckKeyPairDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPairConfig_encrypted(lName, testKeyPairPubKey1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeyPairExists(rName),
					resource.TestCheckResourceAttrSet(rName, "arn"),
					resource.TestCheckResourceAttrSet(rName, "fingerprint"),
					resource.TestCheckResourceAttrSet(rName, "encrypted_fingerprint"),
					resource.TestCheckResourceAttrSet(rName, "encrypted_private_key"),
					resource.TestCheckResourceAttrSet(rName, "public_key"),
					resource.TestCheckNoResourceAttr(rName, "private_key"),
				),
			},
		},
	})
}

func TestAccKeyPair_namePrefix(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckKeyPairDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPairConfig_prefixed(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeyPairExists("awslightsail_key_pair.lightsail_key_pair_test_omit"),
					testAccCheckKeyPairExists("awslightsail_key_pair.lightsail_key_pair_test_prefixed"),
					resource.TestCheckResourceAttrSet("awslightsail_key_pair.lightsail_key_pair_test_omit", "name"),
					resource.TestCheckResourceAttrSet("awslightsail_key_pair.lightsail_key_pair_test_prefixed", "name"),
				),
			},
		},
	})
}

func testAccCheckKeyPairExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No KeyPair set")
		}

		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn

		respKeyPair, err := conn.GetKeyPair(context.TODO(), &lightsail.GetKeyPairInput{
			KeyPairName: aws.String(rs.Primary.Attributes["name"]),
		})

		if err != nil {
			return err
		}

		if respKeyPair == nil || respKeyPair.KeyPair == nil {
			return fmt.Errorf("KeyPair (%s) not found", rs.Primary.Attributes["name"])
		}

		return nil
	}
}

func testAccCheckKeyPairDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "awslightsail_key_pair" {
			continue
		}

		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn

		respKeyPair, err := conn.GetKeyPair(context.TODO(), &lightsail.GetKeyPairInput{
			KeyPairName: aws.String(rs.Primary.Attributes["name"]),
		})

		if err == nil {
			if respKeyPair.KeyPair != nil {
				return fmt.Errorf("KeyPair %q still exists", rs.Primary.ID)
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

func testAccKeyPairConfig_basic(lightsailName string) string {
	return fmt.Sprintf(`
resource "awslightsail_key_pair" "test" {
  name = "%s"
}
`, lightsailName)
}

func testAccKeyPairConfig_imported(lightsailName, publicKey string) string {
	return fmt.Sprintf(`
resource "awslightsail_key_pair" "test" {
  name       = %[1]q
  public_key = "%[2]s"
}
`, lightsailName, publicKey)
}

func testAccKeyPairConfig_encrypted(lightsailName, key string) string {
	return fmt.Sprintf(`
resource "awslightsail_key_pair" "test" {
  name    = %[1]q
  pgp_key = <<EOF
%[2]s
EOF
}
`, lightsailName, key)
}

func testAccKeyPairConfig_prefixed() string {
	return `
resource "awslightsail_key_pair" "lightsail_key_pair_test_omit" {}
resource "awslightsail_key_pair" "lightsail_key_pair_test_prefixed" {
  name_prefix = "cts"
}
`
}

const testKeyPairPubKey1 = `mQENBFXbjPUBCADjNjCUQwfxKL+RR2GA6pv/1K+zJZ8UWIF9S0lk7cVIEfJiprzzwiMwBS5cD0da
rGin1FHvIWOZxujA7oW0O2TUuatqI3aAYDTfRYurh6iKLC+VS+F7H+/mhfFvKmgr0Y5kDCF1j0T/
063QZ84IRGucR/X43IY7kAtmxGXH0dYOCzOe5UBX1fTn3mXGe2ImCDWBH7gOViynXmb6XNvXkP0f
sF5St9jhO7mbZU9EFkv9O3t3EaURfHopsCVDOlCkFCw5ArY+DUORHRzoMX0PnkyQb5OzibkChzpg
8hQssKeVGpuskTdz5Q7PtdW71jXd4fFVzoNH8fYwRpziD2xNvi6HABEBAAG0EFZhdWx0IFRlc3Qg
S2V5IDGJATgEEwECACIFAlXbjPUCGy8GCwkIBwMCBhUIAgkKCwQWAgMBAh4BAheAAAoJEOfLr44B
HbeTo+sH/i7bapIgPnZsJ81hmxPj4W12uvunksGJiC7d4hIHsG7kmJRTJfjECi+AuTGeDwBy84TD
cRaOB6e79fj65Fg6HgSahDUtKJbGxj/lWzmaBuTzlN3CEe8cMwIPqPT2kajJVdOyrvkyuFOdPFOE
A7bdCH0MqgIdM2SdF8t40k/ATfuD2K1ZmumJ508I3gF39jgTnPzD4C8quswrMQ3bzfvKC3klXRlB
C0yoArn+0QA3cf2B9T4zJ2qnvgotVbeK/b1OJRNj6Poeo+SsWNc/A5mw7lGScnDgL3yfwCm1gQXa
QKfOt5x+7GqhWDw10q+bJpJlI10FfzAnhMF9etSqSeURBRW5AQ0EVduM9QEIAL53hJ5bZJ7oEDCn
aY+SCzt9QsAfnFTAnZJQrvkvusJzrTQ088eUQmAjvxkfRqnv981fFwGnh2+I1Ktm698UAZS9Jt8y
jak9wWUICKQO5QUt5k8cHwldQXNXVXFa+TpQWQR5yW1a9okjh5o/3d4cBt1yZPUJJyLKY43Wvptb
6EuEsScO2DnRkh5wSMDQ7dTooddJCmaq3LTjOleRFQbu9ij386Do6jzK69mJU56TfdcydkxkWF5N
ZLGnED3lq+hQNbe+8UI5tD2oP/3r5tXKgMy1R/XPvR/zbfwvx4FAKFOP01awLq4P3d/2xOkMu4Lu
9p315E87DOleYwxk+FoTqXEAEQEAAYkCPgQYAQIACQUCVduM9QIbLgEpCRDny6+OAR23k8BdIAQZ
AQIABgUCVduM9QAKCRAID0JGyHtSGmqYB/4m4rJbbWa7dBJ8VqRU7ZKnNRDR9CVhEGipBmpDGRYu
lEimOPzLUX/ZXZmTZzgemeXLBaJJlWnopVUWuAsyjQuZAfdd8nHkGRHG0/DGum0l4sKTta3OPGHN
C1z1dAcQ1RCr9bTD3PxjLBczdGqhzw71trkQRBRdtPiUchltPMIyjUHqVJ0xmg0hPqFic0fICsr0
YwKoz3h9+QEcZHvsjSZjgydKvfLYcm+4DDMCCqcHuJrbXJKUWmJcXR0y/+HQONGrGJ5xWdO+6eJi
oPn2jVMnXCm4EKc7fcLFrz/LKmJ8seXhxjM3EdFtylBGCrx3xdK0f+JDNQaC/rhUb5V2XuX6VwoH
/AtY+XsKVYRfNIupLOUcf/srsm3IXT4SXWVomOc9hjGQiJ3rraIbADsc+6bCAr4XNZS7moViAAcI
PXFv3m3WfUlnG/om78UjQqyVACRZqqAGmuPq+TSkRUCpt9h+A39LQWkojHqyob3cyLgy6z9Q557O
9uK3lQozbw2gH9zC0RqnePl+rsWIUU/ga16fH6pWc1uJiEBt8UZGypQ/E56/343epmYAe0a87sHx
8iDV+dNtDVKfPRENiLOOc19MmS+phmUyrbHqI91c0pmysYcJZCD3a502X1gpjFbPZcRtiTmGnUKd
OIu60YPNE4+h7u2CfYyFPu3AlUaGNMBlvy6PEpU=`
