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
	_ datasource.DataSource              = &UsersDataSource{}
	_ datasource.DataSourceWithConfigure = &UsersDataSource{}
)

type UsersDataSource struct {
	client client.DagsterClient
}

type UsersDataSourceModel struct {
	EmailRegex types.String `tfsdk:"email_regex"`
	Users      types.List   `tfsdk:"users"`
}

//nolint:ireturn // required by Terraform API
func NewUsersDataSource() datasource.DataSource {
	return &UsersDataSource{}
}

// Metadata returns the data source type name.
func (d *UsersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

// Schema defines the schema for the data source.
func (d *UsersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Retrieve information about the Dagster Cloud users.`,
		Attributes: map[string]schema.Attribute{
			"email_regex": schema.StringAttribute{
				Required:    true,
				Computed:    false,
				Description: "Regex filter to select the Dagster Cloud users based on the email id of the user. Regex matching is done using `https://pkg.go.dev/regexp`.",
			},
			"users": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Users",
				NestedObject: schema.NestedAttributeObject{
					Attributes: userAttributes,
				},
			},
		},
	}
}

// Configure adds the provider-configured client to the data source.
func (d *UsersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *UsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UsersDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	users, err := d.client.UsersClient.GetUsersByRegex(ctx, data.EmailRegex.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get users information, got error: %s", err))
		return
	}

	attributeTypes := map[string]attr.Type{
		"id":                  types.Int64Type,
		"name":                types.StringType,
		"email":               types.StringType,
		"picture":             types.StringType,
		"is_scim_provisioned": types.BoolType,
	}

	userObjects := make([]attr.Value, 0, len(users))
	for _, user := range users {
		attributeValues := map[string]attr.Value{
			"id":                  types.Int64Value(int64(user.UserId)),
			"name":                types.StringValue(user.Name),
			"email":               types.StringValue(user.Email),
			"picture":             types.StringValue(user.Picture),
			"is_scim_provisioned": types.BoolValue(user.IsScimProvisioned),
		}

		userObject, diag := types.ObjectValue(attributeTypes, attributeValues)
		resp.Diagnostics.Append(diag...)
		if resp.Diagnostics.HasError() {
			return
		}

		userObjects = append(userObjects, userObject)
	}

	usersAsList, diag := types.ListValue(types.ObjectType{AttrTypes: attributeTypes}, userObjects)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Users = usersAsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
