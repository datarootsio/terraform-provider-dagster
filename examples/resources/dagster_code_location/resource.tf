resource "dagster_code_location" "example" {
  name  = "code_location_example"
  image = "python:3.13"
  code_source = {
    python_file = "my_python_file.py"
  }
}
