package outlook

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				TenantID:     "consumers",
				RedirectURL:  "http://localhost:8080/callback",
			},
			wantErr: false,
		},
		{
			name: "missing client ID",
			config: &Config{
				ClientSecret: "test-client-secret",
				TenantID:     "consumers",
				RedirectURL:  "http://localhost:8080/callback",
			},
			wantErr: true,
			errMsg:  "ClientID is required",
		},
		{
			name: "missing client secret",
			config: &Config{
				ClientID:    "test-client-id",
				TenantID:    "consumers",
				RedirectURL: "http://localhost:8080/callback",
			},
			wantErr: true,
			errMsg:  "ClientSecret is required",
		},
		{
			name: "missing tenant ID",
			config: &Config{
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				RedirectURL:  "http://localhost:8080/callback",
			},
			wantErr: true,
			errMsg:  "TenantID is required",
		},
		{
			name: "missing redirect URL",
			config: &Config{
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				TenantID:     "consumers",
			},
			wantErr: true,
			errMsg:  "RedirectURL is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_ToOAuth2Config(t *testing.T) {
	config := &Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		TenantID:     "consumers",
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{"Mail.Read", "Mail.ReadWrite"},
	}

	oauth2Config := config.ToOAuth2Config()

	assert.NotNil(t, oauth2Config)
	assert.Equal(t, config.ClientID, oauth2Config.ClientID)
	assert.Equal(t, config.ClientSecret, oauth2Config.ClientSecret)
	assert.Equal(t, config.RedirectURL, oauth2Config.RedirectURL)
	assert.Equal(t, config.Scopes, oauth2Config.Scopes)
}

func TestConfig_ToOAuth2Config_DefaultScopes(t *testing.T) {
	config := &Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		TenantID:     "consumers",
		RedirectURL:  "http://localhost:8080/callback",
	}

	oauth2Config := config.ToOAuth2Config()

	assert.NotNil(t, oauth2Config)
	assert.Equal(t, DefaultScopes(), oauth2Config.Scopes)
}

func TestDefaultScopes(t *testing.T) {
	scopes := DefaultScopes()

	assert.Len(t, scopes, 3)
	assert.Contains(t, scopes, "Mail.Read")
	assert.Contains(t, scopes, "Mail.ReadWrite")
	assert.Contains(t, scopes, "offline_access")
}

func TestConfig_String(t *testing.T) {
	config := &Config{
		ClientID:     "1234567890",
		ClientSecret: "secret-value",
		TenantID:     "consumers",
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{"Mail.Read"},
	}

	str := config.String()

	// Should mask the client ID
	assert.Contains(t, str, "1234****")
	assert.NotContains(t, str, "1234567890")

	// Should contain other fields
	assert.Contains(t, str, "consumers")
	assert.Contains(t, str, "http://localhost:8080/callback")
	assert.Contains(t, str, "Mail.Read")
}

func TestConfig_TenantIDOptions(t *testing.T) {
	tests := []struct {
		name     string
		tenantID string
	}{
		{"consumers", "consumers"},
		{"organizations", "organizations"},
		{"common", "common"},
		{"specific tenant", "12345678-1234-1234-1234-123456789012"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				TenantID:     tt.tenantID,
				RedirectURL:  "http://localhost:8080/callback",
			}

			oauth2Config := config.ToOAuth2Config()
			assert.NotNil(t, oauth2Config)
			assert.NotNil(t, oauth2Config.Endpoint)
		})
	}
}

func TestMaskString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"long string", "1234567890", "1234****"},
		{"short string", "123", "****"},
		{"empty string", "", "****"},
		{"exactly 4 chars", "1234", "****"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
