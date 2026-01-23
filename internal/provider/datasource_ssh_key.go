package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marconneves/terraform-provider-dokploy/internal/client"
)

var _ datasource.DataSource = &SSHKeyDataSource{}

func NewSSHKeyDataSource() datasource.DataSource {
	return &SSHKeyDataSource{}
}

type SSHKeyDataSource struct {
	client *client.DokployClient
}

type SSHKeyDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	PublicKey   types.String `tfsdk:"public_key"`
}

func (d *SSHKeyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_key"
}

func (d *SSHKeyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
			"public_key": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *SSHKeyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SSHKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state SSHKeyDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ID.IsNull() && state.Name.IsNull() {
		resp.Diagnostics.AddError("Missing criteria", "Either 'id' or 'name' must be provided.")
		return
	}

	if !state.ID.IsNull() {
		key, err := d.client.GetSSHKey(state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error reading SSH Key", err.Error())
			return
		}

		state.Name = types.StringValue(key.Name)
		state.Description = types.StringValue(key.Description)
		state.PublicKey = types.StringValue(key.PublicKey)
	} else {
		keys, err := d.client.ListSSHKeys()
		if err != nil {
			resp.Diagnostics.AddError("Error listing SSH Keys", err.Error())
			return
		}
		found := false
		targetName := state.Name.ValueString()
		for _, k := range keys {
			if k.Name == targetName {
				state.ID = types.StringValue(k.ID)
				state.Description = types.StringValue(k.Description)
				state.PublicKey = types.StringValue(k.PublicKey)
				found = true
				break
			}
		}
		if !found {
			resp.Diagnostics.AddError("SSH Key not found", fmt.Sprintf("SSH Key with name '%s' not found", targetName))
			return
		}
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
