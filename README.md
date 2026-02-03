# MailBridge Go

A modular, type-safe Go library for interacting with email providers.

## Supported Providers

| Provider | Status | Documentation |
|----------|--------|---------------|
| Gmail | ✅ Ready | [docs/GMAIL.md](docs/gmail/gmail.md) |
| Outlook | ✅ Ready | [docs/OUTLOOK.md](docs/outlook/outlook.md) |

## Installation

```bash
go get github.com/danielrivera/mailbridge-go
```

## Architecture

MailBridge uses a modular architecture where you import only what you need:

```go
import (
    "github.com/danielrivera/mailbridge-go/core"    // Shared types
    "github.com/danielrivera/mailbridge-go/gmail"   // Gmail provider
    "github.com/danielrivera/mailbridge-go/outlook" // Outlook provider
)
```

**Benefits:**
- **Smaller binaries**: Only import what you use
- **Type safety**: Compile-time checks
- **Extensibility**: Easy to add providers
- **Clean API**: No provider enums

## Core Types

All providers use normalized types from `core` package:

```go
type Email struct {
    ID          string
    Subject     string
    From        EmailAddress
    To          []EmailAddress
    Date        time.Time
    Body        EmailBody
    Labels      []string
    Attachments []Attachment
    IsRead      bool
    // ...
}

type ListOptions struct {
    MaxResults int64
    PageToken  string
    Query      string
    Labels     []string
}
```

## Documentation

Each provider has its own comprehensive documentation:

- **[Gmail Integration](docs/GMAIL.md)** - Google OAuth2 setup, authentication, usage examples
- **[Outlook Integration](docs/outlook/outlook.md)** - Microsoft Entra ID setup, authentication, usage examples
- **[Examples](examples/)** - Working code samples for all providers

## Testing

```bash
make test              # All tests
make test-coverage     # With coverage report
make test-unit         # Unit tests only
make pre-commit        # Run all checks
```

## License

MIT License - see [LICENSE](LICENSE)
