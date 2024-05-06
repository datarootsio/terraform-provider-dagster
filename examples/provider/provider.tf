# With Dagster Cloud CLI Env vars
# - DAGSTER_CLOUD_ORGANIZATION
# - DAGSTER_CLOUD_DEPLOYMENT
# - DAGSTER_CLOUD_API_TOKEN
provider "dagster" {}

# Explicitly
# Note: Explicit configuration in the provider block takes precedences over the Env vars from the example above.
provider "dagster" {
  organization = var.organization
  deployment   = var.deployment
  api_token    = var.api_token
}
