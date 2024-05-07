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
	Name             types.String `tfsdk:"name"`
	Image            types.String `tfsdk:"image"`
	CodeSource       types.Object `tfsdk:"code_source"`
	WorkingDirectory types.String `tfsdk:"working_directory"`
	ExecutablePath   types.String `tfsdk:"executable_path"`
	Attribute        types.String `tfsdk:"attribute"`
	Git              types.Object `tfsdk:"git"`
	AgentQueue       types.String `tfsdk:"agent_queue"`
}

func (r *CodeLocationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_code_location"
}

func (r *CodeLocationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Code Location resource",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Code Location name. ",
				Required:            true,
			},
			"image": schema.StringAttribute{
				MarkdownDescription: "Code Location image. Git or Image is a required field (mutually exclusive).",
				Required:            false,
				Optional:            true,
			},
			"code_source": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"module_name": schema.StringAttribute{
						MarkdownDescription: "Code Location code source module name. One of module_name/package_name/python_file is required (mutually exclusive).",
						Required:            false,
						Optional:            true,
					},
					"package_name": schema.StringAttribute{
						MarkdownDescription: "Code Location code source package name. One of module_name/package_name/python_file is required (mutually exclusive).",
						Required:            false,
						Optional:            true,
					},
					"python_file": schema.StringAttribute{
						MarkdownDescription: "Code Location code source python file. One of module_name/package_name/python_file is required (mutually exclusive).",
						Required:            false,
						Optional:            true,
					},
				},
				MarkdownDescription: "Code Location code source",
				Required:            true,
			},
			"working_directory": schema.StringAttribute{
				MarkdownDescription: "Code Location working directory",
				Required:            false,
				Optional:            true,
			},
			"executable_path": schema.StringAttribute{
				MarkdownDescription: "Code Location executable path",
				Required:            false,
				Optional:            true,
			},
			"attribute": schema.StringAttribute{
				MarkdownDescription: "Code Location attribute",
				Required:            false,
				Optional:            true,
			},
			"git": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"commit_hash": schema.StringAttribute{
						MarkdownDescription: "Code Location git commit hash. If git is specified, `commit_hash` is required.",
						Required:            true,
					},
					"url": schema.StringAttribute{
						MarkdownDescription: "Code Location git URL. If git is specified, `url` is required.",
						Required:            true,
					},
				},
				MarkdownDescription: "Code Location git. Git or Image is a required field (mutually exclusive).",
				Required:            false,
				Optional:            true,
			},
			"agent_queue": schema.StringAttribute{
				MarkdownDescription: "Code Location agent queue",
				Required:            false,
				Optional:            true,
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

	// If one of the attributes is nil, it will convert to "" and
	// set the `ok` flag to false. We deliberately ignore the `ok` flag.
	// Some of these values might indeed be nil, but the dagster client
	// can probably handle them. So no need to handle the `ok` flag here.
	// If we don't catch the `ok` flag in a variable, program will panic
	// when converting nil to string. That's why we catch it with `_`.
	moduleName, _ := data.CodeSource.Attributes()["module_name"].(types.String)
	packageName, _ := data.CodeSource.Attributes()["package_name"].(types.String)
	pythonFile, _ := data.CodeSource.Attributes()["python_file"].(types.String)

	commitHash, _ := data.Git.Attributes()["commit_hash"].(types.String)
	url, _ := data.Git.Attributes()["url"].(types.String)

	err := r.client.CodeLocationsClient.AddCodeLocation(
		ctx,
		clientTypes.CodeLocation{
			Name:  data.Name.ValueString(),
			Image: data.Image.ValueString(),
			CodeSource: clientTypes.CodeLocationCodeSource{
				ModuleName:  moduleName.ValueString(),
				PackageName: packageName.ValueString(),
				PythonFile:  pythonFile.ValueString(),
			},
			WorkingDirectory: data.WorkingDirectory.ValueString(),
			ExecutablePath:   data.ExecutablePath.ValueString(),
			Attribute:        data.Attribute.ValueString(),
			Git: clientTypes.CodeLocationGit{
				CommitHash: commitHash.ValueString(),
				URL:        url.ValueString(),
			},
			AgentQueue: data.AgentQueue.ValueString(),
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create code locations, got error: %s", err))
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
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read code locations, got error: %s", err))
		}
		return
	}

	// Code source
	codeSourceAttributeTypes := map[string]attr.Type{
		"module_name":  types.StringType,
		"package_name": types.StringType,
		"python_file":  types.StringType,
	}
	codeSourceAttributeValues := map[string]attr.Value{
		"module_name":  types.StringValue(codeLocation.CodeSource.ModuleName),
		"package_name": types.StringValue(codeLocation.CodeSource.PackageName),
		"python_file":  types.StringValue(codeLocation.CodeSource.PythonFile),
	}
	codeSource, diag := types.ObjectValue(codeSourceAttributeTypes, codeSourceAttributeValues)
	resp.Diagnostics.Append(diag...)

	if resp.Diagnostics.HasError() {
		return
	}

	//Git
	gitAttributeTypes := map[string]attr.Type{
		"commit_hash": types.StringType,
		"url":         types.StringType,
	}
	gitSourceAttributeValues := map[string]attr.Value{
		"commit_hash": types.StringValue(codeLocation.Git.CommitHash),
		"url":         types.StringValue(codeLocation.Git.URL),
	}
	git, diag := types.ObjectValue(gitAttributeTypes, gitSourceAttributeValues)
	resp.Diagnostics.Append(diag...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Image = types.StringValue(codeLocation.Image)
	data.CodeSource = codeSource
	data.WorkingDirectory = types.StringValue(codeLocation.WorkingDirectory)
	data.ExecutablePath = types.StringValue(codeLocation.ExecutablePath)
	data.Attribute = types.StringValue(codeLocation.Attribute)
	data.Git = git
	data.AgentQueue = types.StringValue(codeLocation.AgentQueue)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CodeLocationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CodeLocationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If one of the attributes is nil, it will convert to "" and
	// set the `ok` flag to false. We deliberately ignore the `ok` flag.
	// Some of these values might indeed be nil, but the dagster client
	// can probably handle them. So no need to handle the `ok` flag here.
	// If we don't catch the `ok` flag in a variable, program will panic
	// when converting nil to string. That's why we catch it with `_`.
	moduleName, _ := data.CodeSource.Attributes()["module_name"].(types.String)
	packageName, _ := data.CodeSource.Attributes()["package_name"].(types.String)
	pythonFile, _ := data.CodeSource.Attributes()["python_file"].(types.String)

	commitHash, _ := data.Git.Attributes()["commit_hash"].(types.String)
	url, _ := data.Git.Attributes()["url"].(types.String)

	err := r.client.CodeLocationsClient.UpdateCodeLocation(
		ctx,
		clientTypes.CodeLocation{
			Name:  data.Name.ValueString(),
			Image: data.Image.ValueString(),
			CodeSource: clientTypes.CodeLocationCodeSource{
				ModuleName:  moduleName.ValueString(),
				PackageName: packageName.ValueString(),
				PythonFile:  pythonFile.ValueString(),
			},
			WorkingDirectory: data.WorkingDirectory.ValueString(),
			ExecutablePath:   data.ExecutablePath.ValueString(),
			Attribute:        data.Attribute.ValueString(),
			Git: clientTypes.CodeLocationGit{
				CommitHash: commitHash.ValueString(),
				URL:        url.ValueString(),
			},
			AgentQueue: data.AgentQueue.ValueString(),
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create code locations, got error: %s", err))
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
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete code locations, got error: %s", err))
		}
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted code locations resource with id: %s", data.Name.ValueString()))
}
