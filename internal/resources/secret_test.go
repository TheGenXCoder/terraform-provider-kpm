package resources_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
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
	mock := &client.MockClient{}
	r := resources.NewSecretResource()
	r.(*resources.SecretResource).SetClient(mock)

	ctx := context.Background()

	// Get the schema so we can build a properly typed plan.
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)

	attrTypes := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object).AttributeTypes
	planVal := tftypes.NewValue(tftypes.Object{AttributeTypes: attrTypes}, map[string]tftypes.Value{
		"path":        tftypes.NewValue(tftypes.String, "svc/key"),
		"value":       tftypes.NewValue(tftypes.String, "s3cr3t"),
		"type":        tftypes.NewValue(tftypes.String, nil),
		"description": tftypes.NewValue(tftypes.String, nil),
		"tags":        tftypes.NewValue(attrTypes["tags"], nil),
	})
	req := resource.CreateRequest{}
	req.Plan = tfsdk.Plan{Raw: planVal, Schema: schemaResp.Schema}

	var resp resource.CreateResponse
	resp.State = tfsdk.State{Schema: schemaResp.Schema}
	r.Create(ctx, req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %s", resp.Diagnostics[0].Detail())
	}
	if mock.WrittenPath != "svc/key" {
		t.Errorf("WrittenPath = %q, want %q", mock.WrittenPath, "svc/key")
	}
	if mock.WrittenValue != "s3cr3t" {
		t.Errorf("WrittenValue = %q, want %q", mock.WrittenValue, "s3cr3t")
	}
}
