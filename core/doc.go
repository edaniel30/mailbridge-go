// Package core provides provider-agnostic types and interfaces for email operations.
//
// This package defines the normalized data structures that all email providers
// (Gmail, Outlook, etc.) convert their native types into. This ensures a consistent
// API regardless of which provider you're using.
//
// Core Principle:
//
// The core package has ZERO dependencies on provider packages. Providers depend
// on core types, not the other way around. This architectural decision enables:
//   - Adding new providers without modifying core
//   - Switching between providers with minimal code changes
//   - Testing provider code in isolation
//   - Consistent user experience across all providers
//
// Key Types:
//
// Email - Normalized email message structure:
//
//	type Email struct {
//	    ID          string          // Provider's unique message ID
//	    ThreadID    string          // Conversation/thread ID
//	    Subject     string
//	    From        EmailAddress
//	    To          []EmailAddress
//	    Cc          []EmailAddress
//	    Bcc         []EmailAddress
//	    ReplyTo     []EmailAddress
//	    Date        time.Time
//	    Body        EmailBody       // Text and/or HTML
//	    Snippet     string          // Short preview
//	    Attachments []Attachment    // Metadata only (lazy loading)
//	    Labels      []string        // Folder IDs or label names
//	    IsRead      bool
//	    IsStarred   bool
//	    IsDraft     bool
//	}
//
// ListOptions - Options for listing messages:
//
//	type ListOptions struct {
//	    MaxResults int64      // Maximum messages to return
//	    PageToken  string     // For pagination
//	    Query      string     // Provider-specific search query
//	    Labels     []string   // Filter by labels/folders
//	}
//
// ListResponse - Response from list operations:
//
//	type ListResponse struct {
//	    Emails        []*Email
//	    NextPageToken string    // Use for next page
//	    TotalCount    int64     // Total available (if known)
//	}
//
// Attachment - Email attachment (metadata or full content):
//
//	type Attachment struct {
//	    ID       string
//	    Filename string
//	    MimeType string
//	    Size     int64
//	    Data     []byte  // Only populated when explicitly downloaded
//	}
//
// Label - Folder or label information:
//
//	type Label struct {
//	    ID              string
//	    Name            string
//	    Type            string  // "system" or "user"
//	    TotalMessages   int
//	    UnreadMessages  int
//	}
//
// Message - For sending messages:
//
//	type Message struct {
//	    To          []EmailAddress
//	    Cc          []EmailAddress
//	    Bcc         []EmailAddress
//	    Subject     string
//	    Body        EmailBody
//	    Attachments []Attachment
//	    InReplyTo   string       // Message ID being replied to
//	    References  []string     // Thread references
//	}
//
// Usage Example:
//
//	import (
//	    "github.com/danielrivera/mailbridge-go/core"
//	    "github.com/danielrivera/mailbridge-go/gmail"
//	)
//
//	// Create provider-specific client
//	gmailClient, _ := gmail.New(gmailConfig)
//
//	// Use core types for operations
//	response, err := gmailClient.ListMessages(ctx, &core.ListOptions{
//	    MaxResults: 50,
//	    Query:      "is:unread",
//	})
//
//	// Work with normalized core.Email objects
//	for _, email := range response.Emails {
//	    fmt.Printf("From: %s\n", email.From.Email)
//	    fmt.Printf("Subject: %s\n", email.Subject)
//	}
//
// Provider Independence:
//
// Code written against core types works with any provider:
//
//	// Function that works with any provider
//	func ProcessEmails(client interface{
//	    ListMessages(ctx context.Context, opts *core.ListOptions) (*core.ListResponse, error)
//	}) error {
//	    response, err := client.ListMessages(ctx, &core.ListOptions{
//	        MaxResults: 10,
//	    })
//	    if err != nil {
//	        return err
//	    }
//
//	    for _, email := range response.Emails {
//	        // Process email (same code for Gmail, Outlook, etc.)
//	    }
//	    return nil
//	}
//
//	// Works with Gmail
//	ProcessEmails(gmailClient)
//
//	// Works with Outlook
//	ProcessEmails(outlookClient)
//
// Error Handling:
//
// The core package provides ConfigError for configuration validation:
//
//	type ConfigError struct {
//	    Field   string
//	    Message string
//	}
//
//	// Example usage in provider code:
//	if config.ClientID == "" {
//	    return &core.ConfigError{
//	        Field:   "ClientID",
//	        Message: "client ID is required",
//	    }
//	}
//
// Design Patterns:
//
// 1. Adapter Pattern: Providers convert their native types to core types
//
//	// In provider code:
//	func (c *Client) convertMessage(nativeMsg *provider.Message) *core.Email {
//	    return &core.Email{
//	        ID:      nativeMsg.GetID(),
//	        Subject: nativeMsg.GetSubject(),
//	        // ... map all fields
//	    }
//	}
//
// 2. Lazy Loading: Attachments contain metadata; data loaded on demand
//
//	// List returns metadata only
//	emails, _ := client.ListMessages(ctx, opts)
//	for _, email := range emails {
//	    for _, att := range email.Attachments {
//	        fmt.Println(att.Filename, att.Size) // Metadata available
//	        // att.Data is empty
//	    }
//	}
//
//	// Explicit download when needed
//	attachment, _ := client.GetAttachment(ctx, messageID, attachmentID)
//	// Now attachment.Data contains the file content
//
// 3. Pagination: Use tokens for efficient large result sets
//
//	opts := &core.ListOptions{MaxResults: 100}
//	for {
//	    response, err := client.ListMessages(ctx, opts)
//	    if err != nil {
//	        return err
//	    }
//
//	    // Process response.Emails
//
//	    if response.NextPageToken == "" {
//	        break // No more pages
//	    }
//	    opts.PageToken = response.NextPageToken
//	}
//
// Adding New Providers:
//
// To add a new email provider:
//  1. Create a new package (e.g., yahoo/)
//  2. Implement client with provider-specific authentication
//  3. Implement methods that return core types:
//     - ListMessages(ctx, *core.ListOptions) (*core.ListResponse, error)
//     - GetMessage(ctx, messageID) (*core.Email, error)
//     - GetAttachment(ctx, messageID, attachmentID) (*core.Attachment, error)
//     - And other operations...
//  4. Write conversion functions from provider types to core types
//  5. Add tests using the same patterns as existing providers
//
// Type Conversion Guidelines:
//
// When implementing a provider, follow these guidelines:
//   - Always populate required fields (ID, Subject, From, Date)
//   - Use empty values for missing optional fields (don't use nil)
//   - Convert dates to time.Time in UTC
//   - Normalize email addresses (trim spaces, lowercase domains)
//   - Map provider-specific labels/folders to string array
//   - Include snippet/preview text (truncate if too long)
//   - Don't download attachments in list operations
//
// Thread Safety:
//
// Core types are NOT thread-safe. If you need to access the same Email
// object from multiple goroutines, use proper synchronization.
//
// Zero Values:
//
// All core types have meaningful zero values:
//   - Email{} represents an empty message
//   - ListOptions{} uses provider defaults
//   - EmailAddress{} represents an invalid address
//
// For implementation examples, see:
//   - github.com/danielrivera/mailbridge-go/gmail
//   - github.com/danielrivera/mailbridge-go/outlook
package core
