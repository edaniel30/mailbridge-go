package internal

import (
	"context"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
)

// RealGraphService wraps the Microsoft Graph SDK client.
type RealGraphService struct {
	client *msgraphsdk.GraphServiceClient
}

// NewRealGraphService creates a new RealGraphService.
func NewRealGraphService(client *msgraphsdk.GraphServiceClient) *RealGraphService {
	return &RealGraphService{client: client}
}

// GetMeService returns the service for the authenticated user.
func (r *RealGraphService) GetMeService() MeService {
	return &realMeService{client: r.client}
}

// realMeService implements MeService.
type realMeService struct {
	client *msgraphsdk.GraphServiceClient
}

// GetMessagesService returns the messages service.
func (r *realMeService) GetMessagesService() MessagesService {
	return &realMessagesService{client: r.client}
}

// GetMailFoldersService returns the mail folders service.
func (r *realMeService) GetMailFoldersService() MailFoldersService {
	return &realMailFoldersService{client: r.client}
}

// realMessagesService implements MessagesService.
type realMessagesService struct {
	client *msgraphsdk.GraphServiceClient
}

// List retrieves a list of messages.
func (r *realMessagesService) List(ctx context.Context, config *users.ItemMessagesRequestBuilderGetRequestConfiguration) (models.MessageCollectionResponseable, error) {
	return r.client.Me().Messages().Get(ctx, config)
}

// Get retrieves a specific message by ID.
func (r *realMessagesService) Get(ctx context.Context, messageID string) (models.Messageable, error) {
	return r.client.Me().Messages().ByMessageId(messageID).Get(ctx, nil)
}

// GetAttachments retrieves all attachments for a message.
func (r *realMessagesService) GetAttachments(ctx context.Context, messageID string) ([]models.Attachmentable, error) {
	result, err := r.client.Me().Messages().ByMessageId(messageID).Attachments().Get(ctx, nil)
	if err != nil {
		return nil, err
	}
	return result.GetValue(), nil
}

// GetAttachment retrieves a specific attachment.
func (r *realMessagesService) GetAttachment(ctx context.Context, messageID, attachmentID string) (models.Attachmentable, error) {
	return r.client.Me().Messages().ByMessageId(messageID).Attachments().ByAttachmentId(attachmentID).Get(ctx, nil)
}

// MarkAsRead marks a message as read.
func (r *realMessagesService) MarkAsRead(ctx context.Context, messageID string) error {
	message := models.NewMessage()
	isRead := true
	message.SetIsRead(&isRead)
	_, err := r.client.Me().Messages().ByMessageId(messageID).Patch(ctx, message, nil)
	return err
}

// MarkAsUnread marks a message as unread.
func (r *realMessagesService) MarkAsUnread(ctx context.Context, messageID string) error {
	message := models.NewMessage()
	isRead := false
	message.SetIsRead(&isRead)
	_, err := r.client.Me().Messages().ByMessageId(messageID).Patch(ctx, message, nil)
	return err
}

// Move moves a message to a different folder.
func (r *realMessagesService) Move(ctx context.Context, messageID, destinationFolderID string) error {
	body := users.NewItemMessagesItemMovePostRequestBody()
	body.SetDestinationId(&destinationFolderID)
	_, err := r.client.Me().Messages().ByMessageId(messageID).Move().Post(ctx, body, nil)
	return err
}

// Delete deletes a message.
func (r *realMessagesService) Delete(ctx context.Context, messageID string) error {
	return r.client.Me().Messages().ByMessageId(messageID).Delete(ctx, nil)
}

// realMailFoldersService implements MailFoldersService.
type realMailFoldersService struct {
	client *msgraphsdk.GraphServiceClient
}

// List retrieves all mail folders.
func (r *realMailFoldersService) List(ctx context.Context) (models.MailFolderCollectionResponseable, error) {
	return r.client.Me().MailFolders().Get(ctx, nil)
}

// Get retrieves a specific folder by ID.
func (r *realMailFoldersService) Get(ctx context.Context, folderID string) (models.MailFolderable, error) {
	return r.client.Me().MailFolders().ByMailFolderId(folderID).Get(ctx, nil)
}

// Create creates a new mail folder.
func (r *realMailFoldersService) Create(ctx context.Context, name string) (models.MailFolderable, error) {
	folder := models.NewMailFolder()
	folder.SetDisplayName(&name)
	return r.client.Me().MailFolders().Post(ctx, folder, nil)
}

// Update updates a folder's display name.
func (r *realMailFoldersService) Update(ctx context.Context, folderID, newName string) (models.MailFolderable, error) {
	folder := models.NewMailFolder()
	folder.SetDisplayName(&newName)
	return r.client.Me().MailFolders().ByMailFolderId(folderID).Patch(ctx, folder, nil)
}

// Delete deletes a mail folder.
func (r *realMailFoldersService) Delete(ctx context.Context, folderID string) error {
	return r.client.Me().MailFolders().ByMailFolderId(folderID).Delete(ctx, nil)
}

// GetMessages retrieves messages from a specific folder.
func (r *realMailFoldersService) GetMessages(ctx context.Context, folderID string, config *users.ItemMailFoldersItemMessagesRequestBuilderGetRequestConfiguration) (models.MessageCollectionResponseable, error) {
	return r.client.Me().MailFolders().ByMailFolderId(folderID).Messages().Get(ctx, config)
}
