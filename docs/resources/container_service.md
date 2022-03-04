---
page_title: "AWS Lightsail: awslightsail_container_service"
description: |- 
  Provides a resource to manage Lightsail container service
---

# Resource: awslightsail_container_service

An Amazon Lightsail container service is a highly scalable compute and networking resource on which you can deploy, run,
and manage containers. For more information, see
[Container services in Amazon Lightsail](https://lightsail.aws.amazon.com/ls/docs/en_us/articles/amazon-lightsail-container-services).

**Note:** For more information about the AWS Regions in which you can create Amazon Lightsail container services,
see ["Regions and Availability Zones in Amazon Lightsail"](https://lightsail.aws.amazon.com/ls/docs/overview/article/understanding-regions-and-availability-zones-in-amazon-lightsail).

## Example Usage

```terraform
# create a new Lightsail container service
resource "awslightsail_container_service" "my_container_service" {
  name = "container-service-1"
  power = "nano"
  scale = 1
  is_disabled = false
  tags = {
    foo1 = "bar1"
    foo2 = ""
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name for the container service. Names must be of length 1 to 63, and be
  unique within each AWS Region in your Lightsail account.
* `power` - (Required) The power specification for the container service. The power specifies the amount of memory,
  the number of vCPUs, and the monthly price of each node of the container service.
* `scale` - (Required) The scale specification for the container service. The scale specifies the allocated compute
  nodes of the container service.
* `is_disabled` - (Optional) A Boolean value indicating whether the container service is disabled. Defaults to `false`.
* `tags` - (Optional) Map of container service tags. To tag at launch, specify the tags in the Launch Template. If configured with a provider `default_tags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the container service.
* `availability_zone` - The Availability Zone. Follows the format us-east-2a (case-sensitive).
* `id` - Same as `name`.
* `power_id` - The ID of the power of the container service.
* `principal_arn`- The principal ARN of the container service. The principal ARN can be used to create a trust
  relationship between your standard AWS account and your Lightsail container service. This allows you to give your
  service permission to access resources in your standard AWS account.
* `private_domain_name` - The private domain name of the container service. The private domain name is accessible only
  by other resources within the default virtual private cloud (VPC) of your Lightsail account.
* `region_name` - The AWS Region name.
* `resource_type` - The Lightsail resource type of the container service (i.e., ContainerService).
* `state` - The current state of the container service.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider `default_tags`.

## Import

`awslightsail_container_service` can be imported using their name, e.g.
```shell
$ terraform import awslightsail_container_service.my_container_service container-service-1`
```