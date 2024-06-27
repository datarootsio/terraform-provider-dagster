package resources_test

import (
	"fmt"
	"testing"

	"github.com/datarootsio/terraform-provider-dagster/internal/testutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccResourceCodeLocationFromDocumentConfig(name string, image string, file string) string {
	return fmt.Sprintf(testutils.ProviderConfig+`
resource "dagster_code_location_from_document" "test" {
  document = data.dagster_configuration_document.test.json
}

data "dagster_configuration_document" "test" {
  yaml_body = <<YAML
location_name: "%s"
image: "%s"
code_source:
  python_file: "%s"
YAML
}
`, name, image, file)
}

func TestAccResourceBasicCodeLocationFromDocument(t *testing.T) {
	name := "code-location-as-document-" + acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	image := "python:3.13"
	file := "as_doc/my_python.py"

	updatedImage := "python:3.12"
	updatedName := "code-location-as-document-update-" + acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testCodeLocationDeleted(updatedName),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceCodeLocationFromDocumentConfig(name, image, file),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCodeLocationProperties(name, image, file),
					resource.TestCheckResourceAttr("dagster_code_location_from_document.test", "name", name),
				),
			},
			{
				Config: testAccResourceCodeLocationFromDocumentConfig(name, updatedImage, file),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCodeLocationProperties(name, updatedImage, file),
					resource.TestCheckResourceAttr("dagster_code_location_from_document.test", "name", name),
				),
			},
			{
				Config: testAccResourceCodeLocationFromDocumentConfig(updatedName, image, file),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCodeLocationProperties(updatedName, image, file),
					resource.TestCheckResourceAttr("dagster_code_location_from_document.test", "name", updatedName),
				),
			},
		},
	})
}
