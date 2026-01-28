# MailBridge Go - Developer Context

**Version:** 2.0 | **Language:** Go 1.25 | **Architecture:** Modular Provider Packages

> **Quick Links:** [README](README.md) | [Gmail Docs](docs/GMAIL.md) | [Examples](examples/README.md) | [Makefile](Makefile)

---

## 1. Architecture Overview

### Core Principle: Modular Provider Packages

**v1.x → v2.0 Refactoring (Completed Jan 2025):**
- **Before:** Centralized `mailbridge.Client` with provider enum on every call
- **After:** Provider-specific packages (`gmail`, `outlook`) imported independently
- **Benefits:** Smaller binaries, cleaner API, easier extensibility

```go
// v1.x (OLD)
client.ListMessages(ctx, config.ProviderGmail, opts)

// v2.0 (NEW)
gmailClient.ListMessages(ctx, opts)
```

### Three-Layer Architecture

```
┌─────────────────────────────────────────┐
│  PUBLIC API (gmail/*.go)                │  ← Users interact here
│  - OAuth2, high-level operations        │  ← Converts to core types
└─────────────────────────────────────────┘
            ↓ uses
┌─────────────────────────────────────────┐
│  INTERNAL INTERFACES (gmail/internal/)  │  ← Abstraction for mocking
│  - GmailService, MessagesService        │  ← Enables testability
└─────────────────────────────────────────┘
            ↓ implements
┌─────────────────────────────────────────┐
│  EXTERNAL API (google.golang.org/api)   │  ← Gmail SDK
└─────────────────────────────────────────┘
```

**Key Rule:** `core` package has ZERO dependencies on providers. Providers depend on `core`.

---

## 2. Project Structure

```
mailbridge-go/
├── core/                       # Provider-agnostic types (Email, ListOptions, Attachment)
│   ├── types.go                # Normalized email types
│   └── errors.go               # ConfigError for validation
│
├── gmail/                      # Gmail provider (PUBLIC)
│   ├── client.go               # OAuth2 + connection management
│   ├── config.go               # Gmail config + validation
│   ├── messages.go             # List/get messages, download attachments
│   ├── labels.go               # Label operations, mark read/unread
│   ├── internal/               # Private implementation details
│   │   ├── interfaces.go       # Abstraction interfaces for Gmail API
│   │   └── service_wrapper.go # Real implementations wrapping Gmail SDK
│   └── testing/                # Public mocks for external tests
│       └── mocks.go            # testify/mock implementations
│
├── examples/                   # Runnable examples (see examples/README.md)
├── docs/                       # Provider-specific guides (see docs/GMAIL.md)
└── Makefile                    # Tasks: test, coverage, pre-commit
```

**Critical Files:**
- `gmail/client.go` (136 lines) - Entry point, OAuth2 flow
- `gmail/messages.go` (287 lines) - Message ops, attachment download, conversion to core types
- `gmail/labels.go` (233 lines) - Label management
- `gmail/internal/interfaces.go` (98 lines) - Interface abstractions for mocking

---

## 3. Design Patterns

### 3.1 Adapter Pattern
**Location:** `gmail/messages.go:106` (`convertMessage()`)

Converts Gmail API types to provider-agnostic `core.Email`:
```go
func (c *Client) convertMessage(msg *gmail.Message) *core.Email {
    return &core.Email{
        ID:      msg.Id,
        Subject: parseHeader(msg.Payload.Headers, "subject"),
        From:    parseEmailAddress(headers["from"]),
        // ... normalize all fields
    }
}
```

### 3.2 Wrapper Pattern
**Location:** `gmail/internal/service_wrapper.go`

Wraps concrete Gmail SDK in interfaces for mocking:
```go
type RealGmailService struct {
    service *gmail.Service  // ← Concrete Gmail SDK
}

func (r *RealGmailService) GetUsersService() UsersService {
    return &realUsersService{users: r.service.Users}
}
```

### 3.3 Strategy Pattern
**Location:** `gmail/messages.go:207` (`decodeBody()`)

Try multiple base64 decoding strategies:
```go
// 1. RawURLEncoding (Gmail default)
decoded, err := base64.RawURLEncoding.DecodeString(data)
if err != nil {
    // 2. URLEncoding (with padding)
    decoded, err = base64.URLEncoding.DecodeString(data)
    if err != nil {
        // 3. Standard base64 (fallback)
        decoded, err = base64.StdEncoding.DecodeString(data)
    }
}
```

### 3.4 Factory Method
**Location:** `gmail/client.go:22` (`New()`)

```go
func New(config *Config) (*Client, error) {
    if err := config.Validate(); err != nil {  // Eager validation
        return nil, err
    }
    return &Client{
        config:       config,
        oauth2Config: config.ToOAuth2Config(),
    }, nil
}
```

---

## 4. Dependency Injection

**Interface-Based DI for Testability:**

