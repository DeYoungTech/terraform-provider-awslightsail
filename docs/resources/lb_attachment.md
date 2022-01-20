---
page_title: "AWS Lightsail: awslightsail_lb_attachment"
description: |-
  Provides the ability to register a Lightsail Instance to a Load Balancer
---

# Resource: awslightsail_lb_attachment

Attaches a Lightsail Instance to a load balancer resource.

~> **Note:** Lightsail is currently only supported in a limited number of AWS Regions, please see ["Regions and Availability Zones in Amazon Lightsail"](https://lightsail.aws.amazon.com/ls/docs/overview/article/understanding-regions-and-availability-zones-in-amazon-lightsail) for more details

## Example Usage

```terraform
data "awslightsail_availability_zones" "all" {}

resource "awslightsail_instance" "test" {
  name              = "testingI"
  availability_zone = data.awslightsail_availability_zones.all.names[0]
  blueprint_id      = "amazon_linux"
  bundle_id         = "nano_1_0"
  tags = {
    test  = "test"
    test1 = "test2"
  }
}

resource "awslightsail_lb" "test" {
  name              = "testingLB"
  health_check_path = "/"
  instance_port     = "80"
}

resource "awslightsail_lb_attachment" "test" {
  load_balancer_name = awslightsail_lb.test.name
  instance_name      = awslightsail_instance.test.name
}
```

## Argument Reference

The following arguments are supported:

* `load_balancer_name` - (Required) The name of the Lightsail load balancer.
* `instance_name` - (Required) The name of the instance to attach to the load balancer.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` -  A combination of attributes to create a unique id: `load_balancer_name`_`instance_name`
