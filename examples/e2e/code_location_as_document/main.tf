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

resource "dagster_code_location_from_document" "example" {
  document = data.dagster_configuration_document.example.json
}

data "dagster_configuration_document" "example" {
  yaml_body = <<YAML
location_name: "example_code_location_from_document"
code_source:
  python_file: "a_python_file.py"
YAML
}

output "code_location" {
  value = dagster_code_location_from_document.example
}
