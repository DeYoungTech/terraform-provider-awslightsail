package lightsail

import (
	"context"
	"errors"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/aws-sdk-go-v2/service/lightsail/types"
	"github.com/aws/smithy-go"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/conns"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceContainerDeployment() *schema.Resource {
	return &schema.Resource{
		Create: resourceContainerDeploymentCreate,
		Read:   resourceContainerDeploymentRead,
		Update: resourceContainerDeploymentUpdate,
		Delete: resourceContainerDeploymentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		// CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			// required fields
			"container_service_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"container": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotWhiteSpace,
						},
						"image": {
							Type:     schema.TypeString,
							Required: true,
						},
						"command": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"environment": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:     schema.TypeString,
										Required: true,
									},
									"value": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"port": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"port_number": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"protocol": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"public_endpoint": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"container_port": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"health_check": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"healthy_threshold": {
										Type:     schema.TypeInt,
										Optional: true,
										Default:  2,
									},
									"interval_seconds": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      5,
										ValidateFunc: validation.IntBetween(5, 300),
									},
									"path": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "/",
									},
									"success_codes": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "200-499",
									},
									"timeout_seconds": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      2,
										ValidateFunc: validation.IntBetween(2, 60),
									},
									"unhealthy_threshold": {
										Type:     schema.TypeInt,
										Optional: true,
										Default:  2,
									},
								},
							},
						},
					},
				},
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}
func resourceContainerDeploymentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn
	serviceName := aws.String(d.Get("container_service_name").(string))
	containers := expandLightsailContainerServiceDeploymentContainers(d.Get("container").(*schema.Set).List())
	publicEndpoint := expandLightsailContainerServiceDeploymentPublicEndpoint(d.Get("public_endpoint").([]interface{}))

	ContainerServiceDeploymentInput := lightsail.CreateContainerServiceDeploymentInput{
		ServiceName: serviceName,
	}

	if len(containers) > 0 {
		ContainerServiceDeploymentInput.Containers = containers
	}

	if len(d.Get("public_endpoint").([]interface{})) > 0 {
		ContainerServiceDeploymentInput.PublicEndpoint = publicEndpoint
	}

	if _, err := conn.CreateContainerServiceDeployment(context.TODO(), &ContainerServiceDeploymentInput); err != nil {
		log.Printf("[ERROR] Lightsail Container Service Deployment for Container Service (%s) failed: %s", aws.ToString(serviceName), err)
		return err
	}

	d.SetId(d.Get("container_service_name").(string))
	log.Printf("[INFO] Lightsail Container Service (%s) CreateContainerDeployment call successful, now waiting for ContainerDeploymentState change", d.Id())

	err := waitContainerService(conn, serviceName)
	if err != nil {
		log.Printf("[ERROR] Container Service (%s) failed to become ready: %s", d.Id(), err)
		return err
	}

	log.Printf("[INFO] Lightsail Container Service Deployment (%s) successful", d.Id())

	return resourceContainerDeploymentRead(d, meta)
}

func resourceContainerDeploymentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn

	resp, err := conn.GetContainerServiceDeployments(context.TODO(),
		&lightsail.GetContainerServiceDeploymentsInput{
			ServiceName: aws.String(d.Id()),
		},
	)

	if err != nil {
		var oe *smithy.OperationError
		if errors.As(err, &oe) {
			log.Printf("failed to call service: %s, operation: %s, error: %v", oe.Service(), oe.Operation(), oe.Unwrap())
		}
		d.SetId("")
		return nil
	}

	csd := resp.Deployments[0]

	d.Set("container", flattenLightsailContainerServiceDeploymentContainers(csd.Containers))
	d.Set("public_endpoint", flattenLightsailContainerServiceDeploymentPublicEndpoint(csd.PublicEndpoint))
	d.Set("state", csd.State)
	d.Set("version", csd.Version)

	return nil
}

func resourceContainerDeploymentUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn
	serviceName := aws.String(d.Id())
	containers := d.Get("container").(*schema.Set).List()
	publicendpoint := d.Get("public_endpoint").([]interface{})
	requestUpdate := false

	req := lightsail.CreateContainerServiceDeploymentInput{
		ServiceName: serviceName,
	}

	if d.HasChange("container") {
		requestUpdate = true
	}

	if d.HasChange("public_endpoint") {
		requestUpdate = true
	}

	if requestUpdate {
		req.Containers = expandLightsailContainerServiceDeploymentContainers(containers)
		req.PublicEndpoint = expandLightsailContainerServiceDeploymentPublicEndpoint(publicendpoint)

		_, err := conn.CreateContainerServiceDeployment(context.TODO(), &req)

		if err != nil {
			return err
		}

		err = waitContainerService(conn, serviceName)
		if err != nil {
			log.Printf("[ERROR] Container Service (%s) failed to become ready: %s", d.Id(), err)
			return err
		}
	}

	log.Printf("[INFO] Lightsail Container Service Deployment for Service: (%s) successful", d.Id())
	return resourceContainerDeploymentRead(d, meta)
}

func resourceContainerDeploymentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn
	serviceName := aws.String(d.Id())

	req := lightsail.UpdateContainerServiceInput{
		ServiceName: serviceName,
		IsDisabled:  aws.Bool(true),
	}

	_, err := conn.UpdateContainerService(context.TODO(), &req)

	if err != nil {
		return err
	}

	err = waitContainerService(conn, serviceName)
	if err != nil {
		log.Printf("[ERROR] Container Service (%s) failed to Disable: %s", d.Id(), err)
		return err
	}

	return nil
}

func flattenLightsailContainerServiceDeploymentContainers(containers map[string]types.Container) []interface{} {
	if containers == nil {
		return []interface{}{}
	}

	var rawContainers []interface{}
	for containerName, container := range containers {
		rawContainer := map[string]interface{}{
			"container_name": containerName,
			"image":          aws.ToString(container.Image),
			"command":        aws.StringSlice(container.Command),
			"environment":    flattenLightsailContainerServiceDeploymentEnvironment(container.Environment),
			"port":           flattenLightsailContainerServiceDeploymentPort(container.Ports),
		}

		rawContainers = append(rawContainers, rawContainer)
	}

	return rawContainers
}

func flattenLightsailContainerServiceDeploymentEnvironment(environment map[string]string) []interface{} {
	if len(environment) == 0 {
		return []interface{}{}
	}

	var rawEnvironment []interface{}
	for key, value := range environment {
		rawEnvironment = append(rawEnvironment, map[string]interface{}{
			"key":   key,
			"value": value,
		})
	}
	return rawEnvironment
}

func flattenLightsailContainerServiceDeploymentPort(port map[string]types.ContainerServiceProtocol) []interface{} {
	if len(port) == 0 {
		return []interface{}{}
	}

	var rawPorts []interface{}
	for portNumber, protocol := range port {
		portNumber, err := strconv.Atoi(portNumber)
		if err != nil {
			return []interface{}{}
		}
		rawPorts = append(rawPorts, map[string]interface{}{
			"port_number": portNumber,
			"protocol":    string(protocol),
		})
	}
	return rawPorts
}

func flattenLightsailContainerServiceDeploymentPublicEndpoint(endpoint *types.ContainerServiceEndpoint) []interface{} {
	if endpoint == nil {
		return []interface{}{}
	}

	return []interface{}{
		map[string]interface{}{
			"container_name": aws.ToString(endpoint.ContainerName),
			"container_port": int(aws.ToInt32(endpoint.ContainerPort)),
			"health_check":   flattenLightsailContainerServiceDeploymentPublicEndpointHealthCheck(endpoint.HealthCheck),
		},
	}
}

func flattenLightsailContainerServiceDeploymentPublicEndpointHealthCheck(healthCheck *types.ContainerServiceHealthCheckConfig) []interface{} {
	if healthCheck == nil {
		return []interface{}{}
	}

	return []interface{}{
		map[string]interface{}{
			"healthy_threshold":   int(aws.ToInt32(healthCheck.HealthyThreshold)),
			"interval_seconds":    int(aws.ToInt32(healthCheck.IntervalSeconds)),
			"path":                aws.ToString(healthCheck.Path),
			"success_codes":       aws.ToString(healthCheck.SuccessCodes),
			"timeout_seconds":     int(aws.ToInt32(healthCheck.TimeoutSeconds)),
			"unhealthy_threshold": int(aws.ToInt32(healthCheck.UnhealthyThreshold)),
		},
	}
}

