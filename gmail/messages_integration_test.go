package gmail

import (
	"context"
	"errors"
	"testing"

	"github.com/danielrivera/mailbridge-go/core"
	gmailtest "github.com/danielrivera/mailbridge-go/gmail/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/gmail/v1"
)

func TestClient_ListMessages_Success(t *testing.T) {
	client := newTestClient(t)

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

	client.SetService(mockGmailService)

	// Execute
	opts := &core.ListOptions{
		MaxResults: 10,
	}
	resp, err := client.ListMessages(context.Background(), opts)

	// Assert
	require.NoError(t, err)
	assert.Len(t, resp.Emails, 1)
	assert.Equal(t, "msg-1", resp.Emails[0].ID)
	assert.Equal(t, "next-token", resp.NextPageToken)
	assert.Equal(t, int64(1), resp.TotalCount)
}

func TestClient_ListMessages_NotConnected(t *testing.T) {
	client := newTestClient(t)

	_, err := client.ListMessages(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestClient_ListMessages_APIError(t *testing.T) {
	client := newTestClient(t)

	mockGmailService, mockMessagesService := setupMockMessagesService()
	mockMessagesListCall := &gmailtest.MockMessagesListCall{}

	mockMessagesService.On("List", "me").Return(mockMessagesListCall)
	mockMessagesListCall.On("Context", context.Background()).Return(mockMessagesListCall)
	mockMessagesListCall.On("Do").Return(nil, errors.New("API error"))

	client.SetService(mockGmailService)

	_, err := client.ListMessages(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list messages")
}

func TestClient_ListMessages_WithAllOptions(t *testing.T) {
	client := newTestClient(t)

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

	client.SetService(mockGmailService)

	opts := &core.ListOptions{
		MaxResults: 50,
		PageToken:  "token-123",
		Query:      "is:unread",
		Labels:     []string{"INBOX", "UNREAD"},
	}
	resp, err := client.ListMessages(context.Background(), opts)

	require.NoError(t, err)
	assert.Empty(t, resp.Emails)
}

func TestClient_GetMessage_Success(t *testing.T) {
	client := newTestClient(t)

	mockGmailService, mockMessagesService := setupMockMessagesService()
	mockMessagesGetCall := &gmailtest.MockMessagesGetCall{}

	mockMessagesService.On("Get", "me", "msg-123").Return(mockMessagesGetCall)
	mockMessagesGetCall.On("Format", "full").Return(mockMessagesGetCall)
	mockMessagesGetCall.On("Context", context.Background()).Return(mockMessagesGetCall)
	mockMessagesGetCall.On("Do").Return(&gmail.Message{
		Id:       "msg-123",
		ThreadId: "thread-456",
		Snippet:  "Test message",
		LabelIds: []string{"INBOX", "UNREAD"},
		Payload: &gmail.MessagePart{
			Headers: []*gmail.MessagePartHeader{
				{Name: "Subject", Value: "Test Subject"},
				{Name: "From", Value: "John Doe <john@example.com>"},
				{Name: "To", Value: "jane@example.com"},
			},
			MimeType: "text/plain",
			Body: &gmail.MessagePartBody{
				Data: "SGVsbG8gV29ybGQ=", // "Hello World" in base64
			},
		},
	}, nil)

	client.SetService(mockGmailService)

	email, err := client.GetMessage(context.Background(), "msg-123")

	require.NoError(t, err)
	assert.Equal(t, "msg-123", email.ID)
	assert.Equal(t, "thread-456", email.ThreadID)
	assert.Equal(t, "Test Subject", email.Subject)
	assert.Equal(t, "john@example.com", email.From.Email)
	assert.Equal(t, "John Doe", email.From.Name)
	assert.False(t, email.IsRead)
}

func TestClient_GetMessage_NotConnected(t *testing.T) {
	client := newTestClient(t)

	_, err := client.GetMessage(context.Background(), "msg-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestClient_GetMessage_APIError(t *testing.T) {
	client := newTestClient(t)

	mockGmailService, mockMessagesService := setupMockMessagesService()
	mockMessagesGetCall := &gmailtest.MockMessagesGetCall{}

	mockMessagesService.On("Get", "me", "msg-123").Return(mockMessagesGetCall)
	mockMessagesGetCall.On("Format", "full").Return(mockMessagesGetCall)
	mockMessagesGetCall.On("Context", context.Background()).Return(mockMessagesGetCall)
	mockMessagesGetCall.On("Do").Return(nil, errors.New("message not found"))

	client.SetService(mockGmailService)

	_, err := client.GetMessage(context.Background(), "msg-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get message")
}
