package resources_test

import (
	"fmt"
	"testing"

	"github.com/datarootsio/terraform-provider-dagster/internal/testutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccResourceTeamConfig(teamName string) string {
	return fmt.Sprintf(testutils.ProviderConfig+`
resource "dagster_team" "test" {
	name = "%s"
}
`, teamName)
}

func TestAccResource_team_basic(t *testing.T) {
	teamName := "tar-team-basic/" + acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	renameTeamName := "tar-team-basic/" + acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	var teamId string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			//Create team
			{
				Config: testAccResourceTeamConfig(teamName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagster_team.test", "name", teamName),
					testutils.FetchValueFromState("dagster_team.test", "id", &teamId),
				),
			},
			//Rename team, should have same id as initial create
			{
				Config: testAccResourceTeamConfig(renameTeamName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagster_team.test", "name", renameTeamName),
					resource.TestCheckResourceAttrPtr("dagster_team.test", "id", &teamId),
				),
			},
		},
	})
}
