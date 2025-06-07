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
