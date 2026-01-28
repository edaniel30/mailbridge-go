package messages

import (
	"strings"
	"testing"

	"github.com/danielrivera/mailbridge-go/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		valid bool
	}{
		{"valid simple", "test@example.com", true},
		{"valid with plus", "user+tag@example.com", true},
		{"valid with dash", "first-last@example.com", true},
		{"valid with dot", "first.last@example.com", true},
		{"valid subdomain", "user@mail.example.com", true},
		{"invalid no at", "userexample.com", false},
		{"invalid no domain", "user@", false},
		{"invalid no user", "@example.com", false},
		{"invalid no tld", "user@example", false},
		{"invalid spaces", "user @example.com", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidEmail(tt.email)
			assert.Equal(t, tt.valid, result, "Email: %s", tt.email)
		})
	}
}

func TestFormatEmailAddress(t *testing.T) {
	tests := []struct {
		name     string
		addr     core.EmailAddress
		expected string
	}{
		{
			name:     "email only",
			addr:     core.EmailAddress{Email: "test@example.com"},
			expected: "test@example.com",
		},
		{
			name:     "with simple name",
			addr:     core.EmailAddress{Name: "John Doe", Email: "john@example.com"},
			expected: "John Doe <john@example.com>",
		},
		{
			name:     "name with comma needs quotes",
			addr:     core.EmailAddress{Name: "Doe, John", Email: "john@example.com"},
			expected: "\"Doe, John\" <john@example.com>",
		},
		{
			name:     "name with semicolon needs quotes",
			addr:     core.EmailAddress{Name: "John; Doe", Email: "john@example.com"},
			expected: "\"John; Doe\" <john@example.com>",
		},
		{
			name:     "name with quote needs escaping",
			addr:     core.EmailAddress{Name: "John \"Johnny\" Doe", Email: "john@example.com"},
			expected: "\"John \\\"Johnny\\\" Doe\" <john@example.com>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatEmailAddress(tt.addr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatEmailAddresses(t *testing.T) {
	tests := []struct {
		name     string
		addrs    []core.EmailAddress
		expected string
	}{
		{
			name:     "empty list",
			addrs:    []core.EmailAddress{},
			expected: "",
		},
		{
			name: "single address",
			addrs: []core.EmailAddress{
				{Email: "test@example.com"},
			},
			expected: "test@example.com",
		},
		{
			name: "multiple addresses",
			addrs: []core.EmailAddress{
				{Email: "first@example.com"},
				{Email: "second@example.com"},
			},
			expected: "first@example.com, second@example.com",
		},
		{
			name: "mix of name and email-only",
			addrs: []core.EmailAddress{
				{Name: "John Doe", Email: "john@example.com"},
				{Email: "plain@example.com"},
			},
			expected: "John Doe <john@example.com>, plain@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatEmailAddresses(tt.addrs)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEncodeBase64URL(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "simple string",
			input:    []byte("Hello World"),
			expected: "SGVsbG8gV29ybGQ",
		},
		{
			name:     "empty string",
			input:    []byte(""),
			expected: "",
		},
		{
			name:     "with special chars",
			input:    []byte("test+/="),
			expected: "dGVzdCsvPQ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := encodeBase64URL(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEncodeMIMEHeader(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"ascii only", "Simple Subject"},
		{"with unicode", "Prueba con ñ y acentos"},
		{"with emoji", "Test 📧 Email"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := encodeMIMEHeader(tt.input)
			// Should not be empty
			assert.NotEmpty(t, result)
			// ASCII-only strings should remain unchanged
			if isASCII(tt.input) {
				assert.Equal(t, tt.input, result)
			}
		})
	}
}

func TestGenerateMessageID(t *testing.T) {
	id1 := generateMessageID()
	id2 := generateMessageID()

	// Should start and end with angle brackets
	assert.True(t, strings.HasPrefix(id1, "<"))
	assert.True(t, strings.HasSuffix(id1, ">"))

	// Should be unique
	assert.NotEqual(t, id1, id2)

	// Should contain @ symbol
	assert.Contains(t, id1, "@")
}

func TestGenerateBoundary(t *testing.T) {
	b1 := generateBoundary()
	b2 := generateBoundary()

	// Should have expected prefix
	assert.True(t, strings.HasPrefix(b1, "==boundary_"))

	// Should be unique
	assert.NotEqual(t, b1, b2)
}

func TestValidateDraft(t *testing.T) {
	tests := []struct {
		name    string
		draft   *core.Draft
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid draft - To only",
			draft: &core.Draft{
				To:      []core.EmailAddress{{Email: "test@example.com"}},
				Subject: "Test",
				Body:    core.EmailBody{Text: "Hello"},
			},
			wantErr: false,
		},
		{
			name: "valid draft - Cc only",
			draft: &core.Draft{
				Cc:      []core.EmailAddress{{Email: "test@example.com"}},
				Subject: "Test",
				Body:    core.EmailBody{Text: "Hello"},
			},
			wantErr: false,
		},
		{
			name: "valid draft - Bcc only",
			draft: &core.Draft{
				Bcc:     []core.EmailAddress{{Email: "test@example.com"}},
				Subject: "Test",
				Body:    core.EmailBody{Text: "Hello"},
			},
			wantErr: false,
		},
		{
			name: "valid draft - HTML body",
			draft: &core.Draft{
				To:      []core.EmailAddress{{Email: "test@example.com"}},
				Subject: "Test",
				Body:    core.EmailBody{HTML: "<p>Hello</p>"},
			},
			wantErr: false,
		},
		{
			name: "valid draft - with attachment",
			draft: &core.Draft{
				To:      []core.EmailAddress{{Email: "test@example.com"}},
				Subject: "Test",
				Body:    core.EmailBody{Text: "Hello"},
				Attachments: []core.Attachment{
					{Filename: "test.txt", MimeType: "text/plain", Data: []byte("data")},
				},
			},
			wantErr: false,
		},
		{
			name:    "nil draft",
			draft:   nil,
			wantErr: true,
			errMsg:  "draft is nil",
		},
		{
			name: "no recipients",
			draft: &core.Draft{
				Subject: "Test",
				Body:    core.EmailBody{Text: "Hello"},
			},
			wantErr: true,
			errMsg:  "at least one recipient required",
		},
		{
			name: "invalid email address",
			draft: &core.Draft{
				To:      []core.EmailAddress{{Email: "invalid-email"}},
				Subject: "Test",
				Body:    core.EmailBody{Text: "Hello"},
			},
			wantErr: true,
			errMsg:  "invalid email address",
		},
		{
			name: "empty subject",
			draft: &core.Draft{
				To:      []core.EmailAddress{{Email: "test@example.com"}},
				Subject: "",
				Body:    core.EmailBody{Text: "Hello"},
			},
			wantErr: true,
			errMsg:  "subject is required",
		},
		{
			name: "whitespace-only subject",
			draft: &core.Draft{
				To:      []core.EmailAddress{{Email: "test@example.com"}},
				Subject: "   ",
				Body:    core.EmailBody{Text: "Hello"},
			},
			wantErr: true,
			errMsg:  "subject is required",
		},
		{
			name: "empty body",
			draft: &core.Draft{
				To:      []core.EmailAddress{{Email: "test@example.com"}},
				Subject: "Test",
				Body:    core.EmailBody{},
			},
			wantErr: true,
			errMsg:  "email body required",
		},
		{
			name: "attachment without filename",
			draft: &core.Draft{
				To:      []core.EmailAddress{{Email: "test@example.com"}},
				Subject: "Test",
				Body:    core.EmailBody{Text: "Hello"},
				Attachments: []core.Attachment{
					{Filename: "", MimeType: "text/plain", Data: []byte("data")},
				},
			},
			wantErr: true,
			errMsg:  "attachment filename required",
		},
		{
			name: "attachment without mime type",
			draft: &core.Draft{
				To:      []core.EmailAddress{{Email: "test@example.com"}},
				Subject: "Test",
				Body:    core.EmailBody{Text: "Hello"},
				Attachments: []core.Attachment{
					{Filename: "test.txt", MimeType: "", Data: []byte("data")},
				},
			},
			wantErr: true,
			errMsg:  "attachment MIME type required",
		},
		{
			name: "attachment without data",
			draft: &core.Draft{
				To:      []core.EmailAddress{{Email: "test@example.com"}},
				Subject: "Test",
				Body:    core.EmailBody{Text: "Hello"},
				Attachments: []core.Attachment{
					{Filename: "test.txt", MimeType: "text/plain", Data: []byte{}},
				},
			},
			wantErr: true,
			errMsg:  "has no data",
		},
		{
			name: "attachment too large",
			draft: &core.Draft{
				To:      []core.EmailAddress{{Email: "test@example.com"}},
				Subject: "Test",
				Body:    core.EmailBody{Text: "Hello"},
				Attachments: []core.Attachment{
					{Filename: "huge.bin", MimeType: "application/octet-stream", Data: make([]byte, 26*1024*1024)},
				},
			},
			wantErr: true,
			errMsg:  "exceeds 25MB limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDraft(tt.draft)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBuildSimpleMessage(t *testing.T) {
	tests := []struct {
		name        string
		draft       *core.Draft
		opts        *core.SendOptions
		checkFields []string
	}{
		{
			name: "text only",
			draft: &core.Draft{
				To:      []core.EmailAddress{{Email: "test@example.com"}},
				Subject: "Test Subject",
				Body:    core.EmailBody{Text: "Hello World"},
			},
			checkFields: []string{
				"To: test@example.com",
				"Subject: Test Subject",
				"Content-Type: text/plain",
				"Hello World",
			},
		},
		{
			name: "HTML only",
			draft: &core.Draft{
				To:      []core.EmailAddress{{Email: "test@example.com"}},
				Subject: "Test Subject",
				Body:    core.EmailBody{HTML: "<p>Hello</p>"},
			},
			checkFields: []string{
				"To: test@example.com",
				"Subject: Test Subject",
				"Content-Type: text/html",
				"<p>Hello</p>",
			},
		},
		{
			name: "with Cc and Bcc",
			draft: &core.Draft{
				To:      []core.EmailAddress{{Email: "to@example.com"}},
				Cc:      []core.EmailAddress{{Email: "cc@example.com"}},
				Bcc:     []core.EmailAddress{{Email: "bcc@example.com"}},
				Subject: "Test",
				Body:    core.EmailBody{Text: "Hello"},
			},
			checkFields: []string{
				"To: to@example.com",
				"Cc: cc@example.com",
				"Bcc: bcc@example.com",
			},
		},
		{
			name: "with ReplyTo",
			draft: &core.Draft{
				To:      []core.EmailAddress{{Email: "to@example.com"}},
				ReplyTo: []core.EmailAddress{{Email: "reply@example.com"}},
				Subject: "Test",
				Body:    core.EmailBody{Text: "Hello"},
			},
			checkFields: []string{
				"Reply-To: reply@example.com",
			},
		},
		{
			name: "with custom headers",
			draft: &core.Draft{
				To:      []core.EmailAddress{{Email: "test@example.com"}},
				Subject: "Test",
				Body:    core.EmailBody{Text: "Hello"},
				Headers: map[string]string{
					"X-Priority": "1",
				},
			},
			checkFields: []string{
				"X-Priority: 1",
			},
		},
		{
			name: "with send options headers",
			draft: &core.Draft{
				To:      []core.EmailAddress{{Email: "test@example.com"}},
				Subject: "Test",
				Body:    core.EmailBody{Text: "Hello"},
			},
			opts: &core.SendOptions{
				CustomHeaders: map[string]string{
					"X-Mailer": "MailBridge",
				},
			},
			checkFields: []string{
				"X-Mailer: MailBridge",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := buildSimpleMessage(tt.draft, tt.opts)
			require.NoError(t, err)
			assert.NotEmpty(t, result)

			// Check required fields exist
			assert.Contains(t, result, "From: me")
			assert.Contains(t, result, "Date: ")
			assert.Contains(t, result, "Message-ID: ")
			assert.Contains(t, result, "MIME-Version: 1.0")

			// Check test-specific fields
			for _, field := range tt.checkFields {
				assert.Contains(t, result, field)
			}
		})
	}
}

func TestCreateMIMEMessage(t *testing.T) {
	tests := []struct {
		name        string
		draft       *core.Draft
		checkFields []string
	}{
		{
			name: "text and HTML (alternative)",
			draft: &core.Draft{
				To:      []core.EmailAddress{{Email: "test@example.com"}},
				Subject: "Test",
				Body: core.EmailBody{
					Text: "Plain text version",
					HTML: "<p>HTML version</p>",
				},
			},
			checkFields: []string{
				"Content-Type: multipart/mixed",
				"Content-Type: multipart/alternative",
				"Content-Type: text/plain",
				"Content-Type: text/html",
			},
		},
		{
			name: "with attachment",
			draft: &core.Draft{
				To:      []core.EmailAddress{{Email: "test@example.com"}},
				Subject: "Test",
				Body:    core.EmailBody{Text: "See attachment"},
				Attachments: []core.Attachment{
					{
						Filename: "test.txt",
						MimeType: "text/plain",
						Data:     []byte("Test file content"),
					},
				},
			},
			checkFields: []string{
				"Content-Type: multipart/mixed",
				"Content-Disposition: attachment",
				"filename=\"test.txt\"",
			},
		},
		{
			name: "multiple attachments",
			draft: &core.Draft{
				To:      []core.EmailAddress{{Email: "test@example.com"}},
				Subject: "Test",
				Body:    core.EmailBody{Text: "See attachments"},
				Attachments: []core.Attachment{
					{
						Filename: "file1.txt",
						MimeType: "text/plain",
						Data:     []byte("File 1"),
					},
					{
						Filename: "file2.txt",
						MimeType: "text/plain",
						Data:     []byte("File 2"),
					},
				},
			},
			checkFields: []string{
				"filename=\"file1.txt\"",
				"filename=\"file2.txt\"",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := createMIMEMessage(tt.draft, nil)
			require.NoError(t, err)
			assert.NotEmpty(t, result)

			// Check for expected fields
			for _, field := range tt.checkFields {
				assert.Contains(t, result, field)
			}
		})
	}
}

// Helper function to check if string contains only ASCII
func isASCII(s string) bool {
	for _, r := range s {
		if r > 127 {
			return false
		}
	}
	return true
}
