# Search Operations

Advanced message search using Microsoft Graph query syntax.

> **Setup required**: [OAuth2 configuration](../OUTLOOK.md#setup-oauth2)

## Basic Search

```go
response, err := client.ListMessages(ctx, &core.ListOptions{
    MaxResults: 10,
    Query:      "from:john@example.com",
})
```

## Search Operators

Microsoft Graph supports various search operators:

### From/To/Subject

```go
// From specific sender
Query: "from:john@example.com"

// To specific recipient
Query: "to:alice@example.com"

// Subject contains keyword
Query: "subject:invoice"
```

### Multiple Criteria

```go
// Combine multiple filters
Query: "from:john@example.com subject:invoice"
```

### Attachments

```go
// Messages with attachments
Query: "hasAttachments:true"

// Messages without attachments
Query: "hasAttachments:false"
```

### Importance

```go
// High importance messages
Query: "importance:high"

// Low importance messages
Query: "importance:low"
```

### Read Status

```go
// Unread messages
Query: "isRead:false"

// Read messages
Query: "isRead:true"
```

## Advanced Search Examples

### Recent Messages from Specific Sender

```go
response, err := client.ListMessages(ctx, &core.ListOptions{
    MaxResults: 20,
    Query:      "from:boss@company.com",
})
```

### Unread Messages with Attachments

```go
response, err := client.ListMessages(ctx, &core.ListOptions{
    MaxResults: 10,
    Query:      "isRead:false hasAttachments:true",
})
```

### High Priority Messages

```go
response, err := client.ListMessages(ctx, &core.ListOptions{
    MaxResults: 10,
    Query:      "importance:high",
})
```

### Subject Keyword Search

```go
response, err := client.ListMessages(ctx, &core.ListOptions{
    MaxResults: 10,
    Query:      "subject:meeting subject:quarterly",
})
```

## Search in Specific Folder

```go
// Search in Inbox
messages, err := client.ListMessagesInFolder(ctx, outlook.FolderInbox, &core.ListOptions{
    MaxResults: 10,
    Query:      "from:john@example.com",
})

// Search in custom folder
messages, err := client.ListMessagesInFolder(ctx, customFolderID, &core.ListOptions{
    MaxResults: 10,
    Query:      "hasAttachments:true",
})
```

## Pagination with Search

```go
opts := &core.ListOptions{
    MaxResults: 50,
    Query:      "from:reports@company.com",
}

allEmails := []*core.Email{}

for {
    response, err := client.ListMessages(ctx, opts)
    if err != nil {
        log.Fatal(err)
    }

    allEmails = append(allEmails, response.Emails...)

    if response.NextPageToken == "" {
        break
    }
    opts.PageToken = response.NextPageToken
}

fmt.Printf("Found %d messages\n", len(allEmails))
```

## Query Syntax Reference

| Operator | Example | Description |
|----------|---------|-------------|
| `from:` | `from:user@example.com` | Filter by sender |
| `to:` | `to:recipient@example.com` | Filter by recipient |
| `subject:` | `subject:invoice` | Search in subject |
| `hasAttachments:` | `hasAttachments:true` | Has attachments |
| `importance:` | `importance:high` | Filter by importance |
| `isRead:` | `isRead:false` | Filter by read status |

**Multiple keywords**: Separate with spaces for AND logic

## Complete Examples

- **Search examples**: [`examples/outlook`](../../../examples/outlook/)

## Related

- [Messages](./messages.md) - List and read messages
- [Folders](./folders.md) - Search within folders
- [Microsoft Graph Search Documentation](https://learn.microsoft.com/graph/search-query-parameter)
