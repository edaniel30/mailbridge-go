package internal

import (
	"context"

	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
)

// GraphService is the main interface for Microsoft Graph API operations.
// It provides access to user-specific services.
type GraphService interface {
	GetMeService() MeService
}

// MeService represents operations for the authenticated user.
type MeService interface {
	GetMessagesService() MessagesService
	GetMailFoldersService() MailFoldersService
}

// MessagesService represents operations on email messages.
type MessagesService interface {
	List(ctx context.Context, config *users.ItemMessagesRequestBuilderGetRequestConfiguration) (models.MessageCollectionResponseable, error)
	Get(ctx context.Context, messageID string) (models.Messageable, error)
	GetAttachments(ctx context.Context, messageID string) ([]models.Attachmentable, error)
	GetAttachment(ctx context.Context, messageID, attachmentID string) (models.Attachmentable, error)
	MarkAsRead(ctx context.Context, messageID string) error
	MarkAsUnread(ctx context.Context, messageID string) error
	Move(ctx context.Context, messageID, destinationFolderID string) error
	Delete(ctx context.Context, messageID string) error
}

// MailFoldersService represents operations on mail folders.
type MailFoldersService interface {
	List(ctx context.Context) (models.MailFolderCollectionResponseable, error)
	Get(ctx context.Context, folderID string) (models.MailFolderable, error)
	Create(ctx context.Context, name string) (models.MailFolderable, error)
	Update(ctx context.Context, folderID, newName string) (models.MailFolderable, error)
	Delete(ctx context.Context, folderID string) error
	GetMessages(ctx context.Context, folderID string, config *users.ItemMailFoldersItemMessagesRequestBuilderGetRequestConfiguration) (models.MessageCollectionResponseable, error)
}
