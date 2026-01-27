package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marconneves/terraform-provider-dokploy/internal/client"
)

var _ provider.Provider = &DokployProvider{}
var _ provider.ProviderWithFunctions = &DokployProvider{}

type DokployProvider struct {
	version string
}

type DokployProviderModel struct {
	Host     types.String `tfsdk:"host"`
	ApiKey   types.String `tfsdk:"api_key"`
	Email    types.String `tfsdk:"email"`
	Password types.String `tfsdk:"password"`
}

func (p *DokployProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "dokploy"
	resp.Version = p.version
}

func (p *DokployProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Required:    true,
				Description: "The URL of your Dokploy instance (e.g., https://dokploy.example.com/api)",
			},
			"api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Your Dokploy API Key",
			},
			"email": schema.StringAttribute{
				Optional:    true,
				Description: "The email for authentication (required for creating organizations and API keys)",
			},
			"password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The password for authentication (required for creating organizations and API keys)",
			},
		},
	}
}

func (p *DokployProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config DokployProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddWarning(
			"Missing Host Configuration",
			"While configuring the provider, the Host was unknown. This can happen when the value is not yet known.",
		)
	}

	if config.ApiKey.IsUnknown() && (config.Email.IsUnknown() || config.Password.IsUnknown()) {
		resp.Diagnostics.AddWarning(
			"Missing Authentication Configuration",
			"While configuring the provider, the API Key or Email/Password was unknown. This can happen when the value is not yet known.",
		)
	}

	if config.Host.IsNull() {
		return
	}

	if config.ApiKey.IsNull() && (config.Email.IsNull() || config.Password.IsNull()) {
		// Can't validate here strictly because sometimes config is partial during validation phases,
		// but ideally we need one or the other.
		// For now, allow returning if neither is fully known yet, resources will fail if client is unconfigured properly.
		// However, we should probably instantiate the client if we have partial info.
	}

	apiKey := config.ApiKey.ValueString()
	email := config.Email.ValueString()
	password := config.Password.ValueString()

	// Create client
	c := client.NewDokployClient(config.Host.ValueString(), apiKey, email, password)

	// Make client available to resources
	resp.ResourceData = c
	resp.DataSourceData = c
}

func (p *DokployProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewProjectResource,
		NewEnvironmentResource,
		NewApplicationResource,
		NewComposeResource,
		NewDatabaseResource,
		NewDomainResource,
		NewEnvironmentVariablesResource,
		NewSSHKeyResource,
		NewServerResource,
		NewRegistryResource,
		NewUserResource,
		NewDeploymentResource,
		NewOrganizationResource,
		NewApiKeyResource,
	}
}

func (p *DokployProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewProjectDataSource,
		NewEnvironmentDataSource,
		NewApplicationDataSource,
		NewComposeDataSource,
		NewDatabaseDataSource,
		NewSSHKeyDataSource,
		NewDomainDataSource,
		NewServerDataSource,
		NewRegistryDataSource,
		NewUserDataSource,
		NewDeploymentDataSource,
	}
}

func (p *DokployProvider) Functions(_ context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &DokployProvider{version: version}
	}
}
