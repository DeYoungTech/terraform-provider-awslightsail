---
page_title: "AWS Lightsail: awslightsail_container_public_domain_names"
description: |- 
  Provides a resource to manage Lightsail container public domain names
---

# Resource: awslightsail_container_public_domain_names

Public domain names to be configured as the public endpoint of your container service. Example: www.testdomain.com 

An Amazon Lightsail container service is a highly scalable compute and networking resource on which you can deploy, run,
and manage containers. For more information, see
[Container services in Amazon Lightsail](https://lightsail.aws.amazon.com/ls/docs/en_us/articles/amazon-lightsail-container-services).

**Note:** For more information about the AWS Regions in which you can create Amazon Lightsail container services,
see ["Regions and Availability Zones in Amazon Lightsail"](https://lightsail.aws.amazon.com/ls/docs/overview/article/understanding-regions-and-availability-zones-in-amazon-lightsail).

## Example Usage

```terraform
# create a new Lightsail container service
resource "awslightsail_container_service" "test" {
  name        = "container-service-1"
  power       = "nano"
  scale       = 1
  is_disabled = false
  tags = {
    foo1 = "bar1"
    foo2 = ""
  }
}

resource "awslightsail_certificate" "test" {
  name        = "ExsitingCertificate"
  domain_name = "existingdomain1.com"
  subject_alternative_names = [
    "www.existingdomain1.com"
  ]
}

resource "awslightsail_container_public_domain_names" "test" {
  container_service_name = awslightsail_container_service.test.id
  public_domain_names {
    certificate {
      certificate_name = awslightsail_certificate.test.name
      domain_names = [
        "existingdomain1.com",
        "www.existingdomain1.com",
      ]
    }
  }
}

```

## Argument Reference

The following arguments are supported:

* `container_service_name` - (Required) The name for the container service.
* `public_domain_names` - (Required) The public domain names to use with the container service, such as example.com
  and www.example.com. You can specify up to four public domain names for a container service. The domain names that you
  specify are used when you create a deployment with a container configured as the public endpoint of your container
  service. 
    * `certificate` - (Required) Describes the details of an SSL/TLS certificate for a container service.
        * `certificate_name` - (Required) The certificate's name.
        * `domain_names` - (Required) The domain names. A list of string.

WARNING: You must create and validate an SSL/TLS certificate before you can use public domain names with your container service. For more information, see [Enabling and managing custom domains for your Amazon Lightsail container services](https://lightsail.aws.amazon.com/ls/docs/en_us/articles/amazon-lightsail-creating-container-services-certificates).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Same as `container_service_name`.

## Import

`awslightsail_container_public_domain_names` can be imported using the container service name, e.g.

```shell
$ terraform import awslightsail_container_public_domain_names.test container-service-1
```
