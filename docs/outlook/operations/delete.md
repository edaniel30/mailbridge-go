# Delete Operations

Delete messages and manage the trash folder in Outlook.

> **Setup required**: [OAuth2 configuration](../OUTLOOK.md#setup-oauth2)

## Delete Single Message

```go
err := client.DeleteMessage(ctx, messageID)
if err != nil {
    log.Fatal(err)
}
```

**Note**: Deleted messages are moved to the "Deleted Items" folder, not permanently deleted.

## Delete Multiple Messages

```go
messageIDs := []string{"msg1", "msg2", "msg3"}

for _, id := range messageIDs {
    err := client.DeleteMessage(ctx, id)
    if err != nil {
        log.Printf("Failed to delete %s: %v\n", id, err)
    }
}
```

## Move to Deleted Items Folder

Alternative approach using move operation:

```go
err := client.MoveMessage(ctx, messageID, outlook.FolderDeletedItems)
```

This is equivalent to `DeleteMessage()`.

## Delete with Search

```go
// Find messages to delete
response, err := client.ListMessages(ctx, &core.ListOptions{
    MaxResults: 100,
    Query:      "from:spam@example.com",
})

// Delete each
for _, email := range response.Emails {
    err := client.DeleteMessage(ctx, email.ID)
    if err != nil {
        log.Printf("Failed to delete: %v\n", err)
    }
}
```

## Batch Delete with Rate Limiting

```go
import "time"

func deleteMessages(client *outlook.Client, ctx context.Context, messageIDs []string) error {
    for i, id := range messageIDs {
        err := client.DeleteMessage(ctx, id)
        if err != nil {
            log.Printf("Failed to delete %s: %v\n", id, err)
            continue
        }

        fmt.Printf("Deleted %d/%d\n", i+1, len(messageIDs))

        // Avoid rate limits
        if i < len(messageIDs)-1 {
            time.Sleep(100 * time.Millisecond)
        }
    }
    return nil
}
```

## Delete Old Messages

```go
import "time"

// Delete messages older than 30 days
cutoffDate := time.Now().AddDate(0, 0, -30)

response, err := client.ListMessages(ctx, &core.ListOptions{
    MaxResults: 100,
})

for _, email := range response.Emails {
    if email.Date.Before(cutoffDate) {
        err := client.DeleteMessage(ctx, email.ID)
        if err != nil {
            log.Printf("Failed to delete %s: %v\n", email.ID, err)
        } else {
            fmt.Printf("Deleted: %s\n", email.Subject)
        }
    }
}
```

## Empty Deleted Items Folder

To permanently delete messages, you need to delete them from the Deleted Items folder:

```go
// 1. List messages in Deleted Items
response, err := client.ListMessagesInFolder(ctx, outlook.FolderDeletedItems, &core.ListOptions{
    MaxResults: 100,
})

// 2. Delete each (permanently)
for _, email := range response.Emails {
    err := client.DeleteMessage(ctx, email.ID)
    if err != nil {
        log.Printf("Failed to permanently delete: %v\n", err)
    }
}
```

**Warning**: Messages deleted from "Deleted Items" are permanently removed.

## Best Practices

### 1. Confirm Before Deletion

```go
fmt.Printf("Delete %d messages? (y/n): ", len(messageIDs))
var confirm string
fmt.Scanln(&confirm)

if confirm == "y" {
    // Proceed with deletion
}
```

### 2. Rate Limiting

```go
// Add delays for bulk operations
for _, id := range messageIDs {
    client.DeleteMessage(ctx, id)
    time.Sleep(100 * time.Millisecond)
}
```

### 3. Error Handling

```go
errors := []string{}

for _, id := range messageIDs {
    err := client.DeleteMessage(ctx, id)
    if err != nil {
        errors = append(errors, fmt.Sprintf("%s: %v", id, err))
    }
}

if len(errors) > 0 {
    log.Printf("Failed to delete %d messages:\n", len(errors))
    for _, e := range errors {
        log.Println(e)
    }
}
```

## Comparison with Gmail

| Feature | Outlook | Gmail |
|---------|---------|-------|
| Delete behavior | Moves to Deleted Items | Moves to Trash |
| Permanent delete | Delete from Deleted Items | Delete from Trash |
| Undelete | Move from Deleted Items | Move from Trash |
| Well-known ID | `outlook.FolderDeletedItems` | `gmail.LabelTrash` |

## Complete Examples

- **Delete operations**: [`examples/outlook`](../../../examples/outlook/)

## Related

- [Messages](./messages.md) - List messages
- [Search](./search.md) - Find messages to delete
- [Folders](./folders.md) - Manage Deleted Items folder
