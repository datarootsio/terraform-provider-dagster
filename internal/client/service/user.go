package service

import (
	"context"

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

	return schema.User{}, &types.ErrNotFound{What: "User", Value: email}
}

// AddUser adds a user (identified by an email address) and returns the new user
func (c UsersClient) AddUser(ctx context.Context, email string) (schema.User, error) {
	resp, err := schema.AddUser(ctx, c.client, email)
	if err != nil {
		return schema.User{}, err
	}

	response := (*resp).AddUserToOrganization
	responseName := response.GetTypename()

	var errorMsg string

	if responseName == "AddUserToOrganizationSuccess" {
		user := response.(*schema.AddUserAddUserToOrganizationAddUserToOrganizationSuccess).UserWithGrants.User.User
		return user, nil
	} else if responseName == "PythonError" {
		errorMsg = response.(*schema.AddUserAddUserToOrganizationPythonError).Message
	} else if responseName == "UnauthorizedError" {
		errorMsg = response.(*schema.AddUserAddUserToOrganizationUnauthorizedError).Message
	} else if responseName == "UserLimitError" {
		errorMsg = response.(*schema.AddUserAddUserToOrganizationUserLimitError).Message
	}

	return schema.User{}, &types.ErrApi{Typename: responseName, Message: errorMsg}
}

// RemoveUser removes a user from the organization and returns the email of the user
func (c UsersClient) RemoveUser(ctx context.Context, email string) (string, error) {
	resp, err := schema.RemoveUser(ctx, c.client, email)
	if err != nil {
		return "", err
	}

	response := (*resp).RemoveUserFromOrganization
	responseName := response.GetTypename()
	var errorMsg string

	if responseName == "RemoveUserFromOrganizationSuccess" {
		user := response.(*schema.RemoveUserRemoveUserFromOrganizationRemoveUserFromOrganizationSuccess).Email
		return user, nil
	} else if responseName == "PythonError" {
		errorMsg = response.(*schema.RemoveUserRemoveUserFromOrganizationPythonError).Message
	} else if responseName == "UnauthorizedError" {
		errorMsg = response.(*schema.RemoveUserRemoveUserFromOrganizationUnauthorizedError).Message
	} else if responseName == "CantRemoveAllAdminsError" {
		errorMsg = response.(*schema.RemoveUserRemoveUserFromOrganizationCantRemoveAllAdminsError).Message
	}

	return "", &types.ErrApi{Typename: responseName, Message: errorMsg}
}
