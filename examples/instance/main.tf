terraform {
  required_providers {
    awslightsail = {
      source  = "registry.terraform.io/deyoungtech/awslightsail"
    }
  }
}

provider "awslightsail" {
  region = "us-east-1"    
  default_tags {
   tags = {
     Environment = "Testing"
     Owner       = "TFProviders"
     Project     = "Testproject"
   }
 }
}

data "awslightsail_availability_zones" "all" {}

resource "awslightsail_instance" "test" {
  name              = "testing"
  availability_zone = data.awslightsail_availability_zones.all.names[0]
  blueprint_id      = "amazon_linux"
  bundle_id         = "nano_1_0"
  tags = {
    test = "test"
    test1 = "test2"
  } 
}