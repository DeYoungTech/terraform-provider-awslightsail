---
page_title: "AWS Lightsail: awslightsail_static_ip_attachment"
description: |-
  Provides an Lightsail Static IP Attachment
---

# Resource: awslightsail_static_ip_attachment

Provides a static IP address attachment - relationship between a Lightsail static IP & Lightsail instance.

~> **Note:** Lightsail is currently only supported in a limited number of AWS Regions, please see ["Regions and Availability Zones in Amazon Lightsail"](https://lightsail.aws.amazon.com/ls/docs/overview/article/understanding-regions-and-availability-zones-in-amazon-lightsail) for more details

## Example Usage

```terraform
resource "awslightsail_static_ip_attachment" "test" {
  static_ip_name = awslightsail_static_ip.test.id
  instance_name  = awslightsail_instance.test.id
}

resource "awslightsail_static_ip" "test" {
  name = "example"
}

resource "awslightsail_instance" "test" {
  name              = "example"
  availability_zone = "us-east-1b"
  blueprint_id      = "string"
  bundle_id         = "string"
  key_pair_name     = "some_key_name"
}
```

## Argument Reference

The following arguments are supported:

* `static_ip_name` - (Required) The name of the allocated static IP
* `instance_name` - (Required) The name of the Lightsail instance to attach the IP to

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `ip_address` - The allocated static IP address
