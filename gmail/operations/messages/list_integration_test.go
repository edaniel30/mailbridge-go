package messages

import (
	"context"
	"errors"
	"testing"

	"github.com/danielrivera/mailbridge-go/core"
	"github.com/danielrivera/mailbridge-go/gmail/internal"
	gmailtest "github.com/danielrivera/mailbridge-go/gmail/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/gmail/v1"
)

func TestListMessages_Success(t *testing.T) {
	// Create mocks
	mockGmailService, mockMessagesService := setupMockMessagesService()
	mockMessagesListCall := &gmailtest.MockMessagesListCall{}
	mockMessagesGetCall := &gmailtest.MockMessagesGetCall{}

	// Setup expectations
	mockMessagesService.On("List", "me").Return(mockMessagesListCall)

	mockMessagesListCall.On("MaxResults", int64(10)).Return(mockMessagesListCall)
	mockMessagesListCall.On("Context", context.Background()).Return(mockMessagesListCall)
	mockMessagesListCall.On("Do").Return(&gmail.ListMessagesResponse{
		Messages: []*gmail.Message{
			{Id: "msg-1"},
		},
		NextPageToken:      "next-token",
		ResultSizeEstimate: 1,
	}, nil)

	// Setup GetMessage mock
	mockMessagesService.On("Get", "me", "msg-1").Return(mockMessagesGetCall)
	mockMessagesGetCall.On("Format", "full").Return(mockMessagesGetCall)
	mockMessagesGetCall.On("Context", context.Background()).Return(mockMessagesGetCall)
	mockMessagesGetCall.On("Do").Return(&gmail.Message{
		Id:       "msg-1",
		ThreadId: "thread-1",
		Snippet:  "Test snippet",
		LabelIds: []string{"INBOX"},
		Payload: &gmail.MessagePart{
			Headers: []*gmail.MessagePartHeader{
				{Name: "Subject", Value: "Test Subject"},
				{Name: "From", Value: "sender@example.com"},
			},
		},
	}, nil)

	// Execute
	opts := &core.ListOptions{
		MaxResults: 10,
	}
	resp, err := ListMessages(context.Background(), mockGmailService, opts)

	// Assert
	require.NoError(t, err)
	assert.Len(t, resp.Emails, 1)
	assert.Equal(t, "msg-1", resp.Emails[0].ID)
	assert.Equal(t, "next-token", resp.NextPageToken)
	assert.Equal(t, int64(1), resp.TotalCount)
}

func TestListMessages_APIError(t *testing.T) {
	mockGmailService, mockMessagesService := setupMockMessagesService()
	mockMessagesListCall := &gmailtest.MockMessagesListCall{}

	mockMessagesService.On("List", "me").Return(mockMessagesListCall)
	mockMessagesListCall.On("Context", context.Background()).Return(mockMessagesListCall)
	mockMessagesListCall.On("Do").Return(nil, errors.New("API error"))

	_, err := ListMessages(context.Background(), mockGmailService, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list messages")
}

func TestListMessages_WithAllOptions(t *testing.T) {
	mockGmailService, mockMessagesService := setupMockMessagesService()
	mockMessagesListCall := &gmailtest.MockMessagesListCall{}

	mockMessagesService.On("List", "me").Return(mockMessagesListCall)

	mockMessagesListCall.On("MaxResults", int64(50)).Return(mockMessagesListCall)
	mockMessagesListCall.On("PageToken", "token-123").Return(mockMessagesListCall)
	mockMessagesListCall.On("Q", "is:unread").Return(mockMessagesListCall)
	mockMessagesListCall.On("LabelIds", []string{"INBOX", "UNREAD"}).Return(mockMessagesListCall)
	mockMessagesListCall.On("Context", context.Background()).Return(mockMessagesListCall)
	mockMessagesListCall.On("Do").Return(&gmail.ListMessagesResponse{
		Messages:           []*gmail.Message{},
		ResultSizeEstimate: 0,
	}, nil)

	opts := &core.ListOptions{
		MaxResults: 50,
		PageToken:  "token-123",
		Query:      "is:unread",
		Labels:     []string{"INBOX", "UNREAD"},
	}
	resp, err := ListMessages(context.Background(), mockGmailService, opts)

	require.NoError(t, err)
	assert.Empty(t, resp.Emails)
}

// Helper function to setup mock messages service
func setupMockMessagesService() (internal.GmailService, *gmailtest.MockMessagesService) {
	mockGmailService := &gmailtest.MockGmailService{}
	mockUsersService := &gmailtest.MockUsersService{}
	mockMessagesService := &gmailtest.MockMessagesService{}

	mockGmailService.On("GetUsersService").Return(mockUsersService)
	mockUsersService.On("GetMessagesService").Return(mockMessagesService)

	return mockGmailService, mockMessagesService
}
