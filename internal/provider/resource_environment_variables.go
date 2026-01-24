package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/marconneves/terraform-provider-dokploy/internal/client"
)

var _ resource.Resource = &EnvironmentVariablesResource{}
var _ resource.ResourceWithImportState = &EnvironmentVariablesResource{}

func NewEnvironmentVariablesResource() resource.Resource {
	return &EnvironmentVariablesResource{}
}

type EnvironmentVariablesResource struct {
	client *client.DokployClient
}

type EnvironmentVariablesResourceModel struct {
	ID            types.String `tfsdk:"id"`
	ApplicationID types.String `tfsdk:"application_id"`
	ComposeID     types.String `tfsdk:"compose_id"`
	Variables     types.Map    `tfsdk:"variables"`
	Raw           types.String `tfsdk:"raw"`
	CreateEnvFile types.Bool   `tfsdk:"create_env_file"`
}

func (r *EnvironmentVariablesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment_variables"
}

func (r *EnvironmentVariablesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages all environment variables for a Dokploy application or compose service as a single resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"application_id": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("compose_id")),
				},
			},
			"compose_id": schema.StringAttribute{
				Optional: true,
			},
			"variables": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Sensitive:   true,
				Validators: []validator.Map{
					// We need at least one of variables or raw, but since they are optional and we want exclusive usage if possible,
					// or maybe allow both? If both, which one takes precedence?
					// Let's enforce ExactlyOneOf for clarity.
					// However, ExactlyOneOf on Map validator might not exist or be different.
					// Let's check imports. I didn't import mapvalidator.
					// I'll stick to a simple check in Create/Update logic or assume conflict if I can't find mapvalidator.
					// Actually, simpler: both optional. Logic handles precedence (Raw > Variables).
				},
			},
			"raw": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				// Validators: []validator.String{
				// 	stringvalidator.ConflictsWith(path.MatchRoot("variables")),
				// },
			},
			"create_env_file": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
		},
	}
}

