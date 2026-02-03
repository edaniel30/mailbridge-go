# Outlook Integration Guide

Complete guide for using MailBridge with Microsoft Outlook via Microsoft Graph API.

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
    "github.com/danielrivera/mailbridge-go/outlook"
)

// 1. Configure
cfg := &outlook.Config{
    ClientID:     os.Getenv("OUTLOOK_CLIENT_ID"),
    ClientSecret: os.Getenv("OUTLOOK_CLIENT_SECRET"),
    TenantID:     os.Getenv("OUTLOOK_TENANT_ID"),  // "common", "consumers", or tenant ID
    RedirectURL:  "http://localhost:8080/callback",
}

// 2. Create client
client, _ := outlook.New(cfg)
defer client.Close()

// 3. OAuth flow (first time)
authURL := client.GetAuthURL("state")
// User authorizes → copy code
err := client.ConnectWithAuthCode(ctx, code)

// 4. Use it
messages, _ := client.ListMessages(ctx, &core.ListOptions{
    MaxResults: 10,
    Query:      "isRead:false",
})
```


## Available Operations

### 📨 Message Operations

| Operation | Method | Description |
|-----------|--------|-------------|
| **List Messages** | `ListMessages(ctx, opts)` | List/search emails with filters |
| **Get Message** | `GetMessage(ctx, messageID)` | Get full email details |
| **Get Attachment** | `GetAttachment(ctx, messageID, attachmentID)` | Download attachment data |
| **Mark as Read** | `MarkAsRead(ctx, messageID)` | Mark email as read |
| **Mark as Unread** | `MarkAsUnread(ctx, messageID)` | Mark email as unread |
| **Delete Message** | `DeleteMessage(ctx, messageID)` | Delete email (moves to Deleted Items) |
| **Move Message** | `MoveMessage(ctx, messageID, folderID)` | Move email to folder |

### 📁 Folder Operations

| Operation | Method | Description |
|-----------|--------|-------------|
| **List Folders** | `ListFolders(ctx)` | Get all mail folders |
| **Create Folder** | `CreateFolder(ctx, name)` | Create new folder |
| **Update Folder** | `UpdateFolder(ctx, folderID, newName)` | Rename folder |
| **Delete Folder** | `DeleteFolder(ctx, folderID)` | Delete folder |
| **List Messages in Folder** | `ListMessagesInFolder(ctx, folderID, opts)` | Get messages from specific folder |

### 🔐 Authentication Operations

| Operation | Method | Description |
|-----------|--------|-------------|
| **Get Auth URL** | `GetAuthURL(state)` | Get OAuth2 authorization URL |
| **Connect with Code** | `ConnectWithAuthCode(ctx, code)` | Exchange auth code for token |
| **Connect** | `ConnectWithToken(ctx, token)` | Connect using saved token |
| **Refresh Token** | `RefreshToken(ctx)` | Refresh expired token |
| **Get Token** | `GetToken()` | Get current token |


## Setup OAuth2

### 1. Register Application in Microsoft Entra ID

> **Note**: Microsoft renamed Azure Active Directory to **Microsoft Entra ID** in 2023.

1. Go to [Azure Portal](https://portal.azure.com) or [Entra Portal](https://entra.microsoft.com)
2. Navigate to **Microsoft Entra ID** → **App registrations** → **+ New registration**
3. Configure:
   - **Name**: Your app name (e.g., "MailBridge Outlook")
   - **Supported account types**: Choose based on your needs
   - **Redirect URI**: `http://localhost:8080/callback` (for development)

### 2. Note Credentials

From the app's **Overview** page, save these values:

- **Application (client) ID** → This is your `OUTLOOK_CLIENT_ID`
- **Directory (tenant) ID** → This is your `OUTLOOK_TENANT_ID`

**Tenant ID options:**
- `"consumers"` - Personal Microsoft accounts (Outlook.com, Hotmail)
- `"organizations"` - Work/school accounts only
- `"common"` - Both personal and work/school accounts
- Specific tenant ID - For a specific organization

### 3. Create Client Secret

1. Go to **Certificates & secrets** → **+ New client secret**
2. Add description and expiration period
3. **⚠️ CRITICAL**: Immediately copy the **Value** → This is your `OUTLOOK_CLIENT_SECRET`
   - You cannot view this again!

### 4. Configure API Permissions

