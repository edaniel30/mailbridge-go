package messages

import (
	"context"
	"errors"
	"testing"

	"github.com/danielrivera/mailbridge-go/core"
	gmailtest "github.com/danielrivera/mailbridge-go/gmail/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	gmailapi "google.golang.org/api/gmail/v1"
)

func TestSendMessage_Success(t *testing.T) {
	ctx := context.Background()

	// Create mocks
	mockService := &gmailtest.MockGmailService{}
	mockUsersService := &gmailtest.MockUsersService{}
	mockMessagesService := &gmailtest.MockMessagesService{}
	mockSendCall := &gmailtest.MockMessagesSendCall{}

	// Setup expectations
	mockService.On("GetUsersService").Return(mockUsersService)
	mockUsersService.On("GetMessagesService").Return(mockMessagesService)
	mockMessagesService.On("Send", "me", mock.AnythingOfType("*gmail.Message")).Return(mockSendCall)
	mockSendCall.On("Context", ctx).Return(mockSendCall)
	mockSendCall.On("Do").Return(&gmailapi.Message{
		Id:       "sent-msg-123",
		ThreadId: "thread-456",
	}, nil)

	// Prepare draft
	draft := &core.Draft{
		To:      []core.EmailAddress{{Email: "recipient@example.com"}},
		Subject: "Test Email",
		Body:    core.EmailBody{Text: "Hello World"},
	}

	// Send message
	response, err := SendMessage(ctx, mockService, draft, nil)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "sent-msg-123", response.ID)
	assert.Equal(t, "thread-456", response.ThreadID)

	mockService.AssertExpectations(t)
	mockUsersService.AssertExpectations(t)
	mockMessagesService.AssertExpectations(t)
	mockSendCall.AssertExpectations(t)
}

func TestSendMessage_APIError(t *testing.T) {
	ctx := context.Background()

	// Create mocks
	mockService := &gmailtest.MockGmailService{}
	mockUsersService := &gmailtest.MockUsersService{}
	mockMessagesService := &gmailtest.MockMessagesService{}
	mockSendCall := &gmailtest.MockMessagesSendCall{}

	// Setup expectations with error
	mockService.On("GetUsersService").Return(mockUsersService)
	mockUsersService.On("GetMessagesService").Return(mockMessagesService)
	mockMessagesService.On("Send", "me", mock.AnythingOfType("*gmail.Message")).Return(mockSendCall)
	mockSendCall.On("Context", ctx).Return(mockSendCall)
	mockSendCall.On("Do").Return(nil, errors.New("API error"))

	draft := &core.Draft{
		To:      []core.EmailAddress{{Email: "test@example.com"}},
		Subject: "Test",
		Body:    core.EmailBody{Text: "Hello"},
	}

	response, err := SendMessage(ctx, mockService, draft, nil)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to send message")

	mockService.AssertExpectations(t)
}

func TestSendMessage_WithAttachments(t *testing.T) {
	ctx := context.Background()

	// Create mocks
	mockService := &gmailtest.MockGmailService{}
	mockUsersService := &gmailtest.MockUsersService{}
	mockMessagesService := &gmailtest.MockMessagesService{}
	mockSendCall := &gmailtest.MockMessagesSendCall{}

	// Setup expectations
	mockService.On("GetUsersService").Return(mockUsersService)
	mockUsersService.On("GetMessagesService").Return(mockMessagesService)
	mockMessagesService.On("Send", "me", mock.AnythingOfType("*gmail.Message")).Return(mockSendCall)
	mockSendCall.On("Context", ctx).Return(mockSendCall)
	mockSendCall.On("Do").Return(&gmailapi.Message{
		Id:       "sent-msg-with-att",
		ThreadId: "thread-789",
	}, nil)

	// Draft with attachment
	draft := &core.Draft{
		To:      []core.EmailAddress{{Email: "test@example.com"}},
		Subject: "Email with Attachment",
		Body:    core.EmailBody{Text: "See attached file"},
		Attachments: []core.Attachment{
			{
				Filename: "document.pdf",
				MimeType: "application/pdf",
				Data:     []byte("fake PDF content"),
			},
		},
	}

	response, err := SendMessage(ctx, mockService, draft, nil)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "sent-msg-with-att", response.ID)

	mockService.AssertExpectations(t)
}

func TestSendMessage_HTMLEmail(t *testing.T) {
	ctx := context.Background()

	// Create mocks
	mockService := &gmailtest.MockGmailService{}
	mockUsersService := &gmailtest.MockUsersService{}
	mockMessagesService := &gmailtest.MockMessagesService{}
	mockSendCall := &gmailtest.MockMessagesSendCall{}

	// Setup expectations
	mockService.On("GetUsersService").Return(mockUsersService)
	mockUsersService.On("GetMessagesService").Return(mockMessagesService)
	mockMessagesService.On("Send", "me", mock.AnythingOfType("*gmail.Message")).Return(mockSendCall)
	mockSendCall.On("Context", ctx).Return(mockSendCall)
	mockSendCall.On("Do").Return(&gmailapi.Message{
		Id:       "html-msg-123",
		ThreadId: "thread-html",
	}, nil)

	draft := &core.Draft{
		To:      []core.EmailAddress{{Email: "test@example.com"}},
		Subject: "HTML Email",
		Body:    core.EmailBody{HTML: "<p>Hello <b>World</b></p>"},
	}

	response, err := SendMessage(ctx, mockService, draft, nil)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "html-msg-123", response.ID)

	mockService.AssertExpectations(t)
}

