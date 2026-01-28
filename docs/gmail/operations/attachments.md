# Download Attachments

Download files attached to Gmail messages.

> **Setup required**: [OAuth2 configuration](../gmail/GMAIL.md#setup-oauth2)

## Quick Example

```go
// 1. List messages with attachments
messages, _ := client.ListMessages(ctx, &core.ListOptions{
    Query: "has:attachment",
})

// 2. Get message details
email, _ := client.GetMessage(ctx, messages.Emails[0].ID)

// 3. Download each attachment
for _, att := range email.Attachments {
    data, err := client.GetAttachment(ctx, email.ID, att.ID)
    if err != nil {
        log.Printf("Failed to download %s: %v", att.Filename, err)
        continue
    }

    // Save to file
    os.WriteFile(att.Filename, data, 0644)
    fmt.Printf("Downloaded: %s (%d bytes)\n", att.Filename, len(data))
}
```

## Attachment Metadata

```go
email, _ := client.GetMessage(ctx, messageID)

for _, att := range email.Attachments {
    fmt.Printf("ID: %s\n", att.ID)              // For GetAttachment()
    fmt.Printf("Filename: %s\n", att.Filename)  // Original name
    fmt.Printf("MimeType: %s\n", att.MimeType)  // e.g., "image/png"
    fmt.Printf("Size: %d bytes\n", att.Size)    // File size
}
```

## Filter by Type

```go
// Find PDFs
messages, _ := client.ListMessages(ctx, &core.ListOptions{
    Query: "has:attachment filename:pdf",
})

// Find images
messages, _ := client.ListMessages(ctx, &core.ListOptions{
    Query: "has:attachment (filename:jpg OR filename:png)",
})
```

## Download to Memory

```go
data, err := client.GetAttachment(ctx, messageID, attachmentID)
// data is []byte - process directly or save to file
```

## Complete Example

See [`examples/gmail-attachments`](../../examples/gmail-attachments/) for full code.

```bash
cd examples/gmail-attachments
go run main.go
```

## Related

- [Messages](./messages.md) - List and read emails
- [Search](./search.md) - Find emails with attachments
