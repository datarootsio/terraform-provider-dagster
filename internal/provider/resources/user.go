package resources

import (
	"context"
	"errors"
	"fmt"

	"github.com/datarootsio/terraform-provider-dagster/internal/client"
	clientSchema "github.com/datarootsio/terraform-provider-dagster/internal/client/schema"
	clientTypes "github.com/datarootsio/terraform-provider-dagster/internal/client/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &UserResource{}
var _ resource.ResourceWithImportState = &UserResource{}

func NewUserResource() resource.Resource {
	return &UserResource{}
}

type UserResource struct {
	client client.DagsterClient
}

type UserResourceModel struct {
	Id      types.Int64  `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Email   types.String `tfsdk:"email"`
	Picture types.String `tfsdk:"picture"`
}

func (r *UserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Create a Dagster Cloud user.`,
		Attributes: map[string]schema.Attribute{
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
		},
	}
}

func (r *UserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UserResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.UsersClient.AddUser(ctx, data.Email.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create user, got error: %s", err))
		return
	}

	data.Id = types.Int64Value(int64(user.UserId))
	data.Name = types.StringValue(user.Name)
	data.Email = types.StringValue(user.Email)
	data.Picture = types.StringValue(user.Picture)
	data.Picture = types.StringValue(user.Picture)
	tflog.Trace(ctx, fmt.Sprintf("Created resource with id %v from email %s\n", user.UserId, user.Email))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UserResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var user clientSchema.User
	user, err := r.client.UsersClient.GetUserByEmail(ctx, data.Email.ValueString())
	if err != nil {
		var errComp *clientTypes.ErrNotFound

		if errors.As(err, &errComp) {
			tflog.Trace(ctx, "User not found, probably already deleted manually, removing from state")
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user, got error: %s", err))
		}
		return
	}

	data.Name = types.StringValue(user.Name)
	data.Id = types.Int64Value(int64(user.UserId))
	data.Email = types.StringValue(user.Email)
	data.Picture = types.StringValue(user.Picture)
	data.Picture = types.StringValue(user.Picture)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data UserResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UserResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UsersClient.RemoveUser(ctx, data.Email.ValueString())
	if err != nil {
		var errComp *clientTypes.ErrNotFound
		if errors.As(err, &errComp) {
			tflog.Trace(ctx, "User not found, probably already deleted manually, removing from state")
			resp.State.RemoveResource(ctx)
		} else {
			msg := fmt.Sprintf("Unable to remove user with email %s, got error: %s", data.Email.ValueString(), err)
			resp.Diagnostics.AddError("Client Error", msg)
		}
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("Deleted user with email address %s", data.Email.ValueString()))
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("email"), req, resp)
}
