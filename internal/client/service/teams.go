package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Khan/genqlient/graphql"
	"github.com/datarootsio/terraform-provider-dagster/internal/client/schema"
	"github.com/datarootsio/terraform-provider-dagster/internal/client/types"
	"github.com/datarootsio/terraform-provider-dagster/internal/utils"
)

type TeamsClient struct {
	client graphql.Client
}

func NewTeamsClient(client graphql.Client) TeamsClient {
	return TeamsClient{
		client: client,
	}
}

func (c *TeamsClient) ListTeams(ctx context.Context) ([]schema.Team, error) {
	resp, err := schema.ListTeams(ctx, c.client)
	if err != nil {
		return nil, err
	}

	teams := make([]schema.Team, 0, len(resp.TeamPermissions))

	for _, teamPermission := range resp.TeamPermissions {
		teams = append(teams, teamPermission.Team.Team)
	}

	return teams, nil
}

func (c *TeamsClient) GetTeamByName(ctx context.Context, name string) (schema.Team, error) {
	teams, err := c.ListTeams(ctx)

	if err != nil {
		return schema.Team{}, err
	}

	for _, team := range teams {
		if team.Name == name {
			return team, nil
		}
	}

	return schema.Team{}, &types.ErrNotFound{What: "Team", Key: "name", Value: name}
}

func (c *TeamsClient) GetTeamById(ctx context.Context, id string) (schema.Team, error) {
	teams, err := c.ListTeams(ctx)

	if err != nil {
		return schema.Team{}, err
	}

	for _, team := range teams {
		if team.Id == id {
			return team, nil
		}
	}

	return schema.Team{}, &types.ErrNotFound{What: "Team", Key: "id", Value: id}
}

func (c *TeamsClient) CreateTeam(ctx context.Context, name string) (schema.Team, error) {
	_, err := c.GetTeamByName(ctx, name)

	if err == nil {
		return schema.Team{}, &types.ErrAlreadyExists{What: "Team", Key: "name", Value: name}
	}

	var errComp *types.ErrNotFound
	if !errors.As(err, &errComp) {
		return schema.Team{}, err
	}

	resp, err := schema.CreateTeam(ctx, c.client, name)

	if err != nil {
		return schema.Team{}, err
	}

	switch respCast := resp.CreateOrUpdateTeam.(type) {
	case *schema.CreateTeamCreateOrUpdateTeamCreateOrUpdateTeamSuccess:
		return respCast.Team.Team, nil
	case *schema.CreateTeamCreateOrUpdateTeamPythonError:
		return schema.Team{}, &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	case *schema.CreateTeamCreateOrUpdateTeamUnauthorizedError:
		return schema.Team{}, &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	default:
		return schema.Team{}, fmt.Errorf("unexpected type(%T) of result", resp.CreateOrUpdateTeam)
	}
}

func (c *TeamsClient) DeleteTeam(ctx context.Context, id string) error {
	_, err := c.GetTeamById(ctx, id)
	if err != nil {
		return err
	}

	resp, err := schema.DeleteTeam(ctx, c.client, id)
	if err != nil {
		return err
	}

	switch respCast := resp.DeleteTeam.(type) {
	case *schema.DeleteTeamDeleteTeamDeleteTeamSuccess:
		return nil
	case *schema.DeleteTeamDeleteTeamPythonError:
		return &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	case *schema.DeleteTeamDeleteTeamUnauthorizedError:
		return &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	default:
		return fmt.Errorf("unexpected type(%T) of result", resp.DeleteTeam)
	}
}

func (c *TeamsClient) RenameTeam(ctx context.Context, name string, id string) (schema.Team, error) {
	_, err := c.GetTeamById(ctx, id)
	if err != nil {
		return schema.Team{}, err
	}

	resp, err := schema.RenameTeam(ctx, c.client, name, id)
	if err != nil {
		return schema.Team{}, err
	}

	switch respCast := resp.RenameTeam.(type) {
	case *schema.RenameTeamRenameTeamDagsterCloudTeam:
		return respCast.Team, nil
	case *schema.RenameTeamRenameTeamPythonError:
		return schema.Team{}, &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	case *schema.RenameTeamRenameTeamUnauthorizedError:
		return schema.Team{}, &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	default:
		return schema.Team{}, fmt.Errorf("unexpected type(%T) of result", resp.RenameTeam)
	}
}

func (c *TeamsClient) GetTeamDeploymentGrantByTeamAndDeploymentId(ctx context.Context, teamId string, deploymentId int) (schema.ScopedPermissionGrant, error) {
	resp, err := schema.ListTeamPermissions(ctx, c.client)
	if err != nil {
		return schema.ScopedPermissionGrant{}, err
	}

	for _, teamPermission := range resp.TeamPermissions {
		if teamPermission.Id == teamId {
			for _, grant := range teamPermission.DeploymentPermissionGrants {
				if grant.DeploymentId == deploymentId {
					return grant.ScopedPermissionGrant, nil
				}
			}
		}
	}

	return schema.ScopedPermissionGrant{}, &types.ErrNotFound{What: "DeploymentGrant", Key: "teamId", Value: teamId}
}

