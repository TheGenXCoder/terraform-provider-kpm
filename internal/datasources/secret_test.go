package datasources_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/TheGenXCoder/terraform-provider-kpm/internal/datasources"
)

func TestSecretDataSourceSchema(t *testing.T) {
	ds := datasources.NewSecretDataSource()
	ctx := context.Background()
	var resp datasource.SchemaResponse
	ds.Schema(ctx, datasource.SchemaRequest{}, &resp)

	if _, ok := resp.Schema.Attributes["path"]; !ok {
		t.Error("schema missing 'path' attribute")
	}
	if _, ok := resp.Schema.Attributes["value"]; !ok {
		t.Error("schema missing 'value' attribute")
	}
}
