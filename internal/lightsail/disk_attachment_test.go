package lightsail_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
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

func TestAccDiskAttachment_basic(t *testing.T) {
	rName := "awslightsail_disk_attachment.test"
	ldName := acctest.RandomWithPrefix("tf-acc-test")
	liName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckDiskAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDiskAttachmentConfigBasic(ldName, liName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskAttachmentExists(rName),
				),
			},
		},
	})
}

func TestAccDiskAttachment_disappears(t *testing.T) {
	rName := "awslightsail_disk_attachment.test"
	ldName := acctest.RandomWithPrefix("tf-acc-test")
	liName := acctest.RandomWithPrefix("tf-acc-test")

	testDestroy := func(*terraform.State) error {
		// reach out and detach Instance from the Load Balancer
		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn

		// You must first Stop an instance in order to detach a disk.
		stopresp, stoperr := conn.StopInstance(context.TODO(), &lightsail.StopInstanceInput{
			InstanceName: aws.String(liName),
		})

		if stoperr != nil {
			return stoperr
		}

		stopop := stopresp.Operations[0]

		stoperr = testhelper.WaitLightsailOperation(conn, stopop.Id)
		if stoperr != nil {
			return fmt.Errorf("Error waiting for Instance (%s) to Stop: %s", liName, stoperr)
		}

		resp, err := conn.DetachDisk(context.TODO(), &lightsail.DetachDiskInput{
			DiskName: aws.String(ldName),
		})

		if err != nil {
			return fmt.Errorf("error detaching disk")
		}

		op := resp.Operations[0]

		err = testhelper.WaitLightsailOperation(conn, op.Id)
		if err != nil {
			return fmt.Errorf("Error waiting for disk attatchment (%s) to become ready: %s", ldName, err)
		}

		return nil
	}

	resource.Test(t, resource.TestCase{
		Providers: testhelper.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: testAccDiskAttachmentConfigBasic(ldName, liName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskAttachmentExists(rName),
					testDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDiskAttachmentExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No LightsailDiskAttachment ID is set")
		}

		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn

		id_parts := strings.SplitN(rs.Primary.ID, "_", -1)

		if len(id_parts) != 2 {
			return nil
		}

		ldName := id_parts[0]
		liName := id_parts[1]

		resp, err := conn.GetDisk(context.TODO(), &lightsail.GetDiskInput{
			DiskName: aws.String(ldName),
		})

		if err != nil {
			return err
		}

		if resp == nil || resp.Disk == nil {
			return fmt.Errorf("Disk (%s) not found", ldName)
		}

		if *resp.Disk.AttachedTo != liName {
			return fmt.Errorf("Disk Attachment (%s) not found", ldName)
		}

		return nil
	}
}

func testAccCheckDiskAttachmentDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "awslightsail_disk_attachment" {
			continue
		}

		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn

		id_parts := strings.SplitN(rs.Primary.ID, "_", -1)
		if len(id_parts) != 2 {
			return nil
		}

		ldName := id_parts[0]

		resp, err := conn.GetDisk(context.TODO(), &lightsail.GetDiskInput{
			DiskName: aws.String(ldName),
		})

		// Verify the error
		if err != nil {
			var oe *smithy.OperationError
			if errors.As(err, &oe) {
				log.Printf("failed to call service: %s, operation: %s, error: %v", oe.Service(), oe.Operation(), oe.Unwrap())
			}
			return nil
		}

		if *resp.Disk.AttachedTo == ldName {
			return fmt.Errorf("Disk Attachment (%s) still exists.", ldName)
		}

		return err
	}

	return nil
}

func testAccDiskAttachmentConfigBasic(ldName string, liName string) string {
	return fmt.Sprintf(`
data "awslightsail_availability_zones" "all" {}

resource "awslightsail_disk" "test" {
  name              = %[1]q
  size_in_gb        = "8"
  availability_zone = data.awslightsail_availability_zones.all.names[0]
}

resource "awslightsail_instance" "test" {
  name              = %[2]q
  availability_zone = data.awslightsail_availability_zones.all.names[0]
  blueprint_id      = "amazon_linux"
  bundle_id         = "nano_1_0"
}

resource "awslightsail_disk_attachment" "test" {
  disk_name     = awslightsail_disk.test.name
  instance_name = awslightsail_instance.test.name
  disk_path     = "/dev/xvdf"
}
`, ldName, liName)
}
