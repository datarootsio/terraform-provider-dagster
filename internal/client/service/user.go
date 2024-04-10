package service

import (
	"context"
	"fmt"

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
