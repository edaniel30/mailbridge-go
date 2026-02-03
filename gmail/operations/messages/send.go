package messages

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/danielrivera/mailbridge-go/gmail/operations"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/textproto"
	"regexp"
	"strings"
	"time"

	"github.com/danielrivera/mailbridge-go/core"
	"github.com/danielrivera/mailbridge-go/gmail/internal"
	"google.golang.org/api/gmail/v1"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// SendMessage sends an email message
func SendMessage(ctx context.Context, service internal.GmailService, draft *core.Draft, opts *core.SendOptions) (*core.SendResponse, error) {
	if err := validateDraft(draft); err != nil {
		return nil, fmt.Errorf("invalid draft: %w", err)
	}

	// Build RFC 2822 message
	var rawMessage string
	var err error
	if len(draft.Attachments) > 0 || (draft.Body.Text != "" && draft.Body.HTML != "") {
		rawMessage, err = createMIMEMessage(draft, opts)
	} else {
		rawMessage, err = buildSimpleMessage(draft, opts)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to build message: %w", err)
	}

	// Encode to base64url (Gmail format)
	encoded := encodeBase64URL([]byte(rawMessage))

	// Create Gmail message
	gmailMsg := &gmail.Message{
		Raw: encoded,
	}

	// Send via Gmail API
	messagesService := service.GetUsersService().GetMessagesService()
	call := messagesService.Send(operations.UserIDMe, gmailMsg)
	sent, err := call.Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	return &core.SendResponse{
		ID:       sent.Id,
		ThreadID: sent.ThreadId,
	}, nil
}

// validateDraft validates the draft before sending
func validateDraft(draft *core.Draft) error {
	if draft == nil {
		return fmt.Errorf("draft is nil")
	}

	// At least one recipient required
	if len(draft.To) == 0 && len(draft.Cc) == 0 && len(draft.Bcc) == 0 {
		return fmt.Errorf("at least one recipient required (To, Cc, or Bcc)")
	}

	// Validate all email addresses
	allAddresses := make([]core.EmailAddress, 0, len(draft.To)+len(draft.Cc)+len(draft.Bcc)+len(draft.ReplyTo))
	allAddresses = append(allAddresses, draft.To...)
	allAddresses = append(allAddresses, draft.Cc...)
	allAddresses = append(allAddresses, draft.Bcc...)
	allAddresses = append(allAddresses, draft.ReplyTo...)
	for _, addr := range allAddresses {
		if !isValidEmail(addr.Email) {
			return fmt.Errorf("invalid email address: %s", addr.Email)
		}
	}

	// Subject required
	if strings.TrimSpace(draft.Subject) == "" {
		return fmt.Errorf("subject is required")
	}

	// Body required (text or HTML)
	if draft.Body.Text == "" && draft.Body.HTML == "" {
		return fmt.Errorf("email body required (text or html)")
	}

	// Validate attachments
	for _, att := range draft.Attachments {
		if att.Filename == "" {
			return fmt.Errorf("attachment filename required")
		}
		if att.MimeType == "" {
			return fmt.Errorf("attachment MIME type required for %s", att.Filename)
		}
		if len(att.Data) == 0 {
			return fmt.Errorf("attachment %s has no data", att.Filename)
		}
		// Gmail limit is 25MB per attachment
		if len(att.Data) > 25*1024*1024 {
			return fmt.Errorf("attachment %s exceeds 25MB limit (size: %d bytes)", att.Filename, len(att.Data))
		}
	}

	return nil
}

// buildSimpleMessage builds a simple RFC 2822 message (no attachments, single content type)
//
//nolint:unparam // error return kept for consistency with createMIMEMessage
func buildSimpleMessage(draft *core.Draft, opts *core.SendOptions) (string, error) {
	var buf bytes.Buffer

	// Write headers
	writeHeaders(&buf, draft, opts)

	// Determine content type
	if draft.Body.HTML != "" {
		buf.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	} else {
		buf.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	}
	buf.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
	buf.WriteString("\r\n")

	// Write body
	if draft.Body.HTML != "" {
		buf.WriteString(encodeQuotedPrintable(draft.Body.HTML))
	} else {
		buf.WriteString(encodeQuotedPrintable(draft.Body.Text))
	}

	return buf.String(), nil
}

