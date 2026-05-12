package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/TheGenXCoder/kpm/pkg/tlsutil"
)

// AgentKMSClient is the interface used by all resources and data sources.
// The interface enables mock injection in unit tests.
type AgentKMSClient interface {
	WriteSecret(ctx context.Context, path, value string, tags []string, description, secretType string) error
	GetSecret(ctx context.Context, path string) (string, error)
	DeleteSecret(ctx context.Context, path string) error
	GetLLMCredential(ctx context.Context, provider string) (*CredentialResult, error)
	RegisterGithubApp(ctx context.Context, req RegisterGithubAppRequest) (*GithubAppSummary, error)
	GetGithubApp(ctx context.Context, name string) (*GithubAppSummary, error)
	RemoveGithubApp(ctx context.Context, name string) error
}

// RegisterGithubAppRequest is the payload for POST /github-apps.
type RegisterGithubAppRequest struct {
	Name           string `json:"name"`
	AppID          int64  `json:"app_id"`
	InstallationID int64  `json:"installation_id"`
	PrivateKeyPEM  string `json:"private_key_pem"`
}

// GithubAppSummary is returned by GET /github-apps/{name} and POST /github-apps.
type GithubAppSummary struct {
	Name           string `json:"name"`
	AppID          int64  `json:"app_id"`
	InstallationID int64  `json:"installation_id"`
}

// CredentialResult holds a fetched dynamic credential.
type CredentialResult struct {
	Value     string
	ExpiresAt string
}

// httpAgentKMS is the real AgentKMS HTTP client.
type httpAgentKMS struct {
	base  string
	token string
	http  *http.Client
}

// New builds an AgentKMSClient using mTLS. certPath, keyPath, and caPath are
// file-system paths to PEM-encoded files.
func New(serverURL, certPath, keyPath, caPath string) (AgentKMSClient, error) {
	if serverURL == "" {
		return nil, fmt.Errorf("server URL is required")
	}
	caPEM, err := os.ReadFile(caPath)
	if err != nil {
		return nil, fmt.Errorf("read CA cert %s: %w", caPath, err)
	}
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("read client cert %s: %w", certPath, err)
	}
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("read client key %s: %w", keyPath, err)
	}
	tlsCfg, err := tlsutil.ClientTLSConfig(caPEM, certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("build mTLS config: %w", err)
	}
	return newWithTLS(serverURL, tlsCfg), nil
}

// NewInsecure builds a client with no TLS — only for unit tests against httptest.Server.
func NewInsecure(serverURL string) (AgentKMSClient, error) {
	return newWithTLS(serverURL, &tls.Config{}), nil
}

func newWithTLS(serverURL string, tlsCfg *tls.Config) AgentKMSClient {
	return &httpAgentKMS{
		base: strings.TrimRight(serverURL, "/"),
		http: &http.Client{
			Timeout:   30 * time.Second,
			Transport: &http.Transport{TLSClientConfig: tlsCfg},
		},
	}
}

