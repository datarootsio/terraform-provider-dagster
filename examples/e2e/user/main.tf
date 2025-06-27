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


data "dagster_user" "user" {
  email = "john.doe@dataroots.io"
}

output "user_info" {
  value = data.dagster_user.user
}

output "user_email" {
  value = data.dagster_user.user.email
}

resource "dagster_user" "test" {
  email = "alice.baker@dataroots.io"
  remove_default_permissions = true
}

output "new_user" {
  value = dagster_user.test
}

data "dagster_users" "users" {
  email_regex = "^[^@]+@[^@]+\\.io$"
}

output "filtered_users" {
  value = data.dagster_users.users
}
