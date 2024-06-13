package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Khan/genqlient/graphql"
	"github.com/datarootsio/terraform-provider-dagster/internal/client/schema"
	"github.com/datarootsio/terraform-provider-dagster/internal/client/types"
)

type CodeLocationsClient struct {
	client graphql.Client
}

func NewCodeLocationsClient(client graphql.Client) CodeLocationsClient {
	return CodeLocationsClient{
		client: client,
	}
}

func (c *CodeLocationsClient) GetCodeLocationByName(ctx context.Context, name string) (types.CodeLocation, error) {
	codeLocations, err := c.ListCodeLocations(ctx)
	if err != nil {
		return types.CodeLocation{}, err
	}

	for _, codeLocation := range codeLocations {
		if codeLocation.Name == name {
			return codeLocation, nil
		}
	}

	return types.CodeLocation{}, &types.ErrNotFound{What: "CodeLocation", Key: "name", Value: name}
}

func (c *CodeLocationsClient) AddCodeLocation(ctx context.Context, codeLocation types.CodeLocation) error {
	_, err := c.GetCodeLocationByName(ctx, codeLocation.Name)
	if err == nil {
		return &types.ErrAlreadyExists{What: "CodeLocation", Key: "name", Value: codeLocation.Name}
	}

	var errComp *types.ErrNotFound
	if !errors.As(err, &errComp) {
		return err
	}

	resp, err := schema.AddOrUpdateCodeLocation(
		ctx,
		c.client,
		codeLocation.Name,
		codeLocation.Image,
		codeLocation.CodeSource.PythonFile,
		codeLocation.CodeSource.PackageName,
		codeLocation.CodeSource.ModuleName,
		codeLocation.WorkingDirectory,
		codeLocation.ExecutablePath,
		codeLocation.Attribute,
		codeLocation.Git.CommitHash,
		codeLocation.Git.URL,
		codeLocation.AgentQueue,
	)

	if err != nil {
		return err
	}

	switch respCast := resp.AddOrUpdateLocation.(type) {
	case *schema.AddOrUpdateCodeLocationAddOrUpdateLocationWorkspaceEntry:
		return nil
	case *schema.AddOrUpdateCodeLocationAddOrUpdateLocationInvalidLocationError:
		return &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	case *schema.AddOrUpdateCodeLocationAddOrUpdateLocationPythonError:
		return &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	case *schema.AddOrUpdateCodeLocationAddOrUpdateLocationUnauthorizedError:
		return &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	default:
		return fmt.Errorf("unexpected type(%T) of result", resp.AddOrUpdateLocation)
	}
}

func (c *CodeLocationsClient) UpdateCodeLocation(ctx context.Context, codeLocation types.CodeLocation) error {
	_, err := c.GetCodeLocationByName(ctx, codeLocation.Name)
	if err != nil {
		return err
	}

	resp, err := schema.AddOrUpdateCodeLocation(
		ctx,
		c.client,
		codeLocation.Name,
		codeLocation.Image,
		codeLocation.CodeSource.PythonFile,
		codeLocation.CodeSource.PackageName,
		codeLocation.CodeSource.ModuleName,
		codeLocation.WorkingDirectory,
		codeLocation.ExecutablePath,
		codeLocation.Attribute,
		codeLocation.Git.CommitHash,
		codeLocation.Git.URL,
		codeLocation.AgentQueue,
	)

	if err != nil {
		return err
	}

	switch respCast := resp.AddOrUpdateLocation.(type) {
	case *schema.AddOrUpdateCodeLocationAddOrUpdateLocationWorkspaceEntry:
		return nil
	case *schema.AddOrUpdateCodeLocationAddOrUpdateLocationInvalidLocationError:
		return &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	case *schema.AddOrUpdateCodeLocationAddOrUpdateLocationPythonError:
		return &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	case *schema.AddOrUpdateCodeLocationAddOrUpdateLocationUnauthorizedError:
		return &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	default:
		return fmt.Errorf("unexpected type(%T) of result", resp.AddOrUpdateLocation)
	}
}

