package resources

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/datarootsio/terraform-provider-dagster/internal/client"
	clientTypes "github.com/datarootsio/terraform-provider-dagster/internal/client/types"
	"github.com/datarootsio/terraform-provider-dagster/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &DeploymentResource{}
	_ resource.ResourceWithImportState = &DeploymentResource{}
)

func NewDeploymentResource() resource.Resource {
	return &DeploymentResource{}
}

type DeploymentResource struct {
	client client.DagsterClient
}

type DeploymentResourceModel struct {
	Name         types.String `tfsdk:"name"`
	Id           types.Int64  `tfsdk:"id"`
	Status       types.String `tfsdk:"status"`
	Type         types.String `tfsdk:"type"`
	Settings     types.String `tfsdk:"settings_document"`
	ForceDestroy types.Bool   `tfsdk:"force_destroy"`
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Deployment id",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Deployment status (`ACTIVE` or `PENDNG_DELETION`)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Deployment type (`PRODUCTION`, `DEV` or `BRANCH`)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"settings_document": schema.StringAttribute{
				Required:            false,
				Computed:            true,
				Optional:            true,
				Default:             stringdefault.StaticString("{}"),
				MarkdownDescription: "Deployment settings as a JSON document. We recommend using a `dagster_configuration_document` to generate this instead of composing a JSON document yourself. Leaving this attribute empty or partially filled in, will result in Dagster (partially) applying default settings to your deployment. This leads to perpetual changes in this resource.",
			},
			"force_destroy": schema.BoolAttribute{
				Required:    false,
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "When `false`, will check if there are code locations associated with the deployment, if there are, it will block the delete of the deployment. When `true` ignore the code locations check. This is done because when you delete a deployment, you delete all the resources/metadata of that deployment and this is not recoverable. DEFAULT `false`",
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

	// Check if deployment already exists
	_, err := r.client.DeploymentClient.GetDeploymentByName(ctx, data.Name.ValueString())
	if err != nil {
		var errComp *clientTypes.ErrNotFound
		if errors.As(err, &errComp) {
			// Deployment does not exist, continue
		} else {
			resp.Diagnostics.AddError(
				"Error reading deployment",
				err.Error(),
			)
		}
	} else {
		resp.Diagnostics.AddError(
			"Client error",
			fmt.Sprintf("Deployment with name %v already exists", data.Name.ValueString()),
		)
	}

	// Create deployment
	deployment, err := r.client.DeploymentClient.CreateHybridDeployment(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating deployment",
			err.Error(),
		)
	}

	// Apply settings
	settings, err := r.client.DeploymentClient.SetDeploymentSettings(ctx, deployment.DeploymentId, json.RawMessage(data.Settings.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error setting deployment settings",
			err.Error(),
		)
	}

	// Unmarshal+Marshal settings result to make sure it's uniform
	settingsStr, err := utils.MakeJSONStringUniform(settings)
	if err != nil {
		resp.Diagnostics.AddError(
			"JSON Format error",
			fmt.Sprintf("Trying to parse JSON: %s: %s", settings, err.Error()),
		)
	}

	data.Name = types.StringValue(deployment.DeploymentName)
	data.Id = types.Int64Value(int64(deployment.DeploymentId))
	data.Status = types.StringValue(string(deployment.DeploymentStatus))
	data.Type = types.StringValue(string(deployment.DeploymentType))
	data.Settings = types.StringValue(settingsStr)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DeploymentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if deployment still exists
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

	settings, err := utils.MakeJSONStringUniform(deployment.DeploymentSettings.Settings)
	if err != nil {
		resp.Diagnostics.AddError(
			"JSON Format error",
			fmt.Sprintf("Trying to parse JSON: %s: %s", settings, err.Error()),
		)
	}

	data.Settings = types.StringValue(settings)
	data.Name = types.StringValue(deployment.DeploymentName)
	data.Id = types.Int64Value(int64(deployment.DeploymentId))
	data.Status = types.StringValue(string(deployment.DeploymentStatus))
	data.Type = types.StringValue(string(deployment.DeploymentType))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DeploymentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deploymentName := plan.Name.ValueString()
	deploymentSettings := json.RawMessage(plan.Settings.ValueString())
	deploy, err := r.client.DeploymentClient.GetDeploymentByName(ctx, deploymentName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client error",
			fmt.Sprintf("Error while fetching deployment with name %v", deploymentName),
		)
	}

	// Set new settings
	tflog.Trace(ctx, fmt.Sprintf("Applying settings to deployment %v: %v", deploymentName, string(deploymentSettings)))
	settings, err := r.client.DeploymentClient.SetDeploymentSettings(ctx, deploy.DeploymentId, deploymentSettings)
	if err != nil {
		tflog.Trace(ctx, fmt.Sprintf("Unable to set deployment settings: %s", err.Error()))
	}

	settingsStr, err := utils.MakeJSONStringUniform(settings)
	if err != nil {
		resp.Diagnostics.AddError(
			"JSON Format error",
			fmt.Sprintf("Trying to parse JSON: %s: %s", settings, err.Error()),
		)
	}

	plan.Id = types.Int64Value(int64(deploy.DeploymentId))
	plan.Status = types.StringValue(string(deploy.DeploymentStatus))
	plan.Name = types.StringValue(deploy.DeploymentName)
	plan.Type = types.StringValue(string(deploy.DeploymentType))
	plan.Settings = types.StringValue(settingsStr)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DeploymentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DeploymentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Skip this check in tests, where we create temporary deployments
	if os.Getenv("TF_ACC") == "" && data.Name.ValueString() != r.client.Deployment {
		resp.Diagnostics.AddError("Client Error", "Unable to delete deployment, got error: can't delete deployment with a client configured with another deployment.")
		return
	}

	codeLocations, err := r.client.CodeLocationsClient.ListCodeLocations(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete deployment, got error: %s", err))
		return
	}

	if !data.ForceDestroy.ValueBool() && len(codeLocations) > 0 {
		resp.Diagnostics.AddError("Client Error", "Unable to delete deployment, got error: deployment still contains code locations, use the `force_destroy` flag if you want to force delete it.")
		return
	}

	err = r.client.DeploymentClient.DeleteDeployment(ctx, int(data.Id.ValueInt64()))
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
