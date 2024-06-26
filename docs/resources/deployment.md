---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "dagster_deployment Resource - dagster"
subcategory: ""
description: |-
  Creates a deployment.
---

# dagster_deployment (Resource)

Creates a deployment.

## Example Usage

```terraform
resource "dagster_deployment" "this" {
  name              = "test-deploy"
  settings_document = data.dagster_configuration_document.this.json
}

data "dagster_configuration_document" "this" {
  yaml_body = <<YAML
run_queue:
  max_concurrent_runs: 30
  tag_concurrency_limits: []
run_monitoring:
  start_timeout_seconds: 1200
  cancel_timeout_seconds: 1400
  free_slots_after_run_end_seconds: 300
run_retries:
  max_retries: 0
  retry_on_asset_or_op_failure: true
sso_default_role: VIEWER
non_isolated_runs:
  max_concurrent_non_isolated_runs: 1
auto_materialize:
  run_tags: {}
  respect_materialization_data_versions: false
  use_sensors: false
YAML
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Deployment name

### Optional

- `force_destroy` (Boolean) When `false`, will check if there are code locations associated with the deployment, if there are, it will block the delete of the deployment. When `true` ignore the code locations check. This is done because when you delete a deployment, you delete all the resources/metadata of that deployment and this is not recoverable. DEFAULT `false`
- `settings_document` (String) Deployment settings as a JSON document. We recommend using a `dagster_configuration_document` to generate this instead of composing a JSON document yourself. Leaving this attribute empty or partially filled in, will result in Dagster (partially) applying default settings to your deployment. This leads to perpetual changes in this resource.

### Read-Only

- `id` (Number) Deployment id
- `status` (String) Deployment status (`ACTIVE` or `PENDNG_DELETION`)
- `type` (String) Deployment type (`PRODUCTION`, `DEV` or `BRANCH`)

## Import

Import is supported using the following syntax:

```shell
# Dagster Deployments can be imported via name
terraform import dagster_deployment.this "test-deployment"
```
