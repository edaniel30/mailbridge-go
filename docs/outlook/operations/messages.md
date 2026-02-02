# Message Operations

Basic operations for listing, reading, and managing Outlook messages.

> **Setup required**: [OAuth2 configuration](../OUTLOOK.md#setup-oauth2)

## List Messages

```go
response, err := client.ListMessages(ctx, &core.ListOptions{
    MaxResults: 10,
})

for _, email := range response.Emails {
    fmt.Printf("From: %s\n", email.From.Email)
    fmt.Printf("Subject: %s\n", email.Subject)
    fmt.Printf("Snippet: %s\n", email.Snippet)
}
```

**Options**:
- `MaxResults`: Number of messages (default: 100)
- `Query`: Microsoft Graph search syntax (e.g., `"from:user@example.com"`)
- `Labels`: Filter by folder IDs
- `PageToken`: For pagination

## Pagination

```go
opts := &core.ListOptions{MaxResults: 10}

for {
    response, err := client.ListMessages(ctx, opts)
    if err != nil {
        log.Fatal(err)
    }

    // Process messages
    for _, email := range response.Emails {
        fmt.Println(email.Subject)
    }

    // Next page
    if response.NextPageToken == "" {
        break
    }
    opts.PageToken = response.NextPageToken
}
```

## Get Message Details

```go
email, err := client.GetMessage(ctx, messageID)

fmt.Printf("Subject: %s\n", email.Subject)
fmt.Printf("From: %s <%s>\n", email.From.Name, email.From.Email)
fmt.Printf("Date: %s\n", email.Date)
fmt.Printf("Body (HTML): %s\n", email.Body.HTML)
fmt.Printf("Body (Text): %s\n", email.Body.Text)

// Attachments metadata (not downloaded yet)
for _, att := range email.Attachments {
    fmt.Printf("Attachment: %s (%d bytes)\n", att.Filename, att.Size)
}
```

## Mark as Read/Unread

```go
// Mark as read
err := client.MarkAsRead(ctx, messageID)

// Mark as unread
err := client.MarkAsUnread(ctx, messageID)
```

## List Messages in Folder

```go
// Using well-known folder
messages, err := client.ListMessagesInFolder(ctx, outlook.FolderInbox, &core.ListOptions{
    MaxResults: 10,
})

// Using custom folder
messages, err := client.ListMessagesInFolder(ctx, customFolderID, &core.ListOptions{
    MaxResults: 10,
})
```

## Complete Examples

- **Basic usage**: [`examples/outlook`](../../../examples/outlook/)

## Related

- [Attachments](./attachments.md) - Download files
- [Search](./search.md) - Advanced queries
- [Delete](./delete.md) - Remove messages
- [Folders](./folders.md) - Manage folders
