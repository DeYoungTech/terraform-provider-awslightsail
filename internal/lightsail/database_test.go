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

func TestAccDatabase_basic(t *testing.T) {
	rName := "awslightsail_database.test"
	lName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckAWSDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfigBasic(lName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSDatabaseExists(rName),
					resource.TestCheckResourceAttr(rName, "name", lName),
					resource.TestCheckResourceAttr(rName, "blueprint_id", "mysql_8_0"),
					resource.TestCheckResourceAttr(rName, "bundle_id", "micro_1_0"),
					resource.TestCheckResourceAttr(rName, "master_database_name", "testdatabasename"),
					resource.TestCheckResourceAttr(rName, "master_username", "test"),
					resource.TestCheckResourceAttrSet(rName, "availability_zone"),
					resource.TestCheckResourceAttrSet(rName, "arn"),
					resource.TestCheckResourceAttrSet(rName, "created_at"),
					resource.TestCheckResourceAttrSet(rName, "engine"),
					resource.TestCheckResourceAttrSet(rName, "engine_version"),
					resource.TestCheckResourceAttrSet(rName, "cpu_count"),
					resource.TestCheckResourceAttrSet(rName, "ram_size"),
					resource.TestCheckResourceAttrSet(rName, "disk_size"),
					resource.TestCheckResourceAttrSet(rName, "master_endpoint_port"),
					resource.TestCheckResourceAttrSet(rName, "master_endpoint_address"),
					resource.TestCheckResourceAttrSet(rName, "support_code"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
		},
	})
}

func TestAccDatabase_Name(t *testing.T) {
	rName := "awslightsail_database.test"
	lName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckAWSDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfigBasic(lName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSDatabaseExists(rName),
					resource.TestCheckResourceAttr(rName, "name", lName),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
		},
	})
}

func TestAccDatabase_MasterDatabaseName(t *testing.T) {
	lName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())
	rName := "awslightsail_database.test"

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckAWSDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfigMasterDatabaseName(lName, "databasename1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSDatabaseExists(rName),
					resource.TestCheckResourceAttr(rName, "master_database_name", "databasename1"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
			{
				Config: testAccDatabaseConfigMasterDatabaseName(lName, "databasename2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSDatabaseExists(rName),
					resource.TestCheckResourceAttr(rName, "master_database_name", "databasename2"),
				),
			},
		},
	})
}

func TestAccDatabase_masterUserName(t *testing.T) {
	lName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())
	rName := "awslightsail_database.test"

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckAWSDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfigmasterUserName(lName, "uselName1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDatabaseExists(rName),
					resource.TestCheckResourceAttr(rName, "master_username", "uselName1"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
			{
				Config: testAccDatabaseConfigmasterUserName(lName, "uselName2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDatabaseExists(rName),
					resource.TestCheckResourceAttr(rName, "master_username", "uselName2"),
				),
			},
		},
	})
}

func TestAccDatabase_PreferredBackupWindow(t *testing.T) {
	lName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())
	rName := "awslightsail_database.test"

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckAWSDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfigPreferredBackupWindow(lName, "09:30-10:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDatabaseExists(rName),
					resource.TestCheckResourceAttr(rName, "preferred_backup_window", "09:30-10:00"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
			{
				Config: testAccDatabaseConfigPreferredBackupWindow(lName, "09:45-10:15"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDatabaseExists(rName),
					resource.TestCheckResourceAttr(rName, "preferred_backup_window", "09:45-10:15"),
				),
			},
		},
	})
}

func TestAccDatabase_PreferredMaintenanceWindow(t *testing.T) {
	lName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())
	rName := "awslightsail_database.test"

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckAWSDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfigPreferredMaintenanceWindow(lName, "tue:04:30-tue:05:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDatabaseExists(rName),
					resource.TestCheckResourceAttr(rName, "preferred_maintenance_window", "tue:04:30-tue:05:00"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
			{
				Config: testAccDatabaseConfigPreferredMaintenanceWindow(lName, "wed:06:00-wed:07:30"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDatabaseExists(rName),
					resource.TestCheckResourceAttr(rName, "preferred_maintenance_window", "wed:06:00-wed:07:30"),
				),
			},
		},
	})
}

func TestAccDatabase_PubliclyAccessible(t *testing.T) {
	lName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())
	rName := "awslightsail_database.test"

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckAWSDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfigPubliclyAccessible(lName, "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDatabaseExists(rName),
					resource.TestCheckResourceAttr(rName, "publicly_accessible", "true"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
			{
				Config: testAccDatabaseConfigPubliclyAccessible(lName, "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDatabaseExists(rName),
					resource.TestCheckResourceAttr(rName, "publicly_accessible", "false"),
				),
			},
		},
	})
}

func TestAccDatabase_BackupRetentionEnabled(t *testing.T) {
	lName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())
	rName := "awslightsail_database.test"

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckAWSDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfigBackupRetentionEnabled(lName, "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDatabaseExists(rName),
					resource.TestCheckResourceAttr(rName, "backup_retention_enabled", "true"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
			{
				Config: testAccDatabaseConfigBackupRetentionEnabled(lName, "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDatabaseExists(rName),
					resource.TestCheckResourceAttr(rName, "backup_retention_enabled", "false"),
				),
			},
		},
	})
}

