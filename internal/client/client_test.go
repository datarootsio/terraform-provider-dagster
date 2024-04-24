package client

import (
	"context"
	"os"
	"testing"

	"github.com/datarootsio/terraform-provider-dagster/internal/client/types"
	"github.com/stretchr/testify/assert"
)

func TestDagsterClient_TeamsClient_BasicCRUD(t *testing.T) {
	organization := os.Getenv("TERRAFORM_PROVIDER_DAGSTER_TESTING_ORGANIZATION")
	deployment := os.Getenv("TERRAFORM_PROVIDER_DAGSTER_TESTING_DEPLOYMENT")
	apiToken := os.Getenv("TERRAFORM_PROVIDER_DAGSTER_TESTING_API_TOKEN")

	ctx := context.Background()

	client, err := NewDagsterClient(organization, deployment, apiToken)
	assert.NoError(t, err)

	teamsClient := client.TeamsClient

	var errNotFound *types.ErrNotFound
	teamName := "testing/my_team"
	teamNameRenamed := "testing/my_team_renamed"

	// Ensure no teams with the test names exist
	_, err = teamsClient.GetTeamByName(ctx, teamName)
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
