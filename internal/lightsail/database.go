package lightsail

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/smithy-go"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/tftags"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceDatabase() *schema.Resource {
	return &schema.Resource{
		Create: ResourceDatabaseCreate,
		Read:   ResourceDatabaseRead,
		Update: ResourceDatabaseUpdate,
		Delete: ResourceDatabaseDelete,
		Importer: &schema.ResourceImporter{
			State: ResourceDatabaseImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"master_database_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"master_password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"master_username": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"blueprint_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"bundle_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			// Optional attributes
			"preferred_backup_window": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"preferred_maintenance_window": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"publicly_accessible": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"apply_immediately": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"backup_retention_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"skip_final_snapshot": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"final_snapshot_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			// additional info returned from the API
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ca_certificate_identifier": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cpu_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"ram_size": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"disk_size": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"master_endpoint_port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"master_endpoint_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"secondary_availability_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"support_code": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": TagsSchema(),
		},
	}
}

func ResourceDatabaseCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*lightsail.Client)
	tags := tftags.New(d.Get("tags").(map[string]interface{}))

	req := lightsail.CreateRelationalDatabaseInput{
		MasterDatabaseName:            aws.String(d.Get("master_database_name").(string)),
		MasterUsername:                aws.String(d.Get("master_username").(string)),
		RelationalDatabaseBlueprintId: aws.String(d.Get("blueprint_id").(string)),
		RelationalDatabaseBundleId:    aws.String(d.Get("bundle_id").(string)),
		RelationalDatabaseName:        aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("availability_zone"); ok {
		req.AvailabilityZone = aws.String(v.(string))
	}

	if v, ok := d.GetOk("master_password"); ok {
		req.MasterUserPassword = aws.String(v.(string))
	}

	if v, ok := d.GetOk("preferred_backup_window"); ok {
		req.PreferredBackupWindow = aws.String(v.(string))
	}

	if v, ok := d.GetOk("preferred_maintenance_window"); ok {
		req.PreferredMaintenanceWindow = aws.String(v.(string))
	}

	if v, ok := d.GetOk("publicly_accessible"); ok {
		req.PubliclyAccessible = aws.Bool(v.(bool))
	}

	if len(tags) > 0 {
		req.Tags = Tags(tags.IgnoreAWS())
	}

	resp, err := conn.CreateRelationalDatabase(context.TODO(), &req)
	if err != nil {
		return err
	}

	if len(resp.Operations) == 0 {
		return fmt.Errorf("No operations found for Create Relational Database request")
	}

	op := resp.Operations[0]
	d.SetId(d.Get("name").(string))

	err = waitLightsailOperation(conn, op.Id)
	if err != nil {
		return fmt.Errorf("Error waiting for Relational Database (%s) to become ready: %s", d.Id(), err)
	}

	// Backup Retention is not a value you can pass on creation and defaults to true.
	// Forcing an update of the value after creation if the backup_retention_enabled value is false.
	if v := d.Get("backup_retention_enabled"); v == false {
		log.Printf("[DEBUG] Lightsail Database (%s) backup_retention_enabled setting is false. Updating value.", d.Id())
		req := lightsail.UpdateRelationalDatabaseInput{
			ApplyImmediately:       aws.Bool(true),
			RelationalDatabaseName: aws.String(d.Id()),
			DisableBackupRetention: aws.Bool(true),
		}

		resp, err := conn.UpdateRelationalDatabase(context.TODO(), &req)
		if err != nil {
			return err
		}

		if len(resp.Operations) == 0 {
			return fmt.Errorf("No operations found for Update Relational Database request")
		}

		op := resp.Operations[0]

		err = waitLightsailOperation(conn, op.Id)
		if err != nil {
			return fmt.Errorf("Error waiting for Relational Database (%s) to become ready: %s", d.Id(), err)
		}

		err = waitDatabaseBackupRetentionModified(conn, aws.String(d.Id()), aws.Bool(v.(bool)))
		if err != nil {
			return fmt.Errorf("Error waiting for Relational Database (%s) Backup Retention to be updated: %s", d.Id(), err)
		}

	}

	// Some Operations can complete before the Database enters the Available state. Added a waiter to make sure the Database is available before continuing.
	err = waitDatabaseModified(conn, aws.String(d.Id()))
	if err != nil {
		return fmt.Errorf("Error waiting for Relational Database (%s) to become available: %s", d.Id(), err)
	}

	return ResourceDatabaseRead(d, meta)
}

func ResourceDatabaseRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*lightsail.Client)

	// Some Operations can complete before the Database enters the Available state. Added a waiter to make sure the Database is available before continuing.
	// This is to support importing a resource that is not in a ready state.
	err := waitDatabaseModified(conn, aws.String(d.Id()))
	if err != nil {
		var oe *smithy.OperationError
		if errors.As(err, &oe) {
			log.Printf("failed to call service: %s, operation: %s, error: %v", oe.Service(), oe.Operation(), oe.Unwrap())
		}
		d.SetId("")
		return nil
	}

	resp, err := conn.GetRelationalDatabase(context.TODO(), &lightsail.GetRelationalDatabaseInput{
		RelationalDatabaseName: aws.String(d.Id()),
	})

	if err != nil {
		var oe *smithy.OperationError
		if errors.As(err, &oe) {
			log.Printf("failed to call service: %s, operation: %s, error: %v", oe.Service(), oe.Operation(), oe.Unwrap())
		}
		d.SetId("")
		return nil
	}

	if resp == nil {
		log.Printf("[WARN] Lightsail Relational Database (%s) not found, nil response from server, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	rd := resp.RelationalDatabase

	//manditory attributes
	d.Set("name", rd.Name)
	d.Set("availability_zone", rd.Location.AvailabilityZone)
	d.Set("master_database_name", rd.MasterDatabaseName)
	d.Set("master_username", rd.MasterUsername)
	d.Set("blueprint_id", rd.RelationalDatabaseBlueprintId)
	d.Set("bundle_id", rd.RelationalDatabaseBundleId)
	d.Set("backup_retention_enabled", rd.BackupRetentionEnabled)

	// optional attributes
	d.Set("preferred_backup_window", rd.PreferredBackupWindow)
	d.Set("preferred_maintenance_window", rd.PreferredMaintenanceWindow)
	d.Set("publicly_accessible", rd.PubliclyAccessible)

	// additional attributes
	d.Set("arn", rd.Arn)
	d.Set("ca_certificate_identifier", rd.CaCertificateIdentifier)
	d.Set("engine", rd.Engine)
	d.Set("created_at", rd.CreatedAt.Format(time.RFC3339))
	d.Set("cpu_count", rd.Hardware.CpuCount)
	d.Set("ram_size", rd.Hardware.RamSizeInGb)
	d.Set("disk_size", rd.Hardware.DiskSizeInGb)
	d.Set("engine_version", rd.EngineVersion)
	d.Set("master_endpoint_port", rd.MasterEndpoint.Port)
	d.Set("master_endpoint_address", rd.MasterEndpoint.Address)
	d.Set("secondary_availability_zone", rd.SecondaryAvailabilityZone)
	d.Set("support_code", rd.SupportCode)

	tags := KeyValueTags(rd.Tags).IgnoreAWS()

	//lintignore:AWSR002
	if err := d.Set("tags", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func ResourceDatabaseDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*lightsail.Client)

	// Some Operations can complete before the Database enters the Available state. Added a waiter to make sure the Database is available before continuing.
	err := waitDatabaseModified(conn, aws.String(d.Id()))
	if err != nil {
		return fmt.Errorf("Error waiting for Relational Database (%s) to become available: %s", d.Id(), err)
	}

	skipFinalSnapshot := d.Get("skip_final_snapshot").(bool)

	req := lightsail.DeleteRelationalDatabaseInput{
		RelationalDatabaseName: aws.String(d.Id()),
		SkipFinalSnapshot:      aws.Bool(skipFinalSnapshot),
	}

	if !skipFinalSnapshot {
		if name, present := d.GetOk("final_snapshot_name"); present {
			req.FinalRelationalDatabaseSnapshotName = aws.String(name.(string))
		} else {
			return fmt.Errorf("Lightsail Database FinalRelationalDatabaseSnapshotName is required when a final snapshot is required")
		}
	}

	resp, err := conn.DeleteRelationalDatabase(context.TODO(), &req)

	if err != nil {
		return err
	}

	op := resp.Operations[0]
	d.SetId(d.Get("name").(string))

	err = waitLightsailOperation(conn, op.Id)
	if err != nil {
		return fmt.Errorf("Error waiting for Relational Database (%s) to Delete: %s", d.Id(), err)
	}

	return nil
}

func ResourceDatabaseImport(
	d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Neither skip_final_snapshot nor final_snapshot_identifier can be fetched
	// from any API call, so we need to default skip_final_snapshot to true so
	// that final_snapshot_identifier is not required
	d.Set("skip_final_snapshot", true)
	return []*schema.ResourceData{d}, nil
}

func ResourceDatabaseUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*lightsail.Client)
	requestUpdate := false

	req := lightsail.UpdateRelationalDatabaseInput{
		ApplyImmediately:       aws.Bool(d.Get("apply_immediately").(bool)),
		RelationalDatabaseName: aws.String(d.Id()),
	}

	if d.HasChange("ca_certificate_identifier") {
		req.CaCertificateIdentifier = aws.String(d.Get("ca_certificate_identifier").(string))
		requestUpdate = true
	}

	if d.HasChange("backup_retention_enabled") {
		if d.Get("backup_retention_enabled").(bool) {
			req.EnableBackupRetention = aws.Bool(d.Get("backup_retention_enabled").(bool))
		} else {
			req.DisableBackupRetention = aws.Bool(true)
		}
		requestUpdate = true
	}

	if d.HasChange("master_password") {
		req.MasterUserPassword = aws.String(d.Get("master_password").(string))
		requestUpdate = true
	}

	if d.HasChange("preferred_backup_window") {
		req.PreferredBackupWindow = aws.String(d.Get("preferred_backup_window").(string))
		requestUpdate = true
	}

	if d.HasChange("preferred_maintenance_window") {
		req.PreferredMaintenanceWindow = aws.String(d.Get("preferred_maintenance_window").(string))
		requestUpdate = true
	}

	if d.HasChange("publicly_accessible") {
		req.PubliclyAccessible = aws.Bool(d.Get("publicly_accessible").(bool))
		requestUpdate = true
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating Lightsail Database (%s) tags: %s", d.Id(), err)
		}
	}

	if requestUpdate {
		resp, err := conn.UpdateRelationalDatabase(context.TODO(), &req)
		if err != nil {
			return err
		}

		if len(resp.Operations) == 0 {
			return fmt.Errorf("No operations found for Update Relational Database request")
		}

		op := resp.Operations[0]
		d.SetId(d.Get("name").(string))

		err = waitLightsailOperation(conn, op.Id)
		if err != nil {
			return fmt.Errorf("Error waiting for Relational Database (%s) to become ready: %s", d.Id(), err)
		}

		if d.HasChange("backup_retention_enabled") {
			err = waitDatabaseBackupRetentionModified(conn, aws.String(d.Id()), aws.Bool(d.Get("backup_retention_enabled").(bool)))
			if err != nil {
				return fmt.Errorf("Error waiting for Relational Database (%s) Backup Retention to be updated: %s", d.Id(), err)
			}
		}

		// Some Operations can complete before the Database enters the Available state. Added a waiter to make sure the Database is available before continuing.
		err = waitDatabaseModified(conn, aws.String(d.Id()))
		if err != nil {
			return fmt.Errorf("Error waiting for Relational Database (%s) to become available: %s", d.Id(), err)
		}
	}

	return ResourceDatabaseRead(d, meta)
}
