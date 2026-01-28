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

// Draft represents a message being composed for sending
type Draft struct {
	To          []EmailAddress        `json:"to,omitempty"`
	Cc          []EmailAddress        `json:"cc,omitempty"`
	Bcc         []EmailAddress        `json:"bcc,omitempty"`
	Subject     string                `json:"subject"`
	Body        EmailBody             `json:"body"`
	Attachments []Attachment          `json:"attachments,omitempty"`
	ReplyTo     []EmailAddress        `json:"reply_to,omitempty"`
	Headers     map[string]string     `json:"headers,omitempty"`
}

// SendOptions contains options for sending emails
type SendOptions struct {
	CustomHeaders map[string]string `json:"custom_headers,omitempty"`
}

// SendResponse contains the result of sending an email
type SendResponse struct {
	ID       string `json:"id"`
	ThreadID string `json:"thread_id,omitempty"`
}

// BatchModifyRequest contains options for batch modifying messages
type BatchModifyRequest struct {
	MessageIDs     []string `json:"message_ids"`
	AddLabelIDs    []string `json:"add_label_ids,omitempty"`
	RemoveLabelIDs []string `json:"remove_label_ids,omitempty"`
}

// WatchRequest contains options for setting up push notifications
type WatchRequest struct {
	TopicName           string   `json:"topic_name"`                     // Required: Google Cloud Pub/Sub topic
	LabelIDs            []string `json:"label_ids,omitempty"`            // Optional: filter by labels
	LabelFilterBehavior string   `json:"label_filter_behavior,omitempty"` // "include" or "exclude"
}

// WatchResponse contains information about the watch
type WatchResponse struct {
	HistoryID  string `json:"history_id"`
	Expiration int64  `json:"expiration"` // Unix timestamp in milliseconds
}

// HistoryRequest contains options for fetching history
type HistoryRequest struct {
	StartHistoryID string   `json:"start_history_id"`
	MaxResults     int64    `json:"max_results,omitempty"`
	PageToken      string   `json:"page_token,omitempty"`
	LabelID        string   `json:"label_id,omitempty"`
	HistoryTypes   []string `json:"history_types,omitempty"` // messageAdded, messageDeleted, labelAdded, labelRemoved
}

// HistoryResponse contains history records
type HistoryResponse struct {
	History       []*HistoryRecord `json:"history"`
	NextPageToken string           `json:"next_page_token,omitempty"`
	HistoryID     string           `json:"history_id"`
}

// HistoryRecord represents a change in the mailbox
type HistoryRecord struct {
	ID              string   `json:"id"`
	Messages        []*Email `json:"messages,omitempty"`
	MessagesAdded   []*HistoryMessageAdded `json:"messages_added,omitempty"`
	MessagesDeleted []*HistoryMessageDeleted `json:"messages_deleted,omitempty"`
	LabelsAdded     []*HistoryLabelChange `json:"labels_added,omitempty"`
	LabelsRemoved   []*HistoryLabelChange `json:"labels_removed,omitempty"`
}

// HistoryMessageAdded represents a message that was added
type HistoryMessageAdded struct {
	Message *Email `json:"message"`
}

// HistoryMessageDeleted represents a message that was deleted
type HistoryMessageDeleted struct {
	Message *Email `json:"message"`
}

// HistoryLabelChange represents a label change on a message
type HistoryLabelChange struct {
	Message  *Email   `json:"message"`
	LabelIDs []string `json:"label_ids"`
}
