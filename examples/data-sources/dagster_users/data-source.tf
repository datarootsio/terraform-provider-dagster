data "dagster_users" "users" {
  email_regex = "^[^@]+@[^@]+\\.io$"
}
