---
page_title: "AWS: awslightsail_static_ip"
description: |-
  Provides an Lightsail Static IP
---

# Resource: awslightsail_static_ip

Allocates a static IP address.

~> **Note:** Lightsail is currently only supported in a limited number of AWS Regions, please see ["Regions and Availability Zones in Amazon Lightsail"](https://lightsail.aws.amazon.com/ls/docs/overview/article/understanding-regions-and-availability-zones-in-amazon-lightsail) for more details

## Example Usage

```terraform
resource "awslightsail_static_ip" "test" {
  name = "example"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name for the allocated static IP

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the Lightsail static IP
* `ip_address` - The allocated static IP address
* `support_code` - The support code.

## Import

Lightsail Static IP addresses can be imported using their name, e.g.,

``` shell
terraform import awslightsail_static_ip.gitlab_test 'gitlab-ip'
```
