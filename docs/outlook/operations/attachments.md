# Attachment Operations

Download and manage email attachments in Outlook.

> **Setup required**: [OAuth2 configuration](../OUTLOOK.md#setup-oauth2)

## Download Attachment

```go
// Get message with attachments
email, err := client.GetMessage(ctx, messageID)
if err != nil {
    log.Fatal(err)
}

// Download each attachment
for _, att := range email.Attachments {
    fmt.Printf("Downloading: %s (%d bytes)\n", att.Filename, att.Size)

    attachment, err := client.GetAttachment(ctx, messageID, att.ID)
    if err != nil {
        log.Printf("Failed: %v\n", err)
        continue
    }

    // Save to file
    err = os.WriteFile(att.Filename, attachment.Data, 0644)
    if err != nil {
        log.Printf("Save failed: %v\n", err)
    }
}
```

## Attachment Metadata

When listing messages, attachments contain metadata only (not data):

```go
response, err := client.ListMessages(ctx, &core.ListOptions{
    MaxResults: 10,
})

for _, email := range response.Emails {
    for _, att := range email.Attachments {
        // Available: ID, Filename, MimeType, Size
        fmt.Printf("Attachment: %s (%s)\n", att.Filename, att.MimeType)

        // NOT available yet: Data
        // fmt.Println(att.Data) // Empty!
    }
}
```

**Lazy Loading**: Attachment data is only downloaded when explicitly requested via `GetAttachment()`.

## Filter Messages with Attachments

```go
// Using Microsoft Graph query syntax
response, err := client.ListMessages(ctx, &core.ListOptions{
    MaxResults: 10,
    Query:      "hasAttachments:true",
})
```

## Download All Attachments from Message

```go
func downloadAllAttachments(client *outlook.Client, ctx context.Context, messageID string, dir string) error {
    // Get message
    email, err := client.GetMessage(ctx, messageID)
    if err != nil {
        return err
    }

    // Create directory
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }

    // Download each
    for _, att := range email.Attachments {
        attachment, err := client.GetAttachment(ctx, messageID, att.ID)
        if err != nil {
            log.Printf("Failed to download %s: %v\n", att.Filename, err)
            continue
        }

        path := filepath.Join(dir, att.Filename)
        if err := os.WriteFile(path, attachment.Data, 0644); err != nil {
            log.Printf("Failed to save %s: %v\n", att.Filename, err)
            continue
        }

        fmt.Printf("Downloaded: %s\n", path)
    }

    return nil
}
```

## Complete Examples

- **Download attachments**: [`examples/outlook`](../../../examples/outlook/)

## Related

- [Messages](./messages.md) - List and read messages
- [Search](./search.md) - Find messages with attachments
