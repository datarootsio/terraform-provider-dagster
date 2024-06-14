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

resource "dagster_user" "this" {
  email = "john.doe@dataroots.io"
}

resource "dagster_team" "this" {
  name = "my_example_team"
}

resource "dagster_team_membership" "this" {
  user_id = dagster_user.this.id
  team_id = dagster_team.this.id
}
