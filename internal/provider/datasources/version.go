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
	_ datasource.DataSource              = &VersionDataSource{}
	_ datasource.DataSourceWithConfigure = &VersionDataSource{}
)

type VersionDataSource struct {
	client client.DagsterClient
}

type VersionDataSourceModel struct {
	Version types.String `tfsdk:"version"`
}

//nolint:ireturn // required by Terraform API
func NewVersionDataSource() datasource.DataSource {
	return &VersionDataSource{}
}

// Metadata returns the data source type name.
func (d *VersionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_version"
}

// Schema defines the schema for the data source.
func (d *VersionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Retrieves the version of the Dagster Cloud instance.`,
		Attributes: map[string]schema.Attribute{
			"version": schema.StringAttribute{
				Computed:    true,
				Description: "Version of the Dagster Cloud instance",
			},
		},
	}
}

// Configure adds the provider-configured client to the data source.
func (d *VersionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *VersionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VersionDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	version, err := d.client.InstanceClient.GetDagsterCloudVersion(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get instance version, got error: %s", err))
		return
	}

	data.Version = types.StringValue(version)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
