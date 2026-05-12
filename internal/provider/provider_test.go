package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/TheGenXCoder/terraform-provider-kpm/internal/provider"
)

func TestProviderSchema(t *testing.T) {
	factory := providerserver.NewProtocol6WithError(provider.New("test")())
	server, err := factory()
	if err != nil {
		t.Fatalf("provider server initialisation failed: %v", err)
	}
	if server == nil {
		t.Fatal("provider server is nil")
	}
}
