package resources

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"github.com/datarootsio/terraform-provider-dagster/internal/client"
	clientTypes "github.com/datarootsio/terraform-provider-dagster/internal/client/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &DeploymentResource{}
var _ resource.ResourceWithImportState = &DeploymentResource{}

func NewDeploymentResource() resource.Resource {
	return &DeploymentResource{}
}

type DeploymentResource struct {
	client client.DagsterClient
}

type DeploymentResourceModel struct {
	Name     types.String `tfsdk:"name"`
	Id       types.Int64  `tfsdk:"id"`
	Status   types.String `tfsdk:"status"`
	Type     types.String `tfsdk:"type"`
	Settings types.String `tfsdk:"settings_document"`
}

func (r *DeploymentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment"
}

func (r *DeploymentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Deployment resource",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Deployment name",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-zA-Z0-9-]+$`),
						"Deployment name must contain only A-Z, a-z, 0-9 or -",
					),
				},
			},
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Deployment id",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Deployment status (`ACTIVE` or `PENDNG_DELETION`)",
			},
			"type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Deployment type (`PRODUCTION`, `DEV` or `BRANCH`)",
			},
			"settings_document": schema.StringAttribute{
				Required:            true,
				Computed:            false,
				Optional:            false,
				MarkdownDescription: "Deployment settings as a YAML document",
			},
		},
	}
}

func (r *DeploymentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(client.DagsterClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Configuration",
			fmt.Sprintf("Expected client.DagsterClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *DeploymentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DeploymentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	deployment, err := r.client.DeploymentClient.CreateHybridDeployment(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating deployment",
			err.Error(),
		)
	}

	var settings map[string]interface{}
	err = json.Unmarshal([]byte(data.Settings.ValueString()), &settings)
	settingsStr, _ := json.Marshal(settings)
	r.client.DeploymentClient.SetDeploymentSettings(ctx, deployment.DeploymentId, json.RawMessage(settingsStr))

	data.Name = types.StringValue(deployment.DeploymentName)
	data.Id = types.Int64Value(int64(deployment.DeploymentId))
	data.Status = types.StringValue(string(deployment.DeploymentStatus))
	data.Type = types.StringValue(string(deployment.DeploymentType))
	data.Settings = types.StringValue(string(settingsStr))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DeploymentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	deployment, err := r.client.DeploymentClient.GetDeploymentByName(ctx, data.Name.ValueString())
	if err != nil {
		var errComp *clientTypes.ErrNotFound
		if errors.As(err, &errComp) {
			tflog.Trace(ctx, "Deployment not found, probably already deleted, removing from state")
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Error reading deployment",
				err.Error(),
			)
		}
		return
	}

	var settings map[string]interface{}
	err = json.Unmarshal([]byte(data.Settings.ValueString()), &settings)
	settingsStr, _ := json.Marshal(settings)
	r.client.DeploymentClient.SetDeploymentSettings(ctx, deployment.DeploymentId, json.RawMessage(settingsStr))

	data.Name = types.StringValue(deployment.DeploymentName)
	data.Id = types.Int64Value(int64(deployment.DeploymentId))
	data.Status = types.StringValue(string(deployment.DeploymentStatus))
	data.Type = types.StringValue(string(deployment.DeploymentType))
	data.Settings = types.StringValue(string(settingsStr))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DeploymentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeploymentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DeploymentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeploymentClient.DeleteDeployment(ctx, int(data.Id.ValueInt64()))
	if err != nil {
		var errComp *clientTypes.ErrNotFound
		if errors.As(err, &errComp) {
			tflog.Trace(ctx, "Deployment not found, probably already deleted manually, removing from state")
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete deployment, got error: %s", err))
		}
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted deployment %s with id: %v", data.Name.ValueString(), data.Id.ValueInt64()))
}

func (r *DeploymentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