func TestSendMessage_WithCCBCC(t *testing.T) {
	ctx := context.Background()

	// Create mocks
	mockService := &gmailtest.MockGmailService{}
	mockUsersService := &gmailtest.MockUsersService{}
	mockMessagesService := &gmailtest.MockMessagesService{}
	mockSendCall := &gmailtest.MockMessagesSendCall{}

	// Setup expectations
	mockService.On("GetUsersService").Return(mockUsersService)
	mockUsersService.On("GetMessagesService").Return(mockMessagesService)
	mockMessagesService.On("Send", "me", mock.AnythingOfType("*gmail.Message")).Return(mockSendCall)
	mockSendCall.On("Context", ctx).Return(mockSendCall)
	mockSendCall.On("Do").Return(&gmailapi.Message{
		Id:       "ccbcc-msg-123",
		ThreadId: "thread-ccbcc",
	}, nil)

	draft := &core.Draft{
		To:      []core.EmailAddress{{Email: "primary@example.com"}},
		Cc:      []core.EmailAddress{{Email: "cc@example.com"}},
		Bcc:     []core.EmailAddress{{Email: "bcc@example.com"}},
		Subject: "Email with CC and BCC",
		Body:    core.EmailBody{Text: "Hello"},
	}

	response, err := SendMessage(ctx, mockService, draft, nil)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "ccbcc-msg-123", response.ID)

	mockService.AssertExpectations(t)
}

func TestSendMessage_ValidationErrors(t *testing.T) {
	ctx := context.Background()
	mockService := &gmailtest.MockGmailService{}

	tests := []struct {
		name   string
		draft  *core.Draft
		errMsg string
	}{
		{
			name:   "no recipients",
			draft:  &core.Draft{Subject: "Test", Body: core.EmailBody{Text: "Hello"}},
			errMsg: "at least one recipient required",
		},
		{
			name: "invalid email",
			draft: &core.Draft{
				To:      []core.EmailAddress{{Email: "invalid"}},
				Subject: "Test",
				Body:    core.EmailBody{Text: "Hello"},
			},
			errMsg: "invalid email address",
		},
		{
			name: "no subject",
			draft: &core.Draft{
				To:   []core.EmailAddress{{Email: "test@example.com"}},
				Body: core.EmailBody{Text: "Hello"},
			},
			errMsg: "subject is required",
		},
		{
			name: "no body",
			draft: &core.Draft{
				To:      []core.EmailAddress{{Email: "test@example.com"}},
				Subject: "Test",
			},
			errMsg: "email body required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := SendMessage(ctx, mockService, tt.draft, nil)

			assert.Error(t, err)
			assert.Nil(t, response)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestSendMessage_WithCustomHeaders(t *testing.T) {
	ctx := context.Background()

	// Create mocks
	mockService := &gmailtest.MockGmailService{}
	mockUsersService := &gmailtest.MockUsersService{}
	mockMessagesService := &gmailtest.MockMessagesService{}
	mockSendCall := &gmailtest.MockMessagesSendCall{}

	// Setup expectations
	mockService.On("GetUsersService").Return(mockUsersService)
	mockUsersService.On("GetMessagesService").Return(mockMessagesService)
	mockMessagesService.On("Send", "me", mock.AnythingOfType("*gmail.Message")).Return(mockSendCall)
	mockSendCall.On("Context", ctx).Return(mockSendCall)
	mockSendCall.On("Do").Return(&gmailapi.Message{
		Id:       "custom-headers-msg",
		ThreadId: "thread-custom",
	}, nil)

	draft := &core.Draft{
		To:      []core.EmailAddress{{Email: "test@example.com"}},
		Subject: "Test",
		Body:    core.EmailBody{Text: "Hello"},
		Headers: map[string]string{
			"X-Priority": "1",
		},
	}

	opts := &core.SendOptions{
		CustomHeaders: map[string]string{
			"X-Mailer": "MailBridge",
		},
	}

	response, err := SendMessage(ctx, mockService, draft, opts)

	require.NoError(t, err)
	assert.NotNil(t, response)

	mockService.AssertExpectations(t)
}

func TestSendMessage_TextAndHTML(t *testing.T) {
	ctx := context.Background()

	// Create mocks
	mockService := &gmailtest.MockGmailService{}
	mockUsersService := &gmailtest.MockUsersService{}
	mockMessagesService := &gmailtest.MockMessagesService{}
	mockSendCall := &gmailtest.MockMessagesSendCall{}

	// Setup expectations
	mockService.On("GetUsersService").Return(mockUsersService)
	mockUsersService.On("GetMessagesService").Return(mockMessagesService)
	mockMessagesService.On("Send", "me", mock.AnythingOfType("*gmail.Message")).Return(mockSendCall)
	mockSendCall.On("Context", ctx).Return(mockSendCall)
	mockSendCall.On("Do").Return(&gmailapi.Message{
		Id:       "multipart-alt-msg",
		ThreadId: "thread-alt",
	}, nil)

	// Draft with both text and HTML (creates multipart/alternative)
	draft := &core.Draft{
		To:      []core.EmailAddress{{Email: "test@example.com"}},
		Subject: "Multipart Email",
		Body: core.EmailBody{
			Text: "Plain text version",
			HTML: "<p>HTML version</p>",
		},
	}

	response, err := SendMessage(ctx, mockService, draft, nil)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "multipart-alt-msg", response.ID)

	mockService.AssertExpectations(t)
}
