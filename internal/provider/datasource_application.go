package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marconneves/terraform-provider-dokploy/internal/dokploy"
)

var _ datasource.DataSource = &ApplicationDataSource{}

func NewApplicationDataSource() datasource.DataSource {
	return &ApplicationDataSource{}
}

type ApplicationDataSource struct {
	client *dokploy.Client
}

type ApplicationDataSourceModel struct {
	ID                 types.String `tfsdk:"id"`
	ProjectID          types.String `tfsdk:"project_id"`
	EnvironmentID      types.String `tfsdk:"environment_id"`
	Name               types.String `tfsdk:"name"`
	RepositoryURL      types.String `tfsdk:"repository_url"`
	Branch             types.String `tfsdk:"branch"`
	BuildType          types.String `tfsdk:"build_type"`
	DockerfilePath     types.String `tfsdk:"dockerfile_path"`
	DockerContextPath  types.String `tfsdk:"docker_context_path"`
	DockerBuildStage   types.String `tfsdk:"docker_build_stage"`
	CustomGitUrl       types.String `tfsdk:"custom_git_url"`
	CustomGitBranch    types.String `tfsdk:"custom_git_branch"`
	CustomGitSSHKeyID  types.String `tfsdk:"custom_git_ssh_key_id"`
	CustomGitBuildPath types.String `tfsdk:"custom_git_build_path"`
	SourceType         types.String `tfsdk:"source_type"`
	Username           types.String `tfsdk:"username"`
	AutoDeploy         types.Bool   `tfsdk:"auto_deploy"`
	// GitHub Provider fields
	GithubRepository types.String `tfsdk:"github_repository"`
	GithubOwner      types.String `tfsdk:"github_owner"`
	GithubBranch     types.String `tfsdk:"github_branch"`
	GithubBuildPath  types.String `tfsdk:"github_build_path"`
	GithubID         types.String `tfsdk:"github_id"`
	GithubWatchPaths types.List   `tfsdk:"github_watch_paths"`
	EnableSubmodules types.Bool   `tfsdk:"enable_submodules"`
	TriggerType      types.String `tfsdk:"trigger_type"`
}

func (d *ApplicationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (d *ApplicationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
			"repository_url": schema.StringAttribute{
				Computed: true,
			},
			"branch": schema.StringAttribute{
				Computed: true,
			},
			"build_type": schema.StringAttribute{
				Computed: true,
			},
			"dockerfile_path": schema.StringAttribute{
				Computed: true,
			},
			"docker_context_path": schema.StringAttribute{
				Computed: true,
			},
			"docker_build_stage": schema.StringAttribute{
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
			"custom_git_build_path": schema.StringAttribute{
				Computed: true,
			},
			"source_type": schema.StringAttribute{
				Computed: true,
			},
			"username": schema.StringAttribute{
				Computed: true,
			},
			"auto_deploy": schema.BoolAttribute{
				Computed: true,
			},
			"github_repository": schema.StringAttribute{
				Computed: true,
			},
			"github_owner": schema.StringAttribute{
				Computed: true,
			},
			"github_branch": schema.StringAttribute{
				Computed: true,
			},
			"github_build_path": schema.StringAttribute{
				Computed: true,
			},
			"github_id": schema.StringAttribute{
				Computed: true,
			},
			"github_watch_paths": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"enable_submodules": schema.BoolAttribute{
				Computed: true,
			},
			"trigger_type": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *ApplicationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*dokploy.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Type", fmt.Sprintf("Expected *dokploy.Client, got: %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *ApplicationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ApplicationDataSourceModel
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
		app, err := d.client.GetApplication(state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error reading application", err.Error())
			return
		}
		d.populateState(&state, app, ctx)
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
			for _, app := range env.Applications {
				if app.Name == targetName {
					d.populateState(&state, &app, ctx)
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
				resp.Diagnostics.AddError("Application not found", fmt.Sprintf("Application '%s' not found in project '%s' and environment '%s'", targetName, state.ProjectID.ValueString(), targetEnvID))
			} else {
				resp.Diagnostics.AddError("Application not found", fmt.Sprintf("Application '%s' not found in project '%s'", targetName, state.ProjectID.ValueString()))
			}
			return
		}
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (d *ApplicationDataSource) populateState(state *ApplicationDataSourceModel, app *dokploy.Application, ctx context.Context) {
	state.ID = types.StringValue(app.ID)
	state.Name = types.StringValue(app.Name)
	state.ProjectID = types.StringValue(app.ProjectID)
	state.EnvironmentID = types.StringValue(app.EnvironmentID)
	state.RepositoryURL = types.StringValue(app.RepositoryURL)
	state.Branch = types.StringValue(app.Branch)
	state.BuildType = types.StringValue(app.BuildType)
	state.DockerfilePath = types.StringValue(app.DockerfilePath)
	state.DockerContextPath = types.StringValue(app.DockerContextPath)
	state.DockerBuildStage = types.StringValue(app.DockerBuildStage)
	state.CustomGitUrl = types.StringValue(app.CustomGitUrl)
	state.CustomGitBranch = types.StringValue(app.CustomGitBranch)
	state.CustomGitSSHKeyID = types.StringValue(app.CustomGitSSHKeyId)
	state.CustomGitBuildPath = types.StringValue(app.CustomGitBuildPath)
	state.SourceType = types.StringValue(app.SourceType)
	state.Username = types.StringValue(app.Username)
	state.AutoDeploy = types.BoolValue(app.AutoDeploy)
	state.GithubRepository = types.StringValue(app.GithubRepository)
	state.GithubOwner = types.StringValue(app.GithubOwner)
	state.GithubBranch = types.StringValue(app.GithubBranch)
	state.GithubBuildPath = types.StringValue(app.GithubBuildPath)
	state.GithubID = types.StringValue(app.GithubID)

	if len(app.GithubWatchPaths) > 0 {
		list, _ := types.ListValueFrom(ctx, types.StringType, app.GithubWatchPaths)
		state.GithubWatchPaths = list
	} else {
		state.GithubWatchPaths = types.ListNull(types.StringType)
	}

	state.EnableSubmodules = types.BoolValue(app.EnableSubmodules)
	state.TriggerType = types.StringValue(app.TriggerType)
}
