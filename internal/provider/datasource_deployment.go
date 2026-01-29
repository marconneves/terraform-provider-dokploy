package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marconneves/terraform-provider-dokploy/internal/dokploy"
)

var _ datasource.DataSource = &DeploymentDataSource{}

func NewDeploymentDataSource() datasource.DataSource {
	return &DeploymentDataSource{}
}

type DeploymentDataSource struct {
	client *dokploy.Client
}

type DeploymentDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	ApplicationID types.String `tfsdk:"application_id"`
	Status        types.String `tfsdk:"status"`
	CreatedAt     types.String `tfsdk:"created_at"`
	Log           types.String `tfsdk:"log"`
}

func (d *DeploymentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment"
}

func (d *DeploymentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required: true,
			},
			"application_id": schema.StringAttribute{
				Computed: true,
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
			"log": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *DeploymentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DeploymentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state DeploymentDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deployment, err := d.client.GetDeployment(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading deployment", err.Error())
		return
	}

	state.ApplicationID = types.StringValue(deployment.ApplicationID)
	state.Status = types.StringValue(deployment.Status)
	state.CreatedAt = types.StringValue(deployment.CreatedAt)
	if deployment.Log != "" {
		state.Log = types.StringValue(deployment.Log)
	} else {
		state.Log = types.StringNull()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
