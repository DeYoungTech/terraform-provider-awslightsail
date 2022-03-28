package lightsail_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/conns"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/testhelper"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// Lightsail limits to 5 pending containers at a time
func TestAccContainerDeployment_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"ContainerDeployment": {
			"basic":                     testAccContainerDeployment_basic,
			"disappears":                testAccContainerDeployment_disappears_Service,
			"container_environment":     testAccContainerDeployment_ContainerEnvironment,
			"container_port":            testAccContainerDeployment_ContainerPort,
			"container_public_endpoint": testAccContainerServiceDeployment_PublicEndpoint,
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

func testAccContainerDeployment_basic(t *testing.T) {
	lName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rName := "awslightsail_container_deployment.test"

	resource.Test(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckContainerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerDeploymentConfigBasic(lName, "test1", "amazon/amazon-lightsail:hello-world"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerDeploymentExists(rName),
					resource.TestCheckResourceAttr(rName, "state", "ACTIVE"),
					resource.TestCheckResourceAttr(rName, "version", "1"),
					resource.TestCheckResourceAttr(rName, "container.#", "1"),
					resource.TestCheckResourceAttr(rName, "container.0.container_name", "test1"),
					resource.TestCheckResourceAttr(rName, "container.0.image", "amazon/amazon-lightsail:hello-world"),
					resource.TestCheckResourceAttr(rName, "container.0.command.#", "0"),
					resource.TestCheckResourceAttr(rName, "container.0.environment.#", "0"),
					resource.TestCheckResourceAttr(rName, "container.0.port.#", "0"),
				),
			},
			{
				Config: testAccContainerDeploymentConfig2Containers(lName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerDeploymentExists(rName),
					resource.TestCheckResourceAttr(rName, "state", "ACTIVE"),
					resource.TestCheckResourceAttr(rName, "version", "2"),
					resource.TestCheckResourceAttr(rName, "container.#", "2"),
					resource.TestCheckResourceAttr(rName, "container.0.container_name", "test1"),
					resource.TestCheckResourceAttr(rName, "container.0.image", "amazon/amazon-lightsail:hello-world"),
					resource.TestCheckResourceAttr(rName, "container.0.command.#", "0"),
					resource.TestCheckResourceAttr(rName, "container.0.environment.#", "0"),
					resource.TestCheckResourceAttr(rName, "container.0.port.#", "0"),
					resource.TestCheckResourceAttr(rName, "container.1.container_name", "test2"),
					resource.TestCheckResourceAttr(rName, "container.1.image", "redis:latest"),
					resource.TestCheckResourceAttr(rName, "container.1.command.#", "0"),
					resource.TestCheckResourceAttr(rName, "container.1.environment.#", "0"),
					resource.TestCheckResourceAttr(rName, "container.1.port.#", "0"),
				),
			},
			{
				Config: testAccContainerDeploymentConfigBasic(lName, "test3", "nginx:latest"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerDeploymentExists(rName),
					resource.TestCheckResourceAttr(rName, "state", "ACTIVE"),
					resource.TestCheckResourceAttr(rName, "version", "3"),
					resource.TestCheckResourceAttr(rName, "container.#", "1"),
					resource.TestCheckResourceAttr(rName, "container.0.container_name", "test3"),
					resource.TestCheckResourceAttr(rName, "container.0.image", "nginx:latest"),
					resource.TestCheckResourceAttr(rName, "container.0.command.#", "0"),
					resource.TestCheckResourceAttr(rName, "container.0.environment.#", "0"),
					resource.TestCheckResourceAttr(rName, "container.0.port.#", "0"),
				),
			},
		},
	})
}

