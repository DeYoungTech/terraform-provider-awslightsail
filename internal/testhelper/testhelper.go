package testhelper

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	tflightsail "github.com/deyoungtech/terraform-provider-awslightsail/internal/lightsail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
)

// initialized once, so it can be shared by each acceptance test
var provider = tflightsail.Provider()

// GetProvider returns the lightsail provider
func GetProvider() *schema.Provider {
	return provider
}

func GetProviders() map[string]*schema.Provider {
	return map[string]*schema.Provider{
		"awslightsail": GetProvider(),
	}
}

// DefaultEmailAddress is the default email address to set as a
// resource or data source parameter for acceptance tests.
const DefaultEmailAddress = "no-reply@deyoung.dev"

// initialized once, so it can be shared by each acceptance test
// waitLightsailOperation waits for an Operation to return Succeeded or Compleated
func WaitLightsailOperation(conn *lightsail.Client, oid *string) error {
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

// func testAccPreCheck(t *testing.T) {
// 	conn := provider.Meta(*lightsail.Client)

// 	input := &lightsail.GetInstancesInput{}

// 	_, err := conn.GetInstances(input)

// 	if err != nil {
// 		t.Fatalf("unexpected PreCheck error: %s", err)
// 	}
// }
