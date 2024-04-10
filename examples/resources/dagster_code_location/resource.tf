resource "dagster_code_location" "rbac" {
  name  = "rbac_code_location"
  image = "python:3.13"
  code_source = {
    python_file = "my_python_file.py"
  }
}
