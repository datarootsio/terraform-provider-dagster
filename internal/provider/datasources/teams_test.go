package datasources_test

import (
	"fmt"
	"testing"

	"github.com/datarootsio/terraform-provider-dagster/internal/testutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccTeamsConfig(regexFilter string) string {
	return fmt.Sprintf(testutils.ProviderConfig+`
data "dagster_teams" "this" {
    regex_filter = "%s"
}
`, regexFilter)
}

func TestAccTeams(t *testing.T) {
	regexFilter := "^test-team"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTeamsConfig(regexFilter),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dagster_teams.this", "teams.#", "2"), // '.#' means 'length' see: https://developer.hashicorp.com/terraform/plugin/sdkv2/testing/acceptance-tests/teststep#builtin-check-functions
					resource.TestCheckResourceAttr("data.dagster_teams.this", "teams.0.name", "test-team"),
					resource.TestCheckResourceAttr("data.dagster_teams.this", "teams.1.name", "test-team-2"),
				),
			},
		},
	})
}