func (c *TeamsClient) CreateOrUpdateTeamDeploymentGrant(ctx context.Context, teamId string, deploymentId int, grant schema.PermissionGrant, locationGrants []schema.LocationScopedGrant) (schema.ScopedPermissionGrant, error) {
	locationGrantsInput := make([]schema.LocationScopedGrantInput, 0, len(locationGrants))

	deploymentGrantIdx := utils.IndexOf(types.DeploymentGrantEnumValues(), string(grant))

	for _, locationGrant := range locationGrants {
		locationGrantIdx := utils.IndexOf(types.DeploymentGrantEnumValues(), string(locationGrant.Grant))

		if utils.IndexOf(types.LocationGrantEnumValues(), string(locationGrant.Grant)) == -1 {
			return schema.ScopedPermissionGrant{}, &types.ErrInvalid{
				What: "TeamDeploymentGrant",
				Message: fmt.Sprintf(
					"LocationGrant must be one of %v",
					types.LocationGrantEnumValues(),
				),
			}
		}

		if deploymentGrantIdx >= locationGrantIdx {
			return schema.ScopedPermissionGrant{}, &types.ErrInvalid{
				What:    "TeamDeploymentGrant",
				Message: "LocationGrant can't be less permissive than DeploymentGrant",
			}
		}

		locationGrantsInput = append(
			locationGrantsInput,
			schema.LocationScopedGrantInput(locationGrant),
		)
	}

	resp, err := schema.CreateOrUpdateTeamPermission(
		ctx,
		c.client,
		deploymentId,
		schema.PermissionDeploymentScopeDeployment,
		grant,
		locationGrantsInput,
		teamId,
	)
	if err != nil {
		return schema.ScopedPermissionGrant{}, err
	}

	// Get the Updated DeploymentGrant => Id changes every update
	updatedPermissionGrant, err := c.GetTeamDeploymentGrantByTeamAndDeploymentId(ctx, teamId, deploymentId)
	if err != nil {
		return schema.ScopedPermissionGrant{}, err
	}

	switch respCast := resp.CreateOrUpdateTeamPermission.(type) {
	case *schema.CreateOrUpdateTeamPermissionCreateOrUpdateTeamPermissionCreateOrUpdateTeamPermissionSuccess:
		return updatedPermissionGrant, nil
	case *schema.CreateOrUpdateTeamPermissionCreateOrUpdateTeamPermissionPythonError:
		return schema.ScopedPermissionGrant{}, &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	case *schema.CreateOrUpdateTeamPermissionCreateOrUpdateTeamPermissionUnauthorizedError:
		return schema.ScopedPermissionGrant{}, &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	case *schema.CreateOrUpdateTeamPermissionCreateOrUpdateTeamPermissionUserLimitError:
		return schema.ScopedPermissionGrant{}, &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	default:
		return schema.ScopedPermissionGrant{}, fmt.Errorf("unexpected type(%T) of result", resp.CreateOrUpdateTeamPermission)
	}
}

func (c *TeamsClient) RemoveTeamDeploymentGrant(ctx context.Context, teamId string, deploymentId int) error {
	resp, err := schema.RemoveTeamPermission(
		ctx,
		c.client,
		deploymentId,
		schema.PermissionDeploymentScopeDeployment,
		teamId,
	)
	if err != nil {
		return err
	}

	switch respCast := resp.RemoveTeamPermission.(type) {
	case *schema.RemoveTeamPermissionRemoveTeamPermissionRemoveTeamPermissionSuccess:
		return nil
	case *schema.RemoveTeamPermissionRemoveTeamPermissionPythonError:
		return &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	case *schema.RemoveTeamPermissionRemoveTeamPermissionUnauthorizedError:
		return &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	case *schema.RemoveTeamPermissionRemoveTeamPermissionCantRemoveAllAdminsError:
		return &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	default:
		return fmt.Errorf("unexpected type(%T) of result", resp.RemoveTeamPermission)
	}
}

func (c *TeamsClient) IsUserInTeam(ctx context.Context, userId int, teamId string) (bool, error) {
	resp, err := schema.ListTeamPermissions(ctx, c.client)
	if err != nil {
		return false, err
	}

	for _, teamPermission := range resp.TeamPermissions {
		if teamPermission.Id == teamId {
			for _, user := range teamPermission.Team.Members {
				if user.UserId == userId {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

func (c *TeamsClient) AddUserToTeam(ctx context.Context, userId int, teamId string) error {
	resp, err := schema.AddMemberToTeam(ctx, c.client, userId, teamId)
	if err != nil {
		return err
	}

	switch respCast := resp.AddMemberToTeam.(type) {
	case *schema.AddMemberToTeamAddMemberToTeamAddMemberToTeamSuccess:
		return nil
	case *schema.AddMemberToTeamAddMemberToTeamPythonError:
		return &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	case *schema.AddMemberToTeamAddMemberToTeamUnauthorizedError:
		return &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	case *schema.AddMemberToTeamAddMemberToTeamUserLimitError:
		return &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	default:
		return fmt.Errorf("unexpected type(%T) of result", resp.AddMemberToTeam)
	}
}

func (c *TeamsClient) RemoveUserFromTeam(ctx context.Context, userId int, teamId string) error {
	resp, err := schema.RemoveMemberFromTeam(ctx, c.client, userId, teamId)
	if err != nil {
		return err
	}

	switch respCast := resp.RemoveMemberFromTeam.(type) {
	case *schema.RemoveMemberFromTeamRemoveMemberFromTeamRemoveMemberFromTeamSuccess:
		return nil
	case *schema.RemoveMemberFromTeamRemoveMemberFromTeamPythonError:
		return &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	case *schema.RemoveMemberFromTeamRemoveMemberFromTeamUnauthorizedError:
		return &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	default:
		return fmt.Errorf("unexpected type(%T) of result", resp.RemoveMemberFromTeam)
	}
}