```go
// gmail/client.go
type Client struct {
    service internal.GmailService  // ← Interface, not *gmail.Service
}

// Production: inject real implementation
func (c *Client) ConnectWithToken(ctx, token) {
    gmailService := gmail.New(...)
    c.service = internal.NewRealGmailService(gmailService)
}

// Testing: inject mock
func TestClient_Operation(t *testing.T) {
    mockService := &gmailtest.MockGmailService{}
    client.SetService(mockService)  // ← Swap implementation
}
```

**Result:** Tests never hit real Gmail API. Fast, deterministic, no credentials needed.

---

## 5. Coding Conventions

### Naming
- **Exported:** `PascalCase` → `ListMessages`, `EmailAddress`, `ConfigError`
- **Unexported:** `camelCase` → `convertMessage`, `messageID`, `oauth2Config`
- **Acronyms:** All caps when exported (`ID`, `URL`, `HTML`), lowercase when unexported

### File Organization
```go
package gmail

// 1. Imports (stdlib → external → internal)
import (
    "context"
    "fmt"

    "golang.org/x/oauth2"

    "github.com/danielrivera/mailbridge-go/core"
)

// 2. Types
type Client struct { ... }

// 3. Constructor
func New(config *Config) (*Client, error) { ... }

// 4. Public methods (alphabetical or logical grouping)
func (c *Client) ListMessages(...) {...}

// 5. Private helpers (end of file)
func convertMessage(...) {...}
```

### Comments
- **Exported:** Must have godoc starting with name
- **Complex logic:** Explain "why", not "what"
- **Special cases:** Mark with inline comments (e.g., `// Gmail uses base64url without padding`)

### Error Handling
```go
// ✅ DO: Wrap with context
return fmt.Errorf("failed to get message %s: %w", messageID, err)

// ❌ DON'T: Generic errors
return fmt.Errorf("error occurred")

// ❌ DON'T: Swallow errors silently
if err != nil {
    continue  // User never knows why it failed
}
```

---

## 6. Testing Strategy

**Coverage:** 86.1% (threshold: 74%)
**Test Count:** 58 tests
**Commands:** See [Makefile](Makefile)

### Test Types

**1. Unit Tests (`*_test.go`)**
- Pure functions: parsers, converters, validators
- Example: `gmail/messages_test.go` (parseEmailAddress, decodeBody, extractBody)

**2. Integration Tests (`*_integration_test.go`)**
- Client operations with mocked services
- Example: `gmail/messages_integration_test.go` (ListMessages, GetMessage, GetAttachment)

### Mocking Pattern
```go
// 1. Create mocks
mockService := &gmailtest.MockGmailService{}
mockCall := &gmailtest.MockMessagesGetCall{}

// 2. Configure expectations
mockService.On("GetMessagesService").Return(mockMessagesService)
mockCall.On("Context", ctx).Return(mockCall)
mockCall.On("Do").Return(expectedMessage, nil)

// 3. Inject
client.SetService(mockService)

// 4. Execute & assert
result, err := client.GetMessage(ctx, "msg-123")
assert.NoError(t, err)
mockService.AssertExpectations(t)
```

**No E2E Tests:** Examples serve as manual E2E validation.

---

## 7. Data Flow

### Message Retrieval Lifecycle

```
1. User calls gmailClient.ListMessages(ctx, opts)
   ↓
2. Client validates connection (IsConnected())
   ↓
3. service.GetUsersService().GetMessagesService().List("me")
   ↓ (interface call to gmail/internal/)
4. RealMessagesService.List() → wraps Gmail API
   ↓
5. Gmail API returns *gmail.ListMessagesResponse
   ↓
6. For each message: GetMessage(ctx, msg.Id)
   ↓
7. convertMessage(*gmail.Message) → *core.Email
   ↓
8. Return core.ListResponse with []*core.Email
```

**Key Transformation:** `gmail.Message` (DTO) → `core.Email` (Domain Model) at step 7

---

## 8. Critical Business Rules

### 1. Provider Independence
- **Rule:** `core/` MUST NOT import provider packages
- **Enforcement:** Manual review, import path checks
- **Why:** Enables adding providers without modifying core

### 2. Normalized Types
- **Rule:** All provider types MUST convert to `core.*` types before returning to user
- **Why:** Consistent API across providers
- **Implementation:** Adapter pattern in `convertMessage()`, etc.

### 3. No Logging in Library
- **Rule:** Library code MUST NOT log
- **Why:** Users control logging, not libraries
- **Exception:** Examples can log for demonstration

### 4. Interface-Based External Calls
- **Rule:** All external API calls MUST go through interfaces
- **Why:** Enables mocking for tests
- **Implementation:** `gmail/internal/interfaces.go` layer

### 5. Lazy Attachment Loading
- **Rule:** Attachments not downloaded in `ListMessages()`, only metadata
- **Why:** Avoid downloading large files unnecessarily
- **Usage:** Explicit `GetAttachment(ctx, messageID, attachmentID)` call

---

## 9. Anti-Patterns to Avoid

