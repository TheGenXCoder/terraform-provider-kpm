package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/TheGenXCoder/terraform-provider-kpm/internal/client"
	"github.com/TheGenXCoder/terraform-provider-kpm/internal/datasources"
	"github.com/TheGenXCoder/terraform-provider-kpm/internal/resources"
)

var _ provider.Provider = &KPMProvider{}

// KPMProvider is the Terraform provider implementation.
type KPMProvider struct{ version string }

type kpmProviderModel struct {
	Server types.String `tfsdk:"server"`
	Cert   types.String `tfsdk:"cert"`
	Key    types.String `tfsdk:"key"`
	CACert types.String `tfsdk:"ca_cert"`
}

// New returns a provider constructor.
func New(version string) func() provider.Provider {
	return func() provider.Provider { return &KPMProvider{version: version} }
}

func (p *KPMProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "kpm"
	resp.Version = p.version
}

func (p *KPMProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with KPM and AgentKMS to manage secrets and GitHub App registrations.",
		Attributes: map[string]schema.Attribute{
			"server": schema.StringAttribute{
				Optional:    true,
				Description: "AgentKMS server URL. Overridden by KPM_SERVER env var.",
			},
			"cert": schema.StringAttribute{
				Optional:    true,
				Description: "Path to mTLS client certificate PEM file. Overridden by KPM_CERT env var.",
			},
			"key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Path to mTLS client key PEM file. Overridden by KPM_KEY env var.",
			},
			"ca_cert": schema.StringAttribute{
				Optional:    true,
				Description: "Path to CA certificate PEM file. Overridden by KPM_CA_CERT env var.",
			},
		},
	}
}

func (p *KPMProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var cfg kpmProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	server := first(cfg.Server.ValueString(), os.Getenv("KPM_SERVER"))
	cert := first(cfg.Cert.ValueString(), os.Getenv("KPM_CERT"))
	key := first(cfg.Key.ValueString(), os.Getenv("KPM_KEY"))
	caCert := first(cfg.CACert.ValueString(), os.Getenv("KPM_CA_CERT"))

	c, err := client.New(server, cert, key, caCert)
	if err != nil {
		resp.Diagnostics.AddError("Failed to configure KPM provider", err.Error())
		return
	}

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *KPMProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewSecretResource,
		resources.NewGithubAppResource,
	}
}

func (p *KPMProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewSecretDataSource,
		datasources.NewCredentialDataSource,
	}
}

// first returns the first non-empty string from vals.
func first(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
