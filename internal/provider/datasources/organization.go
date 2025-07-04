package datasources

import (
	"context"
	"fmt"

	"github.com/datarootsio/terraform-provider-dagster/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &OrganizationDataSource{}
	_ datasource.DataSourceWithConfigure = &OrganizationDataSource{}
)

type OrganizationDataSource struct {
	client client.DagsterClient
}

type OrganizationDataSourceModel struct {
	Id            types.Int64  `tfsdk:"id"`
	PublicId      types.String `tfsdk:"public_id"`
	Name          types.String `tfsdk:"name"`
	Status        types.String `tfsdk:"status"`
	AccountReview types.String `tfsdk:"account_review"`
}

//nolint:ireturn // required by Terraform API
func NewOrganizationDataSource() datasource.DataSource {
	return &OrganizationDataSource{}
}

// Metadata returns the data source type name.
func (d *OrganizationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

// Schema defines the schema for the data source.
func (d *OrganizationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Retrieve information about a Dagster Organization.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "Organization id",
			},
			"public_id": schema.StringAttribute{
				Computed:    true,
				Description: "Public ID of the Dagster Organization",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Name the Dagster Organization",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Status of the Dagster Organization",
			},
			"account_review": schema.StringAttribute{
				Computed:    true,
				Description: "Account review status of the Dagster Organization",
			},
		},
	}
}

// Configure adds the provider-configured client to the data source.
func (d *OrganizationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *OrganizationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OrganizationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	organization, err := d.client.InstanceClient.GetDagsterOrganization(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get instance Organization, got error: %s", err))
		return
	}

	data.Id = types.Int64Value(int64(organization.Id))
	data.PublicId = types.StringValue(organization.PublicId)
	data.Name = types.StringValue(organization.Name)
	data.Status = types.StringValue(string(organization.Status))
	data.AccountReview = types.StringValue(string(organization.AccountReview.Status))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
