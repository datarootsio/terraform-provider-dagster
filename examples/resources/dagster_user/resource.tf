resource "dagster_user" "test" {
  email                      = "foo.bar@dataroots.io"
  remove_default_permissions = true
}
