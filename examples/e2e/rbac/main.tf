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

resource "dagster_team" "rbac" {
  name = "rbac_team"
}

resource "dagster_code_location" "rbac" {
  name  = "rbac_code_location"
  image = "python:3.13"
  code_source = {
    python_file = "my_python.py"
  }
}

resource "dagster_team_deployment_grant" "rbac" {
  deployment_id = data.dagster_current_deployment.current.id
  team_id       = dagster_team.rbac.id

  grant = "VIEWER" # One of ["VIEWER" "LAUNCHER" "EDITOR" "ADMIN" ]

  code_location_grants = [
    {
      name  = dagster_code_location.rbac.name
      grant = "LAUNCHER" # One of ["LAUNCHER" "EDITOR" "ADMIN"]
    },
  ]
}

output "team" {
  value = dagster_team.rbac
}

output "team_deployment_grant" {
  value = dagster_team_deployment_grant.rbac
}

output "code_location" {
  value = dagster_code_location.rbac
}
