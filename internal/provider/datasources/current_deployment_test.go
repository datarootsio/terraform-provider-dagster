package datasources_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/datarootsio/terraform-provider-dagster/internal/testutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccCurrentDeploymentConfig() string {
	return fmt.Sprintf(testutils.ProviderConfig + `
data "dagster_current_deployment" "this" {}
`)
}

func TestAccResourceBasicCodeLocation(t *testing.T) {
	var deploymentName string
	var deploymentId string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCurrentDeploymentConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testutils.FetchValueFromState("data.dagster_current_deployment.this", "name", &deploymentName),
					testutils.FetchValueFromState("data.dagster_current_deployment.this", "id", &deploymentId),
					testDeployProperties(&deploymentName, &deploymentId),
				),
			},
		},
	})
}

func testDeployProperties(nameFromState *string, idFromState *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		client := testutils.GetDagsterClientFromEnvVars()
		deploymentNameFromEnv := client.Deployment

		deploymentFromClient, err := client.DeploymentClient.GetCurrentDeployment(context.Background())
		if err != nil {
			return err
		}

		if *nameFromState != deploymentNameFromEnv {
			return fmt.Errorf("expected deployment name to be %s, got %s", *nameFromState, deploymentNameFromEnv)
		}
		if *nameFromState != deploymentFromClient.DeploymentName {
			return fmt.Errorf("expected deployment name to be %s, got %s", *nameFromState, deploymentFromClient.DeploymentName)
		}

		if *idFromState != strconv.Itoa(deploymentFromClient.DeploymentId) {
			return fmt.Errorf("expected deployment id to be %v, got %v", *idFromState, deploymentFromClient.DeploymentId)
		}

		return nil
	}
}
