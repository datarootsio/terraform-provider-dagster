package datasources

import (
	"context"
	"fmt"

	"github.com/datarootsio/terraform-provider-dagster/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"sigs.k8s.io/yaml"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &DeploymentSettingsDataSource{}
	_ datasource.DataSourceWithConfigure = &DeploymentSettingsDataSource{}
)

type DeploymentSettingsDataSource struct {
	client client.DagsterClient
}

type DeploymentSettingsDataSourceModel struct {
	YAMLBody types.String `tfsdk:"yaml_body"`
	JSONBody types.String `tfsdk:"json"`
}

func NewDeploymentSettingsDataSource() datasource.DataSource {
	return &DeploymentSettingsDataSource{}
}

func (d *DeploymentSettingsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment_settings_document"
}

// Schema defines the schema for the data source.
func (d *DeploymentSettingsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Stores deployment settings`,
		Attributes: map[string]schema.Attribute{
			"yaml_body": schema.StringAttribute{
				Required:    true,
				Computed:    false,
				Description: "Deployment settings as YAML document",
			},
			"json": schema.StringAttribute{
				Computed:    true,
				Description: "Deployment settings as JSON document",
			},
		},
	}
}

// Configure adds the provider-configured client to the data source.
func (d *DeploymentSettingsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(client.DagsterClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected client.DagsterClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *DeploymentSettingsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DeploymentSettingsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert YAML to JSON and store in state
	yamlString := data.YAMLBody.ValueString()
	json, err := yaml.YAMLToJSON([]byte(yamlString))
	if err != nil {
		resp.Diagnostics.AddError("Unable to parse YAML", err.Error())
		return
	}

	data.JSONBody = types.StringValue(string(json))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
