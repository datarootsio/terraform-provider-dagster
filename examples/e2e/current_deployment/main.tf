variable "testing_dagster_organization" {
  type = string
}

variable "testing_dagster_deployment" {
  type = string
}

variable "testing_dagster_api_token" {
  type = string
}

terraform {
  required_version = ">= 1.3.5"

  required_providers {
    dagster = {
      source = "datarootsio/dagster"
    }
  }

  backend "local" {
    path = "terraform.tfstate"
  }
}

provider "dagster" {
  organization = var.testing_dagster_organization
  deployment   = var.testing_dagster_deployment
  api_token    = var.testing_dagster_api_token
}

data "dagster_current_deployment" "current" {}

output "testing_dagster_deployment" {
  value = data.dagster_current_deployment.current
}
