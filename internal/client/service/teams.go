package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Khan/genqlient/graphql"
	"github.com/datarootsio/terraform-provider-dagster/internal/client/schema"
	"github.com/datarootsio/terraform-provider-dagster/internal/client/types"
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

	return schema.Team{}, &types.ErrNotFound{What: "Team", Value: name}
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

	return schema.Team{}, &types.ErrNotFound{What: "Team", Value: id}
}

func (c *TeamsClient) CreateTeam(ctx context.Context, name string) (schema.Team, error) {
	_, err := c.GetTeamByName(ctx, name)

	if err == nil {
		return schema.Team{}, &types.ErrAlreadyExists{What: "Team", Value: name}
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

func (c *TeamsClient) DeleteTeam(ctx context.Context, id string) (string, error) {
	_, err := c.GetTeamById(ctx, id)
	if err != nil {
		return "", err
	}

	resp, err := schema.DeleteTeam(ctx, c.client, id)
	if err != nil {
		return "", err
	}

	switch respCast := resp.DeleteTeam.(type) {
	case *schema.DeleteTeamDeleteTeamDeleteTeamSuccess:
		return respCast.TeamId, nil
	case *schema.DeleteTeamDeleteTeamPythonError:
		return "", &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	case *schema.DeleteTeamDeleteTeamUnauthorizedError:
		return "", &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	default:
		return "", fmt.Errorf("unexpected type(%T) of result", resp.DeleteTeam)
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

func (c *TeamsClient) GetTeamDeploymentGrant(ctx context.Context, teamId string, deploymentId int) (schema.ScopedPermissionGrant, error) {
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

	return schema.ScopedPermissionGrant{}, &types.ErrNotFound{What: "DeploymentGrant", Value: teamId}
}

func (c *TeamsClient) CreateOrUpdateTeamDeploymentGrant(ctx context.Context, teamId string, deploymentId int, grant schema.PermissionGrant) (schema.ScopedPermissionGrant, error) {
	// TODO: check if team exists and check if deployment exists => to return specific errors
	existingPermissionGrant, err := c.GetTeamDeploymentGrant(ctx, teamId, deploymentId)

	locationGrants := make([]schema.LocationScopedGrantInput, 0)

	var errComp *types.ErrNotFound
	if errors.As(err, &errComp) {
		// TeamDeploymentGrant does not exist, initialize empty list
		// Do nothing
	} else if err != nil {
		// error fetching GetTeamDeploymentGrant, return error
		return schema.ScopedPermissionGrant{}, err
	} else {
		// exists, transform existing LocationScopedGrant into LocationScopedGrantInput
		for _, locationGrant := range existingPermissionGrant.LocationGrants {
			locationGrants = append(
				locationGrants,
				schema.LocationScopedGrantInput{
					LocationName: locationGrant.LocationName,
					Grant:        locationGrant.Grant,
				},
			)
		}
	}

	resp, err := schema.CreateOrUpdateTeamPermission(
		ctx,
		c.client,
		deploymentId,
		schema.PermissionDeploymentScopeDeployment,
		grant,
		locationGrants,
		teamId,
	)
	if err != nil {
		return schema.ScopedPermissionGrant{}, err
	}

	// At this point the DeploymentGrant should exist so fetch the last state from the API
	updatedPermissionGrant, err := c.GetTeamDeploymentGrant(ctx, teamId, deploymentId)
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
	// TODO: check if team exists and check if deployment exists => to return specific errors
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
