data "dagster_current_deployment" "current" {}

resource "dagster_team" "example" {
  name = "example_team"
}

resource "dagster_team_deployment_grant" "example" {
  deployment_id = data.dagster_current_deployment.current.id
  team_id       = dagster_team.example.id

  grant = "VIEWER" # One of ["LAUNCHER" "EDITOR" "ADMIN" "AGENT" "VIEWER"]
}
