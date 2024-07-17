package datasources_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/datarootsio/terraform-provider-dagster/internal/testutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccTeamConfig(name string) string {
	return fmt.Sprintf(testutils.ProviderConfig+`
data "dagster_team" "this" {
    name = "%s"
}
`, name)
}

func TestAccTeam(t *testing.T) {
	name := "test-team"
	var teamId string
	var teamName string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTeamConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testutils.FetchValueFromState("data.dagster_team.this", "id", &teamId),
					testutils.FetchValueFromState("data.dagster_team.this", "name", &teamName),
					testTeamProperties(&teamName, &teamId),
				),
			},
		},
	})
}

func testTeamProperties(name *string, id *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		client := testutils.GetDagsterClientFromEnvVars()

		team, err := client.TeamsClient.GetTeamByName(context.Background(), *name)
		if err != nil {
			return err
		}

		if team.Name != *name {
			return fmt.Errorf("expected team name to be %s, got %s", *name, team.Name)
		}

		if team.Id != *id {
			return fmt.Errorf("expected team id to be %s, got %s", *id, team.Id)
		}

		return nil
	}
}
