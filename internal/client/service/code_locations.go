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
	resp, err := schema.ListCodeLocations(ctx, c.client)
	if err != nil {
		return types.CodeLocation{}, err
	}

	codeLocationsAsBytes, err := resp.LocationsAsDocument.Document.MarshalJSON()
	if err != nil {
		return types.CodeLocation{}, err
	}

	codeLocations := types.CodeLocationsAsDocumentResponse{}

	err = json.Unmarshal(codeLocationsAsBytes, &codeLocations)
	if err != nil {
		return types.CodeLocation{}, err
	}

	for _, codeLocation := range codeLocations.Locations {
		if codeLocation.Name == name {
			return codeLocation, nil
		}
	}

	return types.CodeLocation{}, &types.ErrNotFound{What: "CodeLocation", Key: "name", Value: name}
}

func (c *CodeLocationsClient) AddCodeLocation(ctx context.Context, codeLocation types.CodeLocation) (types.CodeLocation, error) {
	_, err := c.GetCodeLocationByName(ctx, codeLocation.Name)

	if err == nil {
		return types.CodeLocation{}, &types.ErrAlreadyExists{What: "CodeLocation", Key: "name", Value: codeLocation.Name}
	}

	var errComp *types.ErrNotFound
	if !errors.As(err, &errComp) {
		return types.CodeLocation{}, err
	}

	resp, err := schema.AddOrUpdateCodeLocation(ctx, c.client, schema.LocationSelector{
		Name:       codeLocation.Name,
		Image:      codeLocation.Image,
		PythonFile: codeLocation.CodeSource.PythonFile,
	})

	if err != nil {
		return types.CodeLocation{}, err
	}

	switch respCast := resp.AddOrUpdateLocation.(type) {
	case *schema.AddOrUpdateCodeLocationAddOrUpdateLocationWorkspaceEntry:
		return codeLocation, nil
	case *schema.AddOrUpdateCodeLocationAddOrUpdateLocationInvalidLocationError:
		return types.CodeLocation{}, &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	case *schema.AddOrUpdateCodeLocationAddOrUpdateLocationPythonError:
		return types.CodeLocation{}, &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	case *schema.AddOrUpdateCodeLocationAddOrUpdateLocationUnauthorizedError:
		return types.CodeLocation{}, &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	default:
		return types.CodeLocation{}, fmt.Errorf("unexpected type(%T) of result", resp.AddOrUpdateLocation)
	}
}

func (c *CodeLocationsClient) UpdateCodeLocation(ctx context.Context, codeLocation types.CodeLocation) (types.CodeLocation, error) {
	_, err := c.GetCodeLocationByName(ctx, codeLocation.Name)
	if err != nil {
		return types.CodeLocation{}, err
	}

	resp, err := schema.AddOrUpdateCodeLocation(ctx, c.client, schema.LocationSelector{
		Name:       codeLocation.Name,
		Image:      codeLocation.Image,
		PythonFile: codeLocation.CodeSource.PythonFile,
	})

	if err != nil {
		return types.CodeLocation{}, err
	}

	switch respCast := resp.AddOrUpdateLocation.(type) {
	case *schema.AddOrUpdateCodeLocationAddOrUpdateLocationWorkspaceEntry:
		return codeLocation, nil
	case *schema.AddOrUpdateCodeLocationAddOrUpdateLocationInvalidLocationError:
		return types.CodeLocation{}, &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	case *schema.AddOrUpdateCodeLocationAddOrUpdateLocationPythonError:
		return types.CodeLocation{}, &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	case *schema.AddOrUpdateCodeLocationAddOrUpdateLocationUnauthorizedError:
		return types.CodeLocation{}, &types.ErrApi{Typename: respCast.Typename, Message: respCast.Message}
	default:
		return types.CodeLocation{}, fmt.Errorf("unexpected type(%T) of result", resp.AddOrUpdateLocation)
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
