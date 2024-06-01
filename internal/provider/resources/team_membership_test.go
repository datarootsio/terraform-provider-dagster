package resources_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/datarootsio/terraform-provider-dagster/internal/testutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccResourceTeamMembershipConfig(userEmail string, teamName string) string {
	return fmt.Sprintf(testutils.ProviderConfig+`
data "dagster_user" "this" {
    email = "%s"
}

resource "dagster_team" "this" {
  name = "%s"
}

resource "dagster_team_membership" "this" {
  user_id = data.dagster_user.this.id
  team_id = dagster_team.this.id
}

`, userEmail, teamName)
}

func TestAccResourceTeamMembershipBasic(t *testing.T) {
	teamName := "tar-team-basic/" + acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	var teamId string
	var userId string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccTeamMembershipDeleted(userId, teamId),
		Steps: []resource.TestStep{
			// Create team membership
			{
				Config: testAccResourceTeamMembershipConfig("test-user@dataroots.io", teamName),
				Check: resource.ComposeTestCheckFunc(
					testutils.FetchValueFromState("data.dagster_user.this", "id", &userId),
					testutils.FetchValueFromState("dagster_team.this", "id", &teamId),
					resource.TestCheckResourceAttrPtr("dagster_team_membership.this", "team_id", &teamId),
					resource.TestCheckResourceAttrPtr("dagster_team_membership.this", "user_id", &userId),
				),
			},
		},
	})
}

func testAccTeamMembershipDeleted(userId string, teamId string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		client := testutils.GetDagsterClientFromEnvVars()
		userIdStr, _ := strconv.Atoi(userId)
		inTeam, err := client.TeamsClient.IsUserInTeam(context.Background(), userIdStr, teamId)

		if inTeam || err != nil {
			return errors.New("Team membership not deleted or error")
		}
		return nil
	}
}
