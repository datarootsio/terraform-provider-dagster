package resources_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	clientTypes "github.com/datarootsio/terraform-provider-dagster/internal/client/types"
	"github.com/datarootsio/terraform-provider-dagster/internal/testutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccResourceCodeLocationConfig(name string, image string, file string) string {
	return fmt.Sprintf(testutils.ProviderConfig+`
resource "dagster_code_location" "test" {
  name          = "%s"
  image         = "%s"
  code_source   = {
    python_file = "%s"
  }
}
`, name, image, file)
}

func TestAccResourceBasicCodeLocation(t *testing.T) {
	name := "code-location-" + acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	image := "python:3.13"
	file := "my_python.py"

	updatedImage := "python:3.12"
	updatedName := "code-location-as-document-update-" + acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testCodeLocationDeleted(updatedName),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceCodeLocationConfig(name, image, file),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCodeLocationProperties(name, image, file),
					resource.TestCheckResourceAttr("dagster_code_location.test", "name", name),
					resource.TestCheckResourceAttr("dagster_code_location.test", "image", image),
					resource.TestCheckResourceAttr("dagster_code_location.test", "code_source.python_file", file),
				),
			},
			{
				Config: testAccResourceCodeLocationConfig(name, updatedImage, file),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCodeLocationProperties(name, updatedImage, file),
					resource.TestCheckResourceAttr("dagster_code_location.test", "name", name),
					resource.TestCheckResourceAttr("dagster_code_location.test", "image", updatedImage),
					resource.TestCheckResourceAttr("dagster_code_location.test", "code_source.python_file", file),
				),
			},
			{
				Config: testAccResourceCodeLocationConfig(updatedName, image, file),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCodeLocationProperties(updatedName, image, file),
					resource.TestCheckResourceAttr("dagster_code_location.test", "name", updatedName),
					resource.TestCheckResourceAttr("dagster_code_location.test", "image", image),
					resource.TestCheckResourceAttr("dagster_code_location.test", "code_source.python_file", file),
				),
			},
		},
	})
}

func testCodeLocationProperties(name string, image string, file string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		client := testutils.GetDagsterClientFromEnvVars()
		codeLocation, err := client.CodeLocationsClient.GetCodeLocationByName(context.Background(), name)
		if err != nil {
			return err
		}
		if codeLocation.Image != image {
			return fmt.Errorf("expected image to be %s, got %s", image, codeLocation.Image)
		}
		if codeLocation.CodeSource.PythonFile != file {
			return fmt.Errorf("expected file to be %s, got %s", file, codeLocation.CodeSource.PythonFile)
		}
		return nil
	}
}

func testCodeLocationDeleted(name string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		client := testutils.GetDagsterClientFromEnvVars()
		_, err := client.CodeLocationsClient.GetCodeLocationByName(context.Background(), name)

		notFound := &clientTypes.ErrNotFound{}
		if errors.As(err, &notFound) {
			return nil
		}
		return err
	}
}
