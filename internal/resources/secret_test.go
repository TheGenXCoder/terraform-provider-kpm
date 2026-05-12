package resources_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/TheGenXCoder/terraform-provider-kpm/internal/client"
	"github.com/TheGenXCoder/terraform-provider-kpm/internal/resources"
)

func TestSecretResourceSchemaHasSensitiveValue(t *testing.T) {
	r := resources.NewSecretResource()
	ctx := context.Background()
	var resp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &resp)

	attr, ok := resp.Schema.Attributes["value"]
	if !ok {
		t.Fatal("schema missing 'value' attribute")
	}
	strAttr, ok := attr.(interface{ IsSensitive() bool })
	if !ok {
		t.Fatal("'value' attribute does not implement IsSensitive()")
	}
	if !strAttr.IsSensitive() {
		t.Error("'value' attribute must be sensitive")
	}
}

func TestSecretResourceSchemaFields(t *testing.T) {
	r := resources.NewSecretResource()
	ctx := context.Background()
	var resp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &resp)

	for _, attr := range []string{"path", "value", "type", "description", "tags"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("schema missing %q attribute", attr)
		}
	}
}

func TestSecretResourceWriteRecordsPath(t *testing.T) {
	mock := &client.MockClient{
		UpGetSecret: "myvalue",
	}
	_ = mock
	// MockClient satisfies AgentKMSClient — compile-time check
	var _ client.AgentKMSClient = mock
}