func (c *CodeLocationsClient) DeleteCodeLocation(ctx context.Context, name string) error {
	_, err := c.GetCodeLocationByName(ctx, name)
	if err != nil {
		return err
	}

	resp, err := schema.DeleteCodeLocation(ctx, c.client, name)
	if err != nil {
		return err
	}

	switch respCast := resp.DeleteLocation.(type) {
	case *schema.DeleteCodeLocationDeleteLocationDeleteLocationSuccess:
		return nil
	case *schema.DeleteCodeLocationDeleteLocationPythonError:
		return &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	case *schema.DeleteCodeLocationDeleteLocationUnauthorizedError:
		return &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	default:
		return fmt.Errorf("unexpected type(%T) of result", resp.DeleteLocation)
	}
}

func (c *CodeLocationsClient) ListCodeLocations(ctx context.Context) ([]types.CodeLocation, error) {
	resp, err := schema.ListCodeLocations(ctx, c.client)
	if err != nil {
		return []types.CodeLocation{}, err
	}

	codeLocationsAsBytes, err := resp.LocationsAsDocument.Document.MarshalJSON()
	if err != nil {
		return []types.CodeLocation{}, err
	}

	codeLocations := types.CodeLocationsAsDocumentResponse{}

	err = json.Unmarshal(codeLocationsAsBytes, &codeLocations)
	if err != nil {
		return []types.CodeLocation{}, err
	}

	return codeLocations.Locations, nil
}

func (c *CodeLocationsClient) AddCodeLocationAsDocument(ctx context.Context, codeLocationsAsDocument json.RawMessage) error {
	codeLocationName, err := GetLocationNameFromDocument(codeLocationsAsDocument)
	if err != nil {
		return err
	}

	_, err = c.GetCodeLocationByName(ctx, codeLocationName)
	if err == nil {
		return &types.ErrAlreadyExists{What: "CodeLocation", Key: "name", Value: codeLocationName}
	}

	resp, err := schema.AddOrUpdateLocationFromDocument(
		ctx,
		c.client,
		codeLocationsAsDocument,
	)
	if err != nil {
		return err
	}

	switch respCast := resp.AddOrUpdateLocationFromDocument.(type) {
	case *schema.AddOrUpdateLocationFromDocumentAddOrUpdateLocationFromDocumentWorkspaceEntry:
		return nil
	case *schema.AddOrUpdateLocationFromDocumentAddOrUpdateLocationFromDocumentInvalidLocationError:
		return &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	case *schema.AddOrUpdateLocationFromDocumentAddOrUpdateLocationFromDocumentUnauthorizedError:
		return &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	case *schema.AddOrUpdateLocationFromDocumentAddOrUpdateLocationFromDocumentPythonError:
		return &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	default:
		return fmt.Errorf("unexpected type(%T) of result", resp.AddOrUpdateLocationFromDocument)
	}
}

func (c *CodeLocationsClient) UpdateCodeLocationAsDocument(ctx context.Context, codeLocationsAsDocument json.RawMessage) error {
	codeLocationName, err := GetLocationNameFromDocument(codeLocationsAsDocument)
	if err != nil {
		return err
	}

	_, err = c.GetCodeLocationByName(ctx, codeLocationName)
	if err != nil {
		return err
	}

	resp, err := schema.AddOrUpdateLocationFromDocument(
		ctx,
		c.client,
		codeLocationsAsDocument,
	)
	if err != nil {
		return err
	}

	switch respCast := resp.AddOrUpdateLocationFromDocument.(type) {
	case *schema.AddOrUpdateLocationFromDocumentAddOrUpdateLocationFromDocumentWorkspaceEntry:
		return nil
	case *schema.AddOrUpdateLocationFromDocumentAddOrUpdateLocationFromDocumentInvalidLocationError:
		return &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	case *schema.AddOrUpdateLocationFromDocumentAddOrUpdateLocationFromDocumentUnauthorizedError:
		return &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	case *schema.AddOrUpdateLocationFromDocumentAddOrUpdateLocationFromDocumentPythonError:
		return &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	default:
		return fmt.Errorf("unexpected type(%T) of result", resp.AddOrUpdateLocationFromDocument)
	}
}

func GetLocationNameFromDocument(codeLocationsAsDocument json.RawMessage) (string, error) {
	var codeLocation map[string]interface{}
	err := json.Unmarshal(codeLocationsAsDocument, &codeLocation)

	if err != nil {
		return "", err
	}

	locationNameRaw, ok := codeLocation["location_name"]
	if !ok {
		return "", fmt.Errorf("location_name not found in codeLocationsAsDocument json")
	}

	locationName, ok := locationNameRaw.(string)
	if !ok {
		return "", fmt.Errorf("could not parse locationName in to string")
	}

	return locationName, nil
}
