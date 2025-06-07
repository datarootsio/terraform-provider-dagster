package datasources_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/datarootsio/terraform-provider-dagster/internal/testutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testVersionConfig() string {
	return testutils.ProviderConfig + `
data "dagster_version" "this" {
}
`
}

func TestAccVersion(t *testing.T) {
	var version string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testVersionConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testutils.FetchValueFromState("data.dagster_version.this", "version", &version),
					testVersion(&version),
				),
			},
		},
	})
}

func testVersion(version *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		re := `^[a-zA-Z0-9]{8}-[a-zA-Z0-9]{8}$`
		if regexp.MustCompile(re).MatchString(*version) {
			return nil
		}
		return fmt.Errorf("expected version to match %s, got %s", re, *version)
	}
}
