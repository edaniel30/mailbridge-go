# Gmail Attachments Download Example

This example demonstrates how to search for emails with attachments and download them using MailBridge.

## What This Example Does

1. Connects to Gmail using OAuth2
2. Searches for messages with attachments using `has:attachment` query
3. Lists all attachments with their metadata (filename, type, size)
4. Downloads each attachment to disk
5. Organizes downloads by message ID in subdirectories
6. Shows a summary with total attachments and bytes downloaded

## Prerequisites

- Gmail API credentials (Client ID and Client Secret)
- See [Gmail Setup Guide](../../docs/GMAIL.md#setup-oauth2-credentials) for instructions

## Setup

1. Set your Gmail credentials:

```bash
export GMAIL_CLIENT_ID="your-client-id.apps.googleusercontent.com"
export GMAIL_CLIENT_SECRET="your-client-secret"
```

2. Run the example:

```bash
go run main.go
```

3. On first run:
   - Visit the authorization URL displayed
   - Sign in and authorize the app
   - Copy the authorization code from the redirect URL
   - Paste it into the terminal
   - Token will be saved to `token.json` for future runs

## Output

The example will:
- Create an `attachments/` directory
- Download attachments organized by message ID:
  ```
  attachments/
  ├── msg-abc123/
  │   ├── document.pdf
  │   └── image.png
  └── msg-def456/
      └── invoice.xlsx
  ```

## Example Output

```
✓ Successfully connected to Gmail!

🔍 Searching for messages with attachments...
Found 3 messages with attachments

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📧 Message 1/3
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
From:    John Doe <john@example.com>
Subject: Invoice for January
Date:    2025-01-20 14:30:45
Attachments: 2

  [1/2] invoice_january.pdf
        Type: application/pdf
        Size: 245.3 KB
        ✓ Downloaded: attachments/msg-abc123/invoice_january.pdf
        Actual size: 245.3 KB

  [2/2] receipt.png
        Type: image/png
        Size: 89.5 KB
        ✓ Downloaded: attachments/msg-abc123/receipt.png
        Actual size: 89.5 KB

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📊 Download Summary
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Total attachments downloaded: 5
Total size: 1.2 MB
Download directory: attachments
```

## Customization

You can modify the search query and behavior:

```go
// Change the number of messages to scan
const maxMessagesToScan = 20

// Modify the search query
listOpts := &core.ListOptions{
    MaxResults: maxMessagesToScan,
    Query:      "has:attachment from:finance@company.com", // Only from specific sender
    // Query:   "has:attachment after:2024/01/01",         // After specific date
    // Query:   "has:attachment subject:invoice",          // Subject contains keyword
}
```

## Cleanup

To reset and re-authenticate:

```bash
rm token.json
```

To clear downloaded attachments:

```bash
rm -rf attachments/
```

## Learn More

- [Gmail Integration Guide](../../docs/GMAIL.md)
- [Gmail Search Operators](https://support.google.com/mail/answer/7190)
- [Core API Documentation](../../core/)
