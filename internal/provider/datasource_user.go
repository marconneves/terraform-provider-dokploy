package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marconneves/terraform-provider-dokploy/internal/dokploy"
)

var _ datasource.DataSource = &UserDataSource{}

func NewUserDataSource() datasource.DataSource {
	return &UserDataSource{}
}

type UserDataSource struct {
	client *dokploy.Client
}

type UserDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	Email          types.String `tfsdk:"email"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Role           types.String `tfsdk:"role"`
}

func (d *UserDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *UserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"email": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"organization_id": schema.StringAttribute{
				Computed: true,
			},
			"role": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *UserDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state UserDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ID.IsNull() && state.Email.IsNull() {
		resp.Diagnostics.AddError("Missing Criteria", "Either id or email must be provided")
		return
	}

	if !state.ID.IsNull() {
		user, err := d.client.GetUserByID(state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error reading user", err.Error())
			return
		}
		state.Email = types.StringValue(user.Email)
		state.OrganizationID = types.StringValue(user.OrganizationID)
		if user.Role != "" {
			state.Role = types.StringValue(user.Role)
		}
	} else {
		users, err := d.client.ListUsers()
		if err != nil {
			resp.Diagnostics.AddError("Error listing users", err.Error())
			return
		}
		found := false
		for _, u := range users {
			if u.Email == state.Email.ValueString() {
				state.ID = types.StringValue(u.ID)
				state.OrganizationID = types.StringValue(u.OrganizationID)
				if u.Role != "" {
					state.Role = types.StringValue(u.Role)
				}
				found = true
				break
			}
		}
		if !found {
			resp.Diagnostics.AddError("User not found", fmt.Sprintf("User with email '%s' not found", state.Email.ValueString()))
			return
		}
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
