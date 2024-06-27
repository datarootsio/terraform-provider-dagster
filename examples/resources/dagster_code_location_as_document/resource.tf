resource "dagster_code_location_from_document" "example" {
  document = data.dagster_configuration_document.example.json
}

data "dagster_configuration_document" "example" {
  yaml_body = <<YAML
location_name: "example_code_location_from_document"
code_source:
  python_file: "a_python_file.py"
YAML
}
