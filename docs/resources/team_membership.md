---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "dagster_team_membership Resource - dagster"
subcategory: ""
description: |-
  Adds a Dagster user to a team.
---

# dagster_team_membership (Resource)

Adds a Dagster user to a team.

## Example Usage

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `team_id` (String) Team id
- `user_id` (Number) User id