package service_test

import (
	"context"
	"testing"

	"github.com/datarootsio/terraform-provider-dagster/internal/client/types"
	"github.com/datarootsio/terraform-provider-dagster/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestCodeLocationService_BasicCRUD(t *testing.T) {
	dagsterClient := testutils.GetDagsterClientFromEnvVars()
	var errNotFound *types.ErrNotFound

	ctx := context.Background()
	client := dagsterClient.CodeLocationsClient

	codeLocation := types.CodeLocation{
		Name:  "testing-codelocation",
		Image: "test123",
		CodeSource: types.CodeLocationCodeSource{
			PythonFile: "test.py",
		},
	}
	updatedCodeLocation := types.CodeLocation{
		Name:  "testing-codelocation",
		Image: "test123-2",
		CodeSource: types.CodeLocationCodeSource{
			PythonFile: "test.py-2",
		},
	}

	t.Cleanup(func() {
		client.DeleteCodeLocation(ctx, codeLocation.Name)
		client.DeleteCodeLocation(ctx, updatedCodeLocation.Name)
	})

	// Check that code location doesn't exist
	_, err := client.GetCodeLocationByName(ctx, codeLocation.Name)
	assert.Error(t, err)
	assert.ErrorAs(t, err, &errNotFound)

	// Create code location
	err = client.AddCodeLocation(ctx, codeLocation)
	assert.Nil(t, err)

	// Read code location
	result, err := client.GetCodeLocationByName(ctx, codeLocation.Name)
	assert.NoError(t, err)
	assert.Equal(t, result, codeLocation)

	// Update code location and check result
	err = client.UpdateCodeLocation(ctx, updatedCodeLocation)
	assert.NoError(t, err)
	result, err = client.GetCodeLocationByName(ctx, codeLocation.Name)
	assert.NoError(t, err)
	assert.Equal(t, updatedCodeLocation, result)

	// Delete code location
	err = client.DeleteCodeLocation(ctx, updatedCodeLocation.Name)
	assert.NoError(t, err)

	// Check everything cleaned up
	_, err = client.GetCodeLocationByName(ctx, codeLocation.Name)
	assert.Error(t, err)
	assert.ErrorAs(t, err, &errNotFound)

	_, err = client.GetCodeLocationByName(ctx, updatedCodeLocation.Name)
	assert.Error(t, err)
	assert.ErrorAs(t, err, &errNotFound)
}
