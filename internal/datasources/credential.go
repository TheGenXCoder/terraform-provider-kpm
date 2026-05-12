package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/TheGenXCoder/terraform-provider-kpm/internal/client"
)

var _ datasource.DataSource = &CredentialDataSource{}

// CredentialDataSource reads a dynamic LLM credential from AgentKMS.
type CredentialDataSource struct {
	client client.AgentKMSClient
}

// NewCredentialDataSource returns the constructor used by the provider.
func NewCredentialDataSource() datasource.DataSource { return &CredentialDataSource{} }

type credentialDataModel struct {
	Provider  types.String `tfsdk:"provider"`
	Value     types.String `tfsdk:"value"`
	ExpiresAt types.String `tfsdk:"expires_at"`
}

func (d *CredentialDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_credential"
}

func (d *CredentialDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a dynamic LLM credential (API key) from AgentKMS.",
		Attributes: map[string]schema.Attribute{
			"provider": schema.StringAttribute{
				Required:    true,
				Description: "The LLM provider name (e.g. 'openai', 'anthropic').",
			},
			"value": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The credential value (API key) retrieved from AgentKMS.",
			},
			"expires_at": schema.StringAttribute{
				Computed:    true,
				Description: "The expiration timestamp of the credential, if applicable.",
			},
		},
	}
}

func (d *CredentialDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CredentialDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config credentialDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cred, err := d.client.GetLLMCredential(ctx, config.Provider.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading credential", err.Error())
		return
	}

	config.Value = types.StringValue(cred.Value)
	config.ExpiresAt = types.StringValue(cred.ExpiresAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
