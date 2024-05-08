package service_test

import (
	"context"
	"os"
	"testing"

	"github.com/datarootsio/terraform-provider-dagster/internal/client/schema"
	"github.com/datarootsio/terraform-provider-dagster/internal/client/types"
	"github.com/datarootsio/terraform-provider-dagster/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestTeamsService_BasicCRUD(t *testing.T) {
	client := testutils.GetDagsterClientFromEnvVars()

	ctx := context.Background()

	teamsClient := client.TeamsClient

	var errNotFound *types.ErrNotFound
	teamName := "testing/my_team"
	teamNameRenamed := "testing/my_team_renamed"

	// Ensure no teams with the test names exist
	_, err := teamsClient.GetTeamByName(ctx, teamName)
	assert.ErrorAs(t, err, &errNotFound)

	_, err = teamsClient.GetTeamByName(ctx, teamNameRenamed)
	assert.ErrorAs(t, err, &errNotFound)

	teamCreated, err := teamsClient.CreateTeam(ctx, teamName)
	assert.NoError(t, err)
	assert.Equal(t, teamName, teamCreated.Name, "Expected team names to be the same.")

	teamById, err := teamsClient.GetTeamById(ctx, teamCreated.Id)
	assert.NoError(t, err)
	assert.Equal(t, teamName, teamById.Name, "Expected team names to be the same.")

	teamByName, err := teamsClient.GetTeamByName(ctx, teamName)
	assert.NoError(t, err)
	assert.Equal(t, teamName, teamByName.Name, "Expected team names to be the same.")

	_, err = teamsClient.RenameTeam(ctx, teamNameRenamed, teamCreated.Id)
	assert.NoError(t, err)

	teamRenamed, err := teamsClient.GetTeamByName(ctx, teamNameRenamed)
	assert.NoError(t, err)
	assert.Equal(t, teamNameRenamed, teamRenamed.Name, "Expected team names to be the same.")
	assert.Equal(t, teamCreated.Id, teamRenamed.Id, "Expected team ids to be the same.")

	err = teamsClient.DeleteTeam(ctx, teamCreated.Id)
	assert.NoError(t, err)

	// Ensure everything is cleaned up
	_, err = teamsClient.GetTeamByName(ctx, teamName)
	assert.ErrorAs(t, err, &errNotFound)

	_, err = teamsClient.GetTeamByName(ctx, teamNameRenamed)
	assert.ErrorAs(t, err, &errNotFound)
}

func TestTeamsDeploymentGrants(t *testing.T) {
	client := testutils.GetDagsterClientFromEnvVars()

	ctx := context.Background()
	deploymentClient := client.DeploymentClient
	teamsClient := client.TeamsClient
	codelocationClient := client.CodeLocationsClient

	deploymentName := os.Getenv("TF_VAR_testing_dagster_deployment")
	if deploymentName == "" {
		t.Errorf("Deployment with name %s not found!", deploymentName)
	}
	deployment, err := deploymentClient.GetDeploymentByName(ctx, deploymentName)
	if err != nil {
		t.Errorf("Error getting deployment with name %s", deploymentName)
	}

	teamName := "test_team3"
	codeLocation := types.CodeLocation{
		Name:  "testing_codelocation",
		Image: "test123_2",
		CodeSource: types.CodeLocationCodeSource{
			PythonFile: "test.py-2",
		},
	}
	var errNotFound *types.ErrNotFound

	// Ensure no codelocation or team with the test name exist
	_, err = teamsClient.GetTeamByName(ctx, teamName)
	assert.ErrorAs(t, err, &errNotFound)
	_, err = codelocationClient.GetCodeLocationByName(ctx, codeLocation.Name)
	assert.ErrorAs(t, err, &errNotFound)

	// Create team, code location and deployment
	team, err := teamsClient.CreateTeam(ctx, teamName)
	assert.NoError(t, err)
	err = codelocationClient.AddCodeLocation(ctx, codeLocation)
	assert.Nil(t, err)
	t.Cleanup(func() {
		teamsClient.DeleteTeam(ctx, team.Id)
		codelocationClient.DeleteCodeLocation(ctx, codeLocation.Name)
	})

	// Check current deployment grant doesn't exist
	_, err = teamsClient.GetTeamDeploymentGrantByTeamAndDeploymentId(ctx, team.Id, deployment.DeploymentId)
	assert.ErrorAs(t, err, &errNotFound)

	// Set grant to ADMIN and ADMIN on the codelocation and check result
	grant, err := teamsClient.CreateOrUpdateTeamDeploymentGrant(
		ctx,
		team.Id,
		deployment.DeploymentId,
		schema.PermissionGrantViewer,
		[]schema.LocationScopedGrant{
			{LocationName: codeLocation.Name, Grant: schema.PermissionGrantLauncher},
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, grant.DeploymentId, deployment.DeploymentId)
	assert.Equal(t, grant.Grant, schema.PermissionGrantViewer)

	for _, locationGrant := range grant.LocationGrants {
		if locationGrant.LocationName == codeLocation.Name {
			assert.Equal(t, schema.PermissionGrantLauncher, locationGrant.Grant)
		}
	}

	// Read the grant again and ensure the results are consistent
	grantRead, err := teamsClient.GetTeamDeploymentGrantByTeamAndDeploymentId(ctx, team.Id, deployment.DeploymentId)
	assert.NoError(t, err)
	assert.Equal(t, grantRead.DeploymentId, deployment.DeploymentId)
	assert.Equal(t, grantRead.Grant, schema.PermissionGrantViewer)
	for _, locationGrant := range grantRead.LocationGrants {
		if locationGrant.LocationName == codeLocation.Name {
			assert.Equal(t, schema.PermissionGrantLauncher, locationGrant.Grant)
		}
	}

	// Remove the grant and check the result
	err = teamsClient.RemoveTeamDeploymentGrant(ctx, team.Id, deployment.DeploymentId)
	assert.NoError(t, err)
	_, err = teamsClient.GetTeamDeploymentGrantByTeamAndDeploymentId(ctx, team.Id, deployment.DeploymentId)
	assert.ErrorAs(t, err, &errNotFound)
}
