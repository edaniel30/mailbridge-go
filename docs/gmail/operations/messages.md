# Message Operations

Get message details and content from Gmail.

> **Setup required**: [OAuth2 configuration](../gmail.md#setup-oauth2)

## Get Message Details

Retrieve a specific email message by its ID.

```go
email, err := client.GetMessage(ctx, messageID)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Subject: %s\n", email.Subject)
fmt.Printf("From: %s <%s>\n", email.From.Name, email.From.Email)
fmt.Printf("Date: %s\n", email.Date)
fmt.Printf("Body (Text): %s\n", email.Body.Text)
fmt.Printf("Body (HTML): %s\n", email.Body.HTML)

// Attachments metadata (not downloaded yet)
for _, att := range email.Attachments {
    fmt.Printf("Attachment: %s (%d bytes)\n", att.Filename, att.Size)
}
```

**Returns**:
- `*core.Email` with complete message details
- Message ID, thread ID, and snippet
- Sender, recipients (To, Cc, Bcc)
- Subject and date
- Body content (text and HTML)
- Attachment metadata (use GetAttachment to download)
- Labels applied to the message
- Read/starred status

## Example Usage

```go
// Get a specific message
email, err := client.GetMessage(ctx, "18c8f8a7b2e3d4f5")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("From: %s\n", email.From.Email)
fmt.Printf("Subject: %s\n", email.Subject)
fmt.Printf("Body: %s\n", email.Body.Text)

// Check if message has attachments
if len(email.Attachments) > 0 {
    fmt.Printf("Has %d attachment(s)\n", len(email.Attachments))
}
```

## Related Operations

- [Attachments](./attachments.md) - Download files from messages
- [Delete](./delete.md) - Trash or permanently delete messages
- [Notifications](./notifications.md) - Monitor mailbox for new messages
