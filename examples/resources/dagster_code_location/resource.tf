resource "dagster_code_location" "example" {
  name  = "example_code_location"
  image = "python:3.13"
  code_source = {
    python_file = "my_python.py"
  }
}
