package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/TheGenXCoder/terraform-provider-kpm/internal/client"
)

var _ datasource.DataSource = &SecretDataSource{}

// SecretDataSource reads an existing KPM secret without managing its lifecycle.
type SecretDataSource struct {
	client client.AgentKMSClient
}

// NewSecretDataSource returns the constructor used by the provider.
func NewSecretDataSource() datasource.DataSource { return &SecretDataSource{} }

type secretDataModel struct {
	Path  types.String `tfsdk:"path"`
	Value types.String `tfsdk:"value"`
}

func (d *SecretDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

func (d *SecretDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing KPM secret value from AgentKMS.",
		Attributes: map[string]schema.Attribute{
			"path": schema.StringAttribute{
				Required:    true,
				Description: "Secret path in service/name format (e.g. 'db/password').",
			},
			"value": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The secret value retrieved from AgentKMS.",
			},
		},
	}
}

func (d *SecretDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(client.AgentKMSClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type",
			fmt.Sprintf("expected client.AgentKMSClient, got %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *SecretDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config secretDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	value, err := d.client.GetSecret(ctx, config.Path.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading secret", err.Error())
		return
	}

	config.Value = types.StringValue(value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
