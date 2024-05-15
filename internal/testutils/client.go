package testutils

import (
	"os"

	"github.com/datarootsio/terraform-provider-dagster/internal/client"
)

func GetDagsterClientFromEnvVars() client.DagsterClient {
	organization := os.Getenv("TF_VAR_testing_dagster_organization")
	deployment := os.Getenv("TF_VAR_testing_dagster_deployment")
	apiToken := os.Getenv("TF_VAR_testing_dagster_api_token")

	client, err := client.NewDagsterClient(organization, deployment, apiToken)
	if err != nil {
		panic(err)
	}

	return client
}
