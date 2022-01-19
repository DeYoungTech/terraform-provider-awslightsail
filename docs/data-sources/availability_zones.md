---
page_title: "AWS Lightsail: awslightsail_availability_zones"
description: |-
  Provides a list of Availability Zones which can be used by an AWS account.
---

# Data Source: awslightsail_availability_zones

The Availability Zones data source allows access to the list of AWS
Availability Zones which can be accessed by an AWS account within the region
configured in the provider.

## Example Usage

``` hcl
data "awslightsail_availability_zones" "all" {}
```

## Argument Reference

There are no arguments for this data source.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Region name configured in the provider.
* `names` - A list of the Availability Zone names available to the account.
* `database_names` - A list of the Availability Zone names available to the account for Databases.
