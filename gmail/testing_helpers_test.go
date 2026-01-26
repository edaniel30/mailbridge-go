package gmail

import (
	"testing"

	gmailtest "github.com/danielrivera/mailbridge-go/gmail/testing"
	"github.com/stretchr/testify/require"
)

// newTestConfig creates a test configuration with standard test values.
// Use this instead of manually creating Config in every test.
func newTestConfig() *Config {
	return &Config{
		ClientID:     "test-id",
		ClientSecret: "test-secret",
		RedirectURL:  "http://localhost",
		Scopes:       []string{"test-scope"},
	}
}

// newTestClient creates a test client with standard configuration.
// Returns an initialized but not connected client.
func newTestClient(t *testing.T) *Client {
	t.Helper()
	client, err := New(newTestConfig())
	require.NoError(t, err)
	return client
}

// setupMockGmailService creates and wires up a basic mock Gmail service hierarchy.
// Returns mockGmailService and mockUsersService with GetUsersService already configured.
func setupMockGmailService() (*gmailtest.MockGmailService, *gmailtest.MockUsersService) {
	mockGmailService := &gmailtest.MockGmailService{}
	mockUsersService := &gmailtest.MockUsersService{}
	mockGmailService.On("GetUsersService").Return(mockUsersService)
	return mockGmailService, mockUsersService
}

// setupMockMessagesService creates a full mock chain for messages operations.
// Returns mockGmailService and mockMessagesService (mockUsersService is wired internally).
func setupMockMessagesService() (*gmailtest.MockGmailService, *gmailtest.MockMessagesService) {
	mockGmailService, mockUsersService := setupMockGmailService()
	mockMessagesService := &gmailtest.MockMessagesService{}
	mockUsersService.On("GetMessagesService").Return(mockMessagesService)
	return mockGmailService, mockMessagesService
}

// setupMockLabelsService creates a full mock chain for labels operations.
// Returns mockGmailService and mockLabelsService (mockUsersService is wired internally).
func setupMockLabelsService() (*gmailtest.MockGmailService, *gmailtest.MockLabelsService) {
	mockGmailService, mockUsersService := setupMockGmailService()
	mockLabelsService := &gmailtest.MockLabelsService{}
	mockUsersService.On("GetLabelsService").Return(mockLabelsService)
	return mockGmailService, mockLabelsService
}

// setupMockLabelsAndMessagesService creates a full mock chain for operations that need both labels and messages.
// Returns mockGmailService, mockLabelsService, and mockMessagesService (mockUsersService is wired internally).
func setupMockLabelsAndMessagesService() (*gmailtest.MockGmailService, *gmailtest.MockLabelsService, *gmailtest.MockMessagesService) {
	mockGmailService, mockUsersService := setupMockGmailService()
	mockLabelsService := &gmailtest.MockLabelsService{}
	mockMessagesService := &gmailtest.MockMessagesService{}
	mockUsersService.On("GetLabelsService").Return(mockLabelsService)
	mockUsersService.On("GetMessagesService").Return(mockMessagesService)
	return mockGmailService, mockLabelsService, mockMessagesService
}
