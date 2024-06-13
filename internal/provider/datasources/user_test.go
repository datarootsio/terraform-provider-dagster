package datasources_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/datarootsio/terraform-provider-dagster/internal/testutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccUserConfig(email string) string {
	return fmt.Sprintf(testutils.ProviderConfig+`
data "dagster_user" "this" {
    email = "%s"
}
`, email)
}

func TestAccUser(t *testing.T) {
	email := "test-user@dataroots.io"
	var userId string
	var userName string
	var userEmail string
	var userPicture string
	var userIsScimProvisioned string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig(email),
				Check: resource.ComposeAggregateTestCheckFunc(
					testutils.FetchValueFromState("data.dagster_user.this", "id", &userId),
					testutils.FetchValueFromState("data.dagster_user.this", "name", &userName),
					testutils.FetchValueFromState("data.dagster_user.this", "email", &userEmail),
					testutils.FetchValueFromState("data.dagster_user.this", "picture", &userPicture),
					testutils.FetchValueFromState("data.dagster_user.this", "is_scim_provisioned", &userIsScimProvisioned),
					testUserProperties(email, &userId, &userName, &userEmail, &userPicture, &userIsScimProvisioned),
				),
			},
		},
	})
}

func testUserProperties(inEmail string, id *string, name *string, email *string, picture *string, isScimProvisioned *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		client := testutils.GetDagsterClientFromEnvVars()

		user, err := client.UsersClient.GetUserByEmail(context.Background(), inEmail)
		if err != nil {
			return err
		}

		if user.Email != *email {
			return fmt.Errorf("expected user email to be %s, got %s", *email, user.Email)
		}

		if user.Name != *name {
			return fmt.Errorf("expected user name to be %s, got %s", *name, user.Name)
		}

		intValue, err := strconv.Atoi(*id)
		if err != nil {
			return err
		}
		if user.UserId != intValue {
			return fmt.Errorf("expected user id to be %v, got %v", *id, user.UserId)
		}

		if user.Picture != *picture {
			return fmt.Errorf("expected user picture to be %s, got %s", *picture, user.Picture)
		}

		boolValue, err := strconv.ParseBool(*isScimProvisioned)
		if err != nil {
			return err
		}
		if user.IsScimProvisioned != boolValue {
			return fmt.Errorf("expected user is_scim_provisioned to be %v, got %v", *isScimProvisioned, user.IsScimProvisioned)
		}
		return nil
	}
}
