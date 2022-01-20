terraform {
  required_providers {
    awslightsail = {
      source  = "deyoungtech/awslightsail"
    }
  }
}

provider "awslightsail" {
    region = "us-east-1"
}

data "awslightsail_availability_zones" "all" {}

resource "awslightsail_domain" "test" {
  domain_name = "testing-lightsail.com"
  tags = {
    test = "test"
    test1 = "test2"
  } 
}

resource "awslightsail_domain_entry" "test" {
  domain_name = awslightsail_domain.test.domain_name
  name        = "test123"
  type        = "A"
  target      = "127.0.0.1"
  is_alias    = false
}

resource "awslightsail_domain_entry" "testing" {
  domain_name = awslightsail_domain.test.domain_name
  name        = "testing"
  type        = "A"
  target      = "127.0.0.1"
  is_alias    = false
}