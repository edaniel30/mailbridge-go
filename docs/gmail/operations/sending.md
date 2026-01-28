# Sending Emails

Send emails with plain text, HTML, and attachments.

> **Setup required**: [OAuth2 configuration](../gmail/GMAIL.md#setup-oauth2)

## Simple Text Email

```go
response, err := client.SendMessage(ctx, &core.Draft{
    To:      []core.EmailAddress{{Address: "user@example.com"}},
    Subject: "Hello from MailBridge",
    Body: &core.EmailBody{
        Plain: "This is a plain text email.",
    },
}, nil)

fmt.Printf("Sent! ID: %s\n", response.ID)
```

## HTML Email

```go
draft := &core.Draft{
    To:      []core.EmailAddress{{Address: "user@example.com"}},
    Subject: "HTML Newsletter",
    Body: &core.EmailBody{
        HTML: `
            <h1>Welcome!</h1>
            <p>This is an <strong>HTML</strong> email.</p>
        `,
    },
}

client.SendMessage(ctx, draft, nil)
```

## With Both Plain & HTML

```go
draft := &core.Draft{
    To:      []core.EmailAddress{{Address: "user@example.com"}},
    Subject: "Multi-part Email",
    Body: &core.EmailBody{
        Plain: "This is the plain text version",
        HTML:  "<p>This is the <strong>HTML</strong> version</p>",
    },
}
// Email clients will choose which to display
```

## Multiple Recipients

```go
draft := &core.Draft{
    To: []core.EmailAddress{
        {Name: "Alice", Address: "alice@example.com"},
        {Name: "Bob", Address: "bob@example.com"},
    },
    Cc: []core.EmailAddress{
        {Address: "manager@example.com"},
    },
    Bcc: []core.EmailAddress{
        {Address: "archive@example.com"},
    },
    Subject: "Team Meeting",
    Body: &core.EmailBody{
        Plain: "Meeting at 3pm today.",
    },
}
```

## Reply to Message

```go
// Get original message
original, _ := client.GetMessage(ctx, originalMessageID)

// Send reply
draft := &core.Draft{
    To:      []core.EmailAddress{{Address: original.From.Address}},
    Subject: "Re: " + original.Subject,
    Body: &core.EmailBody{
        Plain: "Thanks for your email!",
    },
}

client.SendMessage(ctx, draft, nil)
```

## Complete Example

See [`examples/gmail-send`](../../examples/gmail-send/) for interactive sending demo.

```bash
cd examples/gmail-send
go run main.go
```

## Email Address Format

```go
// Simple
{Address: "user@example.com"}

// With name
{Name: "John Doe", Address: "john@example.com"}
// Renders as: John Doe <john@example.com>
```

## Related

- [Messages](./messages.md) - Read and list emails
- [Attachments](./attachments.md) - Download files