func testAccContainerDeployment_ContainerEnvironment(t *testing.T) {
	lName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rName := "awslightsail_container_deployment.test"

	resource.Test(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckContainerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerDeploymentConfigEnvironment(lName, "A", "a"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerDeploymentExists(rName),
					resource.TestCheckResourceAttr(rName, "state", "ACTIVE"),
					resource.TestCheckResourceAttr(rName, "version", "1"),
					resource.TestCheckResourceAttr(rName, "container.#", "1"),
					resource.TestCheckResourceAttr(rName, "container.0.container_name", "test"),
					resource.TestCheckResourceAttr(rName, "container.0.image", "amazon/amazon-lightsail:hello-world"),
					resource.TestCheckResourceAttr(rName, "container.0.command.#", "0"),
					resource.TestCheckResourceAttr(rName, "container.0.environment.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(rName, "container.0.environment.*",
						map[string]string{
							"key":   "A",
							"value": "a",
						}),
					resource.TestCheckResourceAttr(rName, "container.0.port.#", "0"),
				),
			},
			{
				Config: testAccContainerDeploymentConfigEnvironment(lName, "B", "b"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerDeploymentExists(rName),
					resource.TestCheckResourceAttr(rName, "state", "ACTIVE"),
					resource.TestCheckResourceAttr(rName, "version", "2"),
					resource.TestCheckResourceAttr(rName, "container.#", "1"),
					resource.TestCheckResourceAttr(rName, "container.0.container_name", "test"),
					resource.TestCheckResourceAttr(rName, "container.0.image", "amazon/amazon-lightsail:hello-world"),
					resource.TestCheckResourceAttr(rName, "container.0.command.#", "0"),
					resource.TestCheckResourceAttr(rName, "container.0.environment.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(rName, "container.0.environment.*",
						map[string]string{
							"key":   "B",
							"value": "b",
						}),
					resource.TestCheckResourceAttr(rName, "container.0.port.#", "0"),
				),
			},
			{
				Config: testAccContainerDeploymentConfig2Environments(lName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerDeploymentExists(rName),
					resource.TestCheckResourceAttr(rName, "state", "ACTIVE"),
					resource.TestCheckResourceAttr(rName, "version", "3"),
					resource.TestCheckResourceAttr(rName, "container.#", "1"),
					resource.TestCheckResourceAttr(rName, "container.0.container_name", "test"),
					resource.TestCheckResourceAttr(rName, "container.0.image", "amazon/amazon-lightsail:hello-world"),
					resource.TestCheckResourceAttr(rName, "container.0.command.#", "0"),
					resource.TestCheckResourceAttr(rName, "container.0.environment.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(rName, "container.0.environment.*",
						map[string]string{
							"key":   "C",
							"value": "c",
						}),
					resource.TestCheckTypeSetElemNestedAttrs(rName, "container.0.environment.*",
						map[string]string{
							"key":   "D",
							"value": "d",
						}),
					resource.TestCheckResourceAttr(rName, "container.0.port.#", "0"),
				),
			},
			{
				Config: testAccContainerDeploymentConfigBasic(lName, "test", "amazon/amazon-lightsail:hello-world"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerDeploymentExists(rName),
					resource.TestCheckResourceAttr(rName, "state", "ACTIVE"),
					resource.TestCheckResourceAttr(rName, "version", "4"),
					resource.TestCheckResourceAttr(rName, "container.#", "1"),
					resource.TestCheckResourceAttr(rName, "container.0.container_name", "test"),
					resource.TestCheckResourceAttr(rName, "container.0.image", "amazon/amazon-lightsail:hello-world"),
					resource.TestCheckResourceAttr(rName, "container.0.command.#", "0"),
					resource.TestCheckResourceAttr(rName, "container.0.environment.#", "0"),
					resource.TestCheckResourceAttr(rName, "container.0.port.#", "0"),
				),
			},
		},
	})
}

