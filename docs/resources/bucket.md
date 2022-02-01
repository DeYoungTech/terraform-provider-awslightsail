---
page_title: "AWS Lightsail: awslightsail_bucket"
description: |-
  Provides a lightsail bucket
---

# Resource: awslightsail_bucket

Provides a lightsail bucket.

## Example Usage

```terraform
resource "awslightsail_bucket" "test" {
  name      = "test"
  bundle_id = "small_1_0"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Lightsail load balancer.
* `bundle_id` - (Required) The bundle of specification information (see list below)
* `versioning_enabled` - (Optional) Determines if versioning is enabled for the bucket. 
* `tags` - (Optional) A map of tags to assign to the resource. To create a key-only tag, use an empty string as the value. If configured with a provider `default_tags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.

## Bundles

Lightsail currently supports the following Bundle IDs for Buckets:

* `small_1_0`
* `medium_1_0`
* `large_1_0`

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the lightsail bucket (matches `name`).
* `arn` - The ARN of the Lightsail bucket.
* `url` - The URL of the bucket.
* `created_at` - The timestamp when the instance was created.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider `default_tags` configuration block.
* `support_code` - The support code for the database. Include this code in your email to support when you have questions about a database in Lightsail. This code enables our support team to look up your Lightsail information more easily.
