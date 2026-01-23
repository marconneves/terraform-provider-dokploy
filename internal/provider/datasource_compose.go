package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marconneves/terraform-provider-dokploy/internal/client"
)

var _ datasource.DataSource = &ComposeDataSource{}

func NewComposeDataSource() datasource.DataSource {
	return &ComposeDataSource{}
}

type ComposeDataSource struct {
	client *client.DokployClient
}

type ComposeDataSourceModel struct {
	ID                 types.String `tfsdk:"id"`
	ProjectID          types.String `tfsdk:"project_id"`
	EnvironmentID      types.String `tfsdk:"environment_id"`
	Name               types.String `tfsdk:"name"`
	ComposeFileContent types.String `tfsdk:"compose_file_content"`
	SourceType         types.String `tfsdk:"source_type"`
	CustomGitUrl       types.String `tfsdk:"custom_git_url"`
	CustomGitBranch    types.String `tfsdk:"custom_git_branch"`
	CustomGitSSHKeyID  types.String `tfsdk:"custom_git_ssh_key_id"`
	ComposePath        types.String `tfsdk:"compose_path"`
	AutoDeploy         types.Bool   `tfsdk:"auto_deploy"`
}

func (d *ComposeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_compose"
}

func (d *ComposeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"project_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"environment_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"name": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"compose_file_content": schema.StringAttribute{
				Computed: true,
			},
			"source_type": schema.StringAttribute{
				Computed: true,
			},
			"custom_git_url": schema.StringAttribute{
				Computed: true,
			},
			"custom_git_branch": schema.StringAttribute{
				Computed: true,
			},
			"custom_git_ssh_key_id": schema.StringAttribute{
				Computed: true,
			},
			"compose_path": schema.StringAttribute{
				Computed: true,
			},
			"auto_deploy": schema.BoolAttribute{
				Computed: true,
			},
		},
	}
}

func (d *ComposeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*client.DokployClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Type", fmt.Sprintf("Expected *client.DokployClient, got: %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *ComposeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ComposeDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ID.IsNull() && (state.Name.IsNull() || state.ProjectID.IsNull()) {
		resp.Diagnostics.AddError("Missing criteria", "Either 'id' or ('name' and 'project_id') must be provided.")
		return
	}

	if !state.ID.IsNull() {
		comp, err := d.client.GetCompose(state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error reading compose", err.Error())
			return
		}
		d.populateState(&state, comp)
	} else {
		// Lookup by name
		project, err := d.client.GetProject(state.ProjectID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error fetching project", err.Error())
			return
		}

		found := false
		targetName := state.Name.ValueString()
		targetEnvID := state.EnvironmentID.ValueString()

		for _, env := range project.Environments {
			if targetEnvID != "" && env.ID != targetEnvID {
				continue
			}
			for _, comp := range env.Composes {
				if comp.Name == targetName {
					d.populateState(&state, &comp)
					found = true
					break
				}
			}
			if found {
				break
			}
		}

		if !found {
			if targetEnvID != "" {
				resp.Diagnostics.AddError("Compose stack not found", fmt.Sprintf("Compose stack '%s' not found in project '%s' and environment '%s'", targetName, state.ProjectID.ValueString(), targetEnvID))
			} else {
				resp.Diagnostics.AddError("Compose stack not found", fmt.Sprintf("Compose stack '%s' not found in project '%s'", targetName, state.ProjectID.ValueString()))
			}
			return
		}
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (d *ComposeDataSource) populateState(state *ComposeDataSourceModel, comp *client.Compose) {
	state.ID = types.StringValue(comp.ID)
	state.Name = types.StringValue(comp.Name)
	state.ProjectID = types.StringValue(comp.ProjectID)
	state.EnvironmentID = types.StringValue(comp.EnvironmentID)
	state.ComposeFileContent = types.StringValue(comp.ComposeFile)
	state.SourceType = types.StringValue(comp.SourceType)
	state.CustomGitUrl = types.StringValue(comp.CustomGitUrl)
	state.CustomGitBranch = types.StringValue(comp.CustomGitBranch)
	state.CustomGitSSHKeyID = types.StringValue(comp.CustomGitSSHKeyId)
	state.ComposePath = types.StringValue(comp.ComposePath)
	state.AutoDeploy = types.BoolValue(comp.AutoDeploy)
}
