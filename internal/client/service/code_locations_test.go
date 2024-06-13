package service_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/datarootsio/terraform-provider-dagster/internal/client/service"
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
		_ = client.DeleteCodeLocation(ctx, codeLocation.Name)
		_ = client.DeleteCodeLocation(ctx, updatedCodeLocation.Name)
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

func TestCodeLocationService_AsDocument_BasicCRUD(t *testing.T) {
	dagsterClient := testutils.GetDagsterClientFromEnvVars()
	var errNotFound *types.ErrNotFound

	ctx := context.Background()
	client := dagsterClient.CodeLocationsClient

	codeLocation := json.RawMessage(`{
		"location_name": "testing-codelocation-as-doc",
		"code_source": {
			"python_file": "my_file.py"
		},
		"image": "my_image:first"
	}`)

	updatedCodeLocation := json.RawMessage(`{
		"location_name": "testing-codelocation-as-doc",
		"code_source": {
			"python_file": "my_file_update.py"
		},
		"image": "my_image:updated"
	}`)

	codeLocationName, err := service.GetLocationNameFromDocument(codeLocation)
	assert.NoError(t, err)

	t.Cleanup(func() {
		_ = client.DeleteCodeLocation(ctx, codeLocationName)
		_ = client.DeleteCodeLocation(ctx, codeLocationName)
	})

	// Check that code location doesn't exist
	_, err = client.GetCodeLocationByName(ctx, codeLocationName)
	assert.Error(t, err)
	assert.ErrorAs(t, err, &errNotFound)

	// Create code location
	err = client.AddCodeLocationAsDocument(ctx, codeLocation)
	assert.Nil(t, err)

	// Read code location
	_, err = client.GetCodeLocationByName(ctx, codeLocationName)
	assert.NoError(t, err)

	// Update code location and check result
	err = client.UpdateCodeLocationAsDocument(ctx, updatedCodeLocation)
	assert.NoError(t, err)
	_, err = client.GetCodeLocationByName(ctx, codeLocationName)
	assert.NoError(t, err)

	// Delete code location
	err = client.DeleteCodeLocation(ctx, codeLocationName)
	assert.NoError(t, err)

	// Check everything cleaned up
	_, err = client.GetCodeLocationByName(ctx, codeLocationName)
	assert.Error(t, err)
	assert.ErrorAs(t, err, &errNotFound)
}

func TestCodeLocationService_AsDocument_Errors(t *testing.T) {
	dagsterClient := testutils.GetDagsterClientFromEnvVars()

	ctx := context.Background()
	client := dagsterClient.CodeLocationsClient

	errorInputMissingRequiredField := json.RawMessage(`{
		"location_name": "testing-codelocation-as-doc"
	}`)

	err := client.AddCodeLocationAsDocument(ctx, errorInputMissingRequiredField)
	assert.ErrorContains(t, err, "missing entry")

	errorInputMalformedJSON := json.RawMessage(`{
		"location_name": "testing-codelocation-as-doc",
		"code_source": {
			"python_file": "malformed
		},
		"image": "my_image:updated"
	}`)

	err = client.AddCodeLocationAsDocument(ctx, errorInputMalformedJSON)
	assert.ErrorContains(t, err, "invalid")

}
