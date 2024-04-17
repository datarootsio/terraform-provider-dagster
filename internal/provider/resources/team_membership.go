package resources

import (
	"context"
	"fmt"

	"github.com/datarootsio/terraform-provider-dagster/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &UserResource{}

func NewTeamMembershipResource() resource.Resource {
	return &TeamMembershipResource{}
}

type TeamMembershipResource struct {
	client client.DagsterClient
}

type TeamMembershipResourceModel struct {
	UserId types.Int64  `tfsdk:"user_id"`
	TeamId types.String `tfsdk:"team_id"`
}

func (r *TeamMembershipResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_membership"
}

func (r *TeamMembershipResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Adds a Dagster user to a team.`,
		Attributes: map[string]schema.Attribute{
			"user_id": schema.Int64Attribute{
				Computed:    false,
				Required:    true,
				Description: "User id",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"team_id": schema.StringAttribute{
				Computed:    false,
				Required:    true,
				Description: "Team id",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *TeamMembershipResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(client.DagsterClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected client.DagsterClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *TeamMembershipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TeamMembershipResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	userId := data.UserId.ValueInt64()
	teamId := data.TeamId.ValueString()

	// Check whether user exists
	_, err := r.client.UsersClient.GetUserById(ctx, userId)
	if err != nil {
		resp.Diagnostics.AddError("Client error", fmt.Sprintf("User %v does not exist", userId))
		return
	}

	// Check whether team exists
	_, err = r.client.TeamsClient.GetTeamById(ctx, teamId)
	if err != nil {
		resp.Diagnostics.AddError("Client error", fmt.Sprintf("Team %v does not exist", teamId))
		return
	}

	// Check whether user is not already in team
	in_team, err := r.client.TeamsClient.IsUserInTeam(ctx, int(userId), teamId)
	if err != nil {
		resp.Diagnostics.AddError("Client error", fmt.Sprintf("Failed to check if user %v is in team %v: %v", userId, teamId, err))
		return
	}
	if in_team {
		resp.Diagnostics.AddError("Client error", fmt.Sprintf("User %v is already in team %v", userId, teamId))
		return
	}

	err = r.client.TeamsClient.AddUserToTeam(ctx, int(userId), teamId)
	if err != nil {
		resp.Diagnostics.AddError("Client error", fmt.Sprintf("Failed to add user %v to team %v: %v", userId, teamId, err))
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("Added user %v to team %s", userId, teamId))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TeamMembershipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TeamMembershipResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	userId := int(data.UserId.ValueInt64())
	teamId := data.TeamId.ValueString()

	in_team, err := r.client.TeamsClient.IsUserInTeam(ctx, userId, teamId)
	if err != nil {
		resp.Diagnostics.AddError("Client error", fmt.Sprintf("Failed to check if user %v is in team %v: %v", userId, teamId, err))
		return
	}

	if !in_team {
		tflog.Trace(ctx, fmt.Sprintf("Removing resource: user %v is not in team %s", data.UserId.ValueInt64(), data.TeamId.ValueString()))
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TeamMembershipResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Not implemented: this resource is always recreated
}

func (r *TeamMembershipResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TeamMembershipResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	userId := data.UserId.ValueInt64()
	teamId := data.TeamId.ValueString()

	in_team, err := r.client.TeamsClient.IsUserInTeam(ctx, int(userId), teamId)
	if err != nil {
		resp.Diagnostics.AddError("Client error", fmt.Sprintf("Failed to check if user %v is in team %v: %v", userId, teamId, err))
		return
	}

	if !in_team {
		tflog.Trace(ctx, fmt.Sprintf("User %v is not in team %s", data.UserId.ValueInt64(), data.TeamId.ValueString()))
		resp.State.RemoveResource(ctx)
		return
	}

	err = r.client.TeamsClient.RemoveUserFromTeam(ctx, int(userId), teamId)
	if err != nil {
		resp.Diagnostics.AddError("Client error", fmt.Sprintf("Failed to remove user %v from team %v: %v", userId, teamId, err))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("Removed user %v from team %s", data.UserId.ValueInt64(), data.TeamId.ValueString()))
}
