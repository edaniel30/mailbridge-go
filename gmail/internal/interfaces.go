package internal

import (
	"context"

	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
)

// GmailService is an interface for gmail.Service operations
type GmailService interface {
	GetUsersService() UsersService
}

// UsersService is an interface for gmail users operations
type UsersService interface {
	GetMessagesService() MessagesService
	GetLabelsService() LabelsService
}

// MessagesService is an interface for gmail messages operations
type MessagesService interface {
	List(userID string) MessagesListCall
	Get(userID, messageID string) MessagesGetCall
	Modify(userID, messageID string, req *gmail.ModifyMessageRequest) MessagesModifyCall
	GetAttachment(userID, messageID, attachmentID string) MessagesAttachmentGetCall
}

// LabelsService is an interface for gmail labels operations
type LabelsService interface {
	List(userID string) LabelsListCall
	Get(userID, labelID string) LabelsGetCall
	Create(userID string, label *gmail.Label) LabelsCreateCall
	Delete(userID, labelID string) LabelsDeleteCall
}

// MessagesListCall is an interface for messages list API calls
type MessagesListCall interface {
	MaxResults(maxResults int64) MessagesListCall
	PageToken(token string) MessagesListCall
	Q(query string) MessagesListCall
	LabelIds(labelIds ...string) MessagesListCall
	Context(ctx context.Context) MessagesListCall
	Do() (*gmail.ListMessagesResponse, error)
}

// MessagesGetCall is an interface for messages get API calls
type MessagesGetCall interface {
	Format(format string) MessagesGetCall
	Context(ctx context.Context) MessagesGetCall
	Do() (*gmail.Message, error)
}

// MessagesModifyCall is an interface for messages modify API calls
type MessagesModifyCall interface {
	Context(ctx context.Context) MessagesModifyCall
	Do() (*gmail.Message, error)
}

// MessagesAttachmentGetCall is an interface for attachment get API calls
type MessagesAttachmentGetCall interface {
	Context(ctx context.Context) MessagesAttachmentGetCall
	Do() (*gmail.MessagePartBody, error)
}

// LabelsListCall is an interface for labels list API calls
type LabelsListCall interface {
	Context(ctx context.Context) LabelsListCall
	Do() (*gmail.ListLabelsResponse, error)
}

// LabelsGetCall is an interface for labels get API calls
type LabelsGetCall interface {
	Context(ctx context.Context) LabelsGetCall
	Do() (*gmail.Label, error)
}

// LabelsCreateCall is an interface for labels create API calls
type LabelsCreateCall interface {
	Context(ctx context.Context) LabelsCreateCall
	Do() (*gmail.Label, error)
}

// LabelsDeleteCall is an interface for labels delete API calls
type LabelsDeleteCall interface {
	Context(ctx context.Context) LabelsDeleteCall
	Do() error
}

// OAuth2Config is an interface for oauth2.Config operations
type OAuth2Config interface {
	AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string
	Exchange(ctx context.Context, code string) (*oauth2.Token, error)
	Client(ctx context.Context, token *oauth2.Token) oauth2.TokenSource
	TokenSource(ctx context.Context, token *oauth2.Token) oauth2.TokenSource
}
