resource "dagster_user" "example" {
  email = "foo.bar@dataroots.io"
}

resource "dagster_team" "example" {
  name = "example_team"
}

resource "dagster_team_membership" "example" {
  user_id = dagster_user.example.id
  team_id = dagster_team.example.id
}
