package client

import (
	"fmt"

	"github.com/Khan/genqlient/graphql"
	"github.com/datarootsio/terraform-provider-dagster/internal/client/service"
)

type DagsterClient struct {
	DeploymentClient service.DeploymentClient
}

func NewDagsterClient(organization, deployment, apiToken string) (DagsterClient, error) {
	url := fmt.Sprintf(
		"https://%s.dagster.cloud/%s/graphql",
		organization,
		deployment,
	)

	gqlClient := graphql.NewClient(url, &AuthDoer{
		APIToken: apiToken,
	})

	return DagsterClient{
		DeploymentClient: service.NewDeploymentClient(gqlClient),
	}, nil
}