package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/datarootsio/terraform-provider-dagster/internal/client"
	"github.com/datarootsio/terraform-provider-dagster/internal/provider/datasources"
	"github.com/datarootsio/terraform-provider-dagster/internal/provider/resources"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DagsterProvider struct {
	client client.DagsterClient
}

// DagsterProviderModel maps provider schema data to a Go type.
type DagsterProviderModel struct {
	Organization types.String `tfsdk:"organization"`
	Deployment   types.String `tfsdk:"deployment"`
	APIToken     types.String `tfsdk:"api_token"`
}

var _ = provider.Provider(&DagsterProvider{})

// New returns a new Dagster Provider instance.
//
//nolint:ireturn // required by Terraform API
func New() provider.Provider {
	return &DagsterProvider{}
}

// Metadata returns the provider type name.
func (p *DagsterProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "dagster"
}

// Schema defines the provider-level schema for configuration data.
func (p *DagsterProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"organization": schema.StringAttribute{
				Description: "Dagster Organization. Can also be set via the `DAGSTER_CLOUD_ORGANIZATION` environment variable. Defaults to `https://api.dagster.cloud`",
				Required:    false,
				Optional:    true,
			},
			"deployment": schema.StringAttribute{
				Description: "Dagster Deployment. Can also be set via the `DAGSTER_CLOUD_DEPLOYMENT` environment variable. Defaults to `https://api.dagster.cloud`",
				Required:    false,
				Optional:    true,
			},
			"api_token": schema.StringAttribute{
				Description: "Dagster Cloud API Token. Can also be set via the `DAGSTER_CLOUD_API_TOKEN` environment variable.",
				Sensitive:   true,
				Required:    false,
				Optional:    true,
			},
		},
	}
}

// Configure configures the provider's internal client.
func (p *DagsterProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	config := &DagsterProviderModel{}

	// Populate the model from provider configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var organization string
	if !config.Organization.IsNull() {
		organization = config.Organization.ValueString()
	} else if envVarVal, ok := os.LookupEnv("DAGSTER_CLOUD_ORGANIZATION"); ok {
		organization = envVarVal
	}

	var deployment string
	if !config.Deployment.IsNull() {
		deployment = config.Deployment.ValueString()
	} else if envVarVal, ok := os.LookupEnv("DAGSTER_CLOUD_DEPLOYMENT"); ok {
		deployment = envVarVal
	}

	var apiToken string
	if !config.APIToken.IsNull() {
		apiToken = config.APIToken.ValueString()
	} else if envVarVal, ok := os.LookupEnv("DAGSTER_CLOUD_API_TOKEN"); ok {
		apiToken = envVarVal
	}

	// Ensure that all configuration values passed in to provider are known
	// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/terraform-concepts#unknown-values
	if organization == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("organization"),
			"Unknown Dagster Organization",
			"The Dagster Organization is not known at configuration time. "+
				"Potential resolutions: target apply the source of the value first, set the value statically in the configuration, or set the DAGSTER_CLOUD_ORGANIZATION environment variable.",
		)
	}

	if deployment == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("deployment"),
			"Unknown Dagster Deployment",
			"The Dagster Deployment is not known at configuration time. "+
				"Potential resolutions: target apply the source of the value first, set the value statically in the configuration, or set the DAGSTER_CLOUD_DEPLOYMENT environment variable.",
		)
	}

	if apiToken == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_token"),
			"Unknown Dagster API Token",
			"The Dagster API Token is not known at configuration time. "+
				"Potential resolutions: target apply the source of the value first, set the value statically in the configuration, or set the DAGSTER_CLOUD_API_TOKEN environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	dagsterClient, err := client.NewDagsterClient(
		organization,
		deployment,
		apiToken,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create Dagster API Client",
			fmt.Sprintf("An unexpected error occurred when creating the Dagster API client. This is a bug in the provider, please create an issue against https://github.com/DagsterHQ/terraform-provider-dagster unless it has already been reported. "+
				"Error returned by the client: %s", err),
		)

		return
	}
	p.client = dagsterClient

	// Pass client to DataSource and Resource type Configure methods
	resp.DataSourceData = dagsterClient
	resp.ResourceData = dagsterClient
}

// DataSources defines the data sources implemented in the provider.
func (p *DagsterProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewCurrentDeploymentDataSource,
		datasources.NewUserDataSource,
		datasources.NewUsersDataSource,
		datasources.NewConfigurationDocumentDataSource,
		datasources.NewTeamDataSource,
		datasources.NewTeamsDataSource,
		datasources.NewVersionDataSource,
		datasources.NewOrganizationDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *DagsterProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewUserResource,
		resources.NewTeamResource,
		resources.NewTeamDeploymentGrantResource,
		resources.NewTeamMembershipResource,
		resources.NewCodeLocationResource,
		resources.NewDeploymentResource,
		resources.NewCodeLocationFromDocumentResource,
	}
}
