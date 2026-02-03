# Folder Operations

Manage mail folders (Outlook's equivalent to Gmail labels).

> **Setup required**: [OAuth2 configuration](../OUTLOOK.md#setup-oauth2)

## List All Folders

```go
folders, err := client.ListFolders(ctx)
if err != nil {
    log.Fatal(err)
}

for _, folder := range folders {
    fmt.Printf("%s (ID: %s)\n", folder.Name, folder.ID)
    fmt.Printf("  Type: %s\n", folder.Type)
    fmt.Printf("  Total: %d, Unread: %d\n",
        folder.TotalMessages, folder.UnreadMessages)
}
```

## Create Folder

```go
folder, err := client.CreateFolder(ctx, "My Custom Folder")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Created: %s (ID: %s)\n", folder.Name, folder.ID)
```

## Update Folder (Rename)

```go
folder, err := client.UpdateFolder(ctx, folderID, "New Folder Name")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Renamed to: %s\n", folder.Name)
```

## Delete Folder

```go
err := client.DeleteFolder(ctx, folderID)
if err != nil {
    log.Fatal(err)
}
```

**Warning**: Deleting a folder also deletes all messages in it.

## List Messages in Folder

```go
// Using well-known folder ID
messages, err := client.ListMessagesInFolder(ctx, outlook.FolderInbox, &core.ListOptions{
    MaxResults: 10,
})

// Using custom folder ID
messages, err := client.ListMessagesInFolder(ctx, customFolderID, &core.ListOptions{
    MaxResults: 10,
})
```

## Move Message to Folder

```go
// Move to archive
err := client.MoveMessage(ctx, messageID, outlook.FolderArchive)

// Move to custom folder
err := client.MoveMessage(ctx, messageID, customFolderID)

// Move to inbox
err := client.MoveMessage(ctx, messageID, outlook.FolderInbox)
```

## Well-Known Folder IDs

Use these constants for common folders:

```go
outlook.FolderInbox        // "inbox"
outlook.FolderDrafts       // "drafts"
outlook.FolderSentItems    // "sentitems"
outlook.FolderDeletedItems // "deleteditems"
outlook.FolderJunkEmail    // "junkemail"
outlook.FolderOutbox       // "outbox"
outlook.FolderArchive      // "archive"
```

## Folder vs Labels (Gmail)

| Feature | Outlook Folders | Gmail Labels |
|---------|----------------|--------------|
| Per message | Single folder | Multiple labels |
| Move operation | Changes folder | Adds/removes labels |
| Organization | Hierarchical | Flat with nesting |
| Well-known | "inbox", "drafts" | "INBOX", "DRAFT" |

**Key difference**: In Outlook, a message can only be in one folder at a time. In Gmail, a message can have multiple labels.

## Complete Examples

### Create Archive System

```go
// Create archive folders by year
currentYear := time.Now().Year()
for i := 0; i < 3; i++ {
    year := currentYear - i
    folderName := fmt.Sprintf("Archive %d", year)

    folder, err := client.CreateFolder(ctx, folderName)
    if err != nil {
        log.Printf("Failed to create %s: %v\n", folderName, err)
        continue
    }

    fmt.Printf("Created: %s (ID: %s)\n", folder.Name, folder.ID)
}
```

### Move Old Messages to Archive

```go
// Get archive folder ID
folders, _ := client.ListFolders(ctx)
var archiveFolderID string
for _, f := range folders {
    if f.Name == "Archive 2024" {
        archiveFolderID = f.ID
        break
    }
}

// Move old messages
cutoffDate := time.Now().AddDate(0, -6, 0) // 6 months ago
response, _ := client.ListMessages(ctx, &core.ListOptions{
    MaxResults: 100,
})

for _, email := range response.Emails {
    if email.Date.Before(cutoffDate) {
        err := client.MoveMessage(ctx, email.ID, archiveFolderID)
        if err != nil {
            log.Printf("Failed to move %s: %v\n", email.ID, err)
        }
    }
}
```

### List Unread Count Per Folder

```go
folders, err := client.ListFolders(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Println("Unread messages per folder:")
for _, folder := range folders {
    if folder.UnreadMessages > 0 {
        fmt.Printf("  %s: %d unread\n", folder.Name, folder.UnreadMessages)
    }
}
```

### Organize by Sender

```go
// Create folder
folder, _ := client.CreateFolder(ctx, "Reports")

// Move all messages from specific sender
response, _ := client.ListMessages(ctx, &core.ListOptions{
    MaxResults: 100,
    Query:      "from:reports@company.com",
})

for _, email := range response.Emails {
    err := client.MoveMessage(ctx, email.ID, folder.ID)
    if err != nil {
        log.Printf("Failed to move: %v\n", err)
    }
}
```

## Best Practices

### 1. Cache Folder IDs

```go
// Fetch once
folders, _ := client.ListFolders(ctx)
folderMap := make(map[string]string)
for _, f := range folders {
    folderMap[f.Name] = f.ID
}

// Reuse
archiveID := folderMap["Archive"]
client.MoveMessage(ctx, messageID, archiveID)
```

### 2. Check Before Delete

```go
folder, err := client.GetFolder(ctx, folderID)
if err != nil {
    log.Fatal(err)
}

if folder.TotalMessages > 0 {
    fmt.Printf("Warning: %s contains %d messages\n", folder.Name, folder.TotalMessages)
    fmt.Print("Delete anyway? (y/n): ")
    var confirm string
    fmt.Scanln(&confirm)
    if confirm != "y" {
        return
    }
}

client.DeleteFolder(ctx, folderID)
```

### 3. Hierarchical Naming

```go
// Use prefixes for organization
folders := []string{
    "Projects/ClientA",
    "Projects/ClientB",
    "Archive/2024",
    "Archive/2023",
}

for _, name := range folders {
    client.CreateFolder(ctx, name)
}
```

## Related Documentation

- **Microsoft Graph**: [`examples/outlook`](../../../examples/outlook/)

## Related

- [Messages](./messages.md) - Work with messages
- [Search](./search.md) - Search within folders
- [Delete](./delete.md) - Deleted Items folder
