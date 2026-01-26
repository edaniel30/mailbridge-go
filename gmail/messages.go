package gmail

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/danielrivera/mailbridge-go/core"
	"google.golang.org/api/gmail/v1"
)

// ListMessages lists messages from Gmail
func (c *Client) ListMessages(ctx context.Context, opts *core.ListOptions) (*core.ListResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("client not connected")
	}

	messagesService := c.service.GetUsersService().GetMessagesService()
	call := messagesService.List("me")

	if opts != nil {
		if opts.MaxResults > 0 {
			call = call.MaxResults(opts.MaxResults)
		}
		if opts.PageToken != "" {
			call = call.PageToken(opts.PageToken)
		}
		if opts.Query != "" {
			call = call.Q(opts.Query)
		}
		if len(opts.Labels) > 0 {
			call = call.LabelIds(opts.Labels...)
		}
	}

	resp, err := call.Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	emails := make([]*core.Email, 0, len(resp.Messages))
	for _, msg := range resp.Messages {
		email, err := c.GetMessage(ctx, msg.Id)
		if err != nil {
			// Skip messages that can't be retrieved
			continue
		}
		emails = append(emails, email)
	}

	return &core.ListResponse{
		Emails:        emails,
		NextPageToken: resp.NextPageToken,
		TotalCount:    resp.ResultSizeEstimate,
	}, nil
}

// GetMessage retrieves a specific message by ID
func (c *Client) GetMessage(ctx context.Context, messageID string) (*core.Email, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("client not connected")
	}

	messagesService := c.service.GetUsersService().GetMessagesService()
	call := messagesService.Get("me", messageID)
	msg, err := call.Format("full").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	return c.convertMessage(msg), nil
}

// GetAttachment downloads an attachment by its ID from a specific message
func (c *Client) GetAttachment(ctx context.Context, messageID, attachmentID string) ([]byte, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("client not connected")
	}

	messagesService := c.service.GetUsersService().GetMessagesService()
	attachment, err := messagesService.GetAttachment("me", messageID, attachmentID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get attachment: %w", err)
	}

	// Decode base64url encoded data (Gmail uses base64url without padding)
	data, err := base64.RawURLEncoding.DecodeString(attachment.Data)
	if err != nil {
		// Try with padding if raw fails
		data, err = base64.URLEncoding.DecodeString(attachment.Data)
		if err != nil {
			// Try standard base64
			data, err = base64.StdEncoding.DecodeString(attachment.Data)
			if err != nil {
				return nil, fmt.Errorf("failed to decode attachment data: %w", err)
			}
		}
	}

	return data, nil
}

// convertMessage converts a Gmail message to normalized Email type
func (c *Client) convertMessage(msg *gmail.Message) *core.Email {
	email := &core.Email{
		ID:       msg.Id,
		ThreadID: msg.ThreadId,
		Snippet:  msg.Snippet,
		Labels:   msg.LabelIds,
	}

	// Parse headers
	headers := make(map[string]string)
	for _, header := range msg.Payload.Headers {
		headers[strings.ToLower(header.Name)] = header.Value
	}

	// Extract basic fields
	email.Subject = headers["subject"]
	email.From = parseEmailAddress(headers["from"])
	email.To = parseEmailAddresses(headers["to"])
	email.Cc = parseEmailAddresses(headers["cc"])
	email.Bcc = parseEmailAddresses(headers["bcc"])
	email.ReplyTo = parseEmailAddresses(headers["reply-to"])

	// Parse date
	if dateStr := headers["date"]; dateStr != "" {
		if date, err := parseEmailDate(dateStr); err == nil {
			email.Date = date
		}
	}

	// Extract body
	email.Body = extractBody(msg.Payload)

	// Check flags
	email.IsRead = !contains(msg.LabelIds, "UNREAD")
	email.IsStarred = contains(msg.LabelIds, "STARRED")
	email.IsDraft = contains(msg.LabelIds, "DRAFT")

	// Extract attachments info (without data)
	email.Attachments = extractAttachments(msg.Payload)

	return email
}

// extractBody extracts text and HTML body from message payload
func extractBody(payload *gmail.MessagePart) core.EmailBody {
	body := core.EmailBody{}

	var extractPart func(*gmail.MessagePart)
	extractPart = func(part *gmail.MessagePart) {
		if part == nil {
			return
		}

		// Check current part
		if part.MimeType == "text/plain" && body.Text == "" {
			body.Text = decodeBody(part.Body.Data)
		} else if part.MimeType == "text/html" && body.HTML == "" {
			body.HTML = decodeBody(part.Body.Data)
		}

		// Recursively check parts
		for _, p := range part.Parts {
			extractPart(p)
		}
	}

	extractPart(payload)
	return body
}

// extractAttachments extracts attachment metadata from message payload
func extractAttachments(payload *gmail.MessagePart) []core.Attachment {
	var attachments []core.Attachment

	var extractPart func(*gmail.MessagePart)
	extractPart = func(part *gmail.MessagePart) {
		if part == nil {
			return
		}

		// Check if this part is an attachment
		if part.Filename != "" && part.Body != nil {
			attachments = append(attachments, core.Attachment{
				ID:       part.Body.AttachmentId,
				Filename: part.Filename,
				MimeType: part.MimeType,
				Size:     part.Body.Size,
			})
		}

		// Recursively check parts
		for _, p := range part.Parts {
			extractPart(p)
		}
	}

	extractPart(payload)
	return attachments
}

// decodeBody decodes base64url encoded body data
func decodeBody(data string) string {
	if data == "" {
		return ""
	}

	// Gmail uses base64url without padding
	decoded, err := base64.RawURLEncoding.DecodeString(data)
	if err != nil {
		// Try with padding if raw fails
		decoded, err = base64.URLEncoding.DecodeString(data)
		if err != nil {
			// Try standard base64
			decoded, err = base64.StdEncoding.DecodeString(data)
			if err != nil {
				return ""
			}
		}
	}

	return string(decoded)
}

// parseEmailAddress parses a single email address
func parseEmailAddress(addr string) core.EmailAddress {
	if addr == "" {
		return core.EmailAddress{}
	}

	// Simple parsing: "Name <email@example.com>" or "email@example.com"
	if strings.Contains(addr, "<") && strings.Contains(addr, ">") {
		parts := strings.SplitN(addr, "<", 2)
		name := strings.TrimSpace(parts[0])
		email := strings.Trim(parts[1], ">")
		return core.EmailAddress{
			Name:  strings.Trim(name, "\""),
			Email: email,
		}
	}

	return core.EmailAddress{Email: addr}
}

// parseEmailAddresses parses multiple email addresses separated by comma
func parseEmailAddresses(addrs string) []core.EmailAddress {
	if addrs == "" {
		return nil
	}

	var result []core.EmailAddress
	for _, addr := range strings.Split(addrs, ",") {
		if parsed := parseEmailAddress(strings.TrimSpace(addr)); parsed.Email != "" {
			result = append(result, parsed)
		}
	}

	return result
}

// parseEmailDate parses email date header
func parseEmailDate(dateStr string) (time.Time, error) {
	// Try common email date formats
	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		"Mon, 2 Jan 2006 15:04:05 -0700",
		"2 Jan 2006 15:04:05 -0700",
		time.RFC3339,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// contains checks if a slice contains a value
func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
