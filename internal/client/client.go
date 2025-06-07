package client

import (
	"fmt"

	"github.com/Khan/genqlient/graphql"
	"github.com/datarootsio/terraform-provider-dagster/internal/client/service"
)

type DagsterClient struct {
	Organization string
	Deployment   string

	DeploymentClient    service.DeploymentClient
	UsersClient         service.UsersClient
	TeamsClient         service.TeamsClient
	CodeLocationsClient service.CodeLocationsClient
	InstanceClient      service.InstanceClient
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
		Organization: organization,
		Deployment:   deployment,

		DeploymentClient:    service.NewDeploymentClient(gqlClient),
		UsersClient:         service.NewUsersClient(gqlClient),
		TeamsClient:         service.NewTeamsClient(gqlClient),
		CodeLocationsClient: service.NewCodeLocationsClient(gqlClient),
		InstanceClient:      service.NewInstanceClient(gqlClient),
	}, nil
}
