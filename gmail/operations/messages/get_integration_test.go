package messages

import (
	"context"
	"errors"
	"testing"

	gmailtest "github.com/danielrivera/mailbridge-go/gmail/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/gmail/v1"
)

func TestGetMessage_Success(t *testing.T) {
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

	email, err := GetMessage(context.Background(), mockGmailService, "msg-123")

	require.NoError(t, err)
	assert.Equal(t, "msg-123", email.ID)
	assert.Equal(t, "thread-456", email.ThreadID)
	assert.Equal(t, "Test Subject", email.Subject)
	assert.Equal(t, "john@example.com", email.From.Email)
	assert.Equal(t, "John Doe", email.From.Name)
	assert.False(t, email.IsRead)
}

func TestGetMessage_APIError(t *testing.T) {
	mockGmailService, mockMessagesService := setupMockMessagesService()
	mockMessagesGetCall := &gmailtest.MockMessagesGetCall{}

	mockMessagesService.On("Get", "me", "msg-123").Return(mockMessagesGetCall)
	mockMessagesGetCall.On("Format", "full").Return(mockMessagesGetCall)
	mockMessagesGetCall.On("Context", context.Background()).Return(mockMessagesGetCall)
	mockMessagesGetCall.On("Do").Return(nil, errors.New("message not found"))

	_, err := GetMessage(context.Background(), mockGmailService, "msg-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get message")
}
