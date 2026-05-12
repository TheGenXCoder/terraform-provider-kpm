package client

import "context"

var _ AgentKMSClient = &MockClient{}

// MockClient is a controllable fake for unit tests.
// Set Up* fields to control what each method returns.
type MockClient struct {
	UpWriteSecret    error
	UpGetSecret      string
	UpGetSecretErr   error
	UpDeleteSecret   error
	UpGetLLMCred     *CredentialResult
	UpGetLLMCredErr  error
	UpRegisterApp    *GithubAppSummary
	UpRegisterAppErr error
	UpGetApp         *GithubAppSummary
	UpGetAppErr      error
	UpRemoveApp      error

	// Recorded calls
	WrittenPath  string
	WrittenValue string
	DeletedPath  string
}

func (m *MockClient) WriteSecret(_ context.Context, path, value string, _ []string, _, _ string) error {
	m.WrittenPath = path
	m.WrittenValue = value
	return m.UpWriteSecret
}

func (m *MockClient) GetSecret(_ context.Context, _ string) (string, error) {
	return m.UpGetSecret, m.UpGetSecretErr
}

func (m *MockClient) DeleteSecret(_ context.Context, path string) error {
	m.DeletedPath = path
	return m.UpDeleteSecret
}

func (m *MockClient) GetLLMCredential(_ context.Context, _ string) (*CredentialResult, error) {
	return m.UpGetLLMCred, m.UpGetLLMCredErr
}

func (m *MockClient) RegisterGithubApp(_ context.Context, _ RegisterGithubAppRequest) (*GithubAppSummary, error) {
	return m.UpRegisterApp, m.UpRegisterAppErr
}

func (m *MockClient) GetGithubApp(_ context.Context, _ string) (*GithubAppSummary, error) {
	return m.UpGetApp, m.UpGetAppErr
}

func (m *MockClient) RemoveGithubApp(_ context.Context, _ string) error {
	return m.UpRemoveApp
}
