package service_test

import (
	"context"
	"os"
	"testing"

	"github.com/datarootsio/terraform-provider-dagster/internal/client/types"
	"github.com/datarootsio/terraform-provider-dagster/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestDeploymentService_BasicCRUD(t *testing.T) {
	client := testutils.GetDagsterClientFromEnvVars()

	ctx := context.Background()
	deploymentClient := client.DeploymentClient

	deploymentName := "test-deployment"
	var errNotFound *types.ErrNotFound

	// Ensure no deployment with the test name exist
	_, err := deploymentClient.GetDeploymentByName(ctx, deploymentName)
	assert.ErrorAs(t, err, &errNotFound)

	// Create deployment
	deploymentCreated, err := deploymentClient.CreateHybridDeployment(ctx, deploymentName)
	assert.NoError(t, err)
	assert.Equal(t, deploymentName, deploymentCreated.DeploymentName, "Expected deployment names to be the same.")

	t.Cleanup(func() {
		deploymentClient.DeleteDeployment(ctx, deploymentCreated.DeploymentId)
	})

	// Read deployment by id and name
	teamById, err := deploymentClient.GetDeploymentById(ctx, deploymentCreated.DeploymentId)
	assert.NoError(t, err)
	assert.Equal(t, deploymentName, teamById.DeploymentName, "Expected team names to be the same.")

	teamByName, err := deploymentClient.GetDeploymentByName(ctx, deploymentName)
	assert.NoError(t, err)
	assert.Equal(t, deploymentName, teamByName.DeploymentName, "Expected team names to be the same.")

	// Delete deployment
	err = deploymentClient.DeleteDeployment(ctx, deploymentCreated.DeploymentId)
	assert.NoError(t, err)

	// Ensure everything is cleaned up
	_, err = deploymentClient.GetDeploymentByName(ctx, deploymentName)
	assert.ErrorAs(t, err, &errNotFound)
}

func TestDeploymentSettingsCRUD(t *testing.T) {
	// Todo: create deploy, read settings, set settings, read settings, delete
	client := testutils.GetDagsterClientFromEnvVars()

	ctx := context.Background()
	deploymentClient := client.DeploymentClient

	deploymentName := "test-deployment-settings"
	var errNotFound *types.ErrNotFound

	// Ensure no deployment with the test name exist
	_, err := deploymentClient.GetDeploymentByName(ctx, deploymentName)
	assert.ErrorAs(t, err, &errNotFound)

	// Create deployment
	deployment, err := deploymentClient.CreateHybridDeployment(ctx, deploymentName)
	assert.NoError(t, err)
	settingsAtCreation := deployment.DeploymentSettings

	t.Cleanup(func() {
		deploymentClient.DeleteDeployment(ctx, deployment.DeploymentId)
	})

	// Read deployment by id
	deploymentRead, err := deploymentClient.GetDeploymentById(ctx, deployment.DeploymentId)
	assert.NoError(t, err)
	settingsAtRead := deploymentRead.DeploymentSettings

	assert.Equal(t, settingsAtCreation, settingsAtRead)

	// Modify deployment settings to "sso_default_role: LAUNCHER"
	settingsJSON := testutils.UnmarshalJSONOrPanic(settingsAtRead.Settings)
	settingsJSON["sso_default_role"] = "LAUNCHER"

	settingsToApply := testutils.MarshalJSONOrPanic(settingsJSON)
	settings, err := deploymentClient.SetDeploymentSettings(ctx, deployment.DeploymentId, settingsToApply)

	assert.NoError(t, err)
	assert.Equal(t, settingsToApply, settings)
	assert.Equal(t, testutils.UnmarshalJSONOrPanic(settings)["sso_default_role"], "LAUNCHER")
}

func TestGetCurrentDeployment(t *testing.T) {
	client := testutils.GetDagsterClientFromEnvVars()
	expected := os.Getenv("TF_VAR_testing_dagster_deployment")
	ctx := context.Background()

	current, err := client.DeploymentClient.GetCurrentDeployment(ctx)
	assert.NoError(t, err)

	actual := current.DeploymentName
	assert.Equal(t, expected, actual)
}

func TestGetDeploymentDoesntExist(t *testing.T) {
	client := testutils.GetDagsterClientFromEnvVars()
	ctx := context.Background()

	_, err := client.DeploymentClient.GetDeploymentById(ctx, -1)
	var errNotFound *types.ErrNotFound
	assert.ErrorAs(t, err, &errNotFound)
}

func TestCreateDeploymentAlreadyExists(t *testing.T) {
	client := testutils.GetDagsterClientFromEnvVars()
	ctx := context.Background()

	_, err := client.DeploymentClient.CreateHybridDeployment(ctx, "prod")
	var errExists *types.ErrAlreadyExists
	assert.ErrorAs(t, err, &errExists)
}

func TestCreateDeploymentInvalidName(t *testing.T) {
	client := testutils.GetDagsterClientFromEnvVars()
	ctx := context.Background()

	_, err := client.DeploymentClient.CreateHybridDeployment(ctx, "_%@")
	assert.Error(t, err)
}
