package gmail

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestConfig creates a test configuration with standard test values.
// Use this instead of manually creating Config in every test.
func newTestConfig() *Config {
	return &Config{
		ClientID:     "test-id",
		ClientSecret: "test-secret",
		RedirectURL:  "http://localhost",
		Scopes:       []string{"test-scope"},
	}
}

// newTestClient creates a test client with standard configuration.
// Returns an initialized but not connected client.
func newTestClient(t *testing.T) *Client {
	t.Helper()
	client, err := New(newTestConfig())
	require.NoError(t, err)
	return client
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &Config{
				ClientID:     "test-id",
				ClientSecret: "test-secret",
				RedirectURL:  "http://localhost",
				Scopes:       DefaultScopes(),
			},
			wantErr: false,
		},
		{
			name: "missing client id",
			config: &Config{
				ClientSecret: "test-secret",
				RedirectURL:  "http://localhost",
				Scopes:       DefaultScopes(),
			},
			wantErr: true,
			errMsg:  "client_id",
		},
		{
			name: "missing client secret",
			config: &Config{
				ClientID:    "test-id",
				RedirectURL: "http://localhost",
				Scopes:      DefaultScopes(),
			},
			wantErr: true,
			errMsg:  "client_secret",
		},
		{
			name: "missing redirect url",
			config: &Config{
				ClientID:     "test-id",
				ClientSecret: "test-secret",
				Scopes:       DefaultScopes(),
			},
			wantErr: true,
			errMsg:  "redirect_url",
		},
		{
			name: "missing scopes auto-filled",
			config: &Config{
				ClientID:     "test-id",
				ClientSecret: "test-secret",
				RedirectURL:  "http://localhost",
				Scopes:       []string{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDefaultScopes(t *testing.T) {
	scopes := DefaultScopes()

	assert.NotEmpty(t, scopes)
	assert.Len(t, scopes, 4)
	assert.Contains(t, scopes, "https://www.googleapis.com/auth/gmail.readonly")
	assert.Contains(t, scopes, "https://www.googleapis.com/auth/gmail.send")
	assert.Contains(t, scopes, "https://www.googleapis.com/auth/gmail.modify")
	assert.Contains(t, scopes, "https://www.googleapis.com/auth/gmail.labels")
}

func TestConfig_ToOAuth2Config(t *testing.T) {
	config := &Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080",
		Scopes:       []string{"scope1", "scope2"},
	}

	oauth2Config := config.ToOAuth2Config()

	assert.NotNil(t, oauth2Config)
	assert.Equal(t, "test-client-id", oauth2Config.ClientID)
	assert.Equal(t, "test-client-secret", oauth2Config.ClientSecret)
	assert.Equal(t, "http://localhost:8080", oauth2Config.RedirectURL)
	assert.Equal(t, []string{"scope1", "scope2"}, oauth2Config.Scopes)
	assert.NotNil(t, oauth2Config.Endpoint)
}
