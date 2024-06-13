package testutils

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDagsterClientFromEnvVars(t *testing.T) {
	client := GetDagsterClientFromEnvVars()
	deploy, err := client.DeploymentClient.GetCurrentDeployment(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, deploy.DeploymentName, os.Getenv("TF_VAR_testing_dagster_deployment"))
}
