package resources_test

import (
	"fmt"
	"testing"

	"github.com/datarootsio/terraform-provider-dagster/internal/testutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccResourceDeploymentConfig(name string) string {
	return fmt.Sprintf(testutils.ProviderConfig+`
resource "dagster_deployment" "this" {
  name              = "%s"
  settings_document = data.dagster_configuration_document.this.json
  force_destroy     = true
}

data "dagster_configuration_document" "this" {
  yaml_body = <<YAML
concurrency: {}
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
`, name)
}

func TestAccResourceBasicDeployment(t *testing.T) {
	deploymentName := "deployment-" + acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	var settings string
	var id string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create team
			{
				Config: testAccResourceDeploymentConfig(deploymentName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagster_deployment.this", "name", deploymentName),
					resource.TestCheckResourceAttr("dagster_deployment.this", "force_destroy", "true"),
					testutils.FetchValueFromState("dagster_deployment.this", "settings_document", &settings),
					testutils.FetchValueFromState("dagster_deployment.this", "id", &id),
				),
			},
		},
	})
}
