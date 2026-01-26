package testing

import (
	"context"

	"github.com/danielrivera/mailbridge-go/gmail/internal"
	"github.com/stretchr/testify/mock"
	gmailapi "google.golang.org/api/gmail/v1"
)

// MockGmailService is a mock for GmailService
type MockGmailService struct {
	mock.Mock
}

func (m *MockGmailService) GetUsersService() internal.UsersService {
	args := m.Called()
	return args.Get(0).(internal.UsersService)
}

// MockUsersService is a mock for UsersService
type MockUsersService struct {
	mock.Mock
}

func (m *MockUsersService) GetMessagesService() internal.MessagesService {
	args := m.Called()
	return args.Get(0).(internal.MessagesService)
}

func (m *MockUsersService) GetLabelsService() internal.LabelsService {
	args := m.Called()
	return args.Get(0).(internal.LabelsService)
}

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

// MockLabelsService is a mock for LabelsService
type MockLabelsService struct {
	mock.Mock
}

func (m *MockLabelsService) List(userID string) internal.LabelsListCall {
	args := m.Called(userID)
	return args.Get(0).(internal.LabelsListCall)
}

func (m *MockLabelsService) Get(userID, labelID string) internal.LabelsGetCall {
	args := m.Called(userID, labelID)
	return args.Get(0).(internal.LabelsGetCall)
}

func (m *MockLabelsService) Create(userID string, label *gmailapi.Label) internal.LabelsCreateCall {
	args := m.Called(userID, label)
	return args.Get(0).(internal.LabelsCreateCall)
}

func (m *MockLabelsService) Delete(userID, labelID string) internal.LabelsDeleteCall {
	args := m.Called(userID, labelID)
	return args.Get(0).(internal.LabelsDeleteCall)
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

// MockLabelsListCall is a mock for LabelsListCall
type MockLabelsListCall struct {
	mock.Mock
}

func (m *MockLabelsListCall) Context(ctx context.Context) internal.LabelsListCall {
	m.Called(ctx)
	return m
}

func (m *MockLabelsListCall) Do() (*gmailapi.ListLabelsResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gmailapi.ListLabelsResponse), args.Error(1)
}

// MockLabelsGetCall is a mock for LabelsGetCall
type MockLabelsGetCall struct {
	mock.Mock
}

func (m *MockLabelsGetCall) Context(ctx context.Context) internal.LabelsGetCall {
	m.Called(ctx)
	return m
}

func (m *MockLabelsGetCall) Do() (*gmailapi.Label, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gmailapi.Label), args.Error(1)
}

// MockLabelsCreateCall is a mock for LabelsCreateCall
type MockLabelsCreateCall struct {
	mock.Mock
}

func (m *MockLabelsCreateCall) Context(ctx context.Context) internal.LabelsCreateCall {
	m.Called(ctx)
	return m
}

func (m *MockLabelsCreateCall) Do() (*gmailapi.Label, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gmailapi.Label), args.Error(1)
}

// MockLabelsDeleteCall is a mock for LabelsDeleteCall
type MockLabelsDeleteCall struct {
	mock.Mock
}

func (m *MockLabelsDeleteCall) Context(ctx context.Context) internal.LabelsDeleteCall {
	m.Called(ctx)
	return m
}

func (m *MockLabelsDeleteCall) Do() error {
	args := m.Called()
	return args.Error(0)
}
