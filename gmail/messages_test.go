package gmail

import (
	"testing"
	"time"

	"github.com/danielrivera/mailbridge-go/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/gmail/v1"
)

func TestParseEmailAddress(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.EmailAddress
	}{
		{
			name:  "email with name",
			input: "John Doe <john@example.com>",
			expected: core.EmailAddress{
				Name:  "John Doe",
				Email: "john@example.com",
			},
		},
		{
			name:  "email with quoted name",
			input: "\"Jane Smith\" <jane@example.com>",
			expected: core.EmailAddress{
				Name:  "Jane Smith",
				Email: "jane@example.com",
			},
		},
		{
			name:  "email only",
			input: "test@example.com",
			expected: core.EmailAddress{
				Email: "test@example.com",
			},
		},
		{
			name:     "empty string",
			input:    "",
			expected: core.EmailAddress{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseEmailAddress(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseEmailAddresses(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []core.EmailAddress
	}{
		{
			name:  "single address",
			input: "john@example.com",
			expected: []core.EmailAddress{
				{Email: "john@example.com"},
			},
		},
		{
			name:  "multiple addresses",
			input: "john@example.com, jane@example.com",
			expected: []core.EmailAddress{
				{Email: "john@example.com"},
				{Email: "jane@example.com"},
			},
		},
		{
			name:  "addresses with names",
			input: "John Doe <john@example.com>, Jane Smith <jane@example.com>",
			expected: []core.EmailAddress{
				{Name: "John Doe", Email: "john@example.com"},
				{Name: "Jane Smith", Email: "jane@example.com"},
			},
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseEmailAddresses(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDecodeBody(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "base64url encoded",
			input:    "SGVsbG8gV29ybGQ=", // "Hello World" in base64url with padding
			expected: "Hello World",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "invalid base64",
			input:    "!!!invalid!!!",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := decodeBody(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseEmailDate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "RFC1123Z format",
			input:   "Mon, 02 Jan 2006 15:04:05 -0700",
			wantErr: false,
		},
		{
			name:    "RFC1123 format",
			input:   "Mon, 02 Jan 2006 15:04:05 MST",
			wantErr: false,
		},
		{
			name:    "alternative format",
			input:   "2 Jan 2006 15:04:05 -0700",
			wantErr: false,
		},
		{
			name:    "invalid format",
			input:   "invalid date",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseEmailDate(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, result.IsZero())
			} else {
				assert.NoError(t, err)
				assert.False(t, result.IsZero())
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		value    string
		expected bool
	}{
		{
			name:     "value exists",
			slice:    []string{"INBOX", "UNREAD", "STARRED"},
			value:    "UNREAD",
			expected: true,
		},
		{
			name:     "value does not exist",
			slice:    []string{"INBOX", "STARRED"},
			value:    "UNREAD",
			expected: false,
		},
		{
			name:     "empty slice",
			slice:    []string{},
			value:    "UNREAD",
			expected: false,
		},
		{
			name:     "nil slice",
			slice:    nil,
			value:    "UNREAD",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.slice, tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractBody(t *testing.T) {
	tests := []struct {
		name     string
		payload  *gmail.MessagePart
		expected core.EmailBody
	}{
		{
			name: "plain text only",
			payload: &gmail.MessagePart{
				MimeType: "text/plain",
				Body: &gmail.MessagePartBody{
					Data: "SGVsbG8gV29ybGQ=", // "Hello World" in base64url
				},
			},
			expected: core.EmailBody{
				Text: "Hello World",
			},
		},
		{
			name: "html only",
			payload: &gmail.MessagePart{
				MimeType: "text/html",
				Body: &gmail.MessagePartBody{
					Data: "PGI-SGVsbG88L2I-", // "<b>Hello</b>" in base64url
				},
			},
			expected: core.EmailBody{
				HTML: "<b>Hello</b>",
			},
		},
		{
			name: "multipart with text and html",
			payload: &gmail.MessagePart{
				MimeType: "multipart/alternative",
				Parts: []*gmail.MessagePart{
					{
						MimeType: "text/plain",
						Body: &gmail.MessagePartBody{
							Data: "SGVsbG8gV29ybGQ=",
						},
					},
					{
						MimeType: "text/html",
						Body: &gmail.MessagePartBody{
							Data: "PGI-SGVsbG88L2I-",
						},
					},
				},
			},
			expected: core.EmailBody{
				Text: "Hello World",
				HTML: "<b>Hello</b>",
			},
		},
		{
			name:     "nil payload",
			payload:  nil,
			expected: core.EmailBody{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractBody(tt.payload)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractAttachments(t *testing.T) {
	tests := []struct {
		name     string
		payload  *gmail.MessagePart
		expected int
	}{
		{
			name: "single attachment",
			payload: &gmail.MessagePart{
				Filename: "document.pdf",
				MimeType: "application/pdf",
				Body: &gmail.MessagePartBody{
					AttachmentId: "att-123",
					Size:         1024,
				},
			},
			expected: 1,
		},
		{
			name: "multipart with attachment",
			payload: &gmail.MessagePart{
				Parts: []*gmail.MessagePart{
					{
						MimeType: "text/plain",
						Body: &gmail.MessagePartBody{
							Data: "SGVsbG8",
						},
					},
					{
						Filename: "image.jpg",
						MimeType: "image/jpeg",
						Body: &gmail.MessagePartBody{
							AttachmentId: "att-456",
							Size:         2048,
						},
					},
				},
			},
			expected: 1,
		},
		{
			name: "no attachments",
			payload: &gmail.MessagePart{
				MimeType: "text/plain",
				Body: &gmail.MessagePartBody{
					Data: "SGVsbG8",
				},
			},
			expected: 0,
		},
		{
			name:     "nil payload",
			payload:  nil,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractAttachments(tt.payload)
			assert.Len(t, result, tt.expected)
		})
	}
}

func TestConvertMessage(t *testing.T) {
	now := time.Now()
	dateStr := now.Format(time.RFC1123Z)

	msg := &gmail.Message{
		Id:       "msg-123",
		ThreadId: "thread-456",
		Snippet:  "Test message snippet",
		LabelIds: []string{"INBOX", "UNREAD"},
		Payload: &gmail.MessagePart{
			Headers: []*gmail.MessagePartHeader{
				{Name: "Subject", Value: "Test Subject"},
				{Name: "From", Value: "sender@example.com"},
				{Name: "To", Value: "recipient@example.com"},
				{Name: "Date", Value: dateStr},
			},
			MimeType: "text/plain",
			Body: &gmail.MessagePartBody{
				Data: "VGVzdCBib2R5", // "Test body" in base64url
			},
		},
	}

	client := &Client{}
	email := client.convertMessage(msg)

	require.NotNil(t, email)
	assert.Equal(t, "msg-123", email.ID)
	assert.Equal(t, "thread-456", email.ThreadID)
	assert.Equal(t, "Test Subject", email.Subject)
	assert.Equal(t, "sender@example.com", email.From.Email)
	assert.Equal(t, "Test message snippet", email.Snippet)
	assert.False(t, email.IsRead) // Has UNREAD label
	assert.Contains(t, email.Labels, "INBOX")
	assert.Contains(t, email.Labels, "UNREAD")
}
