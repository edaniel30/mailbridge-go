package gmail

import (
	"context"
	"fmt"

	"github.com/danielrivera/mailbridge-go/core"
	"github.com/danielrivera/mailbridge-go/gmail/internal"
	"github.com/danielrivera/mailbridge-go/gmail/operations/labels"
	"github.com/danielrivera/mailbridge-go/gmail/operations/messages"
	"github.com/danielrivera/mailbridge-go/gmail/operations/watch"
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

// ensureConnected checks if the client is connected and returns an error if not
func (c *Client) ensureConnected() error {
	if !c.IsConnected() {
		return core.ErrNotConnected
	}
	return nil
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

// Message operations - delegate to operations/messages package

// ListMessages lists messages from Gmail
func (c *Client) ListMessages(ctx context.Context, opts *core.ListOptions) (*core.ListResponse, error) {
	if err := c.ensureConnected(); err != nil {
		return nil, err
	}
	return messages.ListMessages(ctx, c.service, opts)
}

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

// SendMessage sends an email message
func (c *Client) SendMessage(ctx context.Context, draft *core.Draft, opts *core.SendOptions) (*core.SendResponse, error) {
	if err := c.ensureConnected(); err != nil {
		return nil, err
	}
	return messages.SendMessage(ctx, c.service, draft, opts)
}

// Label operations - delegate to operations/labels package

// ListLabels lists all labels in the user's mailbox
func (c *Client) ListLabels(ctx context.Context) ([]*labels.Label, error) {
	if err := c.ensureConnected(); err != nil {
		return nil, err
	}
	return labels.ListLabels(ctx, c.service)
}

// GetLabel gets a specific label by ID
func (c *Client) GetLabel(ctx context.Context, labelID string) (*labels.Label, error) {
	if err := c.ensureConnected(); err != nil {
		return nil, err
	}
	return labels.GetLabel(ctx, c.service, labelID)
}

// FindLabelByName finds a label by its name
func (c *Client) FindLabelByName(ctx context.Context, name string) (*labels.Label, error) {
	if err := c.ensureConnected(); err != nil {
		return nil, err
	}
	return labels.FindLabelByName(ctx, c.service, name)
}

// CreateLabel creates a new label (folder)
func (c *Client) CreateLabel(ctx context.Context, name string) (*labels.Label, error) {
	if err := c.ensureConnected(); err != nil {
		return nil, err
	}
	return labels.CreateLabel(ctx, c.service, name)
}

// DeleteLabel deletes a label
func (c *Client) DeleteLabel(ctx context.Context, labelID string) error {
	if err := c.ensureConnected(); err != nil {
		return err
	}
	return labels.DeleteLabel(ctx, c.service, labelID)
}

// AddLabelToMessage adds a label to a message
func (c *Client) AddLabelToMessage(ctx context.Context, messageID string, labelID string) error {
	if err := c.ensureConnected(); err != nil {
		return err
	}
	return labels.AddLabelToMessage(ctx, c.service, messageID, labelID)
}

// RemoveLabelFromMessage removes a label from a message
func (c *Client) RemoveLabelFromMessage(ctx context.Context, messageID string, labelID string) error {
	if err := c.ensureConnected(); err != nil {
		return err
	}
	return labels.RemoveLabelFromMessage(ctx, c.service, messageID, labelID)
}

// MarkAsRead marks a message as read
func (c *Client) MarkAsRead(ctx context.Context, messageID string) error {
	if err := c.ensureConnected(); err != nil {
		return err
	}
	return labels.MarkAsRead(ctx, c.service, messageID)
}

// MarkAsUnread marks a message as unread
func (c *Client) MarkAsUnread(ctx context.Context, messageID string) error {
	if err := c.ensureConnected(); err != nil {
		return err
	}
	return labels.MarkAsUnread(ctx, c.service, messageID)
}

// MoveMessageToFolder moves a message to a specific folder/label
func (c *Client) MoveMessageToFolder(ctx context.Context, messageID string, folderName string) error {
	if err := c.ensureConnected(); err != nil {
		return err
	}
	return labels.MoveMessageToFolder(ctx, c.service, messageID, folderName)
}

// TrashMessage moves a message to trash (reversible)
func (c *Client) TrashMessage(ctx context.Context, messageID string) error {
	if err := c.ensureConnected(); err != nil {
		return err
	}
	return labels.TrashMessage(ctx, c.service, messageID)
}

// UntrashMessage removes a message from trash
func (c *Client) UntrashMessage(ctx context.Context, messageID string) error {
	if err := c.ensureConnected(); err != nil {
		return err
	}
	return labels.UntrashMessage(ctx, c.service, messageID)
}

// BatchTrashMessages moves multiple messages to trash
func (c *Client) BatchTrashMessages(ctx context.Context, messageIDs []string) error {
	if err := c.ensureConnected(); err != nil {
		return err
	}
	return labels.BatchTrashMessages(ctx, c.service, messageIDs)
}

// DeleteMessage permanently deletes a message (not reversible)
func (c *Client) DeleteMessage(ctx context.Context, messageID string) error {
	if err := c.ensureConnected(); err != nil {
		return err
	}
	return messages.DeleteMessage(ctx, c.service, messageID)
}

// BatchDeleteMessages permanently deletes multiple messages
func (c *Client) BatchDeleteMessages(ctx context.Context, messageIDs []string) error {
	if err := c.ensureConnected(); err != nil {
		return err
	}
	return messages.BatchDeleteMessages(ctx, c.service, messageIDs)
}

// BatchModifyMessages modifies labels on multiple messages
func (c *Client) BatchModifyMessages(ctx context.Context, req *core.BatchModifyRequest) error {
	if err := c.ensureConnected(); err != nil {
		return err
	}
	return labels.BatchModifyMessages(ctx, c.service, req.MessageIDs, req.AddLabelIDs, req.RemoveLabelIDs)
}

// BatchMarkAsRead marks multiple messages as read
func (c *Client) BatchMarkAsRead(ctx context.Context, messageIDs []string) error {
	if err := c.ensureConnected(); err != nil {
		return err
	}
	return labels.BatchMarkAsRead(ctx, c.service, messageIDs)
}

// BatchMarkAsUnread marks multiple messages as unread
func (c *Client) BatchMarkAsUnread(ctx context.Context, messageIDs []string) error {
	if err := c.ensureConnected(); err != nil {
		return err
	}
	return labels.BatchMarkAsUnread(ctx, c.service, messageIDs)
}

// BatchMoveToFolder moves multiple messages to a specific folder
func (c *Client) BatchMoveToFolder(ctx context.Context, messageIDs []string, folderName string) error {
	if err := c.ensureConnected(); err != nil {
		return err
	}
	return labels.BatchMoveToFolder(ctx, c.service, messageIDs, folderName)
}

// WatchMailbox sets up push notifications for the mailbox
func (c *Client) WatchMailbox(ctx context.Context, req *core.WatchRequest) (*core.WatchResponse, error) {
	if err := c.ensureConnected(); err != nil {
		return nil, err
	}
	return watch.WatchMailbox(ctx, c.service, req)
}

// StopWatch stops push notifications for the mailbox
func (c *Client) StopWatch(ctx context.Context) error {
	if err := c.ensureConnected(); err != nil {
		return err
	}
	return watch.StopWatch(ctx, c.service)
}

// GetHistory retrieves mailbox history starting from a history ID
func (c *Client) GetHistory(ctx context.Context, req *core.HistoryRequest) (*core.HistoryResponse, error) {
	if err := c.ensureConnected(); err != nil {
		return nil, err
	}
	return watch.GetHistory(ctx, c.service, req)
}