```go
// ❌ 1. Provider Enums (v1.x pattern)
client.ListMessages(ctx, config.ProviderGmail, opts)

// ✅ Use modular imports
gmailClient.ListMessages(ctx, opts)

// ❌ 2. Mixing Provider Logic in Core
package core
func ConvertGmailMessage(msg *gmail.Message) {...}  // NO!

// ✅ Provider handles its own conversions
package gmail
func (c *Client) convertMessage(msg *gmail.Message) *core.Email {...}

// ❌ 3. Logging in Library
log.Printf("Fetching message %s", messageID)  // NO!

// ✅ Return errors, let user log
return fmt.Errorf("failed to get message: %w", err)

// ❌ 4. Concrete Dependencies
type Client struct {
    service *gmail.Service  // Hard to mock
}

// ✅ Interface Dependencies
type Client struct {
    service internal.GmailService  // Easy to mock
}

// ❌ 5. Testing Against Real API
client.ConnectWithToken(ctx, realToken)  // Slow, brittle

// ✅ Inject Mocks
client.SetService(mockService)  // Fast, deterministic
```

---

## 10. Adding New Features

### Adding a New Provider (e.g., Outlook)

1. **Create package:** `mkdir outlook && touch outlook/{client,config,messages}.go`
2. **Define interfaces:** `outlook/internal/interfaces.go`
3. **Implement wrappers:** `outlook/internal/service_wrapper.go`
4. **Convert to core types:** `convertMessage()` in `outlook/messages.go`
5. **Create mocks:** `outlook/testing/mocks.go`
6. **Write tests:** `outlook/*_test.go`, `outlook/*_integration_test.go`
7. **Update docs:** `docs/OUTLOOK.md`, update `README.md` table
8. **Add example:** `examples/outlook/`

**Validation Checklist:**
- [ ] `core` package NOT modified
- [ ] New provider independent of `gmail`
- [ ] All external calls through interfaces
- [ ] All types convert to `core.*`
- [ ] Coverage ≥74%

### Adding Operation to Gmail

**Example: Send Email**

1. Add method: `gmail/client.go` → `func (c *Client) SendEmail(...)`
2. Add interface: `gmail/internal/interfaces.go` → `MessagesService.Send()`
3. Implement wrapper: `gmail/internal/service_wrapper.go` → `realMessagesService.Send()`
4. Add mock: `gmail/testing/mocks.go` → `MockMessagesService.Send()`
5. Write tests: `gmail/messages_test.go` + `gmail/messages_integration_test.go`
6. Update docs: `gmail/doc.go`, `docs/GMAIL.md`
7. Run: `make pre-commit`

---

## 11. Performance Considerations

**1. Attachment Download:**
- Lazy loading (metadata only in `ListMessages()`)
- Explicit download via `GetAttachment()`

**2. Pagination:**
- Use `ListOptions.MaxResults` (default: provider limit)
- Fetch in batches (20-100 recommended)
- Use `NextPageToken` for subsequent pages

**3. Base64 Decoding:**
- Strategy pattern: RawURLEncoding first (fastest for Gmail)
- Fallback to URLEncoding → StdEncoding

**4. No Caching:**
- Current: No caching layer
- Users can add caching externally if needed

---

## 12. Security

**OAuth2 Tokens:**
- Never log tokens
- User responsible for persistence (library doesn't store)
- Implement refresh: `client.RefreshToken(ctx)`

**Credentials:**
- Environment variables recommended (see `examples/`)
- Never commit: `.gitignore` includes `.env`, `token.json`
- Request minimal scopes: `gmail.DefaultScopes()`

---

## 13. Dependencies

**Core:**
- Go 1.25
- `github.com/stretchr/testify` v1.11.1 - Testing/mocking
- `golang.org/x/oauth2` v0.24.0 - OAuth2 flow
- `google.golang.org/api` v0.214.0 - Gmail API

**Dev Tools:**
- `golangci-lint` - 27+ linters (see `.golangci.yml`)
- `pre-commit` - Automated checks (see `.pre-commit-config.yaml`)

---

## 14. Quick Reference

**Commands:**
```bash
make test              # All tests
make test-coverage     # Coverage report (threshold: 74%)
make pre-commit        # Lint + test (runs on commit)
```

**Documentation:**
- Setup & Usage: [README.md](README.md)
- Gmail Guide: [docs/GMAIL.md](docs/GMAIL.md)
- Examples: [examples/README.md](examples/README.md)

**Architecture Diagrams:**
- Layer separation: Section 1
- Data flow: Section 7
- Directory tree: Section 2

---

## 15. Troubleshooting

**"client not connected"**
→ Call `client.ConnectWithToken(ctx, token)` before operations

**"undefined: gmailtest"**
→ Import with alias: `gmailtest "github.com/danielrivera/mailbridge-go/gmail/testing"`

**Coverage below 74%**
→ Run `make test-coverage-html` to identify uncovered code

**Linter errors**
→ Run `make pre-commit` before committing

---

**Document Version:** 2.0
**Last Updated:** January 2025
**Maintained By:** AI-assisted development
