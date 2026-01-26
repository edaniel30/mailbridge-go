# Gmail Integration Guide

Complete guide for using MailBridge with Gmail.

## Quick Start

```go
import (
    "github.com/danielrivera/mailbridge-go/core"
    "github.com/danielrivera/mailbridge-go/gmail"
)

// Configure
cfg := &gmail.Config{
    ClientID:     os.Getenv("GMAIL_CLIENT_ID"),
    ClientSecret: os.Getenv("GMAIL_CLIENT_SECRET"),
    RedirectURL:  "http://localhost",
    Scopes:       gmail.DefaultScopes(),
}

// Create client
client, _ := gmail.New(cfg)
defer client.Close()

// OAuth flow
authURL := client.GetAuthURL("state")
// User authorizes...
token, _ := client.ExchangeCode(ctx, code)
client.ConnectWithToken(ctx, token)

// List messages
response, _ := client.ListMessages(ctx, &core.ListOptions{
    MaxResults: 10,
    Query:      "is:unread",
})
```

## Setup OAuth2 Credentials

### 1. Create Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com)
2. Click **"NEW PROJECT"**
3. Name it (e.g., "MailBridge App")
4. Click **"CREATE"**

### 2. Enable Gmail API

1. Go to **"APIs & Services"** → **"Library"**
2. Search for **"Gmail API"**
3. Click **"ENABLE"**

### 3. Configure OAuth Consent Screen

1. Go to **"APIs & Services"** → **"OAuth consent screen"**
2. Select **"External"** user type
3. Fill required fields:
   - App name: Your app name
   - User support email: Your email
   - Developer contact: Your email
4. Add scopes:
   - `gmail.readonly`
   - `gmail.send`
   - `gmail.modify`
   - `gmail.labels`
5. Add test users (your email)

### 4. Create OAuth Client

1. Go to **"APIs & Services"** → **"Credentials"**
2. Click **"CREATE CREDENTIALS"** → **"OAuth client ID"**
3. Select **"Desktop app"**
4. Name it
5. Click **"CREATE"**
6. Edit the client and add redirect URI: `http://localhost`
7. Save your **Client ID** and **Client Secret**

⚠️ **Security**: Never commit credentials to Git. Use environment variables.

## Usage

### Configuration

```go
cfg := &gmail.Config{
    ClientID:     "your-id.apps.googleusercontent.com",
    ClientSecret: "your-secret",
    RedirectURL:  "http://localhost",
    Scopes:       gmail.DefaultScopes(),
}
```

**Default Scopes:**
- `gmail.readonly` - Read emails
- `gmail.send` - Send emails
- `gmail.modify` - Modify emails
- `gmail.labels` - Manage labels

**Custom Scopes:**
```go
Scopes: []string{
    "https://www.googleapis.com/auth/gmail.readonly",
}
```

### OAuth Flow

**First time:**
```go
// 1. Get auth URL
authURL := client.GetAuthURL("random-state")
fmt.Println("Visit:", authURL)

// 2. User authorizes in browser
// 3. Copy code from redirect URL

// 4. Exchange code for token
token, err := client.ExchangeCode(ctx, code)

// 5. Connect
err = client.ConnectWithToken(ctx, token)

// 6. Save token for reuse
saveToken(token) // Your function
```

**Subsequent runs:**
```go
token := loadToken() // Your function
err := client.ConnectWithToken(ctx, token)

// Refresh if expired
if err != nil {
    token, err = client.RefreshToken(ctx)
}
```

### List Messages

```go
opts := &core.ListOptions{
    MaxResults: 20,
    Query:      "is:unread",
    Labels:     []string{"INBOX"},
}

response, err := client.ListMessages(ctx, opts)

for _, email := range response.Emails {
    fmt.Printf("%s: %s\n", email.From.Email, email.Subject)
}

// Pagination
if response.NextPageToken != "" {
    opts.PageToken = response.NextPageToken
    nextPage, _ := client.ListMessages(ctx, opts)
}
```

### Search Queries

Gmail's search syntax in the `Query` field:

```go
Query: "is:unread"                           // Unread only
Query: "is:starred"                          // Starred only
Query: "from:boss@company.com"               // From sender
Query: "subject:invoice"                     // Subject contains
Query: "has:attachment"                      // Has attachments
Query: "after:2024/01/01"                    // After date
Query: "before:2024/12/31"                   // Before date
Query: "is:unread from:boss@company.com"     // Combined
Query: "in:inbox -label:processed"           // Exclude label
```

### Get Message

```go
email, err := client.GetMessage(ctx, "message-id")

fmt.Println("Subject:", email.Subject)
fmt.Println("From:", email.From.Name, email.From.Email)
fmt.Println("Body:", email.Body.Text)
fmt.Println("Attachments:", len(email.Attachments))
```

### Download Attachments

```go
// First, get the message with attachment metadata
email, err := client.GetMessage(ctx, "message-id")

// Then download each attachment
for _, attachment := range email.Attachments {
    data, err := client.GetAttachment(ctx, email.ID, attachment.ID)
    if err != nil {
        log.Printf("Failed to download %s: %v", attachment.Filename, err)
        continue
    }

    // Save to file
    err = os.WriteFile(attachment.Filename, data, 0644)
    if err != nil {
        log.Printf("Failed to save %s: %v", attachment.Filename, err)
        continue
    }

    fmt.Printf("Downloaded: %s (%d bytes)\n", attachment.Filename, len(data))
}
```

### Message Operations

