package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Khan/genqlient/graphql"
	"github.com/datarootsio/terraform-provider-dagster/internal/client/schema"
	"github.com/datarootsio/terraform-provider-dagster/internal/client/types"
)

type UsersClient struct {
	client graphql.Client
}

func NewUsersClient(client graphql.Client) UsersClient {
	return UsersClient{
		client: client,
	}
}

// GetUserByEmail looks up a user by email address and returns it
func (c UsersClient) GetUserByEmail(ctx context.Context, email string) (schema.User, error) {
	result, err := schema.GetUsers(ctx, c.client)
	if err != nil {
		return schema.User{}, err
	}

	users := result.UsersOrError.(*schema.GetUsersUsersOrErrorDagsterCloudUsersWithScopedPermissionGrants).Users
	for _, user := range users {
		if user.User.Email == email {
			return user.User.User, nil
		}
	}

	return schema.User{}, &types.ErrNotFound{What: "User", Key: "email", Value: email}
}

// GetUserById looks up a user by id and returns it
func (c UsersClient) GetUserById(ctx context.Context, id int64) (schema.User, error) {
	result, err := schema.GetUsers(ctx, c.client)
	if err != nil {
		return schema.User{}, err
	}

	users := result.UsersOrError.(*schema.GetUsersUsersOrErrorDagsterCloudUsersWithScopedPermissionGrants).Users
	for _, user := range users {
		if int64(user.User.UserId) == id {
			return user.User.User, nil
		}
	}

	return schema.User{}, &types.ErrNotFound{What: "User", Key: "id", Value: strconv.FormatInt(id, 10)}
}

// AddUser adds a user (identified by an email address) and returns the new user
func (c UsersClient) AddUser(ctx context.Context, email string) (schema.User, error) {
	resp, err := schema.AddUser(ctx, c.client, email)
	if err != nil {
		return schema.User{}, err
	}

	switch respCast := resp.AddUserToOrganization.(type) {
	case *schema.AddUserAddUserToOrganizationAddUserToOrganizationSuccess:
		return respCast.UserWithGrants.User.User, nil
	case *schema.AddUserAddUserToOrganizationPythonError:
		return schema.User{}, &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	case *schema.AddUserAddUserToOrganizationUnauthorizedError:
		return schema.User{}, &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	case *schema.AddUserAddUserToOrganizationUserLimitError:
		return schema.User{}, &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	default:
		return schema.User{}, fmt.Errorf("unexpected type(%T) of result", resp.AddUserToOrganization)
	}
}

// RemoveUser removes a user from the organization and returns the email of the user
func (c UsersClient) RemoveUser(ctx context.Context, email string) error {
	resp, err := schema.RemoveUser(ctx, c.client, email)
	if err != nil {
		return err
	}

	switch respCast := resp.RemoveUserFromOrganization.(type) {
	case *schema.RemoveUserRemoveUserFromOrganizationRemoveUserFromOrganizationSuccess:
		return nil
	case *schema.RemoveUserRemoveUserFromOrganizationPythonError:
		return &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	case *schema.RemoveUserRemoveUserFromOrganizationUnauthorizedError:
		return &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	case *schema.RemoveUserRemoveUserFromOrganizationCantRemoveAllAdminsError:
		return &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	default:
		return fmt.Errorf("unexpected type(%T) of result", resp.RemoveUserFromOrganization)
	}
}

func (c UsersClient) AddUserToTeam(ctx context.Context, userId int, teamId string) error {
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

func (c UsersClient) RemoveUserFromTeam(ctx context.Context, userId int, teamId string) error {
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
