package resources_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	clientTypes "github.com/datarootsio/terraform-provider-dagster/internal/client/types"
	"github.com/datarootsio/terraform-provider-dagster/internal/testutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccResourceTeamDeploymentGrantConfig(teamName string, clName string, deploymentGrant string, clGrant string) string {
	return fmt.Sprintf(testutils.ProviderConfig+`
data "dagster_current_deployment" "current" {}

resource "dagster_team" "this" {
  name = "%s"
}

resource "dagster_code_location" "test" {
  name          = "%s"
  image         = "python:3.13"
  code_source   = {
    python_file = "test.py"
  }
}

resource "dagster_team_deployment_grant" "test" {
  deployment_id = data.dagster_current_deployment.current.id
  team_id       = dagster_team.this.id

  grant = "%s" # One of ["VIEWER" "LAUNCHER" "EDITOR" "ADMIN" ]

  code_location_grants = [
    {
      name  = dagster_code_location.test.name
      grant = "%s" # One of ["LAUNCHER" "EDITOR" "ADMIN"]
    },
  ]
}
`, teamName, clName, deploymentGrant, clGrant)
}

func TestAccResourceBasicTeamDeploymentGrant(t *testing.T) {
	teamName := "team-" + acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	clName := "code-location-" + acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	var deploymentId string
	var teamId string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testGrantDeleted(&teamId, &deploymentId),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTeamDeploymentGrantConfig(teamName, clName, "VIEWER", "LAUNCHER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testutils.FetchValueFromState("data.dagster_current_deployment.current", "id", &deploymentId),
					testutils.FetchValueFromState("dagster_team.this", "id", &teamId),
					testGrantProperties(&teamId, &deploymentId, "VIEWER"),
					resource.TestCheckResourceAttrPtr("dagster_team_deployment_grant.test", "deployment_id", &deploymentId),
					resource.TestCheckResourceAttrPtr("dagster_team_deployment_grant.test", "team_id", &teamId),
					resource.TestCheckResourceAttr("dagster_team_deployment_grant.test", "grant", "VIEWER"),
				),
			},
			{
				Config: testAccResourceTeamDeploymentGrantConfig(teamName, clName, "VIEWER", "EDITOR"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testutils.FetchValueFromState("data.dagster_current_deployment.current", "id", &deploymentId),
					testutils.FetchValueFromState("dagster_team.this", "id", &teamId),
					testGrantProperties(&teamId, &deploymentId, "VIEWER"),
					resource.TestCheckResourceAttrPtr("dagster_team_deployment_grant.test", "deployment_id", &deploymentId),
					resource.TestCheckResourceAttrPtr("dagster_team_deployment_grant.test", "team_id", &teamId),
					resource.TestCheckResourceAttr("dagster_team_deployment_grant.test", "grant", "VIEWER"),
				),
			},
		},
	})
}

func testGrantProperties(teamId *string, deploymentId *string, expectedGrant string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		client := testutils.GetDagsterClientFromEnvVars()
		deploymentIdInt, err := strconv.Atoi(*deploymentId)
		if err != nil {
			return err
		}

		grant, err := client.TeamsClient.GetTeamDeploymentGrantByTeamAndDeploymentId(context.Background(), *teamId, deploymentIdInt)
		if err != nil {
			return err
		}

		expectedGrantTyped, err := clientTypes.ConvertToGrantEnum(expectedGrant)
		if err != nil {
			return err
		}
		if grant.Grant != expectedGrantTyped {
			return fmt.Errorf("expected grant to be %v, got %v", expectedGrant, grant.Grant)
		}

		return err
	}
}

func testGrantDeleted(teamId *string, deploymentId *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		client := testutils.GetDagsterClientFromEnvVars()
		deploymentIdInt, err := strconv.Atoi(*deploymentId)
		if err != nil {
			return err
		}

		_, err = client.TeamsClient.GetTeamDeploymentGrantByTeamAndDeploymentId(context.Background(), *teamId, deploymentIdInt)

		notFound := &clientTypes.ErrNotFound{}
		if errors.As(err, &notFound) {
			return nil
		}
		return err
	}
}
