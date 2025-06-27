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
	_ datasource.DataSource              = &UserDataSource{}
	_ datasource.DataSourceWithConfigure = &UserDataSource{}
)

type UserDataSource struct {
	client client.DagsterClient
}

type UserDataSourceModel struct {
	Id                types.Int64  `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Email             types.String `tfsdk:"email"`
	Picture           types.String `tfsdk:"picture"`
	IsScimProvisioned types.Bool   `tfsdk:"is_scim_provisioned"`
}

//nolint:ireturn // required by Terraform API
func NewUserDataSource() datasource.DataSource {
	return &UserDataSource{}
}

// Metadata returns the data source type name.
func (d *UserDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

var userAttributes = map[string]schema.Attribute{
	"id": schema.Int64Attribute{
		Computed:    true,
		Description: "User id",
	},
	"name": schema.StringAttribute{
		Computed:    true,
		Description: "Name the Dagster Cloud user",
	},
	"email": schema.StringAttribute{
		Required:    true,
		Computed:    false,
		Description: "Email address used to register the Dagster Cloud user",
	},
	"picture": schema.StringAttribute{
		Computed:    true,
		Description: "URL to user's profile picture",
	},
	"is_scim_provisioned": schema.BoolAttribute{
		Computed:    true,
		Description: "Whether this user was provisioned through SCIM",
	},
}

// Schema defines the schema for the data source.
func (d *UserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Retrieve information about a Dagster Cloud user.`,
		Attributes:  userAttributes,
	}
}

// Configure adds the provider-configured client to the data source.
func (d *UserDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *UserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UserDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	user, err := d.client.UsersClient.GetUserByEmail(ctx, data.Email.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get user information, got error: %s", err))
		return
	}

	data.Id = types.Int64Value(int64(user.UserId))
	data.Email = types.StringValue(user.Email)
	data.Name = types.StringValue(user.Name)
	data.Picture = types.StringValue(user.Picture)
	data.IsScimProvisioned = types.BoolValue(user.IsScimProvisioned)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
