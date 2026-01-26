# MailBridge Examples

This directory contains working examples demonstrating how to use MailBridge with different providers.

## Gmail Examples

### [gmail](gmail/) - Basic Gmail Integration

Complete example showing:
- OAuth2 authentication flow
- Listing messages with filters
- Getting full message details
- Moving messages to folders
- Marking messages as read/unread

```bash
cd gmail
export GMAIL_CLIENT_ID="your-id"
export GMAIL_CLIENT_SECRET="your-secret"
go run main.go
```

### [gmail-attachments](gmail-attachments/) - Download Attachments

Demonstrates attachment handling:
- Searching for messages with attachments
- Listing attachment metadata
- Downloading attachment contents
- Saving files to disk
- Organizing downloads by message

```bash
cd gmail-attachments
export GMAIL_CLIENT_ID="your-id"
export GMAIL_CLIENT_SECRET="your-secret"
go run main.go
```

## Prerequisites

All Gmail examples require:
- Gmail API credentials (Client ID and Client Secret)
- See [Gmail Setup Guide](../docs/GMAIL.md#setup-oauth2-credentials)

## Running Examples

1. Set environment variables:
   ```bash
   export GMAIL_CLIENT_ID="your-client-id.apps.googleusercontent.com"
   export GMAIL_CLIENT_SECRET="your-client-secret"
   ```

2. Navigate to the example directory:
   ```bash
   cd examples/gmail  # or examples/gmail-attachments
   ```

3. Run the example:
   ```bash
   go run main.go
   ```

4. On first run, you'll be prompted to authorize the application in your browser.

## Token Storage

All examples save the OAuth2 token to `token.json` for subsequent runs. To re-authenticate, simply delete this file.

## Learn More

- [Gmail Integration Guide](../docs/GMAIL.md) - Complete Gmail documentation
- [Core Types](../core/) - Shared types used across all providers
