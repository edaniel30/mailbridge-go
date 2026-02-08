# Download Attachments

Download files attached to Gmail messages.

> **Setup required**: [OAuth2 configuration](../gmail.md#setup-oauth2)

## Quick Example

```go
// 1. Get message details (you need the message ID)
email, err := client.GetMessage(ctx, messageID)
if err != nil {
    log.Fatal(err)
}

// 2. Download each attachment
for _, att := range email.Attachments {
    data, err := client.GetAttachment(ctx, email.ID, att.ID)
    if err != nil {
        log.Printf("Failed to download %s: %v", att.Filename, err)
        continue
    }

    // Save to file
    err = os.WriteFile(att.Filename, data, 0644)
    if err != nil {
        log.Printf("Failed to save %s: %v", att.Filename, err)
        continue
    }

    fmt.Printf("Downloaded: %s (%d bytes)\n", att.Filename, len(data))
}
```

## Attachment Metadata

Get attachment information from message without downloading:

```go
email, err := client.GetMessage(ctx, messageID)
if err != nil {
    log.Fatal(err)
}

// Check if message has attachments
if len(email.Attachments) == 0 {
    fmt.Println("No attachments")
    return
}

// View attachment metadata
for _, att := range email.Attachments {
    fmt.Printf("ID: %s\n", att.ID)              // For GetAttachment()
    fmt.Printf("Filename: %s\n", att.Filename)  // Original name
    fmt.Printf("MimeType: %s\n", att.MimeType)  // e.g., "image/png"
    fmt.Printf("Size: %d bytes\n", att.Size)    // File size
}
```

## Download Specific Attachment

```go
// Download attachment by ID
data, err := client.GetAttachment(ctx, messageID, attachmentID)
if err != nil {
    log.Fatal(err)
}

// data is []byte - process directly or save to file
err = os.WriteFile("attachment.pdf", data, 0644)
```

## Filter by MIME Type

```go
email, err := client.GetMessage(ctx, messageID)
if err != nil {
    log.Fatal(err)
}

// Download only PDFs
for _, att := range email.Attachments {
    if att.MimeType == "application/pdf" {
        data, err := client.GetAttachment(ctx, email.ID, att.ID)
        if err != nil {
            continue
        }
        os.WriteFile(att.Filename, data, 0644)
    }
}

// Download only images
for _, att := range email.Attachments {
    if strings.HasPrefix(att.MimeType, "image/") {
        data, err := client.GetAttachment(ctx, email.ID, att.ID)
        if err != nil {
            continue
        }
        os.WriteFile(att.Filename, data, 0644)
    }
}
```

## Example Workflow

```go
import (
    "context"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "strings"
)

func downloadAttachments(ctx context.Context, client *gmail.Client, messageID string) {
    // Get message
    email, err := client.GetMessage(ctx, messageID)
    if err != nil {
        log.Fatal(err)
    }

    // Check for attachments
    if len(email.Attachments) == 0 {
        fmt.Println("No attachments found")
        return
    }

    // Create download directory
    downloadDir := "attachments"
    os.MkdirAll(downloadDir, 0755)

    // Download each attachment
    for i, att := range email.Attachments {
        fmt.Printf("[%d/%d] Downloading %s...\n", i+1, len(email.Attachments), att.Filename)

        data, err := client.GetAttachment(ctx, email.ID, att.ID)
        if err != nil {
            log.Printf("Failed: %v\n", err)
            continue
        }

        // Save to file
        filePath := filepath.Join(downloadDir, att.Filename)
        err = os.WriteFile(filePath, data, 0644)
        if err != nil {
            log.Printf("Failed to save: %v\n", err)
            continue
        }

        fmt.Printf("Saved to %s (%d bytes)\n", filePath, len(data))
    }
}
```

## API Reference

### GetAttachment

```go
func (c *Client) GetAttachment(ctx context.Context, messageID, attachmentID string) ([]byte, error)
```

**Parameters**:
- `messageID` - The message ID containing the attachment
- `attachmentID` - The attachment ID from `email.Attachments[].ID`

**Returns**:
- `[]byte` - Raw attachment data
- `error` - Error if download fails

## Related Operations

- [Messages](./messages.md) - Get message details to find attachments