func testAccContainerDeployment_ContainerPort(t *testing.T) {
	lName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rName := "awslightsail_container_deployment.test"

	resource.Test(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckContainerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerDeploymentConfigContainerPort(lName, 80, "HTTP"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerDeploymentExists(rName),
					resource.TestCheckResourceAttr(rName, "state", "ACTIVE"),
					resource.TestCheckResourceAttr(rName, "version", "1"),
					resource.TestCheckResourceAttr(rName, "container.#", "1"),
					resource.TestCheckResourceAttr(rName, "container.0.container_name", "test"),
					resource.TestCheckResourceAttr(rName, "container.0.image", "amazon/amazon-lightsail:hello-world"),
					resource.TestCheckResourceAttr(rName, "container.0.command.#", "0"),
					resource.TestCheckResourceAttr(rName, "container.0.environment.#", "0"),
					resource.TestCheckResourceAttr(rName, "container.0.port.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(rName, "container.0.port.*",
						map[string]string{
							"port_number": "80",
							"protocol":    "HTTP",
						}),
				),
			},
			{
				Config: testAccContainerDeploymentConfigContainerPort(lName, 90, "TCP"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerDeploymentExists(rName),
					resource.TestCheckResourceAttr(rName, "state", "ACTIVE"),
					resource.TestCheckResourceAttr(rName, "version", "2"),
					resource.TestCheckResourceAttr(rName, "container.#", "1"),
					resource.TestCheckResourceAttr(rName, "container.0.container_name", "test"),
					resource.TestCheckResourceAttr(rName, "container.0.image", "amazon/amazon-lightsail:hello-world"),
					resource.TestCheckResourceAttr(rName, "container.0.command.#", "0"),
					resource.TestCheckResourceAttr(rName, "container.0.environment.#", "0"),
					resource.TestCheckResourceAttr(rName, "container.0.port.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(rName, "container.0.port.*",
						map[string]string{
							"port_number": "90",
							"protocol":    "TCP",
						}),
				),
			},
			{
				Config: testAccContainerDeploymentConfigContainer2Ports(lName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerDeploymentExists(rName),
					resource.TestCheckResourceAttr(rName, "state", "ACTIVE"),
					resource.TestCheckResourceAttr(rName, "version", "3"),
					resource.TestCheckResourceAttr(rName, "container.#", "1"),
					resource.TestCheckResourceAttr(rName, "container.0.container_name", "test"),
					resource.TestCheckResourceAttr(rName, "container.0.image", "amazon/amazon-lightsail:hello-world"),
					resource.TestCheckResourceAttr(rName, "container.0.command.#", "0"),
					resource.TestCheckResourceAttr(rName, "container.0.environment.#", "0"),
					resource.TestCheckResourceAttr(rName, "container.0.port.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(rName, "container.0.port.*",
						map[string]string{
							"port_number": "80",
							"protocol":    "HTTP",
						}),
					resource.TestCheckTypeSetElemNestedAttrs(rName, "container.0.port.*",
						map[string]string{
							"port_number": "90",
							"protocol":    "TCP",
						}),
				),
			},
			{
				Config: testAccContainerDeploymentConfigBasic(lName, "test", "amazon/amazon-lightsail:hello-world"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerDeploymentExists(rName),
					resource.TestCheckResourceAttr(rName, "state", "ACTIVE"),
					resource.TestCheckResourceAttr(rName, "version", "4"),
					resource.TestCheckResourceAttr(rName, "container.#", "1"),
					resource.TestCheckResourceAttr(rName, "container.0.container_name", "test"),
					resource.TestCheckResourceAttr(rName, "container.0.image", "amazon/amazon-lightsail:hello-world"),
					resource.TestCheckResourceAttr(rName, "container.0.command.#", "0"),
					resource.TestCheckResourceAttr(rName, "container.0.environment.#", "0"),
					resource.TestCheckResourceAttr(rName, "container.0.port.#", "0"),
				),
			},
		},
	})
}