// createMIMEMessage creates a multipart MIME message
func createMIMEMessage(draft *core.Draft, opts *core.SendOptions) (string, error) {
	var buf bytes.Buffer

	// Generate boundary for multipart
	boundary := generateBoundary()

	// Write headers
	writeHeaders(&buf, draft, opts)

	// Multipart content type
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", boundary))
	buf.WriteString("\r\n")

	// Create multipart writer
	writer := multipart.NewWriter(&buf)
	if err := writer.SetBoundary(boundary); err != nil {
		return "", fmt.Errorf("failed to set boundary: %w", err)
	}

	// Write body part
	switch {
	case draft.Body.Text != "" && draft.Body.HTML != "":
		// Both text and HTML: use multipart/alternative
		if err := writeAlternativeBody(writer, draft); err != nil {
			return "", fmt.Errorf("failed to write alternative body: %w", err)
		}
	case draft.Body.HTML != "":
		// HTML only
		if err := writeHTMLBody(writer, draft.Body.HTML); err != nil {
			return "", fmt.Errorf("failed to write HTML body: %w", err)
		}
	default:
		// Text only
		if err := writeTextBody(writer, draft.Body.Text); err != nil {
			return "", fmt.Errorf("failed to write text body: %w", err)
		}
	}

	// Write attachments
	for _, att := range draft.Attachments {
		if err := writeAttachment(writer, &att); err != nil {
			return "", fmt.Errorf("failed to write attachment %s: %w", att.Filename, err)
		}
	}

	// Close multipart
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close multipart writer: %w", err)
	}

	return buf.String(), nil
}

// writeHeaders writes RFC 2822 headers
func writeHeaders(buf *bytes.Buffer, draft *core.Draft, opts *core.SendOptions) {
	// From (required by RFC 2822, but Gmail uses authenticated user)
	buf.WriteString("From: me\r\n")

	// To
	if len(draft.To) > 0 {
		buf.WriteString("To: " + formatEmailAddresses(draft.To) + "\r\n")
	}

	// Cc
	if len(draft.Cc) > 0 {
		buf.WriteString("Cc: " + formatEmailAddresses(draft.Cc) + "\r\n")
	}

	// Bcc
	if len(draft.Bcc) > 0 {
		buf.WriteString("Bcc: " + formatEmailAddresses(draft.Bcc) + "\r\n")
	}

	// Reply-To
	if len(draft.ReplyTo) > 0 {
		buf.WriteString("Reply-To: " + formatEmailAddresses(draft.ReplyTo) + "\r\n")
	}

	// Subject
	buf.WriteString("Subject: " + encodeMIMEHeader(draft.Subject) + "\r\n")

	// Date
	buf.WriteString("Date: " + time.Now().Format(time.RFC1123Z) + "\r\n")

	// Message-ID
	buf.WriteString("Message-ID: " + generateMessageID() + "\r\n")

	// MIME-Version
	buf.WriteString("MIME-Version: 1.0\r\n")

	// Custom headers from draft
	for key, value := range draft.Headers {
		fmt.Fprintf(buf, "%s: %s\r\n", key, value)
	}

	// Custom headers from options
	if opts != nil {
		for key, value := range opts.CustomHeaders {
			fmt.Fprintf(buf, "%s: %s\r\n", key, value)
		}
	}
}

// writeAlternativeBody writes multipart/alternative body (text + HTML)
func writeAlternativeBody(parentWriter *multipart.Writer, draft *core.Draft) error {
	// Create alternative part
	altBoundary := generateBoundary()
	headers := textproto.MIMEHeader{}
	headers.Set("Content-Type", fmt.Sprintf("multipart/alternative; boundary=\"%s\"", altBoundary))

	altPart, err := parentWriter.CreatePart(headers)
	if err != nil {
		return err
	}

	// Create nested multipart writer for alternative
	altWriter := multipart.NewWriter(altPart)
	if err := altWriter.SetBoundary(altBoundary); err != nil {
		return err
	}

	// Write text version
	if err := writeTextBody(altWriter, draft.Body.Text); err != nil {
		return err
	}

	// Write HTML version
	if err := writeHTMLBody(altWriter, draft.Body.HTML); err != nil {
		return err
	}

	return altWriter.Close()
}

