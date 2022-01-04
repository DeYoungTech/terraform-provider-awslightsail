terraform {
  required_providers {
    lightsail = {
      source  = "registry.terraform.io/brittandeyoung/lightsail"
      version = "~> 0.1.0"
    }
  }
}

provider "lightsail" {
    // region = "us-east-1"
}

data "lightsail_availability_zones" "all" {}

resource "lightsail_instance" "test" {
  name              = "testing"
  availability_zone = data.lightsail_availability_zones.all.names[0]
  blueprint_id      = "amazon_linux"
  bundle_id         = "nano_1_0"
  tags = {
    test = "test"
    test1 = "test2"
  } 
}