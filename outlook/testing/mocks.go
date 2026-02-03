package testing

import (
	"context"

	"github.com/danielrivera/mailbridge-go/outlook/internal"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
	"github.com/stretchr/testify/mock"
)

// MockGraphService is a mock for GraphService
type MockGraphService struct {
	mock.Mock
}

func (m *MockGraphService) GetMeService() internal.MeService {
	args := m.Called()
	return args.Get(0).(internal.MeService)
}

// MockMeService is a mock for MeService
type MockMeService struct {
	mock.Mock
}

func (m *MockMeService) GetMessagesService() internal.MessagesService {
	args := m.Called()
	return args.Get(0).(internal.MessagesService)
}

func (m *MockMeService) GetMailFoldersService() internal.MailFoldersService {
	args := m.Called()
	return args.Get(0).(internal.MailFoldersService)
}

// MockMessagesService is a mock for MessagesService
type MockMessagesService struct {
	mock.Mock
}

func (m *MockMessagesService) List(ctx context.Context, config *users.ItemMessagesRequestBuilderGetRequestConfiguration) (models.MessageCollectionResponseable, error) {
	args := m.Called(ctx, config)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(models.MessageCollectionResponseable), args.Error(1)
}

func (m *MockMessagesService) Get(ctx context.Context, messageID string) (models.Messageable, error) {
	args := m.Called(ctx, messageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(models.Messageable), args.Error(1)
}

func (m *MockMessagesService) GetAttachments(ctx context.Context, messageID string) ([]models.Attachmentable, error) {
	args := m.Called(ctx, messageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Attachmentable), args.Error(1)
}

func (m *MockMessagesService) GetAttachment(ctx context.Context, messageID, attachmentID string) (models.Attachmentable, error) {
	args := m.Called(ctx, messageID, attachmentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(models.Attachmentable), args.Error(1)
}

func (m *MockMessagesService) MarkAsRead(ctx context.Context, messageID string) error {
	args := m.Called(ctx, messageID)
	return args.Error(0)
}

func (m *MockMessagesService) MarkAsUnread(ctx context.Context, messageID string) error {
	args := m.Called(ctx, messageID)
	return args.Error(0)
}

func (m *MockMessagesService) Move(ctx context.Context, messageID, destinationFolderID string) error {
	args := m.Called(ctx, messageID, destinationFolderID)
	return args.Error(0)
}

func (m *MockMessagesService) Delete(ctx context.Context, messageID string) error {
	args := m.Called(ctx, messageID)
	return args.Error(0)
}

// MockMailFoldersService is a mock for MailFoldersService
type MockMailFoldersService struct {
	mock.Mock
}

func (m *MockMailFoldersService) List(ctx context.Context) (models.MailFolderCollectionResponseable, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(models.MailFolderCollectionResponseable), args.Error(1)
}

func (m *MockMailFoldersService) Get(ctx context.Context, folderID string) (models.MailFolderable, error) {
	args := m.Called(ctx, folderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(models.MailFolderable), args.Error(1)
}

func (m *MockMailFoldersService) Create(ctx context.Context, name string) (models.MailFolderable, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(models.MailFolderable), args.Error(1)
}

func (m *MockMailFoldersService) Update(ctx context.Context, folderID, newName string) (models.MailFolderable, error) {
	args := m.Called(ctx, folderID, newName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(models.MailFolderable), args.Error(1)
}

func (m *MockMailFoldersService) Delete(ctx context.Context, folderID string) error {
	args := m.Called(ctx, folderID)
	return args.Error(0)
}

func (m *MockMailFoldersService) GetMessages(ctx context.Context, folderID string, config *users.ItemMailFoldersItemMessagesRequestBuilderGetRequestConfiguration) (models.MessageCollectionResponseable, error) {
	args := m.Called(ctx, folderID, config)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(models.MessageCollectionResponseable), args.Error(1)
}
