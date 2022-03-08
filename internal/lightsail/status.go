package lightsail

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// statusLightsailOperation is a method to check the status of a Lightsail Operation
func statusLightsailOperation(conn *lightsail.Client, oid *string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &lightsail.GetOperationInput{
			OperationId: oid,
		}

		oidValue := aws.ToString(oid)
		log.Printf("[DEBUG] Checking if Lightsail Operation (%s) is Completed", oidValue)

		output, err := conn.GetOperation(context.TODO(), input)

		if err != nil {
			return output, "FAILED", err
		}

		if output.Operation == nil {
			return nil, "Failed", fmt.Errorf("Error retrieving Operation info for operation (%s)", oidValue)
		}

		log.Printf("[DEBUG] Lightsail Operation (%s) is currently %q", oidValue, output.Operation.Status)
		return output, string(output.Operation.Status), nil
	}
}

// statusLightsailDatabase is a method to check the status of a Lightsail Relational Database
func statusLightsailDatabase(conn *lightsail.Client, db *string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &lightsail.GetRelationalDatabaseInput{
			RelationalDatabaseName: db,
		}

		dbValue := aws.ToString(db)
		log.Printf("[DEBUG] Checking if Lightsail Database (%s) is in an available state.", dbValue)

		output, err := conn.GetRelationalDatabase(context.TODO(), input)

		if err != nil {
			return output, "FAILED", err
		}

		if output.RelationalDatabase == nil {
			return nil, "Failed", fmt.Errorf("Error retrieving Database info for (%s)", dbValue)
		}

		log.Printf("[DEBUG] Lightsail Database (%s) is currently %q", dbValue, *output.RelationalDatabase.State)
		return output, *output.RelationalDatabase.State, nil
	}
}

// statusLightsailDatabase is a method to check the status of a Lightsail Relational Database Backup Retention
func statusLightsailDatabaseBackupRetention(conn *lightsail.Client, db *string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &lightsail.GetRelationalDatabaseInput{
			RelationalDatabaseName: db,
		}

		dbValue := aws.ToString(db)
		log.Printf("[DEBUG] Checking if Lightsail Database (%s) Backup Retention setting has been updated.", dbValue)

		output, err := conn.GetRelationalDatabase(context.TODO(), input)

		if err != nil {
			return output, "FAILED", err
		}

		if output.RelationalDatabase == nil {
			return nil, "Failed", fmt.Errorf("Error retrieving Database info for (%s)", dbValue)
		}

		log.Printf("[DEBUG] Lightsail Database (%s) Backup Retention setting is currently %t", dbValue, *output.RelationalDatabase.BackupRetentionEnabled)
		return output, strconv.FormatBool(*output.RelationalDatabase.BackupRetentionEnabled), nil
	}
}

// call GetContainerServices to check the current state of the container service
func statusLightsailContainerService(conn *lightsail.Client, cs *string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &lightsail.GetContainerServicesInput{
			ServiceName: cs,
		}
		log.Printf("[DEBUG] Checking Lightsail Container Service state changes")
		resp, err := conn.GetContainerServices(context.TODO(), input)
		if err != nil {
			return nil, "", err
		}
		return resp.ContainerServices[0], string(resp.ContainerServices[0].State), nil
	}
}
