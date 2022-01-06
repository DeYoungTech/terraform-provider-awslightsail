---
layout: "awslightsail"
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
      source  = "deyoungtech/awslightsail"
      version = "~> 1.0"
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
  version = "~> 1.0"
  region  = "us-east-1"
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
$ export AWS_ACCESS_KEY_ID="anaccesskey"
$ export AWS_SECRET_ACCESS_KEY="asecretkey"
$ export AWS_REGION="us-west-2"
$ terraform plan
```

## Authentication

Currently the awslightsail provider is using the default configuration for the aws sdk for go v2. This is documented here: https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/#specifying-credentials

The short version is that it will attempt to autheticate using the following order of configurations:

1. Use IAM roles for tasks if your application uses an ECS task definition or RunTask API operation.

2. Use IAM roles for Amazon EC2 (if your application is running on an Amazon EC2 instance).

IAM roles provide applications on the instance temporary security credentials to make AWS calls. IAM roles provide an easy way to distribute and manage credentials on multiple Amazon EC2 instances.

3. Use shared credentials or config files.

The credentials and config files are shared across other AWS SDKs and AWS CLI. As a security best practice, we recommend using credentials file for setting sensitive values such as access key IDs and secret keys. Here are the formatting requirements for each of these files.

4. Use environment variables.

Setting environment variables is useful if you’re doing development work on a machine other than an Amazon EC2 instance.

You can provide your credentials via the `AWS_ACCESS_KEY_ID` and
`AWS_SECRET_ACCESS_KEY`, environment variables, representing your AWS
Access Key and AWS Secret Key, respectively.  Note that setting your
AWS credentials using either these (or legacy) environment variables
will override the use of `AWS_SHARED_CREDENTIALS_FILE` and `AWS_PROFILE`.
The `AWS_DEFAULT_REGION` and `AWS_SESSION_TOKEN` environment variables
are also used, if applicable:
