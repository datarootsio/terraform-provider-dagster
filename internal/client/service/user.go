package service

import (
	"context"
	"errors"

	"github.com/Khan/genqlient/graphql"
	"github.com/datarootsio/terraform-provider-dagster/internal/client/schema"
)

type UsersClient struct {
	client graphql.Client
}

func NewUsersClient(client graphql.Client) UsersClient {
	return UsersClient{
		client: client,
	}
}

func (c UsersClient) GetUserByEmail(ctx context.Context, email string) (schema.User, error) {
	result, err := schema.GetUsers(ctx, c.client)
	if err != nil {
		return schema.User{}, err
	}

	users := result.UsersOrError.(*schema.UsersOrErrorDagsterCloudUsersWithScopedPermissionGrants).Users
	for _, user := range users {
		if user.User.Email == email {
			return user.User, nil
		}
	}

	return schema.User{}, errors.New("no user with email " + email + " found")
}

func (c UsersClient) AddUser(ctx context.Context, email string) error {
	resp, err := schema.AddUser(ctx, c.client, email)
	if err != nil {
		return err
	}

	response := (*resp).AddUserToOrganization
	response_name := response.GetTypename()
	var error_msg string

	if response_name == "AddUserToOrganizationSuccess" {
		return nil
	} else if response_name == "PythonError" {
		error_msg = response.(*schema.AddUserAddUserToOrganizationPythonError).Message
	} else if response_name == "UnauthorizedError" {
		error_msg = response.(*schema.AddUserAddUserToOrganizationUnauthorizedError).Message
	} else if response_name == "UserLimitError" {
		error_msg = response.(*schema.AddUserAddUserToOrganizationUserLimitError).Message
	}

	return errors.New(error_msg)
}
