package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marconneves/terraform-provider-dokploy/internal/dokploy"
)

var _ datasource.DataSource = &DatabaseDataSource{}

func NewDatabaseDataSource() datasource.DataSource {
	return &DatabaseDataSource{}
}

type DatabaseDataSource struct {
	client *dokploy.Client
}

type DatabaseDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	ProjectID     types.String `tfsdk:"project_id"`
	EnvironmentID types.String `tfsdk:"environment_id"`
	Type          types.String `tfsdk:"type"`
	Name          types.String `tfsdk:"name"`
	Version       types.String `tfsdk:"version"`
	InternalPort  types.Int64  `tfsdk:"internal_port"`
	ExternalPort  types.Int64  `tfsdk:"external_port"`
}

func (d *DatabaseDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database"
}

func (d *DatabaseDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"type": schema.StringAttribute{
				Required: true,
			},
			"project_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"environment_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"name": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"version": schema.StringAttribute{
				Computed: true,
			},
			"internal_port": schema.Int64Attribute{
				Computed: true,
			},
			"external_port": schema.Int64Attribute{
				Computed: true,
			},
		},
	}
}

func (d *DatabaseDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DatabaseDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state DatabaseDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ID.IsNull() && (state.Name.IsNull() || state.ProjectID.IsNull() || state.EnvironmentID.IsNull()) {
		resp.Diagnostics.AddError("Missing criteria", "Either 'id' or ('name', 'project_id', 'environment_id') must be provided.")
		return
	}

	if !state.ID.IsNull() {
		db, err := d.client.GetDatabase(state.ID.ValueString(), state.Type.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error reading database", err.Error())
			return
		}

		state.Name = types.StringValue(db.Name)
		state.Type = types.StringValue(db.Type)
		state.ProjectID = types.StringValue(db.ProjectID)
		state.EnvironmentID = types.StringValue(db.EnvironmentID)
		state.Version = types.StringValue(db.Version)
		state.InternalPort = types.Int64Value(db.InternalPort)
		state.ExternalPort = types.Int64Value(db.ExternalPort)
	} else {
		// Lookup by name
		project, err := d.client.GetProject(state.ProjectID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error fetching project", err.Error())
			return
		}

		found := false
		targetEnvID := state.EnvironmentID.ValueString()
		targetName := state.Name.ValueString()
		targetType := state.Type.ValueString()

		for _, env := range project.Environments {
			if env.ID == targetEnvID {
				var dbs []dokploy.Database
				switch targetType {
				case "postgres":
					dbs = env.Postgres
				case "mysql":
					dbs = env.Mysql
				case "mariadb":
					dbs = env.Mariadb
				case "mongo":
					dbs = env.Mongo
				case "redis":
					dbs = env.Redis
				default:
					resp.Diagnostics.AddError("Invalid database type", fmt.Sprintf("Type '%s' is not supported", targetType))
					return
				}

				for _, db := range dbs {
					if db.Name == targetName {
						id := db.PostgresID
						if db.MysqlID != "" {
							id = db.MysqlID
						}
						if db.MariadbID != "" {
							id = db.MariadbID
						}
						if db.MongoID != "" {
							id = db.MongoID
						}
						if db.RedisID != "" {
							id = db.RedisID
						}
						if id == "" && db.ID != "" {
							id = db.ID
						}

						state.ID = types.StringValue(id)
						state.Name = types.StringValue(db.Name)
						state.Type = types.StringValue(targetType)
						state.Version = types.StringValue(db.Version)
						state.InternalPort = types.Int64Value(db.InternalPort)
						state.ExternalPort = types.Int64Value(db.ExternalPort)
						found = true
						break
					}
				}
			}
			if found {
				break
			}
		}

		if !found {
			resp.Diagnostics.AddError("Database not found", fmt.Sprintf("Database '%s' of type '%s' not found in env '%s'", targetName, targetType, targetEnvID))
			return
		}
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
