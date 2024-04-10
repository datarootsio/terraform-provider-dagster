---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "dagster_team_deployment_grant Resource - dagster"
subcategory: ""
description: |-
  Team Deployment Grant resource
---

# dagster_team_deployment_grant (Resource)

Team Deployment Grant resource

## Example Usage

```terraform
data "dagster_current_deployment" "current" {}

resource "dagster_team" "example" {
  name = "example_team"
}

resource "dagster_team_deployment_grant" "example" {
  deployment_id = data.dagster_current_deployment.current.id
  team_id       = dagster_team.example.id

  grant = "VIEWER" # One of ["VIEWER" "LAUNCHER" "EDITOR" "ADMIN" ]

  code_location_grants = [
    {
      name  = "example_code_location"
      grant = "LAUNCHER" # One of ["LAUNCHER" "EDITOR" "ADMIN"]
    },
  ]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `deployment_id` (Number) Team Deployment Grant DeploymentId
- `grant` (String) Team Deployment Grant Grant
- `team_id` (String) Team Deployment Grant TeamId

### Optional

- `code_location_grants` (Attributes Set) (see [below for nested schema](#nestedatt--code_location_grants))

### Read-Only

- `id` (Number) Team Deployment Grant Id

<a id="nestedatt--code_location_grants"></a>
### Nested Schema for `code_location_grants`

Required:

- `grant` (String) Code location Grant
- `name` (String) Code location Name