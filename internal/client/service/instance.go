package service

import (
	"context"

	"github.com/Khan/genqlient/graphql"
	"github.com/datarootsio/terraform-provider-dagster/internal/client/schema"
)

type InstanceClient struct {
	client graphql.Client
}

func NewInstanceClient(client graphql.Client) InstanceClient {
	return InstanceClient{
		client: client,
	}
}

func (c *InstanceClient) GetDagsterCloudVersion(ctx context.Context) (string, error) {
	resp, err := schema.GetDagsterCloudVersion(ctx, c.client)
	if err != nil {
		return "", err
	}

	return resp.Version, nil
}
