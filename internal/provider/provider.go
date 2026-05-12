package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/provider"
)

// New returns a new provider factory function.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return nil
	}
}