func (c *httpAgentKMS) ensureAuth(ctx context.Context) error {
	if c.token != "" {
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/auth/session", nil)
	if err != nil {
		return fmt.Errorf("build auth request: %w", err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("auth: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth: server returned %d", resp.StatusCode)
	}
	var body struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return fmt.Errorf("auth: decode response: %w", err)
	}
	c.token = body.Token
	return nil
}

func (c *httpAgentKMS) doGet(ctx context.Context, path string) (*http.Response, error) {
	if err := c.ensureAuth(ctx); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base+path, nil)
	if err != nil {
		return nil, fmt.Errorf("build GET request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	return c.http.Do(req)
}

func (c *httpAgentKMS) doPost(ctx context.Context, path string, body any) (*http.Response, error) {
	if err := c.ensureAuth(ctx); err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, fmt.Errorf("encode body: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+path, &buf)
	if err != nil {
		return nil, fmt.Errorf("build POST request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	return c.http.Do(req)
}

func (c *httpAgentKMS) doDelete(ctx context.Context, path string) (*http.Response, error) {
	if err := c.ensureAuth(ctx); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.base+path, nil)
	if err != nil {
		return nil, fmt.Errorf("build DELETE request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	return c.http.Do(req)
}

func serverErr(resp *http.Response, op string) error {
	body, _ := io.ReadAll(resp.Body)
	var e struct{ Error string `json:"error"` }
	if json.Unmarshal(body, &e) == nil && e.Error != "" {
		return fmt.Errorf("%s: %s", op, e.Error)
	}
	return fmt.Errorf("%s: server returned %d", op, resp.StatusCode)
}

func (c *httpAgentKMS) WriteSecret(ctx context.Context, path, value string, tags []string, description, secretType string) error {
	body := map[string]any{"value": value}
	resp, err := c.doPost(ctx, "/secrets/"+path, body)
	if err != nil {
		return fmt.Errorf("write secret: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return serverErr(resp, "write secret "+path)
	}
	// Write metadata only if any optional fields are set.
	if description != "" || len(tags) > 0 || secretType != "" {
		meta := map[string]any{}
		if description != "" {
			meta["description"] = description
		}
		if len(tags) > 0 {
			meta["tags"] = tags
		}
		if secretType != "" {
			meta["type"] = secretType
		}
		mresp, err := c.doPost(ctx, "/metadata/"+path, meta)
		if err != nil {
			return fmt.Errorf("write metadata: %w", err)
		}
		defer mresp.Body.Close()
	}
	return nil
}

func (c *httpAgentKMS) GetSecret(ctx context.Context, path string) (string, error) {
	resp, err := c.doGet(ctx, "/secrets/"+path)
	if err != nil {
		return "", fmt.Errorf("get secret: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("not found: %s", path)
	}
	if resp.StatusCode != http.StatusOK {
		return "", serverErr(resp, "get secret "+path)
	}
	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("decode secret: %w", err)
	}
	return body["value"], nil
}

func (c *httpAgentKMS) DeleteSecret(ctx context.Context, path string) error {
	resp, err := c.doDelete(ctx, "/secrets/"+path)
	if err != nil {
		return fmt.Errorf("delete secret: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return serverErr(resp, "delete secret "+path)
	}
	return nil
}

func (c *httpAgentKMS) GetLLMCredential(ctx context.Context, provider string) (*CredentialResult, error) {
	resp, err := c.doGet(ctx, "/credentials/llm/"+provider)
	if err != nil {
		return nil, fmt.Errorf("get credential: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, serverErr(resp, "get credential "+provider)
	}
	var body struct {
		APIKey    string `json:"api_key"`
		ExpiresAt string `json:"expires_at"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode credential: %w", err)
	}
	return &CredentialResult{Value: body.APIKey, ExpiresAt: body.ExpiresAt}, nil
}

func (c *httpAgentKMS) RegisterGithubApp(ctx context.Context, req RegisterGithubAppRequest) (*GithubAppSummary, error) {
	resp, err := c.doPost(ctx, "/github-apps", req)
	if err != nil {
		return nil, fmt.Errorf("register github app: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, serverErr(resp, "register github app "+req.Name)
	}
	var out GithubAppSummary
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode github app response: %w", err)
	}
	return &out, nil
}

func (c *httpAgentKMS) GetGithubApp(ctx context.Context, name string) (*GithubAppSummary, error) {
	resp, err := c.doGet(ctx, "/github-apps/"+name)
	if err != nil {
		return nil, fmt.Errorf("get github app: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("not found: %s", name)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, serverErr(resp, "get github app "+name)
	}
	var out GithubAppSummary
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode github app response: %w", err)
	}
	return &out, nil
}

func (c *httpAgentKMS) RemoveGithubApp(ctx context.Context, name string) error {
	resp, err := c.doDelete(ctx, "/github-apps/"+name)
	if err != nil {
		return fmt.Errorf("remove github app: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return serverErr(resp, "remove github app "+name)
	}
	return nil
}
