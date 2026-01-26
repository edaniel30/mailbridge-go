package core

import "time"

// Email represents a normalized email message across providers
type Email struct {
	ID          string         `json:"id"`
	ThreadID    string         `json:"thread_id"`
	Subject     string         `json:"subject"`
	From        EmailAddress   `json:"from"`
	To          []EmailAddress `json:"to"`
	Cc          []EmailAddress `json:"cc,omitempty"`
	Bcc         []EmailAddress `json:"bcc,omitempty"`
	ReplyTo     []EmailAddress `json:"reply_to,omitempty"`
	Date        time.Time      `json:"date"`
	Body        EmailBody      `json:"body"`
	Snippet     string         `json:"snippet"`
	Labels      []string       `json:"labels,omitempty"`
	Attachments []Attachment   `json:"attachments,omitempty"`
	IsRead      bool           `json:"is_read"`
	IsStarred   bool           `json:"is_starred"`
	IsDraft     bool           `json:"is_draft"`
}

// EmailAddress represents an email address with optional name
type EmailAddress struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

// EmailBody contains the email content in different formats
type EmailBody struct {
	Text string `json:"text,omitempty"`
	HTML string `json:"html,omitempty"`
}

// Attachment represents an email attachment
type Attachment struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	MimeType string `json:"mime_type"`
	Size     int64  `json:"size"`
	Data     []byte `json:"data,omitempty"`
}

// ListOptions contains options for listing emails
type ListOptions struct {
	MaxResults int64    `json:"max_results,omitempty"`
	PageToken  string   `json:"page_token,omitempty"`
	Query      string   `json:"query,omitempty"`
	Labels     []string `json:"labels,omitempty"`
}

// ListResponse contains the result of listing emails
type ListResponse struct {
	Emails        []*Email `json:"emails"`
	NextPageToken string   `json:"next_page_token,omitempty"`
	TotalCount    int64    `json:"total_count"`
}