func testAccContainerServiceDeployment_PublicEndpoint(t *testing.T) {
	lName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rName := "awslightsail_container_deployment.test"

	resource.Test(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckContainerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerDeploymentConfigPublicEndpoint1(lName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerDeploymentExists(rName),
					resource.TestCheckResourceAttr(rName, "state", "ACTIVE"),
					resource.TestCheckResourceAttr(rName, "version", "1"),
					resource.TestCheckResourceAttr(rName, "container.#", "1"),
					resource.TestCheckResourceAttr(rName, "container.0.container_name", "test1"),
					resource.TestCheckResourceAttr(rName, "container.0.image", "amazon/amazon-lightsail:hello-world"),
					resource.TestCheckResourceAttr(rName, "container.0.port.0.port_number", "80"),
					resource.TestCheckResourceAttr(rName, "container.0.port.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.#", "1"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.0.container_name", "test1"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.0.container_port", "80"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.0.health_check.#", "1"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.0.health_check.0.healthy_threshold", "2"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.0.health_check.0.interval_seconds", "5"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.0.health_check.0.path", "/"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.0.health_check.0.success_codes", "200-499"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.0.health_check.0.timeout_seconds", "2"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.0.health_check.0.unhealthy_threshold", "2"),
				),
			},
			{
				Config: testAccContainerDeploymentConfigPublicEndpoint2(lName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerDeploymentExists(rName),
					resource.TestCheckResourceAttr(rName, "state", "ACTIVE"),
					resource.TestCheckResourceAttr(rName, "version", "2"),
					resource.TestCheckResourceAttr(rName, "container.#", "1"),
					resource.TestCheckResourceAttr(rName, "container.0.container_name", "test2"),
					resource.TestCheckResourceAttr(rName, "container.0.image", "amazon/amazon-lightsail:hello-world"),
					resource.TestCheckResourceAttr(rName, "container.0.port.0.port_number", "80"),
					resource.TestCheckResourceAttr(rName, "container.0.port.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.#", "1"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.0.container_name", "test2"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.0.container_port", "80"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.0.health_check.#", "1"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.0.health_check.0.healthy_threshold", "3"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.0.health_check.0.interval_seconds", "6"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.0.health_check.0.path", "/."),
					resource.TestCheckResourceAttr(rName, "public_endpoint.0.health_check.0.success_codes", "200"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.0.health_check.0.timeout_seconds", "3"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.0.health_check.0.unhealthy_threshold", "3"),
				),
			},
			{
				Config: testAccContainerDeploymentConfigPublicEndpoint3(lName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerDeploymentExists(rName),
					resource.TestCheckResourceAttr(rName, "state", "ACTIVE"),
					resource.TestCheckResourceAttr(rName, "version", "3"),
					resource.TestCheckResourceAttr(rName, "container.#", "1"),
					resource.TestCheckResourceAttr(rName, "container.0.container_name", "test2"),
					resource.TestCheckResourceAttr(rName, "container.0.image", "amazon/amazon-lightsail:hello-world"),
					resource.TestCheckResourceAttr(rName, "container.0.port.0.port_number", "80"),
					resource.TestCheckResourceAttr(rName, "container.0.port.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.#", "1"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.0.container_name", "test2"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.0.container_port", "80"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.0.health_check.#", "1"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.0.health_check.0.healthy_threshold", "4"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.0.health_check.0.interval_seconds", "5"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.0.health_check.0.path", "/"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.0.health_check.0.success_codes", "200-499"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.0.health_check.0.timeout_seconds", "2"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.0.health_check.0.unhealthy_threshold", "2"),
				),
			},
			{
				Config: testAccContainerDeploymentConfigPublicEndpoint4(lName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerDeploymentExists(rName),
					resource.TestCheckResourceAttr(rName, "state", "ACTIVE"),
					resource.TestCheckResourceAttr(rName, "version", "4"),
					resource.TestCheckResourceAttr(rName, "container.#", "1"),
					resource.TestCheckResourceAttr(rName, "container.0.container_name", "test2"),
					resource.TestCheckResourceAttr(rName, "container.0.image", "amazon/amazon-lightsail:hello-world"),
					resource.TestCheckResourceAttr(rName, "container.0.port.0.port_number", "80"),
					resource.TestCheckResourceAttr(rName, "container.0.port.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.#", "1"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.0.container_name", "test2"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.0.container_port", "80"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.0.health_check.#", "1"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.0.health_check.0.healthy_threshold", "2"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.0.health_check.0.interval_seconds", "5"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.0.health_check.0.path", "/"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.0.health_check.0.success_codes", "200-499"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.0.health_check.0.timeout_seconds", "2"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.0.health_check.0.unhealthy_threshold", "2"),
				),
			},
			{
				Config: testAccContainerDeploymentConfigPublicEndpoint5(lName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerDeploymentExists(rName),
					resource.TestCheckResourceAttr(rName, "state", "ACTIVE"),
					resource.TestCheckResourceAttr(rName, "version", "5"),
					resource.TestCheckResourceAttr(rName, "container.#", "1"),
					resource.TestCheckResourceAttr(rName, "container.0.container_name", "test2"),
					resource.TestCheckResourceAttr(rName, "container.0.image", "amazon/amazon-lightsail:hello-world"),
					resource.TestCheckResourceAttr(rName, "container.0.port.0.port_number", "80"),
					resource.TestCheckResourceAttr(rName, "container.0.port.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(rName, "public_endpoint.#", "0"),
				),
			},
		},
	})
}

// You cannot Delete a Deployment. Because of this we just need to make sure that terraform does not error if the Parent is deleted.
func testAccContainerDeployment_disappears_Service(t *testing.T) {
	lName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rName := "awslightsail_container_service.test"

	testDestroy := func(*terraform.State) error {
		// reach out and DELETE the Container Service
		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn

		_, err := conn.DeleteContainerService(context.TODO(), &lightsail.DeleteContainerServiceInput{
			ServiceName: aws.String(lName),
		})

		if err != nil {
			return fmt.Errorf("error deleting Lightsail Container Service in disappear test")
		}

		// sleep 7 seconds to give it time, so we don't have to poll
		time.Sleep(7 * time.Second)

		return nil
	}

	resource.Test(t, resource.TestCase{
		Providers:    testhelper.GetProviders(),
		CheckDestroy: testAccCheckContainerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerDeploymentConfigBasic(lName, "test1", "amazon/amazon-lightsail:hello-world"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerDeploymentExists(rName),
					testDestroy),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckContainerDeploymentExists(rName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rName]
		if !ok {
			return fmt.Errorf("not finding Lightsail Container Service (%s)", rName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Lightsail Container Service ID is set")
		}

		conn := testhelper.GetProvider().Meta().(*conns.AWSClient).LightsailConn
		_, err := conn.GetContainerServiceDeployments(context.TODO(), &lightsail.GetContainerServiceDeploymentsInput{
			ServiceName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccContainerDeploymentConfigBasic(lName, cName, iName string) string {
	return fmt.Sprintf(`
resource "awslightsail_container_service" "test" {
  name  = %[1]q
  power = "nano"
  scale = 1
}

resource "awslightsail_container_deployment" "test" {
  container_service_name = awslightsail_container_service.test.id
  container {
    container_name = %[2]q
    image          = %[3]q
  }
}
`, lName, cName, iName)
}

func testAccContainerDeploymentConfig2Containers(lName string) string {
	return fmt.Sprintf(`
resource "awslightsail_container_service" "test" {
  name  = %[1]q
  power = "nano"
  scale = 1
}

resource "awslightsail_container_deployment" "test" {
  container_service_name = awslightsail_container_service.test.id
  container {
    container_name = "test1"
    image          = "amazon/amazon-lightsail:hello-world"
  }
  container {
    container_name = "test2"
    image          = "redis:latest"
  }
}
`, lName)
}

func testAccContainerDeploymentConfigEnvironment(rName, key, value string) string {
	return fmt.Sprintf(`
resource "awslightsail_container_service" "test" {
  name  = %[1]q
  power = "nano"
  scale = 1
}

resource "awslightsail_container_deployment" "test" {
  container_service_name = awslightsail_container_service.test.id
  container {
    container_name = "test"
    image          = "amazon/amazon-lightsail:hello-world"
    environment {
      key   = %[2]q
      value = %[3]q
    }
  }
}
`, rName, key, value)
}

func testAccContainerDeploymentConfig2Environments(rName string) string {
	return fmt.Sprintf(`
resource "awslightsail_container_service" "test" {
  name  = %[1]q
  power = "nano"
  scale = 1
}

resource "awslightsail_container_deployment" "test" {
  container_service_name = awslightsail_container_service.test.id
  container {
    container_name = "test"
    image          = "amazon/amazon-lightsail:hello-world"
    environment {
      key   = "C"
      value = "c"
    }
    environment {
      key   = "D"
      value = "d"
    }
  }
}
`, rName)
}

func testAccContainerDeploymentConfigContainerPort(rName string, pNumber int, protocol string) string {
	return fmt.Sprintf(`
resource "awslightsail_container_service" "test" {
  name  = %[1]q
  power = "nano"
  scale = 1
}

resource "awslightsail_container_deployment" "test" {
  container_service_name = awslightsail_container_service.test.id
  container {
    container_name = "test"
    image          = "amazon/amazon-lightsail:hello-world"
    port {
      port_number = %[2]d
      protocol    = %[3]q
    }
  }
}
`, rName, pNumber, protocol)
}

func testAccContainerDeploymentConfigContainer2Ports(rName string) string {
	return fmt.Sprintf(`
resource "awslightsail_container_service" "test" {
  name  = %[1]q
  power = "nano"
  scale = 1
}

resource "awslightsail_container_deployment" "test" {
  container_service_name = awslightsail_container_service.test.id
  container {
    container_name = "test"
    image          = "amazon/amazon-lightsail:hello-world"
    port {
      port_number = 80
      protocol    = "HTTP"
    }
    port {
      port_number = 90
      protocol    = "TCP"
    }
  }
}
`, rName)
}

func testAccContainerDeploymentConfigPublicEndpoint1(rName string) string {
	return fmt.Sprintf(`
resource "awslightsail_container_service" "test" {
  name  = %[1]q
  power = "nano"
  scale = 1
}

resource "awslightsail_container_deployment" "test" {
  container_service_name = awslightsail_container_service.test.id
  container {
    container_name = "test1"
    image          = "amazon/amazon-lightsail:hello-world"
    port {
      port_number = 80
      protocol    = "HTTP"
    }
  }
  public_endpoint {
    container_name = "test1"
    container_port = 80
    health_check {}
  }
}
`, rName)
}

func testAccContainerDeploymentConfigPublicEndpoint2(rName string) string {
	return fmt.Sprintf(`
resource "awslightsail_container_service" "test" {
  name  = %[1]q
  power = "nano"
  scale = 1
}

resource "awslightsail_container_deployment" "test" {
  container_service_name = awslightsail_container_service.test.id
  container {
    container_name = "test2"
    image          = "amazon/amazon-lightsail:hello-world"
    port {
      port_number = 80
      protocol    = "HTTP"
    }
  }
  public_endpoint {
    container_name = "test2"
    container_port = 80
    health_check {
      healthy_threshold   = 3
      interval_seconds    = 6
      path                = "/."
      success_codes       = "200"
      timeout_seconds     = 3
      unhealthy_threshold = 3
    }
  }
}
`, rName)
}

func testAccContainerDeploymentConfigPublicEndpoint3(rName string) string {
	return fmt.Sprintf(`
resource "awslightsail_container_service" "test" {
  name  = %[1]q
  power = "nano"
  scale = 1
}

resource "awslightsail_container_deployment" "test" {
  container_service_name = awslightsail_container_service.test.id
  container {
    container_name = "test2"
    image          = "amazon/amazon-lightsail:hello-world"
    port {
      port_number = 80
      protocol    = "HTTP"
    }
  }
  public_endpoint {
    container_name = "test2"
    container_port = 80
    health_check {
      healthy_threshold = 4
    }
  }
}
`, rName)
}

func testAccContainerDeploymentConfigPublicEndpoint4(rName string) string {
	return fmt.Sprintf(`
resource "awslightsail_container_service" "test" {
  name  = %[1]q
  power = "nano"
  scale = 1
}

resource "awslightsail_container_deployment" "test" {
  container_service_name = awslightsail_container_service.test.id
  container {
    container_name = "test2"
    image          = "amazon/amazon-lightsail:hello-world"
    port {
      port_number = 80
      protocol    = "HTTP"
    }
  }
  public_endpoint {
    container_name = "test2"
    container_port = 80
    health_check {}
  }
}
`, rName)
}

func testAccContainerDeploymentConfigPublicEndpoint5(rName string) string {
	return fmt.Sprintf(`
resource "awslightsail_container_service" "test" {
  name  = %[1]q
  power = "nano"
  scale = 1
}

resource "awslightsail_container_deployment" "test" {
  container_service_name = awslightsail_container_service.test.id
  container {
    container_name = "test2"
    image          = "amazon/amazon-lightsail:hello-world"
    port {
      port_number = 80
      protocol    = "HTTP"
    }
  }
}
`, rName)
}
