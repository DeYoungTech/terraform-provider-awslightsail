package lightsail

import (
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// OperationStatusNotStarted is a OperationStatus enum value
	OperationStatusNotStarted = "NotStarted"
	// OperationStatusStarted is a OperationStatus enum value
	OperationStatusStarted = "Started"
	// OperationStatusFailed is a OperationStatus enum value
	OperationStatusFailed = "Failed"
	// OperationStatusCompleted is a OperationStatus enum value
	OperationStatusCompleted = "Completed"
	// OperationStatusSucceeded is a OperationStatus enum value
	OperationStatusSucceeded = "Succeeded"

	// OperationTimeout is the Timout Value for Operations
	OperationTimeout = 20 * time.Minute
	// OperationDelay is the Delay Value for Operations
	OperationDelay = 5 * time.Second
	// OperationMinTimeout is the MinTimout Value for Operations
	OperationMinTimeout = 3 * time.Second

	// DatabaseStateModifying is a state value for a Relational Database undergoing a modification
	DatabaseStateModifying = "modifying"
	// DatabaseStateAvailable is a state value for a Relational Database available for modification
	DatabaseStateAvailable = "available"

	// DatabaseTimeout is the Timout Value for Relational Database Modifications
	DatabaseTimeout = 20 * time.Minute
	// DatabaseDelay is the Delay Value for Relational Database Modifications
	DatabaseDelay = 5 * time.Second
	// DatabaseMinTimeout is the MinTimout Value for Relational Database Modifications
	DatabaseMinTimeout = 3 * time.Second

	// The current state of the container service. The following container service
	// * PENDING - The container service is being created.
	ContainerServiceStatePending = "PENDING"
	// READY - The container service is running but it does not have an active container deployment
	ContainerServiceStateReady = "READY"
	// * DEPLOYING - The container service is launching a container deployment
	ContainerServiceStateDeploying = "DEPLOYING"
	//
	// * RUNNING - The container service is running and it has
	// an active container deployment.
	ContainerServiceStateRunning = "RUNNING"
	// * UPDATING - The container service is being updated
	ContainerServiceStateUpdating = "UPDATING"
	// * DELETING - The container service is being deleted.
	ContainerServiceStateDeleting = "DELEATING"
	// * DISABLED - The container service is disabled, and its active
	// deployment and containers, if any, are shut down.
	ContainerServiceStateDisabled = "DISABLED"
)

// waitLightsailOperation waits for an Operation to return Succeeded or Compleated
func waitLightsailOperation(conn *lightsail.Client, oid *string) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{OperationStatusStarted},
		Target:     []string{OperationStatusCompleted, OperationStatusSucceeded},
		Refresh:    statusLightsailOperation(conn, oid),
		Timeout:    OperationTimeout,
		Delay:      OperationDelay,
		MinTimeout: OperationMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if _, ok := outputRaw.(*lightsail.GetOperationOutput); ok {
		return err
	}

	return err
}

// waitDatabaseModified waits for a Modified Database return available
func waitDatabaseModified(conn *lightsail.Client, db *string) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{DatabaseStateModifying},
		Target:     []string{DatabaseStateAvailable},
		Refresh:    statusLightsailDatabase(conn, db),
		Timeout:    DatabaseTimeout,
		Delay:      DatabaseDelay,
		MinTimeout: DatabaseMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if _, ok := outputRaw.(*lightsail.GetRelationalDatabaseOutput); ok {
		return err
	}

	return err
}

// waitDatabaseBackupRetentionModified waits for a Modified  BackupRetention on Database return available

func waitDatabaseBackupRetentionModified(conn *lightsail.Client, db *string, status *bool) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{strconv.FormatBool(!aws.BoolValue(status))},
		Target:     []string{strconv.FormatBool(aws.BoolValue(status))},
		Refresh:    statusLightsailDatabaseBackupRetention(conn, db),
		Timeout:    DatabaseTimeout,
		Delay:      DatabaseDelay,
		MinTimeout: DatabaseMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if _, ok := outputRaw.(*lightsail.GetRelationalDatabaseOutput); ok {
		return err
	}

	return err
}

func waitContainerService(conn *lightsail.Client, cs *string) error {
	// first wait for the completion of creating an empty container service
	stateConf := &resource.StateChangeConf{
		Pending:    []string{ContainerServiceStatePending, ContainerServiceStateUpdating, ContainerServiceStateDeploying, ContainerServiceStateDeleting},
		Target:     []string{ContainerServiceStateReady, ContainerServiceStateRunning, ContainerServiceStateDisabled},
		Refresh:    statusLightsailContainerService(conn, cs),
		Timeout:    25 * time.Minute,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if _, ok := outputRaw.(*lightsail.GetContainerServicesOutput); ok {
		return err
	}

	return err
}
