package service_test

import (
	"context"
	"testing"

	"github.com/datarootsio/terraform-provider-dagster/internal/client/types"
	"github.com/datarootsio/terraform-provider-dagster/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestUserService_BasicCRUD(t *testing.T) {
	client := testutils.GetDagsterClientFromEnvVars().UsersClient
	ctx := context.Background()

	var errNotFound *types.ErrNotFound
	userEmail := "test-user-dagster@test.com"

	// Make sure user doesn't exist
	_, err := client.GetUserByEmail(ctx, userEmail)
	assert.Error(t, err)
	assert.ErrorAs(t, err, &errNotFound)

	// Create user
	createdUser, err := client.AddUser(ctx, userEmail)
	assert.NoError(t, err)
	assert.Equal(t, userEmail, createdUser.Email)

	t.Cleanup(func() {
		_ = client.RemoveUser(ctx, createdUser.Email)
	})

	// Read user by id and email
	userByEmail, err := client.GetUserByEmail(ctx, userEmail)
	assert.NoError(t, err)

	userById, err := client.GetUserById(ctx, int64(createdUser.UserId))
	assert.NoError(t, err)

	assert.Equal(t, userByEmail, userById)

	// Remove user
	err = client.RemoveUser(ctx, createdUser.Email)
	assert.NoError(t, err)

	_, err = client.GetUserById(ctx, int64(createdUser.UserId))
	assert.Error(t, err)
	assert.ErrorAs(t, err, &errNotFound)
}

func TestUserService_Teams(t *testing.T) {
	userClient := testutils.GetDagsterClientFromEnvVars().UsersClient
	teamsClient := testutils.GetDagsterClientFromEnvVars().TeamsClient
	ctx := context.Background()
	var errNotFound *types.ErrNotFound

	userEmail := "test-user-dagster2@test.com"
	teamName := "testing/my_team2"

	// Make sure user and team don't exist
	_, err := userClient.GetUserByEmail(ctx, userEmail)
	assert.ErrorAs(t, err, &errNotFound)
	_, err = teamsClient.GetTeamByName(ctx, teamName)
	assert.ErrorAs(t, err, &errNotFound)

	// Create user and team
	createdUser, err := userClient.AddUser(ctx, userEmail)
	assert.NoError(t, err)
	createdTeam, err := teamsClient.CreateTeam(ctx, teamName)
	assert.NoError(t, err)

	t.Cleanup(func() {
		err = teamsClient.DeleteTeam(ctx, createdTeam.Id)
		err = userClient.RemoveUser(ctx, createdUser.Email)
	})

	// Check that user is not in team
	in_team, err := teamsClient.IsUserInTeam(ctx, createdUser.UserId, createdTeam.Id)
	assert.NoError(t, err)
	assert.False(t, in_team)

	// Add user to team and verify
	err = teamsClient.AddUserToTeam(ctx, createdUser.UserId, createdTeam.Id)
	assert.NoError(t, err)

	in_team, err = teamsClient.IsUserInTeam(ctx, createdUser.UserId, createdTeam.Id)
	assert.NoError(t, err)
	assert.True(t, in_team)

	// Remove user from team and verify
	err = teamsClient.RemoveUserFromTeam(ctx, createdUser.UserId, createdTeam.Id)
	assert.NoError(t, err)

	in_team, err = teamsClient.IsUserInTeam(ctx, createdUser.UserId, createdTeam.Id)
	assert.NoError(t, err)
	assert.False(t, in_team)

	// Remove team
	err = teamsClient.DeleteTeam(ctx, createdTeam.Id)
	assert.NoError(t, err)
	err = userClient.RemoveUser(ctx, createdUser.Email)
	assert.NoError(t, err)
}
