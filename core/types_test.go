package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListOptions_Defaults(t *testing.T) {
	opts := &ListOptions{}

	assert.Equal(t, int64(0), opts.MaxResults)
	assert.Empty(t, opts.PageToken)
	assert.Empty(t, opts.Query)
	assert.Empty(t, opts.Labels)
}

func TestEmail_Fields(t *testing.T) {
	email := &Email{
		ID:      "test-id",
		Subject: "Test Subject",
		From: EmailAddress{
			Email: "sender@example.com",
			Name:  "Sender Name",
		},
		IsRead:    true,
		IsStarred: false,
		IsDraft:   false,
	}

	assert.Equal(t, "test-id", email.ID)
	assert.Equal(t, "Test Subject", email.Subject)
	assert.Equal(t, "sender@example.com", email.From.Email)
	assert.Equal(t, "Sender Name", email.From.Name)
	assert.True(t, email.IsRead)
	assert.False(t, email.IsStarred)
}

func TestEmailAddress(t *testing.T) {
	t.Run("with name", func(t *testing.T) {
		addr := EmailAddress{
			Email: "test@example.com",
			Name:  "Test User",
		}

		assert.Equal(t, "test@example.com", addr.Email)
		assert.Equal(t, "Test User", addr.Name)
	})

	t.Run("without name", func(t *testing.T) {
		addr := EmailAddress{
			Email: "test@example.com",
		}

		assert.Equal(t, "test@example.com", addr.Email)
		assert.Empty(t, addr.Name)
	})
}

func TestEmailBody(t *testing.T) {
	body := EmailBody{
		Text: "Plain text content",
		HTML: "<p>HTML content</p>",
	}

	assert.Equal(t, "Plain text content", body.Text)
	assert.Equal(t, "<p>HTML content</p>", body.HTML)
}

func TestAttachment(t *testing.T) {
	att := Attachment{
		ID:       "att-123",
		Filename: "document.pdf",
		MimeType: "application/pdf",
		Size:     1024,
	}

	assert.Equal(t, "att-123", att.ID)
	assert.Equal(t, "document.pdf", att.Filename)
	assert.Equal(t, "application/pdf", att.MimeType)
	assert.Equal(t, int64(1024), att.Size)
}

func TestListResponse(t *testing.T) {
	emails := []*Email{
		{ID: "1", Subject: "Test 1"},
		{ID: "2", Subject: "Test 2"},
	}

	resp := &ListResponse{
		Emails:        emails,
		NextPageToken: "next-page",
		TotalCount:    100,
	}

	assert.Len(t, resp.Emails, 2)
	assert.Equal(t, "next-page", resp.NextPageToken)
	assert.Equal(t, int64(100), resp.TotalCount)
}

func TestListOptions_WithValues(t *testing.T) {
	opts := &ListOptions{
		MaxResults: 10,
		PageToken:  "next-page",
		Query:      "is:unread",
		Labels:     []string{"INBOX", "IMPORTANT"},
	}

	assert.Equal(t, int64(10), opts.MaxResults)
	assert.Equal(t, "next-page", opts.PageToken)
	assert.Equal(t, "is:unread", opts.Query)
	assert.Len(t, opts.Labels, 2)
	assert.Contains(t, opts.Labels, "INBOX")
	assert.Contains(t, opts.Labels, "IMPORTANT")
}

func TestListResponse_Empty(t *testing.T) {
	resp := &ListResponse{
		Emails:        []*Email{},
		NextPageToken: "",
		TotalCount:    0,
	}

	assert.Empty(t, resp.Emails)
	assert.Empty(t, resp.NextPageToken)
	assert.Equal(t, int64(0), resp.TotalCount)
}

func TestEmail_ComplexFields(t *testing.T) {
	email := &Email{
		ID:       "msg-123",
		ThreadID: "thread-456",
		Subject:  "Test Email",
		From: EmailAddress{
			Name:  "Sender",
			Email: "sender@example.com",
		},
		To: []EmailAddress{
			{Name: "Recipient 1", Email: "rec1@example.com"},
			{Name: "Recipient 2", Email: "rec2@example.com"},
		},
		Cc: []EmailAddress{
			{Email: "cc@example.com"},
		},
		Bcc: []EmailAddress{
			{Email: "bcc@example.com"},
		},
		ReplyTo: []EmailAddress{
			{Email: "replyto@example.com"},
		},
		Body: EmailBody{
			Text: "Plain text",
			HTML: "<p>HTML</p>",
		},
		Labels:      []string{"INBOX", "UNREAD"},
		Attachments: []Attachment{{Filename: "file.pdf"}},
		IsRead:      false,
		IsStarred:   true,
		IsDraft:     false,
	}

	assert.Equal(t, "msg-123", email.ID)
	assert.Len(t, email.To, 2)
	assert.Len(t, email.Cc, 1)
	assert.Len(t, email.Bcc, 1)
	assert.Len(t, email.ReplyTo, 1)
	assert.NotEmpty(t, email.Body.Text)
	assert.NotEmpty(t, email.Body.HTML)
	assert.Len(t, email.Labels, 2)
	assert.Len(t, email.Attachments, 1)
	assert.False(t, email.IsRead)
	assert.True(t, email.IsStarred)
}
