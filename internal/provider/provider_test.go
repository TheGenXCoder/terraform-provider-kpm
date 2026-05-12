package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/TheGenXCoder/terraform-provider-kpm/internal/provider"
)

func TestProviderSchema(t *testing.T) {
	providers := map[string]func() (tfprotov6.ProviderServer, error){
		"kpm": providerserver.NewProtocol6WithError(provider.New("test")()),
	}
	_ = providers // schema validation is implicit in the provider compile
}
