package outlook

import (
	"context"
	"fmt"

	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"

	"github.com/danielrivera/mailbridge-go/core"
)

// ListFolders retrieves all mail folders (similar to Gmail labels).
func (c *Client) ListFolders(ctx context.Context) ([]*core.Label, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("client not connected")
	}

	foldersService := c.service.GetMeService().GetMailFoldersService()
	result, err := foldersService.List(ctx)
	if err != nil {
		return nil, handleODataError(fmt.Errorf("failed to list folders: %w", err))
	}

	folders := result.GetValue()
	labels := make([]*core.Label, 0, len(folders))

	for _, folder := range folders {
		labels = append(labels, convertFolder(folder))
	}

	return labels, nil
}

// GetFolder retrieves a specific folder by its ID.
func (c *Client) GetFolder(ctx context.Context, folderID string) (*core.Label, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("client not connected")
	}

	foldersService := c.service.GetMeService().GetMailFoldersService()
	folder, err := foldersService.Get(ctx, folderID)
	if err != nil {
		return nil, handleODataError(fmt.Errorf("failed to get folder %s: %w", folderID, err))
	}

	return convertFolder(folder), nil
}

// CreateFolder creates a new mail folder.
func (c *Client) CreateFolder(ctx context.Context, name string) (*core.Label, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("client not connected")
	}

	if name == "" {
		return nil, fmt.Errorf("folder name cannot be empty")
	}

	foldersService := c.service.GetMeService().GetMailFoldersService()
	folder, err := foldersService.Create(ctx, name)
	if err != nil {
		return nil, handleODataError(fmt.Errorf("failed to create folder %s: %w", name, err))
	}

	return convertFolder(folder), nil
}

// UpdateFolder updates a folder's display name.
func (c *Client) UpdateFolder(ctx context.Context, folderID, newName string) (*core.Label, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("client not connected")
	}

	if newName == "" {
		return nil, fmt.Errorf("folder name cannot be empty")
	}

	foldersService := c.service.GetMeService().GetMailFoldersService()
	folder, err := foldersService.Update(ctx, folderID, newName)
	if err != nil {
		return nil, handleODataError(fmt.Errorf("failed to update folder %s: %w", folderID, err))
	}

	return convertFolder(folder), nil
}

// DeleteFolder deletes a mail folder.
func (c *Client) DeleteFolder(ctx context.Context, folderID string) error {
	if !c.IsConnected() {
		return fmt.Errorf("client not connected")
	}

	foldersService := c.service.GetMeService().GetMailFoldersService()
	if err := foldersService.Delete(ctx, folderID); err != nil {
		return handleODataError(fmt.Errorf("failed to delete folder %s: %w", folderID, err))
	}

	return nil
}

// ListMessagesInFolder retrieves messages from a specific folder.
func (c *Client) ListMessagesInFolder(ctx context.Context, folderID string, opts *core.ListOptions) (*core.ListResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("client not connected")
	}

	config := &users.ItemMailFoldersItemMessagesRequestBuilderGetRequestConfiguration{}
	queryParams := &users.ItemMailFoldersItemMessagesRequestBuilderGetQueryParameters{}

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

	foldersService := c.service.GetMeService().GetMailFoldersService()
	result, err := foldersService.GetMessages(ctx, folderID, config)
	if err != nil {
		return nil, handleODataError(fmt.Errorf("failed to list messages in folder %s: %w", folderID, err))
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

// convertFolder converts a Microsoft Graph MailFolder to core.Label.
func convertFolder(folder models.MailFolderable) *core.Label {
	label := &core.Label{
		ID:   derefString(folder.GetId()),
		Name: derefString(folder.GetDisplayName()),
	}

	// Map system folders to types
	switch label.Name {
	case "Inbox":
		label.Type = "system"
	case "Sent Items", "Drafts", "Deleted Items", "Junk Email":
		label.Type = "system"
	default:
		label.Type = "user"
	}

	// Message counts
	if totalCount := folder.GetTotalItemCount(); totalCount != nil {
		label.TotalMessages = int(*totalCount)
	}

	if unreadCount := folder.GetUnreadItemCount(); unreadCount != nil {
		label.UnreadMessages = int(*unreadCount)
	}

	return label
}

// Well-known folder IDs in Microsoft Graph
const (
	FolderInbox        = "inbox"
	FolderDrafts       = "drafts"
	FolderSentItems    = "sentitems"
	FolderDeletedItems = "deleteditems"
	FolderJunkEmail    = "junkemail"
	FolderOutbox       = "outbox"
	FolderArchive      = "archive"
)
