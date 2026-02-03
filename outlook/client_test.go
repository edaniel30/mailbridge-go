package outlook

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/danielrivera/mailbridge-go/core"
	outlooktest "github.com/danielrivera/mailbridge-go/outlook/testing"
	"github.com/microsoftgraph/msgraph-sdk-go/models/odataerrors"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

func TestNew(t *testing.T) {
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
				ClientSecret: "test-secret",
				TenantID:     "consumers",
				RedirectURL:  "http://localhost:8080/callback",
			},
			wantErr: false,
		},
		{
			name: "missing client ID",
			config: &Config{
				ClientSecret: "test-secret",
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
				ClientSecret: "test-secret",
				RedirectURL:  "http://localhost:8080/callback",
			},
			wantErr: true,
			errMsg:  "TenantID is required",
		},
		{
			name: "missing redirect URL",
			config: &Config{
				ClientID:     "test-client-id",
				ClientSecret: "test-secret",
				TenantID:     "consumers",
			},
			wantErr: true,
			errMsg:  "RedirectURL is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, tt.config, client.config)
				assert.NotNil(t, client.oauth2Config)
			}
		})
	}
}

func TestClient_GetAuthURL(t *testing.T) {
	config := &Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		TenantID:     "consumers",
		RedirectURL:  "http://localhost:8080/callback",
	}

	client, err := New(config)
	assert.NoError(t, err)

	authURL := client.GetAuthURL("random-state-123")

	// Should contain essential OAuth2 parameters
	assert.Contains(t, authURL, "response_type=code")
	assert.Contains(t, authURL, "client_id=test-client-id")
	assert.Contains(t, authURL, "state=random-state-123")
	assert.Contains(t, authURL, "access_type=offline")
	assert.Contains(t, authURL, "redirect_uri=")
	assert.Contains(t, authURL, "consumers") // Tenant ID in URL
}

func TestClient_IsConnected(t *testing.T) {
	config := &Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		TenantID:     "consumers",
		RedirectURL:  "http://localhost:8080/callback",
	}

	client, err := New(config)
	assert.NoError(t, err)

	// Initially not connected
	assert.False(t, client.IsConnected())

	// After setting service, should be connected
	mockService := &outlooktest.MockGraphService{}
	client.service = mockService
	assert.True(t, client.IsConnected())
}

func TestClient_GetToken(t *testing.T) {
	config := &Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		TenantID:     "consumers",
		RedirectURL:  "http://localhost:8080/callback",
	}

	client, err := New(config)
	assert.NoError(t, err)

	// Initially nil
	assert.Nil(t, client.GetToken())

	// After setting token
	token := &oauth2.Token{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		Expiry:       time.Now().Add(time.Hour),
	}
	client.token = token

	retrievedToken := client.GetToken()
	assert.NotNil(t, retrievedToken)
	assert.Equal(t, token.AccessToken, retrievedToken.AccessToken)
	assert.Equal(t, token.RefreshToken, retrievedToken.RefreshToken)
}

func TestClient_ConnectWithToken_NilToken(t *testing.T) {
	config := &Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		TenantID:     "consumers",
		RedirectURL:  "http://localhost:8080/callback",
	}

	client, err := New(config)
	assert.NoError(t, err)

	err = client.ConnectWithToken(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token cannot be nil")
}

func TestHandleODataError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: "",
		},
		{
			name:     "regular error",
			err:      fmt.Errorf("regular error"),
			expected: "regular error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handleODataError(tt.err)
			if tt.expected == "" {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Contains(t, result.Error(), tt.expected)
			}
		})
	}
}

func TestHandleODataError_ODataError(t *testing.T) {
	// Create an OData error
	odataErr := odataerrors.NewODataError()
	mainErr := odataerrors.NewMainError()

	code := "ErrorItemNotFound"
	message := "The specified object was not found"

	mainErr.SetCode(&code)
	mainErr.SetMessage(&message)
	odataErr.SetErrorEscaped(mainErr)

	// Test handling
	result := handleODataError(odataErr)

	assert.NotNil(t, result)
	assert.Contains(t, result.Error(), "microsoft graph error")
	assert.Contains(t, result.Error(), "ErrorItemNotFound")
	assert.Contains(t, result.Error(), "The specified object was not found")
}

func TestHandleODataError_ODataError_EmptyFields(t *testing.T) {
	// Create an OData error with empty fields
	odataErr := odataerrors.NewODataError()
	mainErr := odataerrors.NewMainError()
	odataErr.SetErrorEscaped(mainErr)

	// Test handling
	result := handleODataError(odataErr)

	assert.NotNil(t, result)
	assert.Contains(t, result.Error(), "microsoft graph error")
}

func TestClient_SetService(t *testing.T) {
	config := &Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		TenantID:     "consumers",
		RedirectURL:  "http://localhost:8080/callback",
	}

	client, err := New(config)
	assert.NoError(t, err)
	assert.False(t, client.IsConnected())

	// Set service
	mockService := &outlooktest.MockGraphService{}
	client.SetService(mockService)

	assert.True(t, client.IsConnected())
	assert.Equal(t, mockService, client.service)
}

func TestConfigError_Compatibility(t *testing.T) {
	// Test that config errors are core.ConfigError
	config := &Config{}

	err := config.Validate()
	assert.Error(t, err)

	// Should be a ConfigError
	var configErr *core.ConfigError
	assert.ErrorAs(t, err, &configErr)
	assert.Equal(t, "ClientID", configErr.Field)
}

func TestClient_OAuth2Config(t *testing.T) {
	config := &Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		TenantID:     "consumers",
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{"Mail.Read", "Mail.ReadWrite"},
	}

	client, err := New(config)
	assert.NoError(t, err)

	oauth2Config := client.oauth2Config
	assert.NotNil(t, oauth2Config)
	assert.Equal(t, config.ClientID, oauth2Config.ClientID)
	assert.Equal(t, config.ClientSecret, oauth2Config.ClientSecret)
	assert.Equal(t, config.RedirectURL, oauth2Config.RedirectURL)
	assert.Equal(t, config.Scopes, oauth2Config.Scopes)
}

func TestClient_DefaultScopes(t *testing.T) {
	// Config without explicit scopes
	config := &Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		TenantID:     "consumers",
		RedirectURL:  "http://localhost:8080/callback",
	}

	client, err := New(config)
	assert.NoError(t, err)

	// Should have default scopes
	assert.Equal(t, DefaultScopes(), client.oauth2Config.Scopes)
}
