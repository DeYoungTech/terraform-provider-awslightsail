---
page_title: "Provider: AWSLightsail"
description: |-
  Use the Amazon Web Services (AWS) Lightsail provider to interact with the many resources supported by AWS Lightsail. You must configure the provider with the proper credentials before you can use it.
---

# AWS Lightsail Provider

Use the Amazon Web Services (AWS) Lightsail provider to interact with the
many resources supported by AWS. You must configure the provider
with the proper credentials before you can use it.

Use the navigation to the left to read about the available resources.

## Example Usage

Terraform 0.13 and later:

```terraform
terraform {
  required_providers {
    aws = {
      source = "deyoungtech/awslightsail"
    }
  }
}

# Configure the Provider
provider "awslightsail" {
  region = "us-east-1"
}

# Create an instance
resource "awslightsail_instance" "test" {
  name              = "test_instance"
  availability_zone = "us-east-1b"
  blueprint_id      = "amazon_linux"
  bundle_id         = "nano_1_0"
  tags = {
    foo = "bar"
  }
}
```

Terraform 0.12 and earlier:

```terraform
# Configure the AWS Provider
provider "awslightsail" {
  region = "us-east-1"
}

# Create an instance
resource "awslightsail_instance" "test" {
  name              = "test_instance"
  availability_zone = "us-east-1b"
  blueprint_id      = "amazon_linux"
  bundle_id         = "nano_1_0"
  tags = {
    foo = "bar"
  }
}
```

Usage:

```sh
export AWS_ACCESS_KEY_ID="anaccesskey"
export AWS_SECRET_ACCESS_KEY="asecretkey"
export AWS_REGION="us-west-2"
terraform plan
```

## Authentication

Currently the awslightsail provider is using the default configuration for the aws sdk for go v2. This is documented here: ["aws-sdk-go-v2 documentation"](https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/#specifying-credentials)

The short version is that it will attempt to autheticate using the following order of configurations:

1. Use IAM roles for tasks if your application uses an ECS task definition or RunTask API operation.

2. Use IAM roles for Amazon EC2 (if your application is running on an Amazon EC2 instance). IAM roles provide applications on the instance temporary security credentials to make AWS calls. IAM roles provide an easy way to distribute and manage credentials on multiple Amazon EC2 instances.

3. Use shared credentials or config files. The credentials and config files are shared across other AWS SDKs and AWS CLI. As a security best practice, we recommend using credentials file for setting sensitive values such as access key IDs and secret keys. Here are the formatting requirements for each of these files.

4. Use environment variables.

Setting environment variables is useful if you’re doing development work on a machine other than an Amazon EC2 instance.

You can provide your credentials via the `AWS_ACCESS_KEY_ID` and
`AWS_SECRET_ACCESS_KEY`, environment variables, representing your AWS
Access Key and AWS Secret Key, respectively.  Note that setting your
AWS credentials using either these (or legacy) environment variables
will override the use of `AWS_SHARED_CREDENTIALS_FILE` and `AWS_PROFILE`.
The `AWS_DEFAULT_REGION` and `AWS_SESSION_TOKEN` environment variables
are also used, if applicable:

## Argument Reference

In addition to [generic `provider` arguments](https://www.terraform.io/docs/configuration/providers.html)
(e.g., `alias` and `version`), the following arguments are supported in the AWS
 `provider` block:

* `region` - (Optional) This is the AWS region. It must be provided, but
  it can also be sourced from the `AWS_DEFAULT_REGION` environment variables, or
  via a shared credentials file if `profile` is specified.

* `default_tags` - (Optional) Configuration block with resource tag settings to apply across all resources handled by this provider. This is designed to replace redundant per-resource `tags` configurations. Provider tags can be overridden with new values, but not excluded from specific resources. To override provider tag values, use the `tags` argument within a resource to configure new tag values for matching keys. See the [`default_tags`](#default_tags-configuration-block) Configuration Block section below for example usage and available arguments. This functionality is supported in all resources that implement `tags`.

* `ignore_tags` - (Optional) Configuration block with resource tag settings to ignore across all resources handled by this provider for situations where external systems are managing certain resource tags. Arguments to the configuration block are described below in the `ignore_tags` Configuration Block section.

### default_tags Configuration Block

Example: Resource with provider default tags

```terraform
provider "awslightsail" {
  default_tags {
    tags = {
      Environment = "Test"
      Name        = "Provider Tag"
    }
  }
}

resource "awslightsail_instance" "example" {
  # ..other configuration...
}

output "instance_all_tags" {
  value = awslightsail_instance.example.tags_all
}
```

Example: Resource with tags and provider default tags

```terraform
provider "awslightsail" {
  default_tags {
    tags = {
      Environment = "Test"
      Name        = "Provider Tag"
    }
  }
}

resource "awslightsail_instance" "example" {
  # ..other configuration...
  tags = {
    Owner = "example"
  }
}

output "instance_resource_tags" {
  value = awslightsail_instance.example.tags
}

output "instance_all_tags" {
  value = awslightsail_instance.example.tags_all
}
```

Example: Resource overriding provider default tags

```terraform
provider "awslightsail" {
  default_tags {
    tags = {
      Environment = "Test"
      Name        = "Provider Tag"
    }
  }
}

resource "awslightsail_instance" "example" {
  # ..other configuration...
  tags = {
    Environment = "example"
  }
}

output "instance_resource_tags" {
  value = awslightsail_instance.example.tags
}

output "instance_all_tags" {
  value = awslightsail_instance.example.tags_all
}
```

The `default_tags` configuration block supports the following argument:

* `tags` - (Optional) Key-value map of tags to apply to all resources.

### ignore_tags Configuration Block

Example:

```terraform
provider "awslightsail" {
  ignore_tags {
    keys = ["TagKey1"]
  }
}
```

The `ignore_tags` configuration block supports the following arguments:

* `keys` - (Optional) List of exact resource tag keys to ignore across all resources handled by this provider. This configuration prevents Terraform from returning the tag in any `tags` attributes and displaying any configuration difference for the tag value. If any resource configuration still has this tag key configured in the `tags` argument, it will display a perpetual difference until the tag is removed from the argument or [`ignore_changes`](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html#ignore_changes) is also used.
* `key_prefixes` - (Optional) List of resource tag key prefixes to ignore across all resources handled by this provider. This configuration prevents Terraform from returning any tag key matching the prefixes in any `tags` attributes and displaying any configuration difference for those tag values. If any resource configuration still has a tag matching one of the prefixes configured in the `tags` argument, it will display a perpetual difference until the tag is removed from the argument or [`ignore_changes`](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html#ignore_changes) is also used.
