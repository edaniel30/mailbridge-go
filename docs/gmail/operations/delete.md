# Delete Operations

Move messages to trash or permanently delete them.

> **Setup required**: [OAuth2 configuration](../gmail/GMAIL.md#setup-oauth2)

## Trash vs Delete

- **Trash**: Reversible (moves to trash folder)
- **Delete**: Permanent (cannot be recovered)

## Trash Messages

### Single Message

```go
err := client.TrashMessage(ctx, messageID)
// Message moved to trash (reversible)
```

### Restore from Trash

```go
err := client.UntrashMessage(ctx, messageID)
// Message restored to original location
```

### Batch Trash

```go
messageIDs := []string{"msg1", "msg2", "msg3"}
err := client.BatchTrashMessages(ctx, messageIDs)
// All moved to trash in one operation
```

## Permanent Delete

⚠️ **Warning**: Cannot be undone!

### Single Message

```go
err := client.DeleteMessage(ctx, messageID)
// Permanently deleted
```

### Batch Delete

```go
messageIDs := []string{"msg1", "msg2", "msg3"}
err := client.BatchDeleteMessages(ctx, messageIDs)
// All permanently deleted
```

## Batch Operations

Efficiently process multiple messages.

### Batch Mark as Read

```go
messageIDs := []string{"msg1", "msg2", "msg3"}
err := client.BatchMarkAsRead(ctx, messageIDs)
```

### Batch Mark as Unread

```go
err := client.BatchMarkAsUnread(ctx, messageIDs)
```

### Batch Move to Folder

```go
err := client.BatchMoveToFolder(ctx, messageIDs, "Work")
// Moves all messages to "Work" label/folder
```

### Batch Modify Labels

```go
err := client.BatchModifyMessages(ctx, &core.BatchModifyRequest{
    MessageIDs:     []string{"msg1", "msg2"},
    AddLabelIDs:    []string{"IMPORTANT"},
    RemoveLabelIDs: []string{"UNREAD"},
})
```

## Example Workflow

### Clean up old emails

```go
// 1. Find old emails
query := gmail.NewQueryBuilder().
    OlderThan("1y").
    InLabel("Promotions").
    Build()

messages, _ := client.ListMessages(ctx, &core.ListOptions{
    Query: query,
})

// 2. Extract IDs
var ids []string
for _, email := range messages.Emails {
    ids = append(ids, email.ID)
}

// 3. Batch trash
if len(ids) > 0 {
    client.BatchTrashMessages(ctx, ids)
    fmt.Printf("Trashed %d old emails\n", len(ids))
}
```

### Archive read emails

```go
// 1. Find read emails in inbox
query := gmail.NewQueryBuilder().
    IsRead().
    InInbox().
    Build()

messages, _ := client.ListMessages(ctx, &core.ListOptions{Query: query})

// 2. Remove INBOX label (archives them)
var ids []string
for _, email := range messages.Emails {
    ids = append(ids, email.ID)
}

if len(ids) > 0 {
    client.BatchModifyMessages(ctx, &core.BatchModifyRequest{
        MessageIDs:     ids,
        RemoveLabelIDs: []string{"INBOX"},
    })
}
```

## Complete Example

See [`examples/gmail-delete`](../../examples/gmail-delete/) for interactive demos.

```bash
cd examples/gmail-delete
go run main.go
```

## Best Practices

1. **Test with trash first**: Use `TrashMessage()` before `DeleteMessage()`
2. **Confirm before batch delete**: Permanent operations need confirmation
3. **Handle errors**: Some messages may fail in batch operations
4. **Use batch for efficiency**: Better than individual operations in loops

## Single vs Batch

```go
// ❌ Slow (N API calls)
for _, id := range messageIDs {
    client.TrashMessage(ctx, id)
}

// ✅ Fast (1 API call)
client.BatchTrashMessages(ctx, messageIDs)
```

## Related

- [Messages](./messages.md) - List messages to delete
- [Search](./search.md) - Find messages to clean up
- [Notifications](./notifications.md) - Track deletions
