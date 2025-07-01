package datasources_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/datarootsio/terraform-provider-dagster/internal/testutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccUsersConfig(regexFilter string) string {
	return fmt.Sprintf(testutils.ProviderConfig+`
data "dagster_users" "this" {
    email_regex = "%s"
}
`, regexFilter)
}

func TestAccUsers(t *testing.T) {
	regexFilter := "test-user"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUsersConfig(regexFilter),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dagster_users.this", "users.#", "1"), // '.#' means 'length' see: https://developer.hashicorp.com/terraform/plugin/sdkv2/testing/acceptance-tests/teststep#builtin-check-functions
					resource.TestMatchResourceAttr("data.dagster_users.this", "users.0.email", regexp.MustCompile("test-user")),
				),
			},
		},
	})
}
