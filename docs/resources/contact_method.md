---
page_title: "AWS Lightsail: awslightsail_contact_method"
description: |-
  Provides an Lightsail Contact Method
---

# Resource: awslightsail_contact_method

Creates a domain entry resource. Each AWS account is limited to a single endpoint per conact protocol.

## Example Usage

```terraform
resource "awslightsail_contact_method" "test" {
  endpoint = "test@test.com"
  protocol        = "Email"
}
```

## Argument Reference

The following arguments are supported:

* `endpoint` - (Required) The destination of the contact method, such as an email address or a mobile phone number. Use the E.164 format when specifying a mobile phone number
* `protocol` - (Required) The protocol of the contact method, such as Email or SMS (text messaging).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The contact method `protocol`
