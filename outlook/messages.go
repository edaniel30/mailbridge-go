package outlook

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"

	"github.com/danielrivera/mailbridge-go/core"
)

// ListMessages retrieves a list of email messages from the user's mailbox.
// It returns provider-agnostic core.Email types.
func (c *Client) ListMessages(ctx context.Context, opts *core.ListOptions) (*core.ListResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("client not connected")
	}

	config := &users.ItemMessagesRequestBuilderGetRequestConfiguration{}
	queryParams := &users.ItemMessagesRequestBuilderGetQueryParameters{}

	// Apply pagination
	if opts != nil {
		if opts.MaxResults > 0 {
			top := int32(opts.MaxResults)
			queryParams.Top = &top
		}
		if opts.PageToken != "" {
			skip := int32(0)
			if _, err := fmt.Sscanf(opts.PageToken, "%d", &skip); err == nil && skip > 0 {
				queryParams.Skip = &skip
			}
		}

		// Note: Label filtering is not supported in Outlook's ListMessages.
		// For folder-specific queries, use ListMessagesInFolder instead.
		// This is a limitation of the Graph API list endpoint.

		// Apply query filter
		if opts.Query != "" {
			queryParams.Search = &opts.Query
		}
	}

	// Select fields to retrieve
	selectFields := []string{
		"id", "subject", "from", "toRecipients", "ccRecipients", "bccRecipients",
		"receivedDateTime", "sentDateTime", "hasAttachments", "isRead", "body",
		"bodyPreview", "parentFolderId",
	}
	queryParams.Select = selectFields

	config.QueryParameters = queryParams

	messagesService := c.service.GetMeService().GetMessagesService()
	result, err := messagesService.List(ctx, config)
	if err != nil {
		return nil, handleODataError(fmt.Errorf("failed to list messages: %w", err))
	}

	messages := result.GetValue()
	emails := make([]*core.Email, 0, len(messages))

	for _, msg := range messages {
		email := c.convertMessage(msg)
		emails = append(emails, email)
	}

	// Calculate next page token
	var nextPageToken string
	if len(messages) > 0 && opts != nil && opts.MaxResults > 0 && int64(len(messages)) == opts.MaxResults {
		skip := int32(0)
		if opts.PageToken != "" {
			_, _ = fmt.Sscanf(opts.PageToken, "%d", &skip)
		}
		nextPageToken = fmt.Sprintf("%d", int64(skip)+opts.MaxResults)
	}

	return &core.ListResponse{
		Emails:        emails,
		NextPageToken: nextPageToken,
	}, nil
}

// GetMessage retrieves a single message by its ID.
func (c *Client) GetMessage(ctx context.Context, messageID string) (*core.Email, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("client not connected")
	}

	messagesService := c.service.GetMeService().GetMessagesService()
	message, err := messagesService.Get(ctx, messageID)
	if err != nil {
		return nil, handleODataError(fmt.Errorf("failed to get message %s: %w", messageID, err))
	}

	return c.convertMessage(message), nil
}

// GetAttachment retrieves a specific attachment from a message.
func (c *Client) GetAttachment(ctx context.Context, messageID, attachmentID string) (*core.Attachment, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("client not connected")
	}

	messagesService := c.service.GetMeService().GetMessagesService()
	attachment, err := messagesService.GetAttachment(ctx, messageID, attachmentID)
	if err != nil {
		return nil, handleODataError(fmt.Errorf("failed to get attachment %s from message %s: %w", attachmentID, messageID, err))
	}

	return convertAttachment(attachment), nil
}

// MarkAsRead marks a message as read.
func (c *Client) MarkAsRead(ctx context.Context, messageID string) error {
	if !c.IsConnected() {
		return fmt.Errorf("client not connected")
	}

	messagesService := c.service.GetMeService().GetMessagesService()
	if err := messagesService.MarkAsRead(ctx, messageID); err != nil {
		return handleODataError(fmt.Errorf("failed to mark message %s as read: %w", messageID, err))
	}

	return nil
}

// MarkAsUnread marks a message as unread.
func (c *Client) MarkAsUnread(ctx context.Context, messageID string) error {
	if !c.IsConnected() {
		return fmt.Errorf("client not connected")
	}

	messagesService := c.service.GetMeService().GetMessagesService()
	if err := messagesService.MarkAsUnread(ctx, messageID); err != nil {
		return handleODataError(fmt.Errorf("failed to mark message %s as unread: %w", messageID, err))
	}

	return nil
}

// MoveMessage moves a message to a different folder.
func (c *Client) MoveMessage(ctx context.Context, messageID, destinationFolderID string) error {
	if !c.IsConnected() {
		return fmt.Errorf("client not connected")
	}

	messagesService := c.service.GetMeService().GetMessagesService()
	if err := messagesService.Move(ctx, messageID, destinationFolderID); err != nil {
		return handleODataError(fmt.Errorf("failed to move message %s to folder %s: %w", messageID, destinationFolderID, err))
	}

	return nil
}

// DeleteMessage deletes a message permanently.
func (c *Client) DeleteMessage(ctx context.Context, messageID string) error {
	if !c.IsConnected() {
		return fmt.Errorf("client not connected")
	}

	messagesService := c.service.GetMeService().GetMessagesService()
	if err := messagesService.Delete(ctx, messageID); err != nil {
		return handleODataError(fmt.Errorf("failed to delete message %s: %w", messageID, err))
	}

	return nil
}

