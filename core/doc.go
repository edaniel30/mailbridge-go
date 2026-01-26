// Package core provides shared types and utilities used across email providers.
//
// This package contains common email data structures and interfaces that are
// provider-agnostic, allowing different email providers (Gmail, Outlook, etc.)
// to use the same normalized format.
//
// # Email Structure
//
// The Email type represents a normalized email message:
//
//	type Email struct {
//		ID          string         // Unique message identifier
//		ThreadID    string         // Thread/conversation identifier
//		Subject     string         // Email subject line
//		From        EmailAddress   // Sender information
//		To          []EmailAddress // Primary recipients
//		Cc          []EmailAddress // CC recipients
//		Bcc         []EmailAddress // BCC recipients
//		ReplyTo     []EmailAddress // Reply-to addresses
//		Date        time.Time      // Send/receive date
//		Body        EmailBody      // Message content
//		Snippet     string         // Short preview text
//		Labels      []string       // Labels/folders
//		Attachments []Attachment   // File attachments
//		IsRead      bool           // Read status
//		IsStarred   bool           // Starred/flagged status
//		IsDraft     bool           // Draft status
//	}
//
// # List Options
//
// ListOptions configures message listing queries:
//
//	opts := &core.ListOptions{
//		MaxResults: 20,                 // Limit number of results
//		PageToken:  "next-page-token",  // For pagination
//		Query:      "is:unread",        // Provider-specific query
//		Labels:     []string{"INBOX"},  // Filter by labels
//	}
//
// # Email Address
//
// EmailAddress represents an email address with optional display name:
//
//	addr := core.EmailAddress{
//		Email: "user@example.com",
//		Name:  "John Doe",
//	}
//
// # Email Body
//
// EmailBody contains message content in different formats:
//
//	body := core.EmailBody{
//		Text: "Plain text version",
//		HTML: "<html>HTML version</html>",
//	}
//
// # Attachments
//
// Attachment represents file attachments:
//
//	attachment := core.Attachment{
//		ID:       "attachment-id",
//		Filename: "document.pdf",
//		MimeType: "application/pdf",
//		Size:     1024,
//		Data:     nil, // Populated when downloaded
//	}
//
// # Error Handling
//
// ConfigError provides structured configuration validation errors:
//
//	err := core.NewConfigFieldError("client_id", "is required")
//
//	var configErr *core.ConfigError
//	if errors.As(err, &configErr) {
//		fmt.Printf("Field: %s, Message: %s\n",
//			configErr.Field, configErr.Message)
//	}
//
// # List Response
//
// ListResponse contains paginated message results:
//
//	response := &core.ListResponse{
//		Emails:        emails,
//		NextPageToken: "token-for-next-page",
//		TotalCount:    estimatedTotal,
//	}
//
// Use NextPageToken for pagination:
//
//	// First page
//	opts := &core.ListOptions{MaxResults: 20}
//	response, err := client.ListMessages(ctx, opts)
//
//	// Next page
//	if response.NextPageToken != "" {
//		opts.PageToken = response.NextPageToken
//		nextPage, err := client.ListMessages(ctx, opts)
//	}
//
// # Provider Independence
//
// Types in this package are designed to work with any email provider.
// Provider-specific clients (like gmail.Client) convert between their
// native formats and these normalized types, ensuring consistent APIs
// across different email services.
package core
