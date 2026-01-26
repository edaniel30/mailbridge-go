package gmail

import (
	"context"
	"fmt"

	"github.com/danielrivera/mailbridge-go/gmail/internal"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// Client represents a Gmail API client
type Client struct {
	config       *Config
	oauth2Config *oauth2.Config
	service      internal.GmailService
	token        *oauth2.Token
}

// New creates a new Gmail client
func New(config *Config) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &Client{
		config:       config,
		oauth2Config: config.ToOAuth2Config(),
	}, nil
}

// GetAuthURL returns the OAuth2 authorization URL
func (c *Client) GetAuthURL(state string) string {
	return c.oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

// ExchangeCode exchanges an authorization code for an access token
func (c *Client) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := c.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	c.token = token
	return token, nil
}

// SetToken sets the OAuth2 token for the client
func (c *Client) SetToken(token *oauth2.Token) {
	c.token = token
}

// Connect establishes connection to Gmail API using the stored token
func (c *Client) Connect(ctx context.Context) error {
	if c.token == nil {
		return fmt.Errorf("no token available, please authenticate first")
	}

	httpClient := c.oauth2Config.Client(ctx, c.token)

	service, err := gmail.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return fmt.Errorf("failed to create gmail service: %w", err)
	}

	c.service = internal.NewRealGmailService(service)
	return nil
}

// SetService sets the Gmail service (used for testing)
func (c *Client) SetService(service internal.GmailService) {
	c.service = service
}

// ConnectWithToken establishes connection using a provided token
func (c *Client) ConnectWithToken(ctx context.Context, token *oauth2.Token) error {
	c.SetToken(token)
	return c.Connect(ctx)
}

// IsConnected returns true if the client is connected to Gmail API
func (c *Client) IsConnected() bool {
	return c.service != nil
}

// GetToken returns the current OAuth2 token
func (c *Client) GetToken() *oauth2.Token {
	return c.token
}

// RefreshToken refreshes the OAuth2 token if needed
func (c *Client) RefreshToken(ctx context.Context) (*oauth2.Token, error) {
	if c.token == nil {
		return nil, fmt.Errorf("no token to refresh")
	}

	tokenSource := c.oauth2Config.TokenSource(ctx, c.token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	c.token = newToken

	// Reconnect with new token
	if err := c.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to reconnect after token refresh: %w", err)
	}

	return newToken, nil
}

// Close closes the Gmail client and cleans up resources
func (c *Client) Close() error {
	c.service = nil
	c.token = nil
	return nil
}
