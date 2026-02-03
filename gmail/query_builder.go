package gmail

import (
	"fmt"
	"strings"
	"time"
)

// QueryBuilder helps construct Gmail search queries
type QueryBuilder struct {
	parts []string
}

// NewQueryBuilder creates a new QueryBuilder instance
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		parts: make([]string, 0),
	}
}

// Build returns the final query string
func (qb *QueryBuilder) Build() string {
	return strings.Join(qb.parts, " ")
}

// Reset clears all query parts
func (qb *QueryBuilder) Reset() *QueryBuilder {
	qb.parts = make([]string, 0)
	return qb
}

// From adds a sender filter
func (qb *QueryBuilder) From(email string) *QueryBuilder {
	qb.parts = append(qb.parts, fmt.Sprintf("from:%s", email))
	return qb
}

// To adds a recipient filter
func (qb *QueryBuilder) To(email string) *QueryBuilder {
	qb.parts = append(qb.parts, fmt.Sprintf("to:%s", email))
	return qb
}

// Subject adds a subject filter
func (qb *QueryBuilder) Subject(subject string) *QueryBuilder {
	// If subject contains spaces, wrap in quotes
	if strings.Contains(subject, " ") {
		qb.parts = append(qb.parts, fmt.Sprintf("subject:\"%s\"", subject))
	} else {
		qb.parts = append(qb.parts, fmt.Sprintf("subject:%s", subject))
	}
	return qb
}

// HasWords adds a search for messages containing specific words
func (qb *QueryBuilder) HasWords(words string) *QueryBuilder {
	if strings.Contains(words, " ") {
		qb.parts = append(qb.parts, fmt.Sprintf("\"%s\"", words))
	} else {
		qb.parts = append(qb.parts, words)
	}
	return qb
}

// IsUnread filters for unread messages
func (qb *QueryBuilder) IsUnread() *QueryBuilder {
	qb.parts = append(qb.parts, "is:unread")
	return qb
}

// IsRead filters for read messages
func (qb *QueryBuilder) IsRead() *QueryBuilder {
	qb.parts = append(qb.parts, "is:read")
	return qb
}

// IsStarred filters for starred messages
func (qb *QueryBuilder) IsStarred() *QueryBuilder {
	qb.parts = append(qb.parts, "is:starred")
	return qb
}

// IsImportant filters for important messages
func (qb *QueryBuilder) IsImportant() *QueryBuilder {
	qb.parts = append(qb.parts, "is:important")
	return qb
}

// IsSnoozed filters for snoozed messages
func (qb *QueryBuilder) IsSnoozed() *QueryBuilder {
	qb.parts = append(qb.parts, "is:snoozed")
	return qb
}

// HasAttachment filters for messages with attachments
func (qb *QueryBuilder) HasAttachment() *QueryBuilder {
	qb.parts = append(qb.parts, "has:attachment")
	return qb
}

// Filename filters by attachment filename
func (qb *QueryBuilder) Filename(filename string) *QueryBuilder {
	qb.parts = append(qb.parts, fmt.Sprintf("filename:%s", filename))
	return qb
}

// After filters for messages after a specific date
func (qb *QueryBuilder) After(date time.Time) *QueryBuilder {
	qb.parts = append(qb.parts, fmt.Sprintf("after:%s", date.Format("2006/01/02")))
	return qb
}

// Before filters for messages before a specific date
func (qb *QueryBuilder) Before(date time.Time) *QueryBuilder {
	qb.parts = append(qb.parts, fmt.Sprintf("before:%s", date.Format("2006/01/02")))
	return qb
}

// OlderThan filters for messages older than a time period
// Examples: "2d" (2 days), "4m" (4 months), "1y" (1 year)
func (qb *QueryBuilder) OlderThan(period string) *QueryBuilder {
	qb.parts = append(qb.parts, fmt.Sprintf("older_than:%s", period))
	return qb
}

// NewerThan filters for messages newer than a time period
// Examples: "2d" (2 days), "4m" (4 months), "1y" (1 year)
func (qb *QueryBuilder) NewerThan(period string) *QueryBuilder {
	qb.parts = append(qb.parts, fmt.Sprintf("newer_than:%s", period))
	return qb
}

// LargerThan filters for messages larger than a specific size in bytes
func (qb *QueryBuilder) LargerThan(bytes int64) *QueryBuilder {
	qb.parts = append(qb.parts, fmt.Sprintf("larger:%d", bytes))
	return qb
}

