package resources

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/datarootsio/terraform-provider-dagster/internal/client"
	clientTypes "github.com/datarootsio/terraform-provider-dagster/internal/client/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &DeploymentResource{}
var _ resource.ResourceWithImportState = &DeploymentResource{}
var _ resource.ResourceWithValidateConfig = &DeploymentResource{}

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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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
				Required:            false,
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Deployment settings as a JSON document",
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
	settingsStr, _ := json.MarshalIndent(settings, "", "  ")

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

	// Update state with new values
	settingsStr, _ := json.MarshalIndent(deployment.DeploymentSettings.Settings, "", "  ")

	data.Settings = types.StringValue(string(settingsStr))
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

	// Todo: can we use DeploymentId from state? TF marks it as unknown = 0. Why?
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
	settings, err := r.client.DeploymentClient.SetDeploymentSettings(ctx, int(deploy.DeploymentId), deploymentSettings)
	if err != nil {
		tflog.Trace(ctx, fmt.Sprintf("Unable to set deployment settings: %s", err.Error()))
	}

	// Store everything in state
	settingsStr, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		resp.Diagnostics.AddError(
			"Client error", fmt.Sprintf("Unable to marshal new settings: %v", string(settings)),
		)
	}

	plan.Id = types.Int64Value(int64(deploy.DeploymentId))
	plan.Status = types.StringValue(string(deploy.DeploymentStatus))
	plan.Name = types.StringValue(deploy.DeploymentName)
	plan.Type = types.StringValue(string(deploy.DeploymentType))
	plan.Settings = types.StringValue(string(settingsStr))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
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

func (r *DeploymentResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data DeploymentResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Check whether JSON is correct formatted
	settings := data.Settings.ValueString()

	if strings.TrimSpace(settings) != settings {
		resp.Diagnostics.AddAttributeError(
			path.Root("settings_document"),
			"Validation error",
			"settings_document has leading or trailing whitespace. Use trimspace() to remove it.",
		)
	}
	var settingsJSON map[string]interface{}
	err := json.Unmarshal([]byte(settings), &settingsJSON)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("settings_document"),
			"Validation error",
			"Unable to marshal settings_document. Make sure it is valid JSON",
		)
	}

	_, err = json.MarshalIndent(settingsJSON, "", "  ")
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("settings_document"),
			"Validation error",
			"Unable to check formatting. Make sure it is valid JSON.",
		)
	}
}

func (r *DeploymentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
