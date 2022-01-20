terraform {
  required_providers {
    awslightsail = {
      source  = "registry.terraform.io/deyoungtech/awslightsail"
    }
  }
}

provider "awslightsail" {
    region = "us-east-1"
}

data "awslightsail_availability_zones" "all" {}

resource "awslightsail_instance" "test" {
  count = 2
  name              = format("testingI-%s", count.index)
  availability_zone = data.awslightsail_availability_zones.all.names[0]
  blueprint_id      = "amazon_linux"
  bundle_id         = "nano_1_0"
  tags = {
    test = "test"
    test1 = "test2"
  } 
}

resource "awslightsail_lb" "test" {
  name              = "testingLB"
  health_check_path = "/"
  instance_port     = "80"
}

resource "awslightsail_lb_attachment" "test" {
  count = 2
  load_balancer_name = awslightsail_lb.test.name
  instance_name      = awslightsail_instance.test[count.index].name
}