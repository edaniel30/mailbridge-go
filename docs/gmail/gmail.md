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
    "github.com/danielrivera/mailbridge-go/core"
    "github.com/danielrivera/mailbridge-go/gmail"
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

// 5. Use it
messages, _ := client.ListMessages(ctx, &core.ListOptions{
    MaxResults: 10,
    Query:      "is:unread",
})
```


## Available Operations

### 📨 Message Operations

| Operation | Method | Description |
|-----------|--------|-------------|
| **List Messages** | `ListMessages(ctx, opts)` | List/search emails with filters |
| **Get Message** | `GetMessage(ctx, messageID)` | Get full email details |
| **Get Attachment** | `GetAttachment(ctx, messageID, attachmentID)` | Download attachment data |
| **Send Message** | `SendMessage(ctx, draft, opts)` | Send email (text/HTML/attachments) |
| **Mark as Read** | `MarkAsRead(ctx, messageID)` | Mark email as read |
| **Mark as Unread** | `MarkAsUnread(ctx, messageID)` | Mark email as unread |
| **Move to Folder** | `MoveMessageToFolder(ctx, messageID, folder)` | Move email to folder (creates if needed) |

### 🏷️ Label Operations

| Operation | Method | Description |
|-----------|--------|-------------|
| **List Labels** | `ListLabels(ctx)` | Get all labels/folders |
| **Get Label** | `GetLabel(ctx, labelID)` | Get label details |
| **Find Label** | `FindLabelByName(ctx, name)` | Find label by name |
| **Create Label** | `CreateLabel(ctx, name)` | Create new label/folder |
| **Delete Label** | `DeleteLabel(ctx, labelID)` | Delete label |
| **Add Label** | `AddLabelToMessage(ctx, messageID, labelID)` | Add label to message |
| **Remove Label** | `RemoveLabelFromMessage(ctx, messageID, labelID)` | Remove label from message |

### 🔐 Authentication Operations

| Operation | Method | Description |
|-----------|--------|-------------|
| **Get Auth URL** | `GetAuthURL(state)` | Get OAuth2 authorization URL |
| **Exchange Code** | `ExchangeCode(ctx, code)` | Exchange auth code for token |
| **Connect** | `ConnectWithToken(ctx, token)` | Connect using saved token |
| **Refresh Token** | `RefreshToken(ctx)` | Refresh expired token |
| **Get Token** | `GetToken()` | Get current token |


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
- **[Messages](./operations/messages.md)** - List, read, and send emails
- **[Attachments](./operations/attachments.md)** - Download files from emails
- **[Sending](./operations/sending.md)** - Send emails with HTML/attachments
- **[Search](./operations/search.md)** - Advanced queries with QueryBuilder
- **[Delete](./operations/delete.md)** - Trash and permanently delete messages

### Advanced Features
- **[Push Notifications](../operations/notifications.md)** - Real-time mailbox monitoring with Pub/Sub

## Usage Examples

See [examples/gmail/](../../examples/gmail/) for a complete, runnable example demonstrating all features.

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

**What it demonstrates:**
- OAuth2 authentication with token persistence
- Listing recent and unread messages
- Advanced search with query builder
- Getting message details
- Managing labels
- Downloading attachments (optional)

### Query Builder Examples

```go
// Search for unread emails from boss with PDF attachments
query := gmail.NewQueryBuilder().
    IsUnread().
    From("boss@company.com").
    Filename("pdf").
    HasAttachment().
    Build()

messages, err := client.ListMessages(ctx, &core.ListOptions{
    Query:      query,
    MaxResults: 10,
})
```

**Available methods:**
- `IsUnread()`, `IsRead()`, `IsStarred()`, `IsImportant()`
- `From(email)`, `To(email)`, `Subject(text)`
- `HasAttachment()`, `Filename(extension)`
- `After(date)`, `Before(date)`
- `InInbox()`, `InSent()`, `InDraft()`
- `Category(name)`, `LargerThan(size)`, `SmallerThan(size)`
- `NOT()`, `OR()`

### Common Search Queries

Use in `ListOptions.Query`:

```go
"is:unread"                          // Unread only
"is:starred"                         // Starred
"from:boss@company.com"              // From sender
"subject:invoice"                    // Subject contains
"has:attachment"                     // Has attachments
"after:2024/01/01"                   // After date
"is:unread from:boss@company.com"    // Combined
"in:inbox -label:processed"          // Exclude label
```


## Rate Limits & Best Practices

**Gmail API Quotas:**
- Queries per day: **1 billion**
- Queries per user/sec: **250**

**Best Practices:**
- ✅ Batch operations when possible
- ✅ Cache label IDs
- ✅ Use pagination for large datasets
- ✅ Implement exponential backoff on errors

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
- 💻 [Working Examples](../examples/)
