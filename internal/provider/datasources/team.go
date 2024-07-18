package datasources

import (
	"context"
	"fmt"

	"github.com/datarootsio/terraform-provider-dagster/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &TeamDataSource{}
	_ datasource.DataSourceWithConfigure = &TeamDataSource{}
)

type TeamDataSource struct {
	client client.DagsterClient
}

type TeamDataSourceModel struct {
	Name types.String `tfsdk:"name"`
	Id   types.String `tfsdk:"id"`
}

//nolint:ireturn // required by Terraform API
func NewTeamDataSource() datasource.DataSource {
	return &TeamDataSource{}
}

// Metadata returns the data source type name.
func (d *TeamDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

var teamAttributes = map[string]schema.Attribute{
	"name": schema.StringAttribute{
		Required:    true,
		Computed:    false,
		Description: "Name of the Dagster Cloud team",
	},
	"id": schema.StringAttribute{
		Computed:    true,
		Description: "Team id",
	},
}

// Schema defines the schema for the data source.
func (d *TeamDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Retrieve information about a Dagster Cloud team.`,
		Attributes:  teamAttributes,
	}
}

// Configure adds the provider-configured client to the data source.
func (d *TeamDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read refreshes the Terraform state with the latest data.
func (d *TeamDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TeamDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	team, err := d.client.TeamsClient.GetTeamByName(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get team information, got error: %s", err))
		return
	}

	data.Id = types.StringValue(team.Id)
	data.Name = types.StringValue(team.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
