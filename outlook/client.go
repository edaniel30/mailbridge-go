// Package outlook provides integration with Microsoft Outlook/Exchange via Microsoft Graph API.
// It follows the same architectural patterns as the Gmail provider, with provider-agnostic types
// from the core package and interface-based dependency injection for testability.
package outlook

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	abstractions "github.com/microsoft/kiota-abstractions-go"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models/odataerrors"
	"golang.org/x/oauth2"

	"github.com/danielrivera/mailbridge-go/outlook/internal"
)

// Client provides access to Microsoft Outlook/Exchange email operations via Microsoft Graph API.
// It uses OAuth2 for authentication and converts all provider-specific types to core.Email types.
type Client struct {
	config       *Config
	oauth2Config *oauth2.Config
	token        *oauth2.Token
	service      internal.GraphService
}

// New creates a new Outlook client with the given configuration.
// It validates the configuration but does not establish a connection.
// Use ConnectWithToken or ConnectWithAuthCode to establish a connection.
func New(config *Config) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &Client{
		config:       config,
		oauth2Config: config.ToOAuth2Config(),
	}, nil
}

// GetAuthURL returns the OAuth2 authorization URL for user consent.
// The state parameter should be a random string to prevent CSRF attacks.
func (c *Client) GetAuthURL(state string) string {
	return c.oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ConnectWithAuthCode exchanges an authorization code for an access token
// and establishes a connection to the Microsoft Graph API.
func (c *Client) ConnectWithAuthCode(ctx context.Context, authCode string) error {
	token, err := c.oauth2Config.Exchange(ctx, authCode)
	if err != nil {
		return fmt.Errorf("failed to exchange auth code for token: %w", err)
	}

	return c.ConnectWithToken(ctx, token)
}

// ConnectWithToken establishes a connection to Microsoft Graph API using an existing OAuth2 token.
// The token should have the required scopes (Mail.Read, Mail.ReadWrite, offline_access).
func (c *Client) ConnectWithToken(ctx context.Context, token *oauth2.Token) error {
	if token == nil {
		return fmt.Errorf("token cannot be nil")
	}

	c.token = token

	// Create an HTTP client with the OAuth2 token
	httpClient := c.oauth2Config.Client(ctx, token)

	// Create authentication provider using the token
	authProvider := &oauth2AuthProvider{
		token:      token,
		httpClient: httpClient,
	}

	// Create Graph client
	adapter, err := msgraphsdk.NewGraphRequestAdapter(authProvider)
	if err != nil {
		return fmt.Errorf("failed to create request adapter: %w", err)
	}

	graphClient := msgraphsdk.NewGraphServiceClient(adapter)

	// Wrap in our interface
	c.service = internal.NewRealGraphService(graphClient)

	return nil
}

// IsConnected returns true if the client is connected to Microsoft Graph API.
func (c *Client) IsConnected() bool {
	return c.service != nil
}

// GetToken returns the current OAuth2 token.
// Users should persist this token for future use.
func (c *Client) GetToken() *oauth2.Token {
	return c.token
}

// RefreshToken refreshes the OAuth2 token if it has expired or is about to expire.
// Returns the new token, which should be persisted.
func (c *Client) RefreshToken(ctx context.Context) (*oauth2.Token, error) {
	if c.token == nil {
		return nil, fmt.Errorf("no token to refresh")
	}

	tokenSource := c.oauth2Config.TokenSource(ctx, c.token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Reconnect with new token
	if err := c.ConnectWithToken(ctx, newToken); err != nil {
		return nil, fmt.Errorf("failed to reconnect with refreshed token: %w", err)
	}

	return newToken, nil
}

// SetService sets the internal Graph service (for testing).
func (c *Client) SetService(service internal.GraphService) {
	c.service = service
}

// handleODataError converts OData errors to readable error messages.
func handleODataError(err error) error {
	if err == nil {
		return nil
	}

	var odataErr *odataerrors.ODataError
	if errors.As(err, &odataErr) {
		if terr := odataErr.GetErrorEscaped(); terr != nil {
			code := ""
			message := ""
			if terr.GetCode() != nil {
				code = *terr.GetCode()
			}
			if terr.GetMessage() != nil {
				message = *terr.GetMessage()
			}
			return fmt.Errorf("microsoft graph error [%s]: %s", code, message)
		}
		return fmt.Errorf("microsoft graph error: %s", odataErr.Error())
	}

	return err
}

// oauth2AuthProvider implements the Kiota authentication provider interface
// using an OAuth2 token for delegated authentication flow.
type oauth2AuthProvider struct {
	token      *oauth2.Token
	httpClient *http.Client
}

// AuthenticateRequest adds the OAuth2 bearer token to the request.
func (p *oauth2AuthProvider) AuthenticateRequest(ctx context.Context, request *abstractions.RequestInformation, additionalAuthenticationContext map[string]interface{}) error {
	if p.token == nil {
		return fmt.Errorf("no token available")
	}

	// Add the bearer token to the Authorization header
	request.Headers.Add("Authorization", "Bearer "+p.token.AccessToken)
	return nil
}