func (r *EnvironmentVariablesResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EnvironmentVariablesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EnvironmentVariablesResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validation
	hasVariables := !plan.Variables.IsNull() && !plan.Variables.IsUnknown()
	hasRaw := !plan.Raw.IsNull() && !plan.Raw.IsUnknown()

	if !hasVariables && !hasRaw {
		resp.Diagnostics.AddError("Missing variables", "Either 'variables' or 'raw' must be specified.")
		return
	}
	if hasVariables && hasRaw {
		resp.Diagnostics.AddError("Conflicting attributes", "Specify either 'variables' or 'raw', not both.")
		return
	}

	if !plan.ApplicationID.IsNull() && !plan.ApplicationID.IsUnknown() {
		appID := plan.ApplicationID.ValueString()
		plan.ID = plan.ApplicationID

		if hasRaw {
			err := r.client.SaveApplicationEnv(appID, plan.Raw.ValueString(), plan.CreateEnvFile.ValueBoolPointer())
			if err != nil {
				resp.Diagnostics.AddError("Error creating environment variables (raw)", err.Error())
				return
			}
		} else {
			envMap := make(map[string]string)
			diags = plan.Variables.ElementsAs(ctx, &envMap, false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			err := r.client.UpdateApplicationEnv(appID, func(m map[string]string) {
				for k, v := range envMap {
					m[k] = v
				}
			}, plan.CreateEnvFile.ValueBoolPointer())

			if err != nil {
				resp.Diagnostics.AddError("Error creating environment variables", err.Error())
				return
			}
		}
	} else if !plan.ComposeID.IsNull() && !plan.ComposeID.IsUnknown() {
		composeID := plan.ComposeID.ValueString()
		plan.ID = plan.ComposeID

		if hasRaw {
			err := r.client.SaveComposeEnv(composeID, plan.Raw.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Error creating compose environment variables (raw)", err.Error())
				return
			}
		} else {
			envMap := make(map[string]string)
			diags = plan.Variables.ElementsAs(ctx, &envMap, false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			err := r.client.UpdateComposeEnv(composeID, func(m map[string]string) {
				for k, v := range envMap {
					m[k] = v
				}
			})

			if err != nil {
				resp.Diagnostics.AddError("Error creating compose environment variables", err.Error())
				return
			}
		}
	} else {
		resp.Diagnostics.AddError("Missing ID", "Either 'application_id' or 'compose_id' must be specified.")
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *EnvironmentVariablesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EnvironmentVariablesResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var envStr string

	if !state.ApplicationID.IsNull() {
		app, err := r.client.GetApplication(state.ApplicationID.ValueString())
		if err != nil {
			if strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
				resp.State.RemoveResource(ctx)
				return
			}
			resp.Diagnostics.AddError("Error reading application", err.Error())
			return
		}
		envStr = app.Env
	} else if !state.ComposeID.IsNull() {
		comp, err := r.client.GetCompose(state.ComposeID.ValueString())
		if err != nil {
			if strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
				resp.State.RemoveResource(ctx)
				return
			}
			resp.Diagnostics.AddError("Error reading compose", err.Error())
			return
		}
		envStr = comp.Env
	} else {
		resp.Diagnostics.AddError("Missing ID", "State is missing application_id or compose_id")
		return
	}

	// Update state based on configuration preference (Raw vs Variables)
	// If state.Raw is set (or was set), we populate Raw.
	// If state.Variables is set, we populate Variables.
	// If both are null (initial import?), maybe try to determine?

	// Since we can't easily know if the user wants raw or map during Read without looking at config (which we don't have in Read),
	// we rely on what's in State. But if it's import, State might be empty/null on these fields?
	// Actually ImportState populates ID. The Read is called afterwards.

	// We will populate both if possible, or stick to what is not null.

	if !state.Raw.IsNull() {
		state.Raw = types.StringValue(envStr)
		// We might also want to nullify Variables to avoid confusion?
		state.Variables = types.MapNull(types.StringType)
	} else {
		// Default to Variables map if Raw is null
		envMap := client.ParseEnv(envStr)
		state.Variables, diags = types.MapValueFrom(ctx, types.StringType, envMap)
		resp.Diagnostics.Append(diags...)
		state.Raw = types.StringNull()
	}

	// If both were null (e.g. import), we probably should prefer Variables as default?
	if state.Raw.IsNull() && state.Variables.IsNull() {
		envMap := client.ParseEnv(envStr)
		state.Variables, diags = types.MapValueFrom(ctx, types.StringType, envMap)
		resp.Diagnostics.Append(diags...)
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *EnvironmentVariablesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state EnvironmentVariablesResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validation
	hasVariables := !plan.Variables.IsNull() && !plan.Variables.IsUnknown()
	hasRaw := !plan.Raw.IsNull() && !plan.Raw.IsUnknown()

	if !hasVariables && !hasRaw {
		resp.Diagnostics.AddError("Missing variables", "Either 'variables' or 'raw' must be specified.")
		return
	}
	if hasVariables && hasRaw {
		resp.Diagnostics.AddError("Conflicting attributes", "Specify either 'variables' or 'raw', not both.")
		return
	}

	if !plan.ApplicationID.IsNull() && !plan.ApplicationID.IsUnknown() {
		appID := plan.ApplicationID.ValueString()
		plan.ID = plan.ApplicationID

		if hasRaw {
			err := r.client.SaveApplicationEnv(appID, plan.Raw.ValueString(), plan.CreateEnvFile.ValueBoolPointer())
			if err != nil {
				resp.Diagnostics.AddError("Error updating environment variables (raw)", err.Error())
				return
			}
		} else {
			envMap := make(map[string]string)
			diags = plan.Variables.ElementsAs(ctx, &envMap, false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			err := r.client.UpdateApplicationEnv(appID, func(m map[string]string) {
				// Clear existing vars and set new ones
				for k := range m {
					delete(m, k)
				}
				for k, v := range envMap {
					m[k] = v
				}
			}, plan.CreateEnvFile.ValueBoolPointer())

			if err != nil {
				resp.Diagnostics.AddError("Error updating environment variables", err.Error())
				return
			}
		}

	} else if !plan.ComposeID.IsNull() && !plan.ComposeID.IsUnknown() {
		composeID := plan.ComposeID.ValueString()
		plan.ID = plan.ComposeID

		if hasRaw {
			err := r.client.SaveComposeEnv(composeID, plan.Raw.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Error updating compose environment variables (raw)", err.Error())
				return
			}
		} else {
			envMap := make(map[string]string)
			diags = plan.Variables.ElementsAs(ctx, &envMap, false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			err := r.client.UpdateComposeEnv(composeID, func(m map[string]string) {
				// Clear existing vars and set new ones
				for k := range m {
					delete(m, k)
				}
				for k, v := range envMap {
					m[k] = v
				}
			})

			if err != nil {
				resp.Diagnostics.AddError("Error updating compose environment variables", err.Error())
				return
			}
		}
	} else {
		resp.Diagnostics.AddError("Missing ID", "Either 'application_id' or 'compose_id' must be specified.")
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *EnvironmentVariablesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EnvironmentVariablesResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !state.ApplicationID.IsNull() {
		err := r.client.UpdateApplicationEnv(state.ApplicationID.ValueString(), func(m map[string]string) {
			for k := range m {
				delete(m, k)
			}
		}, state.CreateEnvFile.ValueBoolPointer())

		if err != nil {
			if strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
				return
			}
			resp.Diagnostics.AddError("Error deleting environment variables", err.Error())
			return
		}
	} else if !state.ComposeID.IsNull() {
		err := r.client.UpdateComposeEnv(state.ComposeID.ValueString(), func(m map[string]string) {
			for k := range m {
				delete(m, k)
			}
		})

		if err != nil {
			if strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
				return
			}
			resp.Diagnostics.AddError("Error deleting compose environment variables", err.Error())
			return
		}
	}
}

func (r *EnvironmentVariablesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Defaults to application_id for import if we can't distinguish.
	// Users can't easily specify "type" during import.
	// We'll set application_id to the imported ID.
	resource.ImportStatePassthroughID(ctx, path.Root("application_id"), req, resp)
}
