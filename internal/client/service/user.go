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

func (c UsersClient) GetUserByEmail(ctx context.Context, email string) (schema.GetUsersUsersOrErrorDagsterCloudUsersWithScopedPermissionGrantsUsersDagsterCloudUserWithScopedPermissionGrantsUserDagsterCloudUser, error) {
	result, err := schema.GetUsers(ctx, c.client)
	if err != nil {
		return schema.GetUsersUsersOrErrorDagsterCloudUsersWithScopedPermissionGrantsUsersDagsterCloudUserWithScopedPermissionGrantsUserDagsterCloudUser{}, err
	}

	users := result.UsersOrError.(*schema.GetUsersUsersOrErrorDagsterCloudUsersWithScopedPermissionGrants).Users
	for _, user := range users {
		if user.User.Email == email {
			return user.User, nil
		}
	}

	return schema.GetUsersUsersOrErrorDagsterCloudUsersWithScopedPermissionGrantsUsersDagsterCloudUserWithScopedPermissionGrantsUserDagsterCloudUser{}, errors.New("no user with email " + email + " found")
}