// writeTextBody writes a text/plain part
func writeTextBody(writer *multipart.Writer, text string) error {
	headers := textproto.MIMEHeader{}
	headers.Set("Content-Type", "text/plain; charset=\"UTF-8\"")
	headers.Set("Content-Transfer-Encoding", "quoted-printable")

	part, err := writer.CreatePart(headers)
	if err != nil {
		return err
	}

	_, err = part.Write([]byte(encodeQuotedPrintable(text)))
	return err
}

// writeHTMLBody writes a text/html part
func writeHTMLBody(writer *multipart.Writer, html string) error {
	headers := textproto.MIMEHeader{}
	headers.Set("Content-Type", "text/html; charset=\"UTF-8\"")
	headers.Set("Content-Transfer-Encoding", "quoted-printable")

	part, err := writer.CreatePart(headers)
	if err != nil {
		return err
	}

	_, err = part.Write([]byte(encodeQuotedPrintable(html)))
	return err
}

// writeAttachment writes an attachment part
func writeAttachment(writer *multipart.Writer, att *core.Attachment) error {
	headers := textproto.MIMEHeader{}
	headers.Set("Content-Type", att.MimeType+"; name=\""+mime.QEncoding.Encode("UTF-8", att.Filename)+"\"")
	headers.Set("Content-Disposition", "attachment; filename=\""+mime.QEncoding.Encode("UTF-8", att.Filename)+"\"")
	headers.Set("Content-Transfer-Encoding", "base64")

	part, err := writer.CreatePart(headers)
	if err != nil {
		return err
	}

	// Encode attachment data as base64
	encoded := base64.StdEncoding.EncodeToString(att.Data)

	// Write in chunks of 76 characters (RFC 2045)
	for i := 0; i < len(encoded); i += 76 {
		end := i + 76
		if end > len(encoded) {
			end = len(encoded)
		}
		if _, err := part.Write([]byte(encoded[i:end] + "\r\n")); err != nil {
			return err
		}
	}

	return nil
}

// formatEmailAddress formats an EmailAddress to RFC 2822 format
func formatEmailAddress(addr core.EmailAddress) string {
	if addr.Name == "" {
		return addr.Email
	}

	// Check if name needs quoting (contains special characters)
	if strings.ContainsAny(addr.Name, ",;\"<>") {
		return fmt.Sprintf("\"%s\" <%s>", strings.ReplaceAll(addr.Name, "\"", "\\\""), addr.Email)
	}

	return fmt.Sprintf("%s <%s>", addr.Name, addr.Email)
}

// formatEmailAddresses formats multiple email addresses
func formatEmailAddresses(addrs []core.EmailAddress) string {
	if len(addrs) == 0 {
		return ""
	}

	formatted := make([]string, len(addrs))
	for i, addr := range addrs {
		formatted[i] = formatEmailAddress(addr)
	}

	return strings.Join(formatted, ", ")
}

// encodeBase64URL encodes data to base64url format (URL-safe, no padding)
func encodeBase64URL(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

// encodeMIMEHeader encodes a header value using MIME Q-encoding if needed
func encodeMIMEHeader(value string) string {
	// Check if encoding is needed (non-ASCII characters)
	needsEncoding := false
	for _, r := range value {
		if r > 127 {
			needsEncoding = true
			break
		}
	}

	if needsEncoding {
		return mime.QEncoding.Encode("UTF-8", value)
	}

	return value
}

// encodeQuotedPrintable encodes text using quoted-printable encoding
func encodeQuotedPrintable(text string) string {
	var buf bytes.Buffer
	writer := quotedprintable.NewWriter(&buf)
	_, _ = writer.Write([]byte(text))
	_ = writer.Close()
	return buf.String()
}

// generateMessageID generates a unique RFC 2822 Message-ID
func generateMessageID() string {
	// Generate random bytes
	b := make([]byte, 16)
	_, _ = rand.Read(b)

	// Format as Message-ID
	return fmt.Sprintf("<%x.%d@mailbridge.local>",
		b,
		time.Now().UnixNano())
}

// generateBoundary generates a unique MIME boundary
func generateBoundary() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("==boundary_%x==", b)
}

// isValidEmail validates an email address format
func isValidEmail(email string) bool {
	if email == "" {
		return false
	}
	return emailRegex.MatchString(email)
}
