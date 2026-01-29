package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marconneves/terraform-provider-dokploy/internal/dokploy"
)

var _ resource.Resource = &ServerResource{}
var _ resource.ResourceWithImportState = &ServerResource{}

func NewServerResource() resource.Resource {
	return &ServerResource{}
}

type ServerResource struct {
	client *dokploy.Client
}

type ServerResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	IPAddress   types.String `tfsdk:"ip_address"`
	Port        types.Int64  `tfsdk:"port"`
	Username    types.String `tfsdk:"username"`
	SSHKeyID    types.String `tfsdk:"ssh_key_id"`
}

func (r *ServerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (r *ServerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
			"ip_address": schema.StringAttribute{
				Required: true,
			},
			"port": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"username": schema.StringAttribute{
				Required: true,
			},
			"ssh_key_id": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (r *ServerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ServerResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	port := int64(22)
	if !plan.Port.IsNull() && !plan.Port.IsUnknown() {
		port = plan.Port.ValueInt64()
	}

	server := dokploy.Server{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		IPAddress:   plan.IPAddress.ValueString(),
		Port:        port,
		Username:    plan.Username.ValueString(),
		SSHKeyID:    plan.SSHKeyID.ValueString(),
	}

	createdServer, err := r.client.CreateServer(server)
	if err != nil {
		resp.Diagnostics.AddError("Error creating server", err.Error())
		return
	}

	plan.ID = types.StringValue(createdServer.ID)
	plan.Name = types.StringValue(createdServer.Name)
	plan.Description = types.StringValue(createdServer.Description)
	plan.IPAddress = types.StringValue(createdServer.IPAddress)
	plan.Port = types.Int64Value(createdServer.Port)
	plan.Username = types.StringValue(createdServer.Username)
	plan.SSHKeyID = types.StringValue(createdServer.SSHKeyID)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ServerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	server, err := r.client.GetServer(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading server", err.Error())
		return
	}

	state.Name = types.StringValue(server.Name)
	state.Description = types.StringValue(server.Description)
	state.IPAddress = types.StringValue(server.IPAddress)
	state.Port = types.Int64Value(server.Port)
	state.Username = types.StringValue(server.Username)
	state.SSHKeyID = types.StringValue(server.SSHKeyID)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *ServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ServerResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ServerResourceModel
	req.State.Get(ctx, &state)

	port := int64(22)
	if !plan.Port.IsNull() && !plan.Port.IsUnknown() {
		port = plan.Port.ValueInt64()
	}

	server := dokploy.Server{
		ID:          state.ID.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		IPAddress:   plan.IPAddress.ValueString(),
		Port:        port,
		Username:    plan.Username.ValueString(),
		SSHKeyID:    plan.SSHKeyID.ValueString(),
	}

	updatedServer, err := r.client.UpdateServer(server)
	if err != nil {
		resp.Diagnostics.AddError("Error updating server", err.Error())
		return
	}

	plan.ID = types.StringValue(updatedServer.ID)
	plan.Name = types.StringValue(updatedServer.Name)
	plan.Description = types.StringValue(updatedServer.Description)
	plan.IPAddress = types.StringValue(updatedServer.IPAddress)
	plan.Port = types.Int64Value(updatedServer.Port)
	plan.Username = types.StringValue(updatedServer.Username)
	plan.SSHKeyID = types.StringValue(updatedServer.SSHKeyID)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ServerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	err := r.client.DeleteServer(state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
			return
		}
		resp.Diagnostics.AddError("Error deleting server", err.Error())
		return
	}
}

func (r *ServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
