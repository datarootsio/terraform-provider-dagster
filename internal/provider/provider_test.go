package provider_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/datarootsio/terraform-provider-dagster/internal/testutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProvider_configViaEnvVars(t *testing.T) {
	t.Setenv("DAGSTER_CLOUD_ORGANIZATION", os.Getenv("TF_VAR_testing_dagster_organization"))
	t.Setenv("DAGSTER_CLOUD_DEPLOYMENT", os.Getenv("TF_VAR_testing_dagster_deployment"))
	t.Setenv("DAGSTER_CLOUD_API_TOKEN", os.Getenv("TF_VAR_testing_dagster_api_token"))

	teamName := "tap-env-var-conf/" + acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					provider "dagster" {}

					resource "dagster_team" "test" {
						name = "%s"
					}
				`, teamName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagster_team.test", "name", teamName),
				),
			},
		},
	})
}