// SmallerThan filters for messages smaller than a specific size in bytes
func (qb *QueryBuilder) SmallerThan(bytes int64) *QueryBuilder {
	qb.parts = append(qb.parts, fmt.Sprintf("smaller:%d", bytes))
	return qb
}

// Label filters by label name
func (qb *QueryBuilder) Label(labelName string) *QueryBuilder {
	qb.parts = append(qb.parts, fmt.Sprintf("label:%s", labelName))
	return qb
}

// InInbox filters for messages in inbox
func (qb *QueryBuilder) InInbox() *QueryBuilder {
	qb.parts = append(qb.parts, "in:inbox")
	return qb
}

// InTrash filters for messages in trash
func (qb *QueryBuilder) InTrash() *QueryBuilder {
	qb.parts = append(qb.parts, "in:trash")
	return qb
}

// InSpam filters for messages in spam
func (qb *QueryBuilder) InSpam() *QueryBuilder {
	qb.parts = append(qb.parts, "in:spam")
	return qb
}

// InSent filters for sent messages
func (qb *QueryBuilder) InSent() *QueryBuilder {
	qb.parts = append(qb.parts, "in:sent")
	return qb
}

// InDrafts filters for draft messages
func (qb *QueryBuilder) InDrafts() *QueryBuilder {
	qb.parts = append(qb.parts, "in:drafts")
	return qb
}

// Category filters by category
// Valid categories: primary, social, promotions, updates, forums
func (qb *QueryBuilder) Category(category string) *QueryBuilder {
	qb.parts = append(qb.parts, fmt.Sprintf("category:%s", category))
	return qb
}

// HasDrive filters for messages with Google Drive attachments
func (qb *QueryBuilder) HasDrive() *QueryBuilder {
	qb.parts = append(qb.parts, "has:drive")
	return qb
}

// HasDocument filters for messages with Google Docs attachments
func (qb *QueryBuilder) HasDocument() *QueryBuilder {
	qb.parts = append(qb.parts, "has:document")
	return qb
}

// HasSpreadsheet filters for messages with Google Sheets attachments
func (qb *QueryBuilder) HasSpreadsheet() *QueryBuilder {
	qb.parts = append(qb.parts, "has:spreadsheet")
	return qb
}

// HasPresentation filters for messages with Google Slides attachments
func (qb *QueryBuilder) HasPresentation() *QueryBuilder {
	qb.parts = append(qb.parts, "has:presentation")
	return qb
}

// HasYouTube filters for messages with YouTube videos
func (qb *QueryBuilder) HasYouTube() *QueryBuilder {
	qb.parts = append(qb.parts, "has:youtube")
	return qb
}

// Cc filters by CC recipient
func (qb *QueryBuilder) Cc(email string) *QueryBuilder {
	qb.parts = append(qb.parts, fmt.Sprintf("cc:%s", email))
	return qb
}

// Bcc filters by BCC recipient
func (qb *QueryBuilder) Bcc(email string) *QueryBuilder {
	qb.parts = append(qb.parts, fmt.Sprintf("bcc:%s", email))
	return qb
}

// DeliveredTo filters by delivered-to header
func (qb *QueryBuilder) DeliveredTo(email string) *QueryBuilder {
	qb.parts = append(qb.parts, fmt.Sprintf("deliveredto:%s", email))
	return qb
}

// List filters by mailing list
func (qb *QueryBuilder) List(list string) *QueryBuilder {
	qb.parts = append(qb.parts, fmt.Sprintf("list:%s", list))
	return qb
}

// Raw adds a raw query string (for advanced users)
func (qb *QueryBuilder) Raw(query string) *QueryBuilder {
	qb.parts = append(qb.parts, query)
	return qb
}

// OR adds an OR operator between the previous query and the next one
func (qb *QueryBuilder) OR() *QueryBuilder {
	qb.parts = append(qb.parts, "OR")
	return qb
}

// NOT negates the next condition
func (qb *QueryBuilder) NOT() *QueryBuilder {
	qb.parts = append(qb.parts, "NOT")
	return qb
}

// Helper functions for common sizes

// MegaBytes converts MB to bytes for size queries
func MegaBytes(mb int64) int64 {
	return mb * 1024 * 1024
}

// KiloBytes converts KB to bytes for size queries
func KiloBytes(kb int64) int64 {
	return kb * 1024
}

// GigaBytes converts GB to bytes for size queries
func GigaBytes(gb int64) int64 {
	return gb * 1024 * 1024 * 1024
}
