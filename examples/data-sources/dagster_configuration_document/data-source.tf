data "dagster_configuration_document" "defaults" {
  yaml_body = <<YAML
run_queue:
  max_concurrent_runs: 30
  tag_concurrency_limits: []
run_monitoring:
  start_timeout_seconds: 1200
  cancel_timeout_seconds: 1200
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
