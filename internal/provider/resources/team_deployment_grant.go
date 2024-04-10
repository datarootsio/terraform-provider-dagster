package resources

import (
	"context"
	"errors"
	"fmt"

	"github.com/datarootsio/terraform-provider-dagster/internal/client"
	clientSchema "github.com/datarootsio/terraform-provider-dagster/internal/client/schema"
	clientTypes "github.com/datarootsio/terraform-provider-dagster/internal/client/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &TeamDeploymentGrantResource{}

func NewTeamDeploymentGrantResource() resource.Resource {
	return &TeamDeploymentGrantResource{}
}

type TeamDeploymentGrantResource struct {
	client client.DagsterClient
}

type TeamDeploymentGrantResourceModel struct {
	DeploymentId       types.Int64  `tfsdk:"deployment_id"`
	TeamId             types.String `tfsdk:"team_id"`
	Grant              types.String `tfsdk:"grant"`
	CodeLocationGrants types.Set    `tfsdk:"code_location_grants"`
	Id                 types.Int64  `tfsdk:"id"`
}

func (r *TeamDeploymentGrantResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_deployment_grant"
}

func (r *TeamDeploymentGrantResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Team Deployment Grant resource",

		Attributes: map[string]schema.Attribute{
			"deployment_id": schema.Int64Attribute{
				MarkdownDescription: "Team Deployment Grant DeploymentId",
				Required:            true,
			},
			"team_id": schema.StringAttribute{
				MarkdownDescription: "Team Deployment Grant TeamId",
				Required:            true,
			},
			"grant": schema.StringAttribute{
				MarkdownDescription: "Team Deployment Grant Grant",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(clientTypes.DeploymentGrantEnumValues()...),
				},
			},
			"code_location_grants": schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Code location Name",
							Required:            true,
						},
						"grant": schema.StringAttribute{
							MarkdownDescription: "Code location Grant",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf(clientTypes.LocationGrantEnumValues()...),
							},
						},
					},
				},
				Optional: true,
			},
			"id": schema.Int64Attribute{
				MarkdownDescription: "Team Deployment Grant Id",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *TeamDeploymentGrantResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TeamDeploymentGrantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TeamDeploymentGrantResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	grantEnum, err := clientTypes.ConvertToGrantEnum(data.Grant.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create team deployment grant, got error: %s", err))
		return
	}

	codeLocationGrants := make([]clientSchema.LocationScopedGrant, 0)

	if !data.CodeLocationGrants.IsNull() {
		for _, codeLocationGrant := range data.CodeLocationGrants.Elements() {
			codeLocationGrantObject := codeLocationGrant.(types.Object)
			attributes := codeLocationGrantObject.Attributes()

			grant, err := clientTypes.ConvertToGrantEnum(attributes["grant"].(types.String).ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create team deployment grant, got error: %s", err))
				return
			}

			codeLocationGrants = append(
				codeLocationGrants,
				clientSchema.LocationScopedGrant{
					LocationName: attributes["name"].(types.String).ValueString(),
					Grant:        grant,
				},
			)
		}
	}

	teamDeploymentGrant, err := r.client.TeamsClient.CreateOrUpdateTeamDeploymentGrant(
		ctx,
		data.TeamId.ValueString(),
		int(data.DeploymentId.ValueInt64()),
		grantEnum,
		codeLocationGrants,
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create team deployment grant, got error: %s", err))
		return
	}

	data.Id = types.Int64Value(int64(teamDeploymentGrant.Id))

	tflog.Trace(ctx, "created team deployment grant resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TeamDeploymentGrantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TeamDeploymentGrantResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	teamDeploymentGrant, err := r.client.TeamsClient.GetTeamDeploymentGrantByTeamAndDeploymentId(
		ctx,
		data.TeamId.ValueString(),
		int(data.DeploymentId.ValueInt64()),
	)
	if err != nil {
		var errComp *clientTypes.ErrNotFound
		// This handles the case when a resource is still in the state but delete from the API
		// in that case we remove the resource from the state so that it gets recreated.
		// We check for !data.Name.IsNull() because if it is null it means we just imported a resource
		// via "tf import" and we want to show an error message instead of resp.State.RemoveResource.
		if errors.As(err, &errComp) {
			tflog.Trace(ctx, "Team not found, probably already deleted manually, removing from state")
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read team, got error: %s", err))
		}
		return
	}

	data.Id = types.Int64Value(int64(teamDeploymentGrant.Id))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TeamDeploymentGrantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TeamDeploymentGrantResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	grantEnum, err := clientTypes.ConvertToGrantEnum(data.Grant.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create team deployment grant, got error: %s", err))
		return
	}

	codeLocationGrants := make([]clientSchema.LocationScopedGrant, 0)

	if !data.CodeLocationGrants.IsNull() {
		for _, codeLocationGrant := range data.CodeLocationGrants.Elements() {
			codeLocationGrantObject := codeLocationGrant.(types.Object)
			attributes := codeLocationGrantObject.Attributes()

			grant, err := clientTypes.ConvertToGrantEnum(attributes["grant"].(types.String).ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create team deployment grant, got error: %s", err))
				return
			}

			codeLocationGrants = append(
				codeLocationGrants,
				clientSchema.LocationScopedGrant{
					LocationName: attributes["name"].(types.String).ValueString(),
					Grant:        grant,
				},
			)
		}
	}

	teamDeploymentGrant, err := r.client.TeamsClient.CreateOrUpdateTeamDeploymentGrant(
		ctx,
		data.TeamId.ValueString(),
		int(data.DeploymentId.ValueInt64()),
		grantEnum,
		codeLocationGrants,
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create team deployment grant, got error: %s", err))
		return
	}

	data.Grant = types.StringValue(string(teamDeploymentGrant.Grant))
	data.Id = types.Int64Value(int64(teamDeploymentGrant.Id))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TeamDeploymentGrantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TeamDeploymentGrantResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.TeamsClient.RemoveTeamDeploymentGrant(
		ctx,
		data.TeamId.ValueString(),
		int(data.DeploymentId.ValueInt64()),
	)
	if err != nil {
		var errComp *clientTypes.ErrNotFound
		if errors.As(err, &errComp) {
			tflog.Trace(ctx, "Team Deployment Grant not found, probably already deleted manually, removing from state")
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete team deployment grant, got error: %s", err))
		}
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted team deployment grant resource with TeamId=%s DeploymentId=%d", data.TeamId.ValueString(), data.DeploymentId.ValueInt64()))
}
