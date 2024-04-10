package resources

import (
	"context"
	"errors"
	"fmt"

	"github.com/datarootsio/terraform-provider-dagster/internal/client"
	clientTypes "github.com/datarootsio/terraform-provider-dagster/internal/client/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &CodeLocationResource{}

func NewCodeLocationResource() resource.Resource {
	return &CodeLocationResource{}
}

type CodeLocationResource struct {
	client client.DagsterClient
}

type CodeLocationResourceModel struct {
	Name       types.String `tfsdk:"name"`
	Image      types.String `tfsdk:"image"`
	CodeSource types.Object `tfsdk:"code_source"`
}

func (r *CodeLocationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_code_location"
}

func (r *CodeLocationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Code Location resource",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Code Location name",
				Required:            true,
			},
			"image": schema.StringAttribute{
				MarkdownDescription: "Code Location image",
				Required:            true,
			},
			"code_source": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"python_file": schema.StringAttribute{
						MarkdownDescription: "Code Location code source python file",
						Required:            true,
					},
				},
				MarkdownDescription: "Code Location code source",
				Required:            true,
			},
		},
	}
}

func (r *CodeLocationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CodeLocationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CodeLocationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.CodeLocationsClient.AddCodeLocation(
		ctx,
		clientTypes.CodeLocation{
			Name:  data.Name.ValueString(),
			Image: data.Image.ValueString(),
			CodeSource: clientTypes.CodeLocationCodeSource{
				PythonFile: data.CodeSource.Attributes()["python_file"].String(),
			},
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create team, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "created code location resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CodeLocationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CodeLocationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	codeLocation, err := r.client.CodeLocationsClient.GetCodeLocationByName(ctx, data.Name.ValueString())

	if err != nil {
		var errComp *clientTypes.ErrNotFound
		if errors.As(err, &errComp) {
			tflog.Trace(ctx, "Code Location not found, probably already deleted manually, removing from state")
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read team, got error: %s", err))
		}
		return
	}

	attributeTypes := map[string]attr.Type{
		"python_file": types.StringType,
	}
	attributeValues := map[string]attr.Value{
		"python_file": types.StringValue(codeLocation.CodeSource.PythonFile),
	}
	codeSource, diag := types.ObjectValue(attributeTypes, attributeValues)
	resp.Diagnostics.Append(diag...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Image = types.StringValue(codeLocation.Image)
	data.CodeSource = codeSource

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CodeLocationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CodeLocationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.CodeLocationsClient.UpdateCodeLocation(
		ctx,
		clientTypes.CodeLocation{
			Name:  data.Name.ValueString(),
			Image: data.Image.ValueString(),
			CodeSource: clientTypes.CodeLocationCodeSource{
				PythonFile: data.CodeSource.Attributes()["python_file"].String(),
			},
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create team, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "created code location resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CodeLocationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CodeLocationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.CodeLocationsClient.DeleteCodeLocation(ctx, data.Name.ValueString())
	if err != nil {
		var errComp *clientTypes.ErrNotFound
		if errors.As(err, &errComp) {
			tflog.Trace(ctx, "Code Location not found, probably already deleted manually, removing from state")
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete team, got error: %s", err))
		}
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted team resource with id: %s", data.Name.ValueString()))
}
