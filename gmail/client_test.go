package gmail

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edaniel30/mailbridge-go/gmail/testing/mocks"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
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

	assert.Nil(t, client.GetToken())

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

	assert.False(t, client.IsConnected())
}

func TestClient_Close(t *testing.T) {
	client := newTestClient(t)

	client.SetToken(&oauth2.Token{AccessToken: "test"})

	err := client.Close()
	assert.NoError(t, err)

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

func TestClient_ConnectWithHTTPClient(t *testing.T) {
	t.Run("successful connection", func(t *testing.T) {
		client := newTestClient(t)
		ctx := context.Background()

		httpClient := &http.Client{}

		err := client.ConnectWithHTTPClient(ctx, httpClient)

		if err != nil {
			assert.Contains(t, err.Error(), "failed to create gmail service")
		} else {
			assert.True(t, client.IsConnected())
		}
	})

	t.Run("validates service is set on success", func(t *testing.T) {
		client := newTestClient(t)

		// Initially not connected
		assert.False(t, client.IsConnected())

		ctx := context.Background()
		httpClient := &http.Client{}
		_ = client.ConnectWithHTTPClient(ctx, httpClient)
	})
}

func TestClient_RevokeToken(t *testing.T) {
	t.Run("nil token", func(t *testing.T) {
		client := newTestClient(t)
		ctx := context.Background()

		err := client.RevokeToken(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no token to revoke")
	})

	t.Run("empty token", func(t *testing.T) {
		client := newTestClient(t)
		ctx := context.Background()

		err := client.RevokeToken(ctx, &oauth2.Token{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no valid token to revoke")
	})

	t.Run("validates request format", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

			body, _ := io.ReadAll(r.Body)
			bodyStr := string(body)
			assert.Contains(t, bodyStr, "token=")
			assert.NotContains(t, r.URL.RawQuery, "token")

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := newTestClient(t)
		ctx := context.Background()

		token := &oauth2.Token{
			AccessToken: "test-access-token",
		}

		err := client.RevokeToken(ctx, token)
		assert.Error(t, err)
	})

	t.Run("prefers access token over refresh token", func(t *testing.T) {
		client := newTestClient(t)
		ctx := context.Background()

		token := &oauth2.Token{
			AccessToken:  "test-access",
			RefreshToken: "test-refresh",
		}

		err := client.RevokeToken(ctx, token)
		assert.Error(t, err)
	})
}

func TestClient_GetUserProfile(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client := newTestClient(t)
		ctx := context.Background()

		_, err := client.GetUserProfile(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})

	t.Run("successful retrieval", func(t *testing.T) {
		client := newTestClient(t)
		ctx := context.Background()

		mockService := &mocks.MockGmailService{}
		mockUsersService := &mocks.MockUsersService{}
		mockProfileCall := &mocks.MockUsersGetProfileCall{}

		mockService.On("GetUsersService").Return(mockUsersService)
		mockUsersService.On("GetProfile", "me").Return(mockProfileCall)
		mockProfileCall.On("Context", ctx).Return(mockProfileCall)
		mockProfileCall.On("Do").Return(&gmail.Profile{
			EmailAddress: "test@example.com",
		}, nil)

		client.service = mockService

		profile, err := client.GetUserProfile(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, profile)
		assert.Equal(t, "test@example.com", profile.EmailAddress)

		mockService.AssertExpectations(t)
		mockUsersService.AssertExpectations(t)
		mockProfileCall.AssertExpectations(t)
	})
}

func TestClient_RefreshToken_Reconnect(t *testing.T) {
	t.Run("no token to refresh", func(t *testing.T) {
		client := newTestClient(t)
		ctx := context.Background()

		_, err := client.RefreshToken(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no token to refresh")
	})

	t.Run("validates reconnection behavior", func(t *testing.T) {
		client := newTestClient(t)

		client.SetToken(&oauth2.Token{
			AccessToken:  "old-token",
			RefreshToken: "refresh-token",
		})

		mockService := &mocks.MockGmailService{}
		client.service = mockService

		ctx := context.Background()
		_, err := client.RefreshToken(ctx)

		if err != nil {
			assert.Contains(t, err.Error(), "failed to")
		}
	})
}
