package gmail

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/edaniel30/mailbridge-go/core"
	"github.com/edaniel30/mailbridge-go/gmail/internal"
	"github.com/edaniel30/mailbridge-go/gmail/operations/labels"
	"github.com/edaniel30/mailbridge-go/gmail/operations/messages"
	"github.com/edaniel30/mailbridge-go/gmail/operations/watch"
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

// ConnectWithToken establishes connection using a provided token
func (c *Client) ConnectWithToken(ctx context.Context, token *oauth2.Token) error {
	c.token = token
	httpClient := c.oauth2Config.Client(ctx, c.token)

	service, err := gmail.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return fmt.Errorf("failed to create gmail service: %w", err)
	}

	c.service = internal.NewRealGmailService(service)
	return nil
}

// ConnectWithHTTPClient establishes connection using a custom HTTP client
// This allows using custom token sources (e.g., with automatic token refresh and persistence)
func (c *Client) ConnectWithHTTPClient(ctx context.Context, httpClient *http.Client) error {
	service, err := gmail.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return fmt.Errorf("failed to create gmail service: %w", err)
	}

	c.service = internal.NewRealGmailService(service)
	return nil
}

// IsConnected returns true if the client is connected to Gmail API
func (c *Client) IsConnected() bool {
	return c.service != nil
}

// GetToken returns the current OAuth2 token
func (c *Client) GetToken() *oauth2.Token {
	return c.token
}

// SetToken sets the OAuth2 token without establishing connection
func (c *Client) SetToken(token *oauth2.Token) {
	c.token = token
}

// Connect establishes connection using the current token
func (c *Client) Connect(ctx context.Context) error {
	if c.token == nil {
		return fmt.Errorf("no token available, call SetToken or ExchangeCode first")
	}
	return c.ConnectWithToken(ctx, c.token)
}

// RefreshToken refreshes the OAuth2 token
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
	return newToken, nil
}

// ensureConnected checks if the client is connected and returns an error if not
func (c *Client) ensureConnected() error {
	if !c.IsConnected() {
		return core.ErrNotConnected
	}
	return nil
}

// RevokeToken revokes the OAuth2 token in Google's servers
// This method accepts the token as a parameter to avoid modifying client state
func (c *Client) RevokeToken(ctx context.Context, token *oauth2.Token) error {
	if token == nil {
		return fmt.Errorf("no token to revoke")
	}

	// Use access token for revocation (Google accepts both access_token and refresh_token)
	tokenToRevoke := token.AccessToken
	if tokenToRevoke == "" && token.RefreshToken != "" {
		tokenToRevoke = token.RefreshToken
	}

	if tokenToRevoke == "" {
		return fmt.Errorf("no valid token to revoke")
	}

	// Google's token revocation endpoint
	revokeURL := fmt.Sprintf("https://oauth2.googleapis.com/revoke?token=%s", tokenToRevoke)

	req, err := http.NewRequestWithContext(ctx, "POST", revokeURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create revoke request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Google returns 200 OK if successful, or if token is already invalid/revoked
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("token revocation failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Close closes the Gmail client and cleans up resources
func (c *Client) Close() error {
	c.service = nil
	c.token = nil
	return nil
}

// UserProfile represents the Gmail user's profile information
type UserProfile struct {
	EmailAddress string
}

// GetUserProfile retrieves the authenticated user's profile
func (c *Client) GetUserProfile(ctx context.Context) (*UserProfile, error) {
	if err := c.ensureConnected(); err != nil {
		return nil, err
	}

	profile, err := c.service.GetUsersService().GetProfile("me").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	return &UserProfile{
		EmailAddress: profile.EmailAddress,
	}, nil
}

// Message operations

// GetMessage retrieves a specific message by ID
func (c *Client) GetMessage(ctx context.Context, messageID string) (*core.Email, error) {
	if err := c.ensureConnected(); err != nil {
		return nil, err
	}
	return messages.GetMessage(ctx, c.service, messageID)
}

// GetAttachment downloads an attachment by its ID from a specific message
func (c *Client) GetAttachment(ctx context.Context, messageID, attachmentID string) ([]byte, error) {
	if err := c.ensureConnected(); err != nil {
		return nil, err
	}
	return messages.GetAttachment(ctx, c.service, messageID, attachmentID)
}

// CreateLabel creates a new label (folder)
func (c *Client) CreateLabel(ctx context.Context, name string) (*labels.Label, error) {
	if err := c.ensureConnected(); err != nil {
		return nil, err
	}
	return labels.CreateLabel(ctx, c.service, name)
}

// MarkAsRead marks a message as read
func (c *Client) MarkAsRead(ctx context.Context, messageID string) error {
	if err := c.ensureConnected(); err != nil {
		return err
	}
	return labels.MarkAsRead(ctx, c.service, messageID)
}

// MoveMessageToFolder moves a message to a specific folder/label
func (c *Client) MoveMessageToFolder(ctx context.Context, messageID string, folderName string) error {
	if err := c.ensureConnected(); err != nil {
		return err
	}
	return labels.MoveMessageToFolder(ctx, c.service, messageID, folderName)
}

// Watch operations for push notifications

// WatchMailbox sets up push notifications for the mailbox
func (c *Client) WatchMailbox(ctx context.Context, req *core.WatchRequest) (*core.WatchResponse, error) {
	if err := c.ensureConnected(); err != nil {
		return nil, err
	}
	return watch.WatchMailbox(ctx, c.service, req)
}

// StopWatch stops push notifications for the mailbox
// userID can be an email address or "me" for the currently authenticated user
func (c *Client) StopWatch(ctx context.Context, userID string) error {
	if err := c.ensureConnected(); err != nil {
		return err
	}
	return watch.StopWatch(ctx, c.service, userID)
}

// GetHistory retrieves mailbox history starting from a history ID
func (c *Client) GetHistory(ctx context.Context, req *core.HistoryRequest) (*core.HistoryResponse, error) {
	if err := c.ensureConnected(); err != nil {
		return nil, err
	}
	return watch.GetHistory(ctx, c.service, req)
}
