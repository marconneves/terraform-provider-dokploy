package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marconneves/terraform-provider-dokploy/internal/dokploy"
)

var _ resource.Resource = &RegistryResource{}
var _ resource.ResourceWithImportState = &RegistryResource{}

func NewRegistryResource() resource.Resource {
	return &RegistryResource{}
}

type RegistryResource struct {
	client *dokploy.Client
}

type RegistryResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Username      types.String `tfsdk:"username"`
	Password      types.String `tfsdk:"password"`
	ImageRegistry types.String `tfsdk:"image_registry"`
	RegistryURL   types.String `tfsdk:"registry_url"`
}

func (r *RegistryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registry"
}

func (r *RegistryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"username": schema.StringAttribute{
				Required: true,
			},
			"password": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
			"image_registry": schema.StringAttribute{
				Required:    true,
				Description: "Type of registry: 'ghcr', 'dockerhub', 'ecr', 'digitalocean', 'google', 'gitlab', 'azure', 'self-hosted'",
			},
			"registry_url": schema.StringAttribute{
				Optional:    true,
				Description: "Required if image_registry is 'self-hosted' or custom",
			},
		},
	}
}

func (r *RegistryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*dokploy.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Type", fmt.Sprintf("Expected *dokploy.Client, got: %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *RegistryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RegistryResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	registry := dokploy.Registry{
		Name:          plan.Name.ValueString(),
		Username:      plan.Username.ValueString(),
		Password:      plan.Password.ValueString(),
		ImageRegistry: plan.ImageRegistry.ValueString(),
	}

	if !plan.RegistryURL.IsNull() {
		registry.RegistryURL = plan.RegistryURL.ValueString()
	}

	created, err := r.client.CreateRegistry(registry)
	if err != nil {
		resp.Diagnostics.AddError("Error creating registry", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *RegistryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RegistryResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	registry, err := r.client.GetRegistry(state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading registry", err.Error())
		return
	}

	state.Name = types.StringValue(registry.Name)
	state.Username = types.StringValue(registry.Username)
	state.ImageRegistry = types.StringValue(registry.ImageRegistry)
	// Password is usually not returned or hashed, so we might want to keep the state's password?
	// But for drift detection, if the API doesn't return it, we can't verify.
	// We'll leave the password in state as is, unless the API returns something.

	if registry.RegistryURL != "" {
		state.RegistryURL = types.StringValue(registry.RegistryURL)
	} else {
		state.RegistryURL = types.StringNull()
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *RegistryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RegistryResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	registry := dokploy.Registry{
		ID:            plan.ID.ValueString(),
		Name:          plan.Name.ValueString(),
		Username:      plan.Username.ValueString(),
		Password:      plan.Password.ValueString(),
		ImageRegistry: plan.ImageRegistry.ValueString(),
	}

	if !plan.RegistryURL.IsNull() {
		registry.RegistryURL = plan.RegistryURL.ValueString()
	}

	updated, err := r.client.UpdateRegistry(registry)
	if err != nil {
		resp.Diagnostics.AddError("Error updating registry", err.Error())
		return
	}

	plan.Name = types.StringValue(updated.Name)
	plan.Username = types.StringValue(updated.Username)
	plan.ImageRegistry = types.StringValue(updated.ImageRegistry)

	if updated.RegistryURL != "" {
		plan.RegistryURL = types.StringValue(updated.RegistryURL)
	} else {
		plan.RegistryURL = types.StringNull()
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *RegistryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RegistryResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteRegistry(state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
			return
		}
		resp.Diagnostics.AddError("Error deleting registry", err.Error())
		return
	}
}

func (r *RegistryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