func expandLightsailContainerServiceDeploymentPublicEndpointHealthCheck(rawHealthCheck []interface{}) *types.ContainerServiceHealthCheckConfig {
	if len(rawHealthCheck) == 0 {
		return nil
	}

	healthCheck := types.ContainerServiceHealthCheckConfig{}

	for _, rhc := range rawHealthCheck {
		rhcMap := rhc.(map[string]interface{})

		healthCheck.HealthyThreshold = aws.Int32(int32(rhcMap["healthy_threshold"].(int)))
		healthCheck.IntervalSeconds = aws.Int32(int32(rhcMap["interval_seconds"].(int)))
		healthCheck.Path = aws.String(rhcMap["path"].(string))
		healthCheck.SuccessCodes = aws.String(rhcMap["success_codes"].(string))
		healthCheck.TimeoutSeconds = aws.Int32(int32(rhcMap["timeout_seconds"].(int)))
		healthCheck.UnhealthyThreshold = aws.Int32(int32(rhcMap["unhealthy_threshold"].(int)))
	}

	return &healthCheck
}

func expandLightsailContainerServiceDeploymentContainers(rawContainers []interface{}) map[string]types.Container {
	if len(rawContainers) == 0 {
		return map[string]types.Container{}
	}

	result := make(map[string]types.Container)

	for _, rawContainer := range rawContainers {
		rawContainerMap := rawContainer.(map[string]interface{})

		containerName := rawContainerMap["container_name"].(string)
		// ignore empty-named container, which means a container is removed from .tf file
		// important to ignore this empty container because we don't need to delete an unwanted container,
		// besides, lightsail.CreateContainerServiceDeployment will throw InvalidInputException with an empty container name
		if containerName == "" {
			continue
		}

		container := types.Container{
			Image: aws.String(rawContainerMap["image"].(string)),
		}

		var commands []string
		// "command" is a []interface{} on top of []string, but Lightsail API needs a []*string
		for _, command := range rawContainerMap["command"].([]interface{}) {
			commands = append(commands, command.(string))
		}
		container.Command = commands

		environmentMap := make(map[string]string)
		rawEnvironments := rawContainerMap["environment"].(*schema.Set).List()
		// rawEnvironment is a map[string]interface{} on top of map[string]string, but Lightsail API needs a map[string]*string
		for _, rawEnvironment := range rawEnvironments {
			rawEnvironmentMap := rawEnvironment.(map[string]interface{})
			environmentMap[rawEnvironmentMap["key"].(string)] = rawEnvironmentMap["value"].(string)
		}
		container.Environment = environmentMap

		portsMap := make(map[string]types.ContainerServiceProtocol)
		rawPorts := rawContainerMap["port"].(*schema.Set).List()
		// rawPort is a map[string]interface{} on top of map[string]string, but Lightsail API needs a map[string]*string
		for _, rawPort := range rawPorts {
			rawPortMap := rawPort.(map[string]interface{})
			portNumber := strconv.Itoa(rawPortMap["port_number"].(int))
			portsMap[portNumber] = types.ContainerServiceProtocol(rawPortMap["protocol"].(string))
		}
		container.Ports = portsMap

		result[containerName] = container
	}

	return result
}

func expandLightsailContainerServiceDeploymentPublicEndpoint(rawEndpoint []interface{}) *types.EndpointRequest {
	if len(rawEndpoint) == 0 {
		return nil
	}

	endpoint := types.EndpointRequest{}

	for _, re := range rawEndpoint {
		reMap := re.(map[string]interface{})

		endpoint.ContainerName = aws.String(reMap["container_name"].(string))

		endpoint.ContainerPort = aws.Int32(int32(reMap["container_port"].(int)))

		healthCheck := reMap["health_check"].([]interface{})
		if len(healthCheck) > 0 {
			endpoint.HealthCheck = expandLightsailContainerServiceDeploymentPublicEndpointHealthCheck(healthCheck)
		}
	}

	return &endpoint
}
