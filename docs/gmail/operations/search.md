# Advanced Search

Build complex Gmail search queries without knowing the syntax.

> **Setup required**: [OAuth2 configuration](../gmail/GMAIL.md#setup-oauth2)

## QueryBuilder

Fluent API for building Gmail queries.

### Basic Search

```go
query := gmail.NewQueryBuilder().
    From("boss@company.com").
    IsUnread().
    Build()

messages, _ := client.ListMessages(ctx, &core.ListOptions{
    Query: query,  // "from:boss@company.com is:unread"
})
```

### Multiple Conditions

```go
query := gmail.NewQueryBuilder().
    From("client@example.com").
    Subject("invoice").
    HasAttachment().
    After(time.Now().AddDate(0, 0, -30)).  // Last 30 days
    Build()
// Result: from:client@example.com subject:invoice has:attachment after:2026/01/01
```

### Size Filters

```go
query := gmail.NewQueryBuilder().
    LargerThan(gmail.MegaBytes(5)).
    HasAttachment().
    Build()
// Find emails with attachments >5MB
```

Helper functions:
- `gmail.MegaBytes(5)` → 5MB in bytes
- `gmail.KiloBytes(500)` → 500KB in bytes
- `gmail.GigaBytes(1)` → 1GB in bytes

### Status Filters

```go
// Unread messages
gmail.NewQueryBuilder().IsUnread().Build()

// Starred messages
gmail.NewQueryBuilder().IsStarred().Build()

// Important messages
gmail.NewQueryBuilder().IsImportant().Build()

// Messages in trash
gmail.NewQueryBuilder().InTrash().Build()
```

### Date Filters

```go
thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

// Messages after date
query := gmail.NewQueryBuilder().After(thirtyDaysAgo).Build()

// Messages before date
query := gmail.NewQueryBuilder().Before(time.Now()).Build()

// Relative dates
query := gmail.NewQueryBuilder().
    NewerThan("2d").  // Last 2 days
    Build()

query := gmail.NewQueryBuilder().
    OlderThan("1y").  // Older than 1 year
    Build()
```

Relative units: `d` (days), `m` (months), `y` (years)

### Logical Operators

```go
// OR condition
query := gmail.NewQueryBuilder().
    From("alice@example.com").
    OR().
    From("bob@example.com").
    Build()
// Result: from:alice@example.com OR from:bob@example.com

// NOT condition
query := gmail.NewQueryBuilder().
    HasAttachment().
    NOT().
    From("spam@example.com").
    Build()
// Result: has:attachment -from:spam@example.com
```

### Location Filters

```go
gmail.NewQueryBuilder().InInbox().Build()
gmail.NewQueryBuilder().InSent().Build()
gmail.NewQueryBuilder().InTrash().Build()
gmail.NewQueryBuilder().InLabel("Work").Build()
```

### Attachment Filters

```go
// Has any attachment
gmail.NewQueryBuilder().HasAttachment().Build()

// Specific filename
gmail.NewQueryBuilder().Filename("report.pdf").Build()

// Google Drive files
gmail.NewQueryBuilder().HasDrive().Build()

// YouTube videos
gmail.NewQueryBuilder().HasYoutube().Build()
```

## Raw Queries

If you know Gmail syntax, use it directly:

```go
query := gmail.NewQueryBuilder().
    Raw("from:user@example.com label:important").
    Build()
```

## Complete Example

See [`examples/gmail-search`](../../examples/gmail-search/) for 7 practical search examples.

```bash
cd examples/gmail-search
go run main.go
```

## Common Patterns

### Unread from specific person
```go
gmail.NewQueryBuilder().
    From("boss@example.com").
    IsUnread().
    Build()
```

### Large attachments
```go
gmail.NewQueryBuilder().
    LargerThan(gmail.MegaBytes(10)).
    HasAttachment().
    Build()
```

### Recent invoices
```go
gmail.NewQueryBuilder().
    Subject("invoice").
    After(time.Now().AddDate(0, -1, 0)).  // Last month
    Build()
```

### Unread in specific label
```go
gmail.NewQueryBuilder().
    IsUnread().
    InLabel("Important").
    Build()
```

## Gmail Search Syntax Reference

If QueryBuilder doesn't cover your use case, use raw Gmail syntax:

- `from:user@example.com`
- `to:user@example.com`
- `subject:meeting`
- `has:attachment`
- `is:unread`
- `is:starred`
- `after:2026/01/01`
- `larger:5M`
- `filename:pdf`

[Full Gmail search syntax](https://support.google.com/mail/answer/7190)

## Related

- [Messages](./messages.md) - List and read results
- [Attachments](./attachments.md) - Download files from results
