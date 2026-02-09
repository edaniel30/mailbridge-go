package mocks

import (
	"github.com/edaniel30/mailbridge-go/gmail/internal"
)

// SetupMockGmailService creates a mock GmailService with a mock UsersService
func SetupMockGmailService() (*MockGmailService, *MockUsersService) {
	mockGmailService := &MockGmailService{}
	mockUsersService := &MockUsersService{}

	mockGmailService.On("GetUsersService").Return(mockUsersService)

	return mockGmailService, mockUsersService
}

// SetupMockMessagesService creates a mock GmailService with MessagesService configured
func SetupMockMessagesService() (internal.GmailService, *MockMessagesService) {
	mockGmailService := &MockGmailService{}
	mockMessagesService := &MockMessagesService{}
	mockUsersService := &MockUsersService{}

	mockUsersService.On("GetMessagesService").Return(mockMessagesService)
	mockGmailService.On("GetUsersService").Return(mockUsersService)

	return mockGmailService, mockMessagesService
}

// SetupMockLabelsService creates a mock GmailService with LabelsService configured
func SetupMockLabelsService() (*MockGmailService, *MockLabelsService) {
	mockGmailService := &MockGmailService{}
	mockLabelsService := &MockLabelsService{}
	mockUsersService := &MockUsersService{}

	mockUsersService.On("GetLabelsService").Return(mockLabelsService)
	mockGmailService.On("GetUsersService").Return(mockUsersService)

	return mockGmailService, mockLabelsService
}
