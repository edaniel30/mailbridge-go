# Gmail Integration Guide

Complete guide for using MailBridge with Gmail.

## Table of Contents

- [Quick Start](#quick-start)
- [Setup OAuth2](#setup-oauth2)
- [Available Operations](#available-operations)
- [Operation Guides](#operation-guides)
- [Usage Examples](#usage-examples)
- [Resources](#resources)


## Quick Start

```go
import (
    "github.com/edaniel30/mailbridge-go/gmail"
)

// 1. Configure
cfg := &gmail.Config{
    ClientID:     os.Getenv("GMAIL_CLIENT_ID"),
    ClientSecret: os.Getenv("GMAIL_CLIENT_SECRET"),
    RedirectURL:  "http://localhost",
    Scopes:       gmail.DefaultScopes(),
}

// 2. Create client
client, _ := gmail.New(cfg)
defer client.Close()

// 3. OAuth flow (first time)
authURL := client.GetAuthURL("state")
// User authorizes → copy code
token, _ := client.ExchangeCode(ctx, code)

// 4. Connect
client.ConnectWithToken(ctx, token)

// 5. Get a message
email, _ := client.GetMessage(ctx, messageID)
fmt.Printf("Subject: %s\n", email.Subject)
```


## Available Operations

### 📨 Message Operations

| Operation | Method | Description |
|-----------|--------|-------------|
| **Get Message** | `GetMessage(ctx, messageID)` | Get full email details |
| **Get Attachment** | `GetAttachment(ctx, messageID, attachmentID)` | Download attachment data |
| **Mark as Read** | `MarkAsRead(ctx, messageID)` | Mark email as read |
| **Move to Folder** | `MoveMessageToFolder(ctx, messageID, folder)` | Move email to folder (creates label if needed) |

### 🏷️ Label Operations

| Operation | Method | Description |
|-----------|--------|-------------|
| **Create Label** | `CreateLabel(ctx, name)` | Create new label/folder |

### 🔔 Watch Operations

| Operation | Method | Description |
|-----------|--------|-------------|
| **Watch Mailbox** | `WatchMailbox(ctx, req)` | Set up push notifications |
| **Stop Watch** | `StopWatch(ctx, userID)` | Stop push notifications |
| **Get History** | `GetHistory(ctx, req)` | Get mailbox changes |

### 🔐 Authentication Operations

| Operation | Method | Description |
|-----------|--------|-------------|
| **Get Auth URL** | `GetAuthURL(state)` | Get OAuth2 authorization URL |
| **Exchange Code** | `ExchangeCode(ctx, code)` | Exchange auth code for token |
| **Connect** | `ConnectWithToken(ctx, token)` | Connect using saved token |
| **Refresh Token** | `RefreshToken(ctx)` | Refresh expired token |
| **Get Token** | `GetToken()` | Get current token |
| **Set Token** | `SetToken(token)` | Set token without connecting |
| **Revoke Token** | `RevokeToken(ctx, token)` | Revoke OAuth2 token |


## Setup OAuth2

### 1. Create Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com)
2. Click **"NEW PROJECT"** → Name it → **"CREATE"**

### 2. Enable Gmail API

1. **"APIs & Services"** → **"Library"**
2. Search **"Gmail API"** → **"ENABLE"**

### 3. Configure OAuth Consent

1. **"OAuth consent screen"** → **"External"**
2. Fill required fields (app name, emails)
3. Add scopes: `gmail.readonly`, `gmail.send`, `gmail.modify`, `gmail.labels`
4. Add test users (your email)

### 4. Create OAuth Client

1. **"Credentials"** → **"CREATE CREDENTIALS"** → **"OAuth client ID"**
2. Select **"Desktop app"** → Name it → **"CREATE"**
3. Edit client → Add redirect URI: `http://localhost`
4. Save **Client ID** and **Client Secret**

⚠️ **Never commit credentials!** Use environment variables:

```bash
export GMAIL_CLIENT_ID="your-id.apps.googleusercontent.com"
export GMAIL_CLIENT_SECRET="your-secret"
```


## Operation Guides

Detailed guides for specific operations:

### Core Operations
- **[Messages](./operations/messages.md)** - Get message details
- **[Attachments](./operations/attachments.md)** - Download files from emails
- **[Delete](./operations/delete.md)** - Trash and permanently delete messages

### Advanced Features
- **[Push Notifications](./operations/notifications.md)** - Real-time mailbox monitoring with Pub/Sub

## Usage Examples

See [examples/gmail/](../../examples/gmail/) for a complete, runnable example demonstrating OAuth2 authentication.

### Running the Example

```bash
# 1. Set credentials
export GMAIL_CLIENT_ID="your-client-id.apps.googleusercontent.com"
export GMAIL_CLIENT_SECRET="your-client-secret"

# 2. Run example
cd examples/gmail
go run main.go
```

**First Run:** You'll be prompted to authorize via browser. The token is saved to `token.json` for future use.

### Basic Usage Examples

#### Get Message Details
```go
email, err := client.GetMessage(ctx, messageID)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Subject: %s\n", email.Subject)
fmt.Printf("From: %s <%s>\n", email.From.Name, email.From.Email)
fmt.Printf("Body: %s\n", email.Body.Text)
```

#### Download Attachment
```go
data, err := client.GetAttachment(ctx, messageID, attachmentID)
if err != nil {
    log.Fatal(err)
}

err = os.WriteFile("attachment.pdf", data, 0644)
```

#### Mark as Read
```go
err := client.MarkAsRead(ctx, messageID)
```

#### Move to Folder
```go
// Creates label "Work" if it doesn't exist
err := client.MoveMessageToFolder(ctx, messageID, "Work")
```

#### Create Label
```go
label, err := client.CreateLabel(ctx, "Important")
fmt.Printf("Created label: %s (ID: %s)\n", label.Name, label.ID)
```

### Gmail Search Queries

You can use Gmail's native search syntax directly:

```go
// Get message IDs using Gmail's search operators
// Note: Use Gmail API directly for searching, then use GetMessage() for details
```

**Common search operators:**
- `is:unread` - Unread messages
- `is:starred` - Starred messages
- `from:user@example.com` - From specific sender
- `subject:invoice` - Subject contains text
- `has:attachment` - Has attachments
- `after:2024/01/01` - After specific date
- `label:INBOX` - In specific label

See [Gmail Search Operators](https://support.google.com/mail/answer/7190) for complete syntax.


## Rate Limits & Best Practices

**Gmail API Quotas:**
- Queries per day: **1 billion**
- Queries per user/sec: **250**

**Best Practices:**
- ✅ Use pagination for large datasets
- ✅ Implement exponential backoff on errors
- ✅ Cache message data when possible

**Security:**
- ✅ Use environment variables for credentials
- ✅ Never commit `token.json` to Git
- ✅ Request minimum required scopes
- ✅ Store tokens securely (encrypted)
- ✅ Use HTTPS redirect URIs in production


## Resources

- 📖 [Gmail API Documentation](https://developers.google.com/gmail/api)
- 🔐 [OAuth2 for Desktop Apps](https://developers.google.com/identity/protocols/oauth2/native-app)
- 🔍 [Gmail Search Operators](https://support.google.com/mail/answer/7190)
- 📊 [API Quotas & Limits](https://developers.google.com/gmail/api/reference/quota)
- 💻 [Working Examples](../../examples/gmail/)
