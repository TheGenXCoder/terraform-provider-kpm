package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ provider.Provider = &KPMProvider{}

// KPMProvider is the Terraform provider implementation. Fully implemented in Task 4.
type KPMProvider struct{ version string }

// New returns a provider constructor.
func New(version string) func() provider.Provider {
	return func() provider.Provider { return &KPMProvider{version: version} }
}

func (p *KPMProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "kpm"
	resp.Version = p.version
}

func (p *KPMProvider) Schema(_ context.Context, _ provider.SchemaRequest, _ *provider.SchemaResponse) {}

func (p *KPMProvider) Configure(_ context.Context, _ provider.ConfigureRequest, _ *provider.ConfigureResponse) {
}

func (p *KPMProvider) Resources(_ context.Context) []func() resource.Resource { return nil }

func (p *KPMProvider) DataSources(_ context.Context) []func() datasource.DataSource { return nil }
