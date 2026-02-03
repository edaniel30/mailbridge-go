# Message Operations

Basic operations for listing, reading, and sending Gmail messages.

> **Setup required**: [OAuth2 configuration](../gmail/GMAIL.md#setup-oauth2)

## List Messages

```go
messages, err := client.ListMessages(ctx, &core.ListOptions{
    MaxResults: 10,
    Query:      "is:unread",  // Gmail search query
    LabelIDs:   []string{"INBOX"},
})

for _, email := range messages.Emails {
    fmt.Printf("From: %s\n", email.From.Email)
    fmt.Printf("Subject: %s\n", email.Subject)
    fmt.Printf("Snippet: %s\n", email.Snippet)
}
```

**Options**:
- `MaxResults`: Number of messages (default: 100)
- `Query`: Gmail search syntax (e.g., `"from:user@example.com"`)
- `LabelIDs`: Filter by labels
- `PageToken`: For pagination

## Get Message Details

```go
email, err := client.GetMessage(ctx, messageID)

fmt.Printf("Subject: %s\n", email.Subject)
fmt.Printf("From: %s <%s>\n", email.From.Name, email.From.Email)
fmt.Printf("Date: %s\n", email.Date)
fmt.Printf("Body (HTML): %s\n", email.Body.HTML)
fmt.Printf("Body (Plain): %s\n", email.Body.Plain)

// Attachments metadata (not downloaded yet)
for _, att := range email.Attachments {
    fmt.Printf("Attachment: %s (%d bytes)\n", att.Filename, att.Size)
}
```

## Send Message

```go
response, err := client.SendMessage(ctx, &core.Draft{
    To:      []core.EmailAddress{{Email: "user@example.com"}},
    Subject: "Hello",
    Body: &core.EmailBody{
        Plain: "This is a plain text email",
        HTML:  "<h1>Or send HTML</h1>",
    },
}, nil)

fmt.Printf("Sent! Message ID: %s\n", response.ID)
```

**With multiple recipients**:
```go
draft := &core.Draft{
    To: []core.EmailAddress{
        {Name: "Alice", Email: "alice@example.com"},
        {Name: "Bob", Email: "bob@example.com"},
    },
    Cc:      []core.EmailAddress{{Email: "manager@example.com"}},
    Subject: "Team Update",
    Body:    &core.EmailBody{Plain: "Meeting at 3pm"},
}
```

## Complete Examples

- **Basic usage**: [`examples/gmail`](../../examples/gmail/)
- **Sending emails**: [`examples/gmail-send`](../../examples/gmail-send/)
- **Download attachments**: [`examples/gmail-attachments`](../../examples/gmail-attachments/)

## Related

- [Attachments](./attachments.md) - Download files
- [Search](./search.md) - Advanced queries
- [Delete](./delete.md) - Remove messages
