package resources_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/TheGenXCoder/terraform-provider-kpm/internal/resources"
)

func TestGithubAppResourceSchemaHasSensitiveKey(t *testing.T) {
	r := resources.NewGithubAppResource()
	ctx := context.Background()
	var resp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &resp)

	key, ok := resp.Schema.Attributes["private_key"]
	if !ok {
		t.Fatal("schema missing 'private_key' attribute")
	}
	if !key.(interface{ IsSensitive() bool }).IsSensitive() {
		t.Error("'private_key' must be sensitive")
	}

	if _, ok := resp.Schema.Attributes["private_key_sha256"]; !ok {
		t.Error("schema missing 'private_key_sha256' attribute for drift detection")
	}
}

func TestGithubAppResourceSchemaFields(t *testing.T) {
	r := resources.NewGithubAppResource()
	ctx := context.Background()
	var resp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &resp)

	for _, attr := range []string{"name", "app_id", "installation_id", "private_key", "private_key_sha256"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("schema missing %q attribute", attr)
		}
	}
}
