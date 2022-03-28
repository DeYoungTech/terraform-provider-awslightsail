---
page_title: "AWS Lightsail: awslightsail_container_deployment"
description: |- 
  Provides a resource to manage Lightsail container service deplloyments
---

# Resource: awslightsail_container_deployment

An Amazon Lightsail container service is a highly scalable compute and networking resource on which you can deploy, run,
and manage containers. For more information, see
[Container services in Amazon Lightsail](https://lightsail.aws.amazon.com/ls/docs/en_us/articles/amazon-lightsail-container-services).

**Note:** For more information about the AWS Regions in which you can create Amazon Lightsail container services,
see ["Regions and Availability Zones in Amazon Lightsail"](https://lightsail.aws.amazon.com/ls/docs/overview/article/understanding-regions-and-availability-zones-in-amazon-lightsail).

## Example Usage

```terraform
# create a new Lightsail container service
resource "awslightsail_container_service" "test" {
  name        = "container-service-1"
  power       = "nano"
  scale       = 1
  is_disabled = false
  tags = {
    foo1 = "bar1"
    foo2 = ""
  }
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
  }
}

```

## Argument Reference

The following arguments are supported:

* `container_service_name` - (Required) The name for the container service.
* `container` - (Required) Describes the configuration for a container of the deployment.
    * `container_name` - (Required) The name for the container.
    * `image` - (Required) The name of the image used for the container. Container images sourced from your Lightsail
      container service, that are registered and stored on your service, start with a colon `:`. For
      example, `:container-service-1.mystaticwebsite.1`. Container images sourced from a public registry like Docker Hub
      don't start with a colon. For example, nginx:latest or nginx.
    * `command` - (Optional) The launch command for the container. A list of string.
    * `environment` - (Optional) The environment variables of the container.
        * `key` - (Required) The environment variable name.
        * `value` - (Required) The environment variable value.
    * `port` - (Optional) The open firewall ports of the container.
        * `port_number` - (Required) The port number.
        * `protocol` - (Required) The protocol. Possible values: `"HTTP"`, `HTTPS`, `TCP`, `UDP`.
* `public_endpoint` - (Optional) Describes the public endpoint configuration of a deployment.
    * `container_name` - (Required) The name of the container for the endpoint.
    * `container_port` - (Required) The port of the container to which traffic is forwarded to.
    * `health_check` - (Required) Describes the health check configuration of the container.
        * `healthy_threshold` - (Optional) The number of consecutive health checks successes required before moving the
          container to the Healthy state. Defaults to `2`.
        * `unhealthy_threshold` - (Optional) The number of consecutive health checks failures required before moving the
          container to the Unhealthy state. Defaults to `2`.
        * `timeout_seconds` - (Optional) The amount of time, in seconds, during which no response means a failed health
          check. You can specify between 2 and 60 seconds. Defaults to `2`.
        * `interval_seconds` - (Optional) The approximate interval, in seconds, between health checks of an individual
          container. You can specify between 5 and 300 seconds. Defaults to `5`.
        * `path` - (Optional) The path on the container on which to perform the health check. Defaults to `"/"`.
        * `success_codes` - (Optional) The HTTP codes to use when checking for a successful response from a container.
          You can specify values between 200 and 499. Defaults to `"200-499"`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Same as `container_service_name`.
* `state` - The current state of the container service.
* `version` - The current version of the container deployment.

## Import

`awslightsail_container_deployment` can be imported using the container service name, e.g.

```shell
$ terraform import awslightsail_container_deployment.test container-service-1
```
