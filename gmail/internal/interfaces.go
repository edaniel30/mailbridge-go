package internal

import (
	"context"

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
	Watch(userID string, req *gmail.WatchRequest) UsersWatchCall
	Stop(userID string) UsersStopCall
	GetHistory(userID string) UsersHistoryListCall
}

// MessagesService is an interface for gmail messages operations
type MessagesService interface {
	List(userID string) MessagesListCall
	Get(userID, messageID string) MessagesGetCall
	Modify(userID, messageID string, req *gmail.ModifyMessageRequest) MessagesModifyCall
	GetAttachment(userID, messageID, attachmentID string) MessagesAttachmentGetCall
	Send(userID string, message *gmail.Message) MessagesSendCall
	Trash(userID, messageID string) MessagesTrashCall
	Untrash(userID, messageID string) MessagesUntrashCall
	Delete(userID, messageID string) MessagesDeleteCall
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

// MessagesSendCall is an interface for messages send API calls
type MessagesSendCall interface {
	Context(ctx context.Context) MessagesSendCall
	Do() (*gmail.Message, error)
}

// MessagesTrashCall is an interface for messages trash API calls
type MessagesTrashCall interface {
	Context(ctx context.Context) MessagesTrashCall
	Do() (*gmail.Message, error)
}

// MessagesUntrashCall is an interface for messages untrash API calls
type MessagesUntrashCall interface {
	Context(ctx context.Context) MessagesUntrashCall
	Do() (*gmail.Message, error)
}

// MessagesDeleteCall is an interface for messages delete API calls
type MessagesDeleteCall interface {
	Context(ctx context.Context) MessagesDeleteCall
	Do() error
}

// UsersWatchCall is an interface for users watch API calls
type UsersWatchCall interface {
	Context(ctx context.Context) UsersWatchCall
	Do() (*gmail.WatchResponse, error)
}

// UsersStopCall is an interface for users stop API calls
type UsersStopCall interface {
	Context(ctx context.Context) UsersStopCall
	Do() error
}

// UsersHistoryListCall is an interface for users history list API calls
type UsersHistoryListCall interface {
	MaxResults(maxResults int64) UsersHistoryListCall
	PageToken(token string) UsersHistoryListCall
	LabelId(labelId string) UsersHistoryListCall
	StartHistoryId(historyId uint64) UsersHistoryListCall
	HistoryTypes(types ...string) UsersHistoryListCall
	Context(ctx context.Context) UsersHistoryListCall
	Do() (*gmail.ListHistoryResponse, error)
}
