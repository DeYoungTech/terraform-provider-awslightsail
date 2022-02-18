---
page_title: "AWS Lightsail: awslightsail_disk_attachment"
description: |-
  Provides a Lightsail Disk attachment
---

# Resource: awslightsail_disk_attachment

Attaches a Lightsail disk to a Lightsail Instance

~> **Note:** Lightsail is currently only supported in a limited number of AWS Regions, please see ["Regions and Availability Zones in Amazon Lightsail"](https://lightsail.aws.amazon.com/ls/docs/overview/article/understanding-regions-and-availability-zones-in-amazon-lightsail) for more details

## Example Usage

```terraform
data "awslightsail_availability_zones" "all" {}

resource "awslightsail_disk" "test" {
	name              = "test"
	size_in_gb = 8
	availability_zone     = data.awslightsail_availability_zones.all.names[0]
  }

resource "awslightsail_instance" "test" {
  name              = "test"
  availability_zone = data.awslightsail_availability_zones.all.names[0]
  blueprint_id      = "amazon_linux"
  bundle_id         = "nano_1_0"
}

resource "awslightsail_disk_attachment" "test" {
  disk_name = awslightsail_disk.test.name
  instance_name      = awslightsail_instance.test.name
  disk_path = "/dev/xvdf"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Lightsail load balancer.
* `size_in_gb` - (Required) The instance port the load balancer will connect.
* `availability_zone` - (Required) The Availability Zone in which to create your disk.
* `tags` - (Optional) A map of tags to assign to the resource. To create a key-only tag, use an empty string as the value. If configured with a provider `default_tags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the Lightsail load balancer.
* `created_at` - The timestamp when the load balancer was created.
* `id` - The name of the disk  (matches `name`).

Lightsail Disks can be imported using their name, e.g.

```shell
$ terraform import awslightsail_disk_attachment.test 'test'
```
