package gmail

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
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
			name: "invalid config",
			config: &Config{
				ClientID: "test-id",
				// Missing required fields
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.NotNil(t, client.config)
				assert.NotNil(t, client.oauth2Config)
			}
		})
	}
}

func TestClient_GetAuthURL(t *testing.T) {
	client := newTestClient(t)

	authURL := client.GetAuthURL("test-state")

	assert.NotEmpty(t, authURL)
	assert.Contains(t, authURL, "test-id")
	assert.Contains(t, authURL, "test-state")
	assert.Contains(t, authURL, "redirect_uri")
}

func TestClient_SetToken(t *testing.T) {
	client := newTestClient(t)

	token := &oauth2.Token{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "Bearer",
	}

	client.SetToken(token)

	assert.Equal(t, token, client.token)
}

func TestClient_GetToken(t *testing.T) {
	client := newTestClient(t)

	// Initially no token
	assert.Nil(t, client.GetToken())

	// Set token
	token := &oauth2.Token{
		AccessToken: "test-token",
	}
	client.SetToken(token)

	// Get token
	retrievedToken := client.GetToken()
	assert.Equal(t, token, retrievedToken)
}

func TestClient_IsConnected(t *testing.T) {
	client := newTestClient(t)

	// Initially not connected
	assert.False(t, client.IsConnected())
}

func TestClient_Close(t *testing.T) {
	client := newTestClient(t)

	// Set some state
	client.SetToken(&oauth2.Token{AccessToken: "test"})

	// Close
	err := client.Close()
	assert.NoError(t, err)

	// Verify state is cleared
	assert.Nil(t, client.service)
	assert.Nil(t, client.token)
}

func TestClient_ConnectWithoutToken(t *testing.T) {
	client := newTestClient(t)

	ctx := context.Background()

	// Try to connect without token
	err := client.Connect(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no token available")
}
