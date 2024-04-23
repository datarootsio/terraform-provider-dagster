package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Khan/genqlient/graphql"
	"github.com/datarootsio/terraform-provider-dagster/internal/client/schema"
	"github.com/datarootsio/terraform-provider-dagster/internal/client/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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

func (c DeploymentClient) GetDeploymentByName(ctx context.Context, name string) (schema.Deployment, error) {
	deployments, err := c.GetAllDeployments(ctx)
	if err != nil {
		return schema.Deployment{}, err
	}

	for _, deploy := range deployments {
		if deploy.DeploymentName == name {
			return deploy, nil
		}
	}

	return schema.Deployment{}, &types.ErrNotFound{What: "deployment", Key: "name", Value: name}
}

func (c DeploymentClient) GetDeploymentById(ctx context.Context, id int) (schema.Deployment, error) {
	deployments, err := c.GetAllDeployments(ctx)
	if err != nil {
		return schema.Deployment{}, err
	}

	for _, deploy := range deployments {
		if deploy.DeploymentId == id {
			return deploy, nil
		}
	}

	return schema.Deployment{}, &types.ErrNotFound{What: "deployment", Key: "name", Value: strconv.Itoa(id)}
}

func (c DeploymentClient) CreateHybridDeployment(ctx context.Context, name string) (schema.Deployment, error) {
	resp, err := schema.CreateHybridDeployment(ctx, c.client, name)
	if err != nil {
		return schema.Deployment{}, fmt.Errorf("Unable to create deployment %s: %w", name, err)
	}

	switch respCast := resp.CreateDeployment.(type) {
	case *schema.CreateHybridDeploymentCreateDeploymentDagsterCloudDeployment:
		return respCast.Deployment, nil
	case *schema.CreateHybridDeploymentCreateDeploymentDeploymentLimitError:
		return schema.Deployment{}, &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	case *schema.CreateHybridDeploymentCreateDeploymentDeploymentNotFoundError:
		// TODO return ErrNotFound, but how can a create trigger this?
		return schema.Deployment{}, &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	case *schema.CreateHybridDeploymentCreateDeploymentPythonError:
		return schema.Deployment{}, &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	case *schema.CreateHybridDeploymentCreateDeploymentDuplicateDeploymentError:
		return schema.Deployment{}, &types.ErrAlreadyExists{What: "deployment", Key: "name", Value: name}
	case *schema.CreateHybridDeploymentCreateDeploymentUnauthorizedError:
		return schema.Deployment{}, &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	default:
		return schema.Deployment{}, fmt.Errorf("unexpected type(%T) of result", resp.CreateDeployment)
	}
}

func (c DeploymentClient) DeleteDeployment(ctx context.Context, id int) error {
	resp, err := schema.DeleteDeployment(ctx, c.client, id)
	if err != nil {
		return fmt.Errorf("Unable to delete deployment %v: %w", id, err)
	}

	switch respCast := resp.DeleteDeployment.(type) {
	case *schema.DeleteDeploymentDeleteDeploymentDagsterCloudDeployment:
		return nil
	case *schema.DeleteDeploymentDeleteDeploymentDeleteFinalDeploymentError:
		return &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	case *schema.DeleteDeploymentDeleteDeploymentDeploymentNotFoundError:
		return &types.ErrNotFound{What: "deployment", Key: "id", Value: strconv.Itoa(id)}
	case *schema.DeleteDeploymentDeleteDeploymentPythonError:
		return &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	case *schema.DeleteDeploymentDeleteDeploymentUnauthorizedError:
		return &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	default:
		return fmt.Errorf("unexpected type(%T) of result", resp.DeleteDeployment)
	}
}

func (c DeploymentClient) SetDeploymentSettings(ctx context.Context, deploymentId int, settings json.RawMessage) (json.RawMessage, error) {
	settingsInput := schema.DeploymentSettingsInput{
		Settings: settings,
	}
	resp, err := schema.SetDeploymentSettings(ctx, c.client, deploymentId, settingsInput)
	if err != nil {
		tflog.Trace(ctx, fmt.Sprintf("Unable to set deployment settings: %v", err.Error()))
		return nil, fmt.Errorf("Unable to set deployment settings: %w", err)
	}

	switch respCast := resp.SetDeploymentSettings.(type) {
	case *schema.SetDeploymentSettingsSetDeploymentSettings:
		return respCast.Settings, nil
	case *schema.SetDeploymentSettingsSetDeploymentSettingsDeleteFinalDeploymentError:
		return nil, &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	case *schema.SetDeploymentSettingsSetDeploymentSettingsDeploymentNotFoundError:
		return nil, &types.ErrNotFound{What: "deployment", Key: "id", Value: strconv.Itoa(deploymentId)}
	case *schema.SetDeploymentSettingsSetDeploymentSettingsDuplicateDeploymentError:
		return nil, &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	case *schema.SetDeploymentSettingsSetDeploymentSettingsPythonError:
		return nil, &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	case *schema.SetDeploymentSettingsSetDeploymentSettingsUnauthorizedError:
		return nil, &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	default:
		return nil, fmt.Errorf("unexpected type(%T) of result", resp.SetDeploymentSettings)
	}
}
