package outlook

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/danielrivera/mailbridge-go/core"
	"github.com/stretchr/testify/assert"
)

func TestDerefString(t *testing.T) {
	tests := []struct {
		name     string
		input    *string
		expected string
	}{
		{
			name:     "non-nil string",
			input:    stringPtr("test"),
			expected: "test",
		},
		{
			name:     "nil string",
			input:    nil,
			expected: "",
		},
		{
			name:     "empty string",
			input:    stringPtr(""),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := derefString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseEmailAddress(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.EmailAddress
	}{
		{
			name:  "full format",
			input: "John Doe <john@example.com>",
			expected: core.EmailAddress{
				Name:  "John Doe",
				Email: "john@example.com",
			},
		},
		{
			name:  "email only",
			input: "john@example.com",
			expected: core.EmailAddress{
				Name:  "",
				Email: "john@example.com",
			},
		},
		{
			name:  "with extra spaces",
			input: "  John Doe  < john@example.com > ",
			expected: core.EmailAddress{
				Name:  "John Doe",
				Email: "john@example.com",
			},
		},
		{
			name:     "empty string",
			input:    "",
			expected: core.EmailAddress{},
		},
		{
			name:  "malformed",
			input: "John Doe",
			expected: core.EmailAddress{
				Name:  "",
				Email: "John Doe",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseEmailAddress(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDecodeBase64(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "standard encoding",
			input:    base64.StdEncoding.EncodeToString([]byte("Hello World")),
			expected: "Hello World",
			wantErr:  false,
		},
		{
			name:     "URL encoding",
			input:    base64.URLEncoding.EncodeToString([]byte("Hello World")),
			expected: "Hello World",
			wantErr:  false,
		},
		{
			name:     "invalid base64",
			input:    "not-valid-base64!!!",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := decodeBase64(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, string(result))
			}
		})
	}
}

func TestFormatDate(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)
	result := formatDate(testTime)

	assert.NotEmpty(t, result)
	assert.Contains(t, result, "2024")
	assert.Contains(t, result, "01")
	assert.Contains(t, result, "15")
}

// Helper functions

func stringPtr(s string) *string {
	return &s
}
