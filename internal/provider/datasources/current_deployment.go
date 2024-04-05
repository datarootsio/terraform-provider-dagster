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
var _ datasource.DataSource = &CurrentDeploymentDataSource{}
var _ datasource.DataSourceWithConfigure = &CurrentDeploymentDataSource{}

type CurrentDeploymentDataSource struct {
	client client.DagsterClient
}

type CurrentDeploymentDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// NewCurrentDeploymentDataSource returns a new CurrentDeploymentDataSource.
//
//nolint:ireturn // required by Terraform API
func NewCurrentDeploymentDataSource() datasource.DataSource {
	return &CurrentDeploymentDataSource{}
}

// Metadata returns the data source type name.
func (d *CurrentDeploymentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_current_deployment"
}

// Schema defines the schema for the data source.
func (d *CurrentDeploymentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Retrieve information about a the CurrentDeployment.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "CurrentDeployment ID",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of CurrentDeployment",
			},
		},
	}
}

// Configure adds the provider-configured client to the data source.
func (d *CurrentDeploymentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *CurrentDeploymentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CurrentDeploymentDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	deployment, err := d.client.DeploymentClient.GetCurrentDeployment(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read current deployment, got error: %s", err))

		return
	}

	data.ID = types.Int64Value(int64(deployment.DeploymentId))
	data.Name = types.StringValue(deployment.DeploymentName)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
