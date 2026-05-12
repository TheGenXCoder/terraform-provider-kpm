package testhelpers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// StubServer is an in-memory AgentKMS stub for unit tests.
// It handles auth, secrets, metadata, github-apps, and credentials endpoints.
type StubServer struct {
	Server   *httptest.Server
	mu       sync.Mutex
	secrets  map[string]string            // path → value
	metadata map[string]map[string]string // path → metadata fields
	apps     map[string]GithubApp         // name → app
}

type GithubApp struct {
	Name           string `json:"name"`
	AppID          int64  `json:"app_id"`
	InstallationID int64  `json:"installation_id"`
}

// NewStubServer starts an httptest.Server and returns the stub.
func NewStubServer(t *testing.T) *StubServer {
	t.Helper()
	s := &StubServer{
		secrets:  make(map[string]string),
		metadata: make(map[string]map[string]string),
		apps:     make(map[string]GithubApp),
	}
	mux := http.NewServeMux()

	// Auth
	mux.HandleFunc("/auth/session", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"token": "test-token"})
	})

	// Secrets: POST and GET /secrets/{path}
	mux.HandleFunc("/secrets/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/secrets/")
		s.mu.Lock()
		defer s.mu.Unlock()
		switch r.Method {
		case http.MethodPost:
			var body map[string]string
			json.NewDecoder(r.Body).Decode(&body)
			s.secrets[path] = body["value"]
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{"path": path, "version": 1, "status": "created"})
		case http.MethodGet:
			v, ok := s.secrets[path]
			if !ok {
				http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(map[string]string{"value": v})
		case http.MethodDelete:
			delete(s.secrets, path)
			w.WriteHeader(http.StatusNoContent)
		}
	})

	// Metadata: POST and GET /metadata/{path}
	mux.HandleFunc("/metadata/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/metadata/")
		s.mu.Lock()
		defer s.mu.Unlock()
		switch r.Method {
		case http.MethodPost:
			var body map[string]string
			json.NewDecoder(r.Body).Decode(&body)
			s.metadata[path] = body
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		case http.MethodGet:
			m := s.metadata[path]
			if m == nil {
				m = map[string]string{}
			}
			json.NewEncoder(w).Encode(m)
		}
	})

	// GitHub Apps
	mux.HandleFunc("/github-apps", func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		switch r.Method {
		case http.MethodPost:
			var body struct {
				Name           string `json:"name"`
				AppID          int64  `json:"app_id"`
				InstallationID int64  `json:"installation_id"`
				PrivateKeyPEM  string `json:"private_key_pem"`
			}
			json.NewDecoder(r.Body).Decode(&body)
			app := GithubApp{Name: body.Name, AppID: body.AppID, InstallationID: body.InstallationID}
			s.apps[body.Name] = app
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(app)
		case http.MethodGet:
			apps := make([]GithubApp, 0, len(s.apps))
			for _, a := range s.apps {
				apps = append(apps, a)
			}
			json.NewEncoder(w).Encode(map[string]any{"apps": apps})
		}
	})

	mux.HandleFunc("/github-apps/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/github-apps/")
		s.mu.Lock()
		defer s.mu.Unlock()
		switch r.Method {
		case http.MethodGet:
			app, ok := s.apps[name]
			if !ok {
				http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(app)
		case http.MethodDelete:
			delete(s.apps, name)
			w.WriteHeader(http.StatusNoContent)
		}
	})

	// LLM credentials
	mux.HandleFunc("/credentials/llm/", func(w http.ResponseWriter, r *http.Request) {
		provider := strings.TrimPrefix(r.URL.Path, "/credentials/llm/")
		json.NewEncoder(w).Encode(map[string]string{
			"provider":   provider,
			"api_key":    "test-api-key-for-" + provider,
			"expires_at": "2099-01-01T00:00:00Z",
		})
	})

	s.Server = httptest.NewServer(mux)
	t.Cleanup(s.Server.Close)
	return s
}

// URL returns the base URL of the stub server.
func (s *StubServer) URL() string {
	return s.Server.URL
}
