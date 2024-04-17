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
  email = "quinten.bruynseraede@telenetgroup.be"
}

output "user_info" {
  value = data.dagster_user.user
  # user_info = {
  #   "email" = "quinten.bruynseraede@telenetgroup.be"
  #   "id" = 104379
  #   "is_scim_provisioned" = false
  #   "name" = ""
  #   "picture" = ""
  # }
}

output "user_email" {
  value = data.dagster_user.user.email
}

resource "dagster_user" "test" {
  email = "foo@bar.be"
}

output "new_user" {
  value = dagster_user.test
  # new_user = {
  #   "email" = "foo@bar.be"
  #   "id" = 108370
  #   "name" = ""
  #   "picture" = ""
}
