---
page_title: "AWS Lightsail: awslightsail_domain_entry"
description: |-
  Provides an Lightsail Domain Entry
---

# Resource: awslightsail_domain_entry

Creates a domain entry resource

## Example Usage

```terraform
resource "awslightsail_domain" "test" {
  domain_name = "mydomain.com"
}
resource "awslightsail_domain_entry" "test" {
  domain_name = awslightsail_domain.test.domain_name
  name        = "test"
  type        = "A"
  target      = "127.0.0.1"
}
```

## Argument Reference

The following arguments are supported:

* `domain_name` - (Required) The name of the Lightsail domain in which to create the entry
* `name` - (Required) Name of the entry record
* `type` - (Required) Type of record
* `target` - (Required) Target of the domain entry
* `is_alias` - (Optional) If the entry should be an alias Defaults to `false`
