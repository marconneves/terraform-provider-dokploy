package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marconneves/terraform-provider-dokploy/internal/client"
)

var _ datasource.DataSource = &RegistryDataSource{}

func NewRegistryDataSource() datasource.DataSource {
	return &RegistryDataSource{}
}

type RegistryDataSource struct {
	client *client.DokployClient
}

type RegistryDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Username      types.String `tfsdk:"username"`
	ImageRegistry types.String `tfsdk:"image_registry"`
	RegistryURL   types.String `tfsdk:"registry_url"`
}

func (d *RegistryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registry"
}

func (d *RegistryDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"username": schema.StringAttribute{
				Computed: true,
			},
			"image_registry": schema.StringAttribute{
				Computed: true,
			},
			"registry_url": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *RegistryDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RegistryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state RegistryDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	registries, err := d.client.ListRegistries()
	if err != nil {
		resp.Diagnostics.AddError("Error reading registries", err.Error())
		return
	}

	found := false
	for _, reg := range registries {
		if reg.Name == state.Name.ValueString() {
			state.ID = types.StringValue(reg.ID)
			state.Username = types.StringValue(reg.Username)
			state.ImageRegistry = types.StringValue(reg.ImageRegistry)
			if reg.RegistryURL != "" {
				state.RegistryURL = types.StringValue(reg.RegistryURL)
			} else {
				state.RegistryURL = types.StringNull()
			}
			found = true
			break
		}
	}

	if !found {
		resp.Diagnostics.AddError("Registry not found", fmt.Sprintf("Registry with name '%s' not found", state.Name.ValueString()))
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
