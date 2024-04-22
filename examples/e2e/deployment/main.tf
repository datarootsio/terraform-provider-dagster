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

resource "dagster_deployment" "test" {
  name              = "test-quinten"
  settings_document = <<JSON
{
  "run_queue": {
    "max_concurrent_runs": 11,
    "tag_concurrency_limits": []
  },
  "run_monitoring": {
    "start_timeout_seconds": 1200,
    "cancel_timeout_seconds": 1200,
    "free_slots_after_run_end_seconds": 300
  },
  "run_retries": {
    "max_retries": 0,
    "retry_on_asset_or_op_failure": true
  },
  "sso_default_role": "VIEWER",
  "non_isolated_runs": {
    "max_concurrent_non_isolated_runs": 1
  },
  "auto_materialize": {
    "run_tags": {},
    "respect_materialization_data_versions": false,
    "use_sensors": false
  }
}
JSON
}

# run_queue:
#   max_concurrent_runs: 10
#   tag_concurrency_limits: []
# run_monitoring:
#   start_timeout_seconds: 1200
#   cancel_timeout_seconds: 1200
#   free_slots_after_run_end_seconds: 300
# run_retries:
#   max_retries: 0
#   retry_on_asset_or_op_failure: true
# sso_default_role: VIEWER
# non_isolated_runs:
#   max_concurrent_non_isolated_runs: 1
# auto_materialize:
#   run_tags: {}
#   respect_materialization_data_versions: false
#   use_sensors: false
