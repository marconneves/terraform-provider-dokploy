package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marconneves/terraform-provider-dokploy/internal/client"
)

var _ datasource.DataSource = &EnvironmentDataSource{}

func NewEnvironmentDataSource() datasource.DataSource {
	return &EnvironmentDataSource{}
}

type EnvironmentDataSource struct {
	client *client.DokployClient
}

type EnvironmentDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	ProjectID   types.String `tfsdk:"project_id"`
}

func (d *EnvironmentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (d *EnvironmentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"name": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"description": schema.StringAttribute{
				Computed: true,
			},
			"project_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
		},
	}
}

func (d *EnvironmentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *EnvironmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state EnvironmentDataSourceModel
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
		env, err := d.client.GetEnvironment(state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error reading environment", err.Error())
			return
		}

		state.Name = types.StringValue(env.Name)
		state.Description = types.StringValue(env.Description)
		state.ProjectID = types.StringValue(env.ProjectID)
	} else {
		// Lookup by name within project
		project, err := d.client.GetProject(state.ProjectID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error fetching project", err.Error())
			return
		}

		found := false
		targetName := state.Name.ValueString()
		for _, e := range project.Environments {
			if e.Name == targetName {
				state.ID = types.StringValue(e.ID)
				state.Description = types.StringValue(e.Description)
				found = true
				break
			}
		}

		if !found {
			resp.Diagnostics.AddError("Environment not found", fmt.Sprintf("Environment '%s' not found in project '%s'", targetName, state.ProjectID.ValueString()))
			return
		}
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
