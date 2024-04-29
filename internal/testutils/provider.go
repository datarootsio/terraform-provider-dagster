package testutils

import (
	"os"
	"testing"

	dagsterProvider "github.com/datarootsio/terraform-provider-dagster/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var TestAccProvider provider.Provider = dagsterProvider.New()

// TestAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var TestAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"dagster": providerserver.NewProtocol6WithError(TestAccProvider),
}

// AccTestPreCheck is a utility hook, which every test suite will call
// in order to verify if the necessary provider configurations are passed
// through the environment variables.
// https://developer.hashicorp.com/terraform/plugin/testing/acceptance-tests/testcase#precheck
func AccTestPreCheck(t *testing.T) {
	t.Helper()

	// Use TF_VARs so that we don't clash with possible
	// real dagster cloud CLI env vars that have been set
	requiredEnvVars := []string{
		"TF_VAR_tf_provider_testing_organization",
		"TF_VAR_tf_provider_testing_deployment",
		"TF_VAR_tf_provider_testing_api_token",
	}

	for _, envVar := range requiredEnvVars {
		if v := os.Getenv(envVar); v == "" {
			t.Fatalf("Env var %s must be set for tf acc tests.", envVar)
		}
	}
}

const ProviderConfig = `
variable "tf_provider_testing_organization" {}
variable "tf_provider_testing_deployment" {}
variable "tf_provider_testing_api_token" {}

provider "dagster" {
	organization = var.tf_provider_testing_organization
	deployment   = var.tf_provider_testing_deployment
	api_token    = var.tf_provider_testing_api_token
}
`