1. Go to **API permissions** → **+ Add a permission**
2. Select **Microsoft Graph** → **Delegated permissions**
3. Add these permissions:
   - ✅ `Mail.Read` - Read user mail
   - ✅ `Mail.ReadWrite` - Read and write user mail
   - ✅ `offline_access` - Maintain access to data (refresh tokens)
4. Click **Grant admin consent** (if you're an administrator)

### 5. Environment Variables

⚠️ **Never commit credentials!** Use environment variables:

```bash
export OUTLOOK_CLIENT_ID="your-application-id"
export OUTLOOK_CLIENT_SECRET="your-client-secret"
export OUTLOOK_TENANT_ID="common"  # or "consumers", "organizations", or tenant ID
export OUTLOOK_REDIRECT_URL="http://localhost:8080/callback"
```


## Operation Guides

Detailed guides for specific operations:

### Core Operations
- **[Messages](./operations/messages.md)** - List, read, and manage emails
- **[Attachments](./operations/attachments.md)** - Download files from emails
- **[Search](./operations/search.md)** - Advanced queries with Microsoft Graph syntax
- **[Delete](./operations/delete.md)** - Delete messages and manage trash
- **[Folders](./operations/folders.md)** - Manage mail folders and organization


## Usage Examples

See [examples/outlook/](../../examples/outlook/) for a complete, runnable example demonstrating all features.

### Running the Example

```bash
# 1. Set credentials
export OUTLOOK_CLIENT_ID="your-application-id"
export OUTLOOK_CLIENT_SECRET="your-client-secret"
export OUTLOOK_TENANT_ID="common"

# 2. Run example
cd examples/outlook
go run main.go
```

**First Run:** Browser will open for OAuth2 authorization. Token is saved to `token.json` for future use.

**What it demonstrates:**
- OAuth2 authentication with token persistence
- Listing recent messages
- Searching messages with Graph queries
- Managing folders
- Error handling

### Common Search Queries

Use in `ListOptions.Query`:

```go
"from:boss@company.com"                  // From sender
"subject:invoice"                        // Subject contains
"hasAttachments:true"                    // Has attachments
"importance:high"                        // High priority
"isRead:false"                           // Unread only
"from:john@example.com subject:report"   // Combined
```

### Well-Known Folder IDs

```go
outlook.FolderInbox        // "inbox"
outlook.FolderDrafts       // "drafts"
outlook.FolderSentItems    // "sentitems"
outlook.FolderDeletedItems // "deleteditems"
outlook.FolderJunkEmail    // "junkemail"
outlook.FolderOutbox       // "outbox"
outlook.FolderArchive      // "archive"
```

### Comparison with Gmail

| Feature | Outlook | Gmail |
|---------|---------|-------|
| Authentication | Microsoft Entra ID OAuth2 | Google OAuth2 |
| Organization | Folders (single per message) | Labels (multiple per message) |
| Well-known IDs | `"inbox"`, `"drafts"` | `"INBOX"`, `"DRAFT"` |
| Search syntax | Microsoft Graph queries | Gmail search operators |


## Resources

**Official Documentation:**
- 📖 [Microsoft Graph Mail API](https://learn.microsoft.com/graph/api/resources/mail-api-overview)
- 🔐 [Microsoft Entra ID](https://learn.microsoft.com/entra/identity/)
- 📝 [App Registration Guide](https://learn.microsoft.com/entra/identity-platform/quickstart-register-app)
- 🔍 [Graph Permissions Reference](https://learn.microsoft.com/graph/permissions-reference)


## Troubleshooting

Common issues and solutions:

| Error | Cause | Solution |
|-------|-------|----------|
| "client not connected" | Not authenticated | Call `ConnectWithToken()` before operations |
| "InvalidAuthenticationToken" | Token expired | Call `RefreshToken(ctx)` |
| "Insufficient privileges" | Missing permissions | Add `Mail.Read`, `Mail.ReadWrite` in Entra ID |
| "redirect_uri_mismatch" | URL mismatch | Update redirect URI in Entra ID |
| "AADSTS65001" | Consent missing | Grant admin consent in API permissions |
| "AADSTS9002346" | Tenant mismatch | Update `OUTLOOK_TENANT_ID` to match app type |

For detailed troubleshooting, see the [complete guide](https://learn.microsoft.com/graph/errors).
