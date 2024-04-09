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

resource "dagster_team" "example" {
  name = "example_team"
}

resource "dagster_team_deployment_grant" "example" {
  deployment_id = data.dagster_current_deployment.current.id
  team_id       = dagster_team.example.id

  grant = "VIEWER" # One of ["LAUNCHER" "EDITOR" "ADMIN" "AGENT" "VIEWER"]
}

# resource "dagster_team_code_location_grant" "example" {
#   deployment_id      = data.dagster_current_deployment.current.id
#   team_id            = dagster_team.example.id
#   code_location_name = "platform-testing-de/dagster-dbt-poc"

#   grant = "VIEWER" # One of ["LAUNCHER" "EDITOR" "ADMIN" "AGENT" "VIEWER"]
# }

output "team" {
  value = dagster_team.example
}

