package service_test

import (
	"context"
	"testing"

	"github.com/datarootsio/terraform-provider-dagster/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestInstanceService_Version(t *testing.T) {
	client := testutils.GetDagsterClientFromEnvVars()

	ctx := context.Background()

	c := client.InstanceClient
	version, err := c.GetDagsterCloudVersion(ctx)
	assert.NoError(t, err)

	re := `^[a-zA-Z0-9]{8}-[a-zA-Z0-9]{8}$`
	assert.Regexp(t, re, version)
}

func TestInstanceService_Organization(t *testing.T) {
	client := testutils.GetDagsterClientFromEnvVars()

	ctx := context.Background()

	c := client.InstanceClient
	organization, err := c.GetDagsterOrganization(ctx)
	assert.NoError(t, err)

	assert.Equal(t, "dataroots-terraform-provider-dagster", organization.Name)
	assert.Equal(t, "dataroots-terraform-provider-dagster", organization.PublicId)
}
