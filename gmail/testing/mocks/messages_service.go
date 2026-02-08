package mocks

import (
	"context"

	"github.com/edaniel30/mailbridge-go/gmail/internal"
	"github.com/stretchr/testify/mock"
	gmailapi "google.golang.org/api/gmail/v1"
)

// MockMessagesService is a mock for MessagesService
type MockMessagesService struct {
	mock.Mock
}

func (m *MockMessagesService) List(userID string) internal.MessagesListCall {
	args := m.Called(userID)
	return args.Get(0).(internal.MessagesListCall)
}

func (m *MockMessagesService) Get(userID, messageID string) internal.MessagesGetCall {
	args := m.Called(userID, messageID)
	return args.Get(0).(internal.MessagesGetCall)
}

func (m *MockMessagesService) Modify(userID, messageID string, req *gmailapi.ModifyMessageRequest) internal.MessagesModifyCall {
	args := m.Called(userID, messageID, req)
	return args.Get(0).(internal.MessagesModifyCall)
}

func (m *MockMessagesService) GetAttachment(userID, messageID, attachmentID string) internal.MessagesAttachmentGetCall {
	args := m.Called(userID, messageID, attachmentID)
	return args.Get(0).(internal.MessagesAttachmentGetCall)
}

func (m *MockMessagesService) Send(userID string, message *gmailapi.Message) internal.MessagesSendCall {
	args := m.Called(userID, message)
	return args.Get(0).(internal.MessagesSendCall)
}

func (m *MockMessagesService) Trash(userID, messageID string) internal.MessagesTrashCall {
	args := m.Called(userID, messageID)
	return args.Get(0).(internal.MessagesTrashCall)
}

func (m *MockMessagesService) Untrash(userID, messageID string) internal.MessagesUntrashCall {
	args := m.Called(userID, messageID)
	return args.Get(0).(internal.MessagesUntrashCall)
}

func (m *MockMessagesService) Delete(userID, messageID string) internal.MessagesDeleteCall {
	args := m.Called(userID, messageID)
	return args.Get(0).(internal.MessagesDeleteCall)
}

// MockMessagesListCall is a mock for MessagesListCall
type MockMessagesListCall struct {
	mock.Mock
}

func (m *MockMessagesListCall) MaxResults(maxResults int64) internal.MessagesListCall {
	m.Called(maxResults)
	return m
}

func (m *MockMessagesListCall) PageToken(token string) internal.MessagesListCall {
	m.Called(token)
	return m
}

func (m *MockMessagesListCall) Q(query string) internal.MessagesListCall {
	m.Called(query)
	return m
}

func (m *MockMessagesListCall) LabelIds(labelIds ...string) internal.MessagesListCall {
	m.Called(labelIds)
	return m
}

func (m *MockMessagesListCall) Context(ctx context.Context) internal.MessagesListCall {
	m.Called(ctx)
	return m
}

func (m *MockMessagesListCall) Do() (*gmailapi.ListMessagesResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gmailapi.ListMessagesResponse), args.Error(1)
}

// MockMessagesGetCall is a mock for MessagesGetCall
type MockMessagesGetCall struct {
	mock.Mock
}

func (m *MockMessagesGetCall) Format(format string) internal.MessagesGetCall {
	m.Called(format)
	return m
}

func (m *MockMessagesGetCall) Context(ctx context.Context) internal.MessagesGetCall {
	m.Called(ctx)
	return m
}

func (m *MockMessagesGetCall) Do() (*gmailapi.Message, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gmailapi.Message), args.Error(1)
}

// MockMessagesModifyCall is a mock for MessagesModifyCall
type MockMessagesModifyCall struct {
	mock.Mock
}

func (m *MockMessagesModifyCall) Context(ctx context.Context) internal.MessagesModifyCall {
	m.Called(ctx)
	return m
}

func (m *MockMessagesModifyCall) Do() (*gmailapi.Message, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gmailapi.Message), args.Error(1)
}

// MockMessagesAttachmentGetCall is a mock for MessagesAttachmentGetCall
type MockMessagesAttachmentGetCall struct {
	mock.Mock
}

func (m *MockMessagesAttachmentGetCall) Context(ctx context.Context) internal.MessagesAttachmentGetCall {
	m.Called(ctx)
	return m
}

func (m *MockMessagesAttachmentGetCall) Do() (*gmailapi.MessagePartBody, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gmailapi.MessagePartBody), args.Error(1)
}

// MockMessagesSendCall is a mock for MessagesSendCall
type MockMessagesSendCall struct {
	mock.Mock
}

func (m *MockMessagesSendCall) Context(ctx context.Context) internal.MessagesSendCall {
	m.Called(ctx)
	return m
}

func (m *MockMessagesSendCall) Do() (*gmailapi.Message, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gmailapi.Message), args.Error(1)
}

// MockMessagesTrashCall is a mock for MessagesTrashCall
type MockMessagesTrashCall struct {
	mock.Mock
}

func (m *MockMessagesTrashCall) Context(ctx context.Context) internal.MessagesTrashCall {
	m.Called(ctx)
	return m
}

func (m *MockMessagesTrashCall) Do() (*gmailapi.Message, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gmailapi.Message), args.Error(1)
}

// MockMessagesUntrashCall is a mock for MessagesUntrashCall
type MockMessagesUntrashCall struct {
	mock.Mock
}

func (m *MockMessagesUntrashCall) Context(ctx context.Context) internal.MessagesUntrashCall {
	m.Called(ctx)
	return m
}

func (m *MockMessagesUntrashCall) Do() (*gmailapi.Message, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gmailapi.Message), args.Error(1)
}

// MockMessagesDeleteCall is a mock for MessagesDeleteCall
type MockMessagesDeleteCall struct {
	mock.Mock
}

func (m *MockMessagesDeleteCall) Context(ctx context.Context) internal.MessagesDeleteCall {
	m.Called(ctx)
	return m
}

func (m *MockMessagesDeleteCall) Do() error {
	args := m.Called()
	return args.Error(0)
}