**Mark as read/unread:**
```go
err = client.MarkAsRead(ctx, messageID)
err = client.MarkAsUnread(ctx, messageID)
```

**Move to folder:**
```go
// Creates folder if doesn't exist
err = client.MoveMessageToFolder(ctx, messageID, "Archive")
err = client.MoveMessageToFolder(ctx, messageID, "Projects/2024")
```

### Label Operations

**List labels:**
```go
labels, err := client.ListLabels(ctx)

for _, label := range labels {
    fmt.Printf("%s: %s\n", label.ID, label.Name)
}
```

**Create label:**
```go
label, err := client.CreateLabel(ctx, "Important")
```

**Add/Remove labels:**
```go
err = client.AddLabelToMessage(ctx, messageID, labelID)
err = client.RemoveLabelFromMessage(ctx, messageID, labelID)
```

## Environment Variables

```bash
export GMAIL_CLIENT_ID="310166051818-xxx.apps.googleusercontent.com"
export GMAIL_CLIENT_SECRET="GOCSPX-xxxxxxxxxxxx"
```

**In code:**
```go
cfg := &gmail.Config{
    ClientID:     os.Getenv("GMAIL_CLIENT_ID"),
    ClientSecret: os.Getenv("GMAIL_CLIENT_SECRET"),
    RedirectURL:  "http://localhost",
    Scopes:       gmail.DefaultScopes(),
}
```

## Example Application

See [examples/gmail](../examples/gmail) for a complete working example.

**Run it:**
```bash
# Set credentials
export GMAIL_CLIENT_ID="your-id"
export GMAIL_CLIENT_SECRET="your-secret"

# Run
go run ./examples/gmail
```

**What it does:**
1. OAuth2 authentication
2. Lists unread messages
3. Shows message details
4. Moves message to folder
5. Marks as read

## Authorization Flow Details

### First Authorization

1. App displays URL: `https://accounts.google.com/o/oauth2/auth?...`
2. User opens URL in browser
3. User signs in and authorizes
4. Browser redirects to: `http://localhost/?code=4/0AY0e-xxx&scope=...`
5. Browser shows "Unable to connect" (normal - no server)
6. User copies `code` value from URL
7. Pastes code in terminal
8. App exchanges code for token
9. Token saved to `token.json`

### Subsequent Uses

1. App loads `token.json`
2. Uses token to authenticate
3. If expired, refreshes automatically
4. Updates `token.json`

### Token Structure

```json
{
  "access_token": "ya29.xxx",
  "token_type": "Bearer",
  "refresh_token": "1//xxx",
  "expiry": "2024-01-24T16:00:00Z"
}
```

## Common Patterns

### Batch Processing

```go
opts := &core.ListOptions{MaxResults: 100}

for {
    response, err := client.ListMessages(ctx, opts)
    if err != nil {
        return err
    }

    for _, email := range response.Emails {
        // Process email
        processEmail(email)
    }

    if response.NextPageToken == "" {
        break
    }
    opts.PageToken = response.NextPageToken
}
```

### Auto-organize Inbox

```go
response, _ := client.ListMessages(ctx, &core.ListOptions{
    Query: "is:unread has:attachment from:finance@company.com",
})

for _, email := range response.Emails {
    // Move to folder
    client.MoveMessageToFolder(ctx, email.ID, "Finance/Invoices")

    // Mark as read
    client.MarkAsRead(ctx, email.ID)
}
```

### Check for New Messages

```go
response, _ := client.ListMessages(ctx, &core.ListOptions{
    MaxResults: 10,
    Query:      "is:unread in:inbox",
})

if len(response.Emails) > 0 {
    fmt.Printf("You have %d new messages\n", len(response.Emails))
}
```

## Troubleshooting

### Error: "invalid config: client_id is required"

**Solution:** Set environment variables
```bash
export GMAIL_CLIENT_ID="your-id"
export GMAIL_CLIENT_SECRET="your-secret"
```

### Error: "failed to exchange code"

**Cause:** Authorization code expired (valid ~10 minutes)

**Solution:** Get a new authorization code

### Error: "failed to connect with token"

**Solutions:**
1. Delete `token.json` and re-authorize
2. Check if scopes changed (requires new authorization)
3. Verify token file isn't corrupted

### Error: "insufficient permissions"

**Cause:** Missing required scope

**Solution:** Add scope to config and re-authorize
```go
Scopes: gmail.DefaultScopes(),
```

### Browser shows "This app isn't verified"

**Cause:** App in testing mode

**Solutions:**
1. Click "Advanced" → "Go to App (unsafe)"
2. Only works for test users added in OAuth consent screen
3. For production, submit app for verification

## Rate Limits

Gmail API has quota limits:
- **Queries per day:** 1 billion
- **Queries per user per second:** 250

**Best practices:**
- Batch operations when possible
- Cache label IDs
- Use pagination instead of large queries
- Implement exponential backoff on errors

## Security Best Practices

1. ✅ Use environment variables for credentials
2. ✅ Never commit `token.json` to Git
3. ✅ Request minimum required scopes
4. ✅ Implement token refresh logic
5. ✅ Store tokens securely (encrypted if possible)
6. ✅ Use HTTPS redirect URIs in production
7. ✅ Rotate credentials periodically

## Learn More

- [Gmail API Documentation](https://developers.google.com/gmail/api)
- [OAuth2 for Desktop Apps](https://developers.google.com/identity/protocols/oauth2/native-app)
- [Gmail Search Operators](https://support.google.com/mail/answer/7190)
- [API Quotas](https://developers.google.com/gmail/api/reference/quota)