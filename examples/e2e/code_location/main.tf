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

resource "dagster_code_location" "example" {
  name  = "example_code_location"
  image = "python:3.13"
  code_source = {
    python_file = "my_python.py"
  }
}

output "code_location" {
  value = dagster_code_location.example
}


