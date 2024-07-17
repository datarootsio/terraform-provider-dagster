package datasources

import (
	"context"
	"fmt"

	"github.com/datarootsio/terraform-provider-dagster/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &TeamsDataSource{}
	_ datasource.DataSourceWithConfigure = &TeamsDataSource{}
)

type TeamsDataSource struct {
	client client.DagsterClient
}

type TeamsDataSourceModel struct {
	RegexFilter types.String `tfsdk:"regex_filter"`
	Teams       types.List   `tfsdk:"teams"`
}

//nolint:ireturn // required by Terraform API
func NewTeamsDataSource() datasource.DataSource {
	return &TeamsDataSource{}
}

// Metadata returns the data source type name.
func (d *TeamsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_teams"
}

// Schema defines the schema for the data source.
func (d *TeamsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Retrieve information about a Dagster Cloud teams.`,
		Attributes: map[string]schema.Attribute{
			"regex_filter": schema.StringAttribute{
				Required:    true,
				Computed:    false,
				Description: "Regex filter to select the Dagster Cloud teams",
			},
			"teams": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Teams",
				NestedObject: schema.NestedAttributeObject{
					Attributes: teamAttributes,
				},
			},
		},
	}
}

// Configure adds the provider-configured client to the data source.
func (d *TeamsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *TeamsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TeamsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	teams, err := d.client.TeamsClient.GetTeamsByRegex(ctx, data.RegexFilter.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get teams information, got error: %s", err))
		return
	}

	attributeTypes := map[string]attr.Type{
		"name": types.StringType,
		"id":   types.StringType,
	}

	teamObjects := make([]attr.Value, 0, len(teams))
	for _, team := range teams {
		attributeValues := map[string]attr.Value{
			"name": types.StringValue(team.Name),
			"id":   types.StringValue(team.Id),
		}

		teamObject, diag := types.ObjectValue(attributeTypes, attributeValues)
		resp.Diagnostics.Append(diag...)
		if resp.Diagnostics.HasError() {
			return
		}

		teamObjects = append(teamObjects, teamObject)
	}

	teamsAsList, diag := types.ListValue(types.ObjectType{AttrTypes: attributeTypes}, teamObjects)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Teams = teamsAsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
