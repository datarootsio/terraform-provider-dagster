package datasources_test

import (
	"testing"

	"github.com/datarootsio/terraform-provider-dagster/internal/testutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testOrganizationConfig defines the Terraform configuration for the organization data source.
func testOrganizationConfig() string {
	return testutils.ProviderConfig + `
data "dagster_organization" "this" {}`
}

// TestAccOrganization performs acceptance tests for the dagster_organization data source.
func TestAccOrganization(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testOrganizationConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check that the data source exists in the state
					resource.TestCheckResourceAttrSet("data.dagster_organization.this", "id"),
					resource.TestCheckResourceAttrSet("data.dagster_organization.this", "public_id"),
					resource.TestCheckResourceAttrSet("data.dagster_organization.this", "name"),
					resource.TestCheckResourceAttrSet("data.dagster_organization.this", "status"),

					// Verify that organization name and status is correct
					resource.TestCheckResourceAttr("data.dagster_organization.this", "public_id", "dataroots-terraform-provider-dagster"),
					resource.TestCheckResourceAttr("data.dagster_organization.this", "name", "dataroots-terraform-provider-dagster"), // Replace with expected name
					resource.TestCheckResourceAttr("data.dagster_organization.this", "status", "ACTIVE"),
				),
			},
		},
	})
}
