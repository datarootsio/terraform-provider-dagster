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

resource "dagster_user" "this" {
  email = "foo.bar@dataroots.io"
}

resource "dagster_team" "this" {
  name = "my_example_team"
}

resource "dagster_team_membership" "this" {
  user_id = dagster_user.this.id
  team_id = dagster_team.this.id
}
