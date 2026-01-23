package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marconneves/terraform-provider-dokploy/internal/client"
)

var _ datasource.DataSource = &DomainDataSource{}

func NewDomainDataSource() datasource.DataSource {
	return &DomainDataSource{}
}

type DomainDataSource struct {
	client *client.DokployClient
}

type DomainDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	ApplicationID types.String `tfsdk:"application_id"`
	ComposeID     types.String `tfsdk:"compose_id"`
	ServiceName   types.String `tfsdk:"service_name"`
	Host          types.String `tfsdk:"host"`
	Path          types.String `tfsdk:"path"`
	Port          types.Int64  `tfsdk:"port"`
	HTTPS         types.Bool   `tfsdk:"https"`
}

func (d *DomainDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func (d *DomainDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required: true,
			},
			"application_id": schema.StringAttribute{
				Computed: true,
			},
			"compose_id": schema.StringAttribute{
				Computed: true,
			},
			"service_name": schema.StringAttribute{
				Computed: true,
			},
			"host": schema.StringAttribute{
				Computed: true,
			},
			"path": schema.StringAttribute{
				Computed: true,
			},
			"port": schema.Int64Attribute{
				Computed: true,
			},
			"https": schema.BoolAttribute{
				Computed: true,
			},
		},
	}
}

func (d *DomainDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state DomainDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain, err := d.client.GetDomain(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading domain", err.Error())
		return
	}

	state.Host = types.StringValue(domain.Host)
	state.Path = types.StringValue(domain.Path)
	state.Port = types.Int64Value(domain.Port)
	state.HTTPS = types.BoolValue(domain.HTTPS)
	state.ServiceName = types.StringValue(domain.ServiceName)
	if domain.ApplicationID != "" {
		state.ApplicationID = types.StringValue(domain.ApplicationID)
	}
	if domain.ComposeID != "" {
		state.ComposeID = types.StringValue(domain.ComposeID)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
