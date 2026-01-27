package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marconneves/terraform-provider-dokploy/internal/client"
)

var _ resource.Resource = &ApiKeyResource{}

func NewApiKeyResource() resource.Resource {
	return &ApiKeyResource{}
}

type ApiKeyResource struct {
	client *client.DokployClient
}

type ApiKeyResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Key            types.String `tfsdk:"key"`
}

func (r *ApiKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func (r *ApiKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dokploy API Key. Note: This resource requires email/password authentication.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the API key.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "The description of the API key.",
			},
			"organization_id": schema.StringAttribute{
				Required:    true,
				Description: "The Organization ID to associate with this API key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The generated API Key. This is only available on creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ApiKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*client.DokployClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Type", fmt.Sprintf("Expected *client.DokployClient, got: %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *ApiKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ApiKeyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	description := plan.Description.ValueString()
	orgID := plan.OrganizationID.ValueString()

	apiKey, err := r.client.CreateApiKey(name, description, orgID)
	if err != nil {
		resp.Diagnostics.AddError("Error creating api key", err.Error())
		return
	}

	plan.ID = types.StringValue(apiKey.ID)
	plan.Name = types.StringValue(apiKey.Name)
	if apiKey.Description != "" {
		plan.Description = types.StringValue(apiKey.Description)
	} else if description != "" {
		plan.Description = types.StringValue(description)
	} else {
		plan.Description = types.StringNull()
	}

	// Assuming key is returned in Create response
	if apiKey.Key != "" {
		plan.Key = types.StringValue(apiKey.Key)
	} else {
		// Should not happen on creation usually
		resp.Diagnostics.AddWarning("API Key empty", "The API key returned was empty.")
		plan.Key = types.StringValue("")
	}

	// Keep OrganizationID from plan as is

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ApiKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApiKeyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey, err := r.client.GetApiKey(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading api key", err.Error())
		return
	}

	if apiKey == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(apiKey.Name)
	if apiKey.Description != "" {
		state.Description = types.StringValue(apiKey.Description)
	}

	// Do not update Key or OrganizationID as they might not be returned
	// UseStateForUnknown on Key handles preserving if plan was unknown, but here we are in Read.
	// We just don't touch state.Key, so it remains as it was in state (passed in via `state` variable).

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *ApiKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "Updating API keys is not currently supported.")
}

func (r *ApiKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ApiKeyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteApiKey(state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
			return
		}
		resp.Diagnostics.AddError("Error deleting api key", err.Error())
		return
	}
}
