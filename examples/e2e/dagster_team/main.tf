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

resource "dagster_team" "example" {
  name = "my_example_team_from_provider_aaa"
}

output "team" {
  value = dagster_team.example
}

