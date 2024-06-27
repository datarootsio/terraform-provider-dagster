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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &UserResource{}
	_ resource.ResourceWithImportState = &UserResource{}
)

func NewUserResource() resource.Resource {
	return &UserResource{}
}

type UserResource struct {
	client client.DagsterClient
}

type UserResourceModel struct {
	Id                       types.Int64  `tfsdk:"id"`
	Name                     types.String `tfsdk:"name"`
	Email                    types.String `tfsdk:"email"`
	Picture                  types.String `tfsdk:"picture"`
	RemoveDefaultPermissions types.Bool   `tfsdk:"remove_default_permissions"`
}

func (r *UserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Registers a new user.`,
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"remove_default_permissions": schema.BoolAttribute{
				Required:    true,
				Computed:    false,
				Description: "Remove the default Viewer permissions on creation",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
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

	// Check if user exists already
	email := data.Email.ValueString()
	user, err := r.client.UsersClient.GetUserByEmail(ctx, email)
	if err == nil { // Exists
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("User with email %s is already registered", email))
		return
	}

	user, err = r.client.UsersClient.AddUser(ctx, email)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create user, got error: %s", err))
		return
	}

	// Remove the default "Viewer" permission on all deployments and codelocations
	if data.RemoveDefaultPermissions.ValueBool() {
		err = removeAllUserPermissions(ctx, r.client, email)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("%s", err))
			return
		}
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
	// Not implemented, changing the email address triggers replacement
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

// removeAllUserPermissions removes the default "Viewer" permission on all (normal and branch) deployments
func removeAllUserPermissions(ctx context.Context, client client.DagsterClient, email string) error {
	deployments, err := client.DeploymentClient.GetAllDeployments(ctx)
	if err != nil {
		return errors.New("Erorr getting a list of deployments")
	}

	var deploymentId int
	for _, deployment := range deployments {
		deploymentId = deployment.DeploymentId
		err := client.UsersClient.RemoveUserPermission(
			ctx,
			email,
			deploymentId,
			clientSchema.PermissionDeploymentScopeDeployment,
		)
		if err != nil {
			return fmt.Errorf("error removing permissions from user %s on deployment %v", email, deploymentId)
		}
	}

	// deploymentId does not matter in this call
	err = client.UsersClient.RemoveUserPermission(ctx, email, 0, clientSchema.PermissionDeploymentScopeAllBranchDeployments)
	if err != nil {
		return fmt.Errorf("error removing permissions from user %s on branch deployments", email)
	}
	return nil
}
