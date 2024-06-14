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

resource "dagster_team" "example" {
  name = "example_team"
}

resource "dagster_team_deployment_grant" "example" {
  deployment_id = data.dagster_current_deployment.current.id
  team_id       = dagster_team.example.id

  grant = "VIEWER" # One of ["VIEWER" "LAUNCHER" "EDITOR" "ADMIN" ]

  code_location_grants = [
    {
      name  = "example_code_location"
      grant = "LAUNCHER" # One of ["LAUNCHER" "EDITOR" "ADMIN"]
    },
  ]
}

output "team" {
  value = dagster_team.example
}

output "team_deployment_grant" {
  value = dagster_team_deployment_grant.example
}
