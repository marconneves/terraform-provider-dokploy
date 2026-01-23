package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marconneves/terraform-provider-dokploy/internal/client"
)

var _ datasource.DataSource = &ServerDataSource{}

func NewServerDataSource() datasource.DataSource {
	return &ServerDataSource{}
}

type ServerDataSource struct {
	client *client.DokployClient
}

type ServerDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	IPAddress   types.String `tfsdk:"ip_address"`
	Port        types.Int64  `tfsdk:"port"`
	Username    types.String `tfsdk:"username"`
	SSHKeyID    types.String `tfsdk:"ssh_key_id"`
}

func (d *ServerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (d *ServerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Computed: true,
			},
			"ip_address": schema.StringAttribute{
				Computed: true,
			},
			"port": schema.Int64Attribute{
				Computed: true,
			},
			"username": schema.StringAttribute{
				Computed: true,
			},
			"ssh_key_id": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *ServerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ServerDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	servers, err := d.client.ListServers()
	if err != nil {
		resp.Diagnostics.AddError("Error reading servers", err.Error())
		return
	}

	found := false
	for _, server := range servers {
		if server.Name == state.Name.ValueString() {
			state.ID = types.StringValue(server.ID)
			state.Description = types.StringValue(server.Description)
			state.IPAddress = types.StringValue(server.IPAddress)
			state.Port = types.Int64Value(server.Port)
			state.Username = types.StringValue(server.Username)
			state.SSHKeyID = types.StringValue(server.SSHKeyID)
			found = true
			break
		}
	}

	if !found {
		resp.Diagnostics.AddError("Server not found", fmt.Sprintf("Server with name '%s' not found", state.Name.ValueString()))
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
