package datasources_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/TheGenXCoder/terraform-provider-kpm/internal/datasources"
)

func TestCredentialDataSourceSchema(t *testing.T) {
	ds := datasources.NewCredentialDataSource()
	ctx := context.Background()
	var resp datasource.SchemaResponse
	ds.Schema(ctx, datasource.SchemaRequest{}, &resp)

	for _, attr := range []string{"type", "path", "value", "expires_at"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("schema missing %q attribute", attr)
		}
	}
}
