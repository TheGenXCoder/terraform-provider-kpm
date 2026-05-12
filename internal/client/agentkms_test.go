package client_test

import (
	"context"
	"testing"

	"github.com/TheGenXCoder/terraform-provider-kpm/internal/client"
	"github.com/TheGenXCoder/terraform-provider-kpm/internal/testhelpers"
)

func TestClientWriteAndReadSecret(t *testing.T) {
	stub := testhelpers.NewStubServer(t)
	c, err := client.NewInsecure(stub.URL())
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if err := c.WriteSecret(ctx, "svc/key", "myvalue", nil, "", ""); err != nil {
		t.Fatalf("WriteSecret: %v", err)
	}

	got, err := c.GetSecret(ctx, "svc/key")
	if err != nil {
		t.Fatalf("GetSecret: %v", err)
	}
	if got != "myvalue" {
		t.Errorf("got %q, want %q", got, "myvalue")
	}
}

func TestClientDeleteSecret(t *testing.T) {
	stub := testhelpers.NewStubServer(t)
	c, _ := client.NewInsecure(stub.URL())
	ctx := context.Background()

	c.WriteSecret(ctx, "svc/del", "val", nil, "", "")
	if err := c.DeleteSecret(ctx, "svc/del"); err != nil {
		t.Fatalf("DeleteSecret: %v", err)
	}
	_, err := c.GetSecret(ctx, "svc/del")
	if err == nil {
		t.Error("expected error for deleted secret, got nil")
	}
}

func TestClientGithubApp(t *testing.T) {
	stub := testhelpers.NewStubServer(t)
	c, _ := client.NewInsecure(stub.URL())
	ctx := context.Background()

	req := client.RegisterGithubAppRequest{
		Name:           "my-app",
		AppID:          12345,
		InstallationID: 67890,
		PrivateKeyPEM:  "fake-pem",
	}
	summary, err := c.RegisterGithubApp(ctx, req)
	if err != nil {
		t.Fatalf("RegisterGithubApp: %v", err)
	}
	if summary.Name != "my-app" {
		t.Errorf("name: got %q, want %q", summary.Name, "my-app")
	}

	got, err := c.GetGithubApp(ctx, "my-app")
	if err != nil {
		t.Fatalf("GetGithubApp: %v", err)
	}
	if got.AppID != 12345 {
		t.Errorf("AppID: got %d, want 12345", got.AppID)
	}

	if err := c.RemoveGithubApp(ctx, "my-app"); err != nil {
		t.Fatalf("RemoveGithubApp: %v", err)
	}
}

func TestClientGetLLMCredential(t *testing.T) {
	stub := testhelpers.NewStubServer(t)
	c, _ := client.NewInsecure(stub.URL())
	ctx := context.Background()

	cred, err := c.GetLLMCredential(ctx, "openai")
	if err != nil {
		t.Fatalf("GetLLMCredential: %v", err)
	}
	if cred.Value == "" {
		t.Error("expected non-empty credential value")
	}
	if cred.ExpiresAt == "" {
		t.Error("expected non-empty expires_at")
	}
}
