package datasources_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/datarootsio/terraform-provider-dagster/internal/testutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceConfigurationDocument(t *testing.T) {
	var json string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testutils.ProviderConfig + `
data "dagster_configuration_document" "this" {
  yaml_body = <<YAML
key: value
list_key:
  - value1
  - value2
YAML
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutils.FetchValueFromState("data.dagster_configuration_document.this", "json", &json),
					testJson(&json),
				),
			},
		},
	})
}

func testJson(jsonString *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		contents := testutils.UnmarshalJSONOrPanic([]byte(*jsonString))

		if contents["key"] != "value" {
			return fmt.Errorf("expected key=value in json")
		}

		if len(contents["list_key"].([]any)) != 2 {
			return fmt.Errorf("expected list_key to have 2 elements")
		}

		if contents["list_key"].([]any)[0] != "value1" || contents["list_key"].([]any)[1] != "value2" {
			return fmt.Errorf("expected list_key to have value1 and value2")
		}

		return nil
	}
}

func TestAccResourceInvalidConfigurationDocument(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testutils.ProviderConfig + `
data "dagster_configuration_document" "this" {
  yaml_body = ",,,,...."
}`,
				ExpectError: regexp.MustCompile(`Unable to parse YAML`),
			},
		},
	})
}