func TestAccDatabase_FinalSnapshotName(t *testing.T) {
	lName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())
	rName := "awslightsail_database.test"
	sName := fmt.Sprintf("%s-snapshot", lName)

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckAWSDatabaseSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfigFinalSnapshotName(lName, sName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDatabaseExists(rName),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
		},
	})
}

func TestAccDatabase_Tags(t *testing.T) {
	lName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())
	rName := "awslightsail_database.test"

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckAWSDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfigTags1(lName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSDatabaseExists(rName),
					resource.TestCheckResourceAttr(rName, "tags.%", "1"),
					resource.TestCheckResourceAttr(rName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
			{
				Config: testAccDatabaseConfigTags2(lName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSDatabaseExists(rName),
					resource.TestCheckResourceAttr(rName, "tags.%", "2"),
					resource.TestCheckResourceAttr(rName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(rName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDatabaseConfigTags1(lName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDatabaseExists(rName),
					resource.TestCheckResourceAttr(rName, "tags.%", "1"),
					resource.TestCheckResourceAttr(rName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccDatabase_disappears(t *testing.T) {
	lName := fmt.Sprintf("tf-test-lightsail-%d", sdkacctest.RandInt())
	rName := "awslightsail_database.test"

	testDestroy := func(*terraform.State) error {
		// reach out and DELETE the Database
		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn

		_, err := conn.DeleteRelationalDatabase(context.TODO(), &lightsail.DeleteRelationalDatabaseInput{
			RelationalDatabaseName: aws.String(lName),
			SkipFinalSnapshot:      aws.Bool(true),
		})

		if err != nil {
			return fmt.Errorf("error deleting Lightsail Database in disappear test")
		}

		// sleep 7 seconds to give it time, so we don't have to poll
		time.Sleep(7 * time.Second)

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckAWSDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfigBasic(lName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDatabaseExists(rName),
					testDestroy),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSDatabaseExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Lightsail Database ID is set")
		}

		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn

		params := lightsail.GetRelationalDatabaseInput{
			RelationalDatabaseName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetRelationalDatabase(context.TODO(), &params)

		if err != nil {
			return err
		}

		if resp == nil || resp.RelationalDatabase == nil {
			return fmt.Errorf("Database (%s) not found", rs.Primary.Attributes["name"])
		}

		return nil
	}
}

func testAccCheckAWSDatabaseDestroy(s *terraform.State) error {
	conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "awslightsail_database" {
			continue
		}

		params := lightsail.GetRelationalDatabaseInput{
			RelationalDatabaseName: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetRelationalDatabase(context.TODO(), &params)

		if err == nil {
			return fmt.Errorf("Lightsail Database %q still exists", rs.Primary.ID)
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

func testAccCheckAWSDatabaseSnapshotDestroy(s *terraform.State) error {
	conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "awslightsail_database" {
			continue
		}

		// Try and delete the snapshot before we check for the cluster not found
		snapshot_identifier := fmt.Sprintf("%s-snapshot", rs.Primary.ID)

		log.Printf("[INFO] Deleting the Snapshot %s", snapshot_identifier)
		_, err := conn.DeleteRelationalDatabaseSnapshot(
			context.TODO(),
			&lightsail.DeleteRelationalDatabaseSnapshotInput{
				RelationalDatabaseSnapshotName: aws.String(snapshot_identifier),
			})

		if err != nil {
			return err
		}

		params := lightsail.GetRelationalDatabaseInput{
			RelationalDatabaseName: aws.String(rs.Primary.ID),
		}

		_, err = conn.GetRelationalDatabase(context.TODO(), &params)

		if err == nil {
			return fmt.Errorf("Lightsail Database %q still exists", rs.Primary.ID)
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

func testAccDatabaseConfigBasic(lName string) string {
	return fmt.Sprintf(`
data "awslightsail_availability_zones" "all" {}

resource "awslightsail_database" "test" {
  name                 = %[1]q
  availability_zone    = data.awslightsail_availability_zones.all.database_names[0]
  master_database_name = "testdatabasename"
  master_password      = "testdatabasepassword"
  master_username      = "test"
  blueprint_id         = "mysql_8_0"
  bundle_id            = "micro_1_0"
  skip_final_snapshot  = true
}
`, lName)
}

func testAccDatabaseConfigMasterDatabaseName(lName string, masterDatabaseName string) string {
	return fmt.Sprintf(`
data "awslightsail_availability_zones" "all" {}

resource "awslightsail_database" "test" {
  name                 = %[1]q
  availability_zone    = data.awslightsail_availability_zones.all.database_names[0]
  master_database_name = %[2]q
  master_password      = "testdatabasepassword"
  master_username      = "test"
  blueprint_id         = "mysql_8_0"
  bundle_id            = "micro_1_0"
  skip_final_snapshot  = true
}
	`, lName, masterDatabaseName)
}

func testAccDatabaseConfigmasterUserName(lName string, masterUserName string) string {
	return fmt.Sprintf(`
data "awslightsail_availability_zones" "all" {}

resource "awslightsail_database" "test" {
  name                 = %[1]q
  availability_zone    = data.awslightsail_availability_zones.all.database_names[0]
  master_database_name = "testdatabasename"
  master_password      = "testdatabasepassword"
  master_username      = %[2]q
  blueprint_id         = "mysql_8_0"
  bundle_id            = "micro_1_0"
  skip_final_snapshot  = true
}
	`, lName, masterUserName)
}

func testAccDatabaseConfigPreferredBackupWindow(lName string, preferredBackupWindow string) string {
	return fmt.Sprintf(`	
data "awslightsail_availability_zones" "all" {}

resource "awslightsail_database" "test" {
  name                    = %[1]q
  availability_zone       = data.awslightsail_availability_zones.all.database_names[0]
  master_database_name    = "testdatabasename"
  master_password         = "testdatabasepassword"
  master_username         = "test"
  blueprint_id            = "mysql_8_0"
  bundle_id               = "micro_1_0"
  preferred_backup_window = %[2]q
  apply_immediately       = true
  skip_final_snapshot     = true
}
`, lName, preferredBackupWindow)
}

func testAccDatabaseConfigPreferredMaintenanceWindow(lName string, preferredMaintenanceWindow string) string {
	return fmt.Sprintf(`	
data "awslightsail_availability_zones" "all" {}

resource "awslightsail_database" "test" {
  name                         = %[1]q
  availability_zone            = data.awslightsail_availability_zones.all.database_names[0]
  master_database_name         = "testdatabasename"
  master_password              = "testdatabasepassword"
  master_username              = "test"
  blueprint_id                 = "mysql_8_0"
  bundle_id                    = "micro_1_0"
  preferred_maintenance_window = %[2]q
  apply_immediately            = true
  skip_final_snapshot          = true
}
`, lName, preferredMaintenanceWindow)
}

func testAccDatabaseConfigPubliclyAccessible(lName string, publiclyAccessible string) string {
	return fmt.Sprintf(`	
data "awslightsail_availability_zones" "all" {}

resource "awslightsail_database" "test" {
  name                 = %[1]q
  availability_zone    = data.awslightsail_availability_zones.all.database_names[0]
  master_database_name = "testdatabasename"
  master_password      = "testdatabasepassword"
  master_username      = "test"
  blueprint_id         = "mysql_8_0"
  bundle_id            = "micro_1_0"
  publicly_accessible  = %[2]q
  apply_immediately    = true
  skip_final_snapshot  = true
}
`, lName, publiclyAccessible)
}

func testAccDatabaseConfigBackupRetentionEnabled(lName string, backupRetentionEnabled string) string {
	return fmt.Sprintf(`
data "awslightsail_availability_zones" "all" {}

resource "awslightsail_database" "test" {
  name                     = %[1]q
  availability_zone        = data.awslightsail_availability_zones.all.database_names[0]
  master_database_name     = "test"
  master_password          = "testdatabasepassword"
  master_username          = "test"
  blueprint_id             = "mysql_8_0"
  bundle_id                = "micro_1_0"
  backup_retention_enabled = %[2]q
  apply_immediately        = true
  skip_final_snapshot      = true
}
`, lName, backupRetentionEnabled)
}

func testAccDatabaseConfigFinalSnapshotName(lName string, sName string) string {
	return fmt.Sprintf(`	
data "awslightsail_availability_zones" "all" {}

resource "awslightsail_database" "test" {
  name                 = %[1]q
  availability_zone    = data.awslightsail_availability_zones.all.database_names[0]
  master_database_name = "test"
  master_password      = "testdatabasepassword"
  master_username      = "test"
  blueprint_id         = "mysql_8_0"
  bundle_id            = "micro_1_0"
  final_snapshot_name  = %[2]q
}
`, lName, sName)
}

func testAccDatabaseConfigTags1(lName string, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`	
data "awslightsail_availability_zones" "all" {}

resource "awslightsail_database" "test" {
  name                 = %[1]q
  availability_zone    = data.awslightsail_availability_zones.all.database_names[0]
  master_database_name = "testdatabasename"
  master_password      = "testdatabasepassword"
  master_username      = "test"
  blueprint_id         = "mysql_8_0"
  bundle_id            = "micro_1_0"
  skip_final_snapshot  = true
  tags = {
    %[2]q = %[3]q
  }
}
`, lName, tagKey1, tagValue1)
}

func testAccDatabaseConfigTags2(lName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`	
data "awslightsail_availability_zones" "all" {}

resource "awslightsail_database" "test" {
  name                 = %[1]q
  availability_zone    = data.awslightsail_availability_zones.all.database_names[0]
  master_database_name = "testdatabasename"
  master_password      = "testdatabasepassword"
  master_username      = "test"
  blueprint_id         = "mysql_8_0"
  bundle_id            = "micro_1_0"
  skip_final_snapshot  = true
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, lName, tagKey1, tagValue1, tagKey2, tagValue2)
}
