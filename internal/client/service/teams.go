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

	return schema.Team{}, &types.ErrTeamNotFound{Name: name}
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

	return schema.Team{}, &types.ErrTeamNotFound{Id: id}
}

func (c *TeamsClient) CreateTeam(ctx context.Context, name string) (schema.Team, error) {
	_, err := c.GetTeamByName(ctx, name)

	if err == nil {
		return schema.Team{}, &types.ErrTeamAlreadyExists{Name: name}
	}

	var errComp *types.ErrTeamNotFound
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
