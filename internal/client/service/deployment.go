package service

import (
	"context"

	"github.com/Khan/genqlient/graphql"
	"github.com/datarootsio/terraform-provider-dagster/internal/client/schema"
)

type DeploymentClient struct {
	client graphql.Client
}

func NewDeploymentClient(client graphql.Client) DeploymentClient {
	return DeploymentClient{
		client: client,
	}
}

func (c DeploymentClient) GetCurrentDeployment(ctx context.Context) (schema.Deployment, error) {
	resp, err := schema.GetCurrentDeployment(ctx, c.client)
	if err != nil {
		return schema.Deployment{}, err
	}

	return resp.CurrentDeployment.Deployment, nil
}

func (c DeploymentClient) GetAllDeployments(ctx context.Context) ([]schema.Deployment, error) {
	resp, err := schema.GetAllDeployments(ctx, c.client)
	if err != nil {
		return []schema.Deployment{}, err
	}

	deployments := make([]schema.Deployment, 0, len(resp.Deployments))
	for _, deployment := range resp.Deployments {
		deployments = append(deployments, deployment.Deployment)
	}

	return deployments, nil
}
