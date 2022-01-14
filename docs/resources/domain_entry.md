---
page_title: "AWS Lightsail: awslightsail_domain_entry"
description: |-
  Provides an Lightsail Domain Entry
---

# Resource: awslightsail_domain_entry

Creates a domain entry resource

## Example Usage, creating a new domain

```hcl
resource "awslightsail_domain" "domain_test" {
  domain_name = "mydomain.com"
}
resource "awslightsail_domain_entry" "entry_test" {
  domain_name = "${awslightsail_domain.domain_test.domain_name}"
  name        = "a3.${awslightsail_domain.domain_test.domain_name}"
  type        = "A"
  target      = "127.0.0.1"
  depends_on  = ["awslightsail_domain.domain_test"]
}
```

## Argument Reference

The following arguments are supported:

* `domain_name` - (Required) The name of the Lightsail domain in which to create the entry
* `name` - (Required) Name of the entry record
* `type` - (Required) Type of record
* `target` - (Required) Target of the domain entry
* `is_alias` - (Optional) If the entry should be an alias Defaults to `false`
