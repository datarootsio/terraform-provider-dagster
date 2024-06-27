package resources

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/datarootsio/terraform-provider-dagster/internal/client"
	"github.com/datarootsio/terraform-provider-dagster/internal/client/service"
	clientTypes "github.com/datarootsio/terraform-provider-dagster/internal/client/types"
	"github.com/datarootsio/terraform-provider-dagster/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &CodeLocationAsDocumentResource{}

func NewCodeLocationAsDocumentResource() resource.Resource {
	return &CodeLocationAsDocumentResource{}
}

type CodeLocationAsDocumentResource struct {
	client client.DagsterClient
}

type CodeLocationAsDocumentResourceModel struct {
	Document types.String `tfsdk:"document"`
	Name     types.String `tfsdk:"name"`
}

func (r *CodeLocationAsDocumentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_code_location_as_document"
}

func (r *CodeLocationAsDocumentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates a code location from a dagster configuration document.",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Code location name",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"document": schema.StringAttribute{
				Required:            false,
				Computed:            true,
				Optional:            true,
				Default:             stringdefault.StaticString("{}"),
				MarkdownDescription: "Code location as a JSON document. We recommend using a `dagster_configuration_document` to generate this instead of composing a JSON document yourself.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(
						replaceIfCodeLocationNameChanges,
						"replace code location when name changes",
						"replace code location when name changes",
					),
				},
			},
		},
	}
}

func replaceIfCodeLocationNameChanges(ctx context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
	stateName, err := service.GetCodeLocationNameFromDocument(json.RawMessage(req.StateValue.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Plan modifier got error: %s", err))
		return
	}

	planName, err := service.GetCodeLocationNameFromDocument(json.RawMessage(req.PlanValue.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Plan modifier got error: %s", err))
		return
	}

	resp.RequiresReplace = stateName != planName
}

func (r *CodeLocationAsDocumentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CodeLocationAsDocumentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CodeLocationAsDocumentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	document := json.RawMessage(data.Document.ValueString())
	codeLocationName, err := service.GetCodeLocationNameFromDocument(document)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create code location, got error: %s", err))
		return
	}

	err = r.client.CodeLocationsClient.AddCodeLocationAsDocument(ctx, document)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create code location, got error: %s", err))
		return
	}

	// Unmarshal+Marshal settings result to make sure it's uniform
	documentString, err := utils.MakeJSONStringUniform(document)
	if err != nil {
		resp.Diagnostics.AddError(
			"JSON Format error",
			fmt.Sprintf("Trying to parse JSON: %s: %s", document, err.Error()),
		)
	}

	data.Name = types.StringValue(codeLocationName)
	data.Document = types.StringValue(documentString)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CodeLocationAsDocumentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CodeLocationAsDocumentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	document := json.RawMessage(data.Document.ValueString())
	codeLocationName, err := service.GetCodeLocationNameFromDocument(document)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read code location, got error: %s", err))
		return
	}

	codeLocationAsDocument, err := r.client.CodeLocationsClient.GetCodeLocationAsDocumentByName(ctx, codeLocationName)
	if err != nil {
		var errComp *clientTypes.ErrNotFound
		if errors.As(err, &errComp) {
			tflog.Trace(ctx, "Code Location as Document not found, probably already deleted manually, removing from state")
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read code location, got error: %s", err))
		}
		return
	}

	documentString, err := utils.MakeJSONStringUniform(codeLocationAsDocument)
	if err != nil {
		resp.Diagnostics.AddError(
			"JSON Format error",
			fmt.Sprintf("Trying to parse JSON: %s: %s", document, err.Error()),
		)
	}

	data.Name = types.StringValue(codeLocationName)
	data.Document = types.StringValue(documentString)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CodeLocationAsDocumentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CodeLocationAsDocumentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	document := json.RawMessage(data.Document.ValueString())
	codeLocationName, err := service.GetCodeLocationNameFromDocument(document)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update code location, got error: %s", err))
		return
	}

	err = r.client.CodeLocationsClient.UpdateCodeLocationAsDocument(
		ctx,
		document,
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update code location, got error: %s", err))
		return
	}

	documentString, err := utils.MakeJSONStringUniform(document)
	if err != nil {
		resp.Diagnostics.AddError(
			"JSON Format error",
			fmt.Sprintf("Trying to parse JSON: %s: %s", document, err.Error()),
		)
	}

	data.Name = types.StringValue(codeLocationName)
	data.Document = types.StringValue(documentString)

	tflog.Trace(ctx, "updated code location resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CodeLocationAsDocumentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CodeLocationAsDocumentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	document := json.RawMessage(data.Document.ValueString())
	codeLocationName, err := service.GetCodeLocationNameFromDocument(document)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete code location, got error: %s", err))
		return
	}

	err = r.client.CodeLocationsClient.DeleteCodeLocation(ctx, codeLocationName)
	if err != nil {
		var errComp *clientTypes.ErrNotFound
		if errors.As(err, &errComp) {
			tflog.Trace(ctx, "Code Location not found, probably already deleted manually, removing from state")
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete code location, got error: %s", err))
		}
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted code location resource with id: %s", data.Name.ValueString()))
}
