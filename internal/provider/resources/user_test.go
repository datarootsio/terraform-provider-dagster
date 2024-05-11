package resources_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	clientTypes "github.com/datarootsio/terraform-provider-dagster/internal/client/types"
	"github.com/datarootsio/terraform-provider-dagster/internal/testutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccResourceUserConfig(email string) string {
	return fmt.Sprintf(testutils.ProviderConfig+`
resource "dagster_user" "test" {
  remove_default_permissions = true
	email                      = "%s"
}`, email)
}

func TestAccResource_user_basic(t *testing.T) {
	userEmail := "acc-test-user@dataroots.io"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccUserDeleted(userEmail),
		Steps: []resource.TestStep{
			// Create user
			{
				Config: testAccResourceUserConfig(userEmail),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagster_user.test", "email", userEmail),
					testAccUserExists(userEmail),
				),
			},
		},
	})
}

func testAccUserExists(email string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		client := testutils.GetDagsterClientFromEnvVars()
		_, err := client.UsersClient.GetUserByEmail(context.Background(), email)
		return err
	}
}

func testAccUserDeleted(email string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		client := testutils.GetDagsterClientFromEnvVars()
		_, err := client.UsersClient.GetUserByEmail(context.Background(), email)

		notFound := &clientTypes.ErrNotFound{}
		if errors.As(err, &notFound) {
			return nil
		}
		return err
	}
}
