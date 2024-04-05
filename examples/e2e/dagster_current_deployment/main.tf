variable "organization" {
  type = string
}

variable "deployment" {
  type = string
}

variable "api_token" {
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
  organization = var.organization
  deployment   = var.deployment
  api_token    = var.api_token
}

data "dagster_current_deployment" "current" {}

output "deployment_id" {
  value = data.dagster_current_deployment.current.id
}

output "deployment_name" {
  value = data.dagster_current_deployment.current.name
}