// convertMessage converts a Microsoft Graph Message to a core.Email.
// This is the adapter pattern implementation.
func (c *Client) convertMessage(msg models.Messageable) *core.Email {
	email := &core.Email{
		ID:      derefString(msg.GetId()),
		Subject: derefString(msg.GetSubject()),
	}

	// From
	if from := msg.GetFrom(); from != nil {
		if emailAddr := from.GetEmailAddress(); emailAddr != nil {
			email.From = core.EmailAddress{
				Name:  derefString(emailAddr.GetName()),
				Email: derefString(emailAddr.GetAddress()),
			}
		}
	}

	// To recipients
	if toRecipients := msg.GetToRecipients(); toRecipients != nil {
		email.To = make([]core.EmailAddress, 0, len(toRecipients))
		for _, recipient := range toRecipients {
			if emailAddr := recipient.GetEmailAddress(); emailAddr != nil {
				email.To = append(email.To, core.EmailAddress{
					Name:  derefString(emailAddr.GetName()),
					Email: derefString(emailAddr.GetAddress()),
				})
			}
		}
	}

	// CC recipients
	if ccRecipients := msg.GetCcRecipients(); ccRecipients != nil {
		email.Cc = make([]core.EmailAddress, 0, len(ccRecipients))
		for _, recipient := range ccRecipients {
			if emailAddr := recipient.GetEmailAddress(); emailAddr != nil {
				email.Cc = append(email.Cc, core.EmailAddress{
					Name:  derefString(emailAddr.GetName()),
					Email: derefString(emailAddr.GetAddress()),
				})
			}
		}
	}

	// BCC recipients
	if bccRecipients := msg.GetBccRecipients(); bccRecipients != nil {
		email.Bcc = make([]core.EmailAddress, 0, len(bccRecipients))
		for _, recipient := range bccRecipients {
			if emailAddr := recipient.GetEmailAddress(); emailAddr != nil {
				email.Bcc = append(email.Bcc, core.EmailAddress{
					Name:  derefString(emailAddr.GetName()),
					Email: derefString(emailAddr.GetAddress()),
				})
			}
		}
	}

	// Dates
	if receivedTime := msg.GetReceivedDateTime(); receivedTime != nil {
		email.Date = *receivedTime
	}
	if sentTime := msg.GetSentDateTime(); sentTime != nil {
		email.Date = *sentTime // Use sent time if available
	}

	// Read status
	if isRead := msg.GetIsRead(); isRead != nil {
		email.IsRead = *isRead
	}

	// Body
	if body := msg.GetBody(); body != nil {
		content := derefString(body.GetContent())
		contentType := body.GetContentType()

		if contentType != nil && *contentType == models.TEXT_BODYTYPE {
			email.Body.Text = content
		} else {
			email.Body.HTML = content
		}
	}

	// Snippet (preview)
	email.Snippet = derefString(msg.GetBodyPreview())

	// Attachments
	if hasAttachments := msg.GetHasAttachments(); hasAttachments != nil && *hasAttachments {
		// Note: Attachment details are not included in list response
		// Users need to call GetAttachment to download
		email.Attachments = []core.Attachment{}
	}

	// Labels (folder ID in Outlook)
	if folderID := msg.GetParentFolderId(); folderID != nil {
		email.Labels = []string{*folderID}
	}

	return email
}

// convertAttachment converts a Microsoft Graph Attachment to core.Attachment.
func convertAttachment(att models.Attachmentable) *core.Attachment {
	attachment := &core.Attachment{
		ID:       derefString(att.GetId()),
		Filename: derefString(att.GetName()),
		MimeType: derefString(att.GetContentType()),
	}

	// Size
	if size := att.GetSize(); size != nil {
		attachment.Size = int64(*size)
	}

	// Data (only for FileAttachment)
	if fileAtt, ok := att.(models.FileAttachmentable); ok {
		if contentBytes := fileAtt.GetContentBytes(); contentBytes != nil {
			attachment.Data = contentBytes
		}
	}

	return attachment
}

// Helper functions

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// parseEmailAddress parses an email address string into name and address components.
// Format: "Name <email@example.com>" or "email@example.com"
func parseEmailAddress(s string) core.EmailAddress {
	s = strings.TrimSpace(s)
	if s == "" {
		return core.EmailAddress{}
	}

	// Check for format: "Name <email@example.com>"
	if strings.Contains(s, "<") && strings.Contains(s, ">") {
		parts := strings.Split(s, "<")
		if len(parts) == 2 {
			name := strings.TrimSpace(parts[0])
			address := strings.TrimSpace(strings.TrimSuffix(parts[1], ">"))
			return core.EmailAddress{
				Name:  name,
				Email: address,
			}
		}
	}

	// Just an email address
	return core.EmailAddress{
		Email: s,
	}
}

// decodeBase64 decodes base64-encoded content.
func decodeBase64(data string) ([]byte, error) {
	// Try standard base64 first
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		// Try URL encoding
		decoded, err = base64.URLEncoding.DecodeString(data)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64: %w", err)
		}
	}
	return decoded, nil
}

// formatDate formats a time.Time to RFC3339 format.
func formatDate(t time.Time) string {
	return t.Format(time.RFC3339)
}
