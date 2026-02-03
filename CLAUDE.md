# MailBridge Go - Provider Implementation Guide

**Version:** 2.0 | **Language:** Go 1.25 | **Architecture:** Modular Provider Packages

> **Purpose:** Guide for implementing email provider integrations in MailBridge Go.

---

## 1. Architecture

### Core Principle
Each provider is an **independent package** that depends on `core/` types. The `core/` package has **ZERO dependencies** on providers.

```
┌─────────────────────────────────────┐
│  PUBLIC API (provider/*.go)         │  ← User-facing interface
│  - OAuth2, configuration            │
└─────────────────────────────────────┘
            ↓
┌─────────────────────────────────────┐
│  INTERNAL (provider/internal/)      │  ← Abstraction for mocking
│  - Interface wrappers for SDK       │
└─────────────────────────────────────┘
            ↓
┌─────────────────────────────────────┐
│  EXTERNAL SDK (3rd party)           │  ← Provider's official SDK
└─────────────────────────────────────┘
```

### Project Structure
```
mailbridge-go/
├── core/                    # Provider-agnostic types
│   ├── types.go             # Email, ListOptions, Attachment, Label
│   └── errors.go            # ConfigError
│
├── <provider>/              # Provider package (gmail, outlook, etc.)
│   ├── client.go            # OAuth2 + connection
│   ├── config.go            # Configuration + validation
│   ├── messages.go          # Message operations
│   ├── folders.go           # Folder operations (optional)
│   ├── internal/
│   │   ├── interfaces.go    # SDK abstractions
│   │   └── service_wrapper.go  # Real implementations
│   └── testing/
│       └── mocks.go         # Mock implementations
│
├── examples/<provider>/     # Runnable example
└── docs/<provider>/         # Provider guide
```

---

## 2. Core Types (core/types.go)

All providers MUST convert to these normalized types:

```go
type Email struct {
    ID          string
    ThreadID    string
    Subject     string
    From        EmailAddress
    To          []EmailAddress
    Cc          []EmailAddress
    Bcc         []EmailAddress
    Date        time.Time
    Body        EmailBody
    Snippet     string
    Attachments []Attachment
    Labels      []string
    IsRead      bool
    IsStarred   bool
}

type ListOptions struct {
    MaxResults int64
    PageToken  string
    Query      string
    Labels     []string
}

type ListResponse struct {
    Emails        []*Email
    NextPageToken string
    TotalCount    int64
}

type Attachment struct {
    ID       string
    Filename string
    MimeType string
    Size     int64
    Data     []byte  // Only populated when explicitly downloaded
}

type Label struct {
    ID             string
    Name           string
    Type           string  // "system" or "user"
    TotalMessages  int
    UnreadMessages int
}
```

---

## 3. Implementation Steps

### Step 1: Create Package Structure
```bash
mkdir <provider> && cd <provider>
touch client.go config.go messages.go
mkdir internal testing
touch internal/{interfaces,service_wrapper}.go testing/mocks.go
```

### Step 2: Configuration (config.go)
```go
package provider

type Config struct {
    ClientID     string
    ClientSecret string
    RedirectURL  string
    Scopes       []string
}

func (c *Config) Validate() error {
    if c.ClientID == "" {
        return &core.ConfigError{Field: "ClientID", Message: "required"}
    }
    // Validate all required fields
    return nil
}

func DefaultScopes() []string {
    return []string{/* provider OAuth scopes */}
}
```

### Step 3: Interfaces (internal/interfaces.go)
```go
package internal

// Main service interface
type ProviderService interface {
    GetMessagesService() MessagesService
}

// Messages interface
type MessagesService interface {
    List(ctx context.Context, opts *core.ListOptions) ([]ProviderMessage, error)
    Get(ctx context.Context, messageID string) (ProviderMessage, error)
    GetAttachment(ctx, messageID, attachmentID string) (ProviderAttachment, error)
    MarkAsRead(ctx context.Context, messageID string) error
    MarkAsUnread(ctx context.Context, messageID string) error
    Delete(ctx context.Context, messageID string) error
}
```

### Step 4: Client (client.go)
```go
package provider

type Client struct {
    config       *Config
    oauth2Config *oauth2.Config
    token        *oauth2.Token
    service      internal.ProviderService
}

func New(config *Config) (*Client, error) {
    if err := config.Validate(); err != nil {
        return nil, err
    }
    return &Client{config: config}, nil
}

func (c *Client) IsConnected() bool {
    return c.service != nil && c.token != nil
}

func (c *Client) GetAuthURL(state string) string {
    return c.oauth2Config.AuthCodeURL(state)
}

func (c *Client) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
    return c.oauth2Config.Exchange(ctx, code)
}

func (c *Client) ConnectWithToken(ctx context.Context, token *oauth2.Token) error {
    // Initialize provider SDK with token
    // Wrap SDK with internal.NewRealProviderService()
    c.token = token
    return nil
}

func (c *Client) RefreshToken(ctx context.Context) (*oauth2.Token, error) {
    tokenSource := c.oauth2Config.TokenSource(ctx, c.token)
    return tokenSource.Token()
}

func (c *Client) Close() error {
    c.service = nil
    return nil
}
```

### Step 5: Message Operations (messages.go)
```go
package provider

func (c *Client) ListMessages(ctx context.Context, opts *core.ListOptions) (*core.ListResponse, error) {
    if !c.IsConnected() {
        return nil, fmt.Errorf("client not connected")
    }

    messagesService := c.service.GetMessagesService()
    providerMessages, err := messagesService.List(ctx, opts)
    if err != nil {
        return nil, fmt.Errorf("failed to list messages: %w", err)
    }

    // Convert provider types to core types
    emails := make([]*core.Email, 0, len(providerMessages))
    for _, msg := range providerMessages {
        emails = append(emails, c.convertMessage(msg))
    }

    return &core.ListResponse{Emails: emails}, nil
}

func (c *Client) GetMessage(ctx context.Context, messageID string) (*core.Email, error) {
    if !c.IsConnected() {
        return nil, fmt.Errorf("client not connected")
    }

    messagesService := c.service.GetMessagesService()
    message, err := messagesService.Get(ctx, messageID)
    if err != nil {
        return nil, fmt.Errorf("failed to get message: %w", err)
    }

    return c.convertMessage(message), nil
}

// ADAPTER PATTERN: Convert provider types to core.Email
func (c *Client) convertMessage(msg ProviderMessage) *core.Email {
    return &core.Email{
        ID:      extractID(msg),
        Subject: extractSubject(msg),
        From:    extractFrom(msg),
        // ... populate all fields
    }
}
```

### Step 6: Service Wrapper (internal/service_wrapper.go)
```go
package internal

type RealProviderService struct {
    client *sdk.Client  // Provider's SDK
}

func NewRealProviderService(client *sdk.Client) ProviderService {
    return &RealProviderService{client: client}
}

func (r *RealProviderService) GetMessagesService() MessagesService {
    return &realMessagesService{client: r.client}
}

type realMessagesService struct {
    client *sdk.Client
}

func (r *realMessagesService) List(ctx context.Context, opts *core.ListOptions) ([]ProviderMessage, error) {
    // Call provider SDK
    // Convert core.ListOptions to provider format
    // Return provider messages
}
```

### Step 7: Mocks (testing/mocks.go)
```go
package testing

import "github.com/stretchr/testify/mock"

type MockProviderService struct {
    mock.Mock
}

func (m *MockProviderService) GetMessagesService() internal.MessagesService {
    return m.Called().Get(0).(internal.MessagesService)
}

type MockMessagesService struct {
    mock.Mock
}

func (m *MockMessagesService) List(ctx context.Context, opts *core.ListOptions) ([]internal.ProviderMessage, error) {
    args := m.Called(ctx, opts)
    return args.Get(0).([]internal.ProviderMessage), args.Error(1)
}
```

### Step 8: Tests
```go
// Unit tests (*_test.go)
func TestConfig_Validate(t *testing.T) { /* ... */ }

// Integration tests (*_integration_test.go)
func TestClient_ListMessages(t *testing.T) {
    mockService := &testing.MockProviderService{}
    mockMessages := &testing.MockMessagesService{}

    mockService.On("GetMessagesService").Return(mockMessages)
    mockMessages.On("List", mock.Anything, mock.Anything).Return([]internal.ProviderMessage{}, nil)

    client := &Client{service: mockService}
    result, err := client.ListMessages(ctx, opts)

    assert.NoError(t, err)
    mockService.AssertExpectations(t)
}
```

---

## 4. Critical Rules

1. **Provider Independence:** `core/` MUST NOT import provider packages
2. **Type Normalization:** All provider types MUST convert to `core.*` types via adapter pattern
3. **No Logging:** Library code MUST NOT log (return errors instead)
4. **Interface-Based Calls:** All SDK calls MUST go through interfaces (enables mocking)
5. **Lazy Loading:** Don't download attachments in `ListMessages()` (use explicit `GetAttachment()`)
6. **Error Wrapping:** Always wrap errors: `fmt.Errorf("failed to X: %w", err)`

---

## 5. Design Patterns

### Adapter Pattern
Convert provider types to `core.*` types in `convertMessage()`:
```go
func (c *Client) convertMessage(providerMsg ProviderMessage) *core.Email {
    return &core.Email{
        ID:      providerMsg.GetID(),
        Subject: providerMsg.GetSubject(),
        // Normalize all fields
    }
}
```

### Wrapper Pattern
Wrap SDK in interfaces for mocking (`internal/service_wrapper.go`):
```go
type RealProviderService struct {
    client *ProviderSDK  // Concrete SDK
}
```

### Dependency Injection
Inject interfaces, not concrete types:
```go
type Client struct {
    service internal.ProviderService  // Interface, not concrete
}
```

---

## 6. Testing

### Coverage Target
- **Minimum:** 74%
- **Commands:** `make test`, `make test-coverage`

### Mock Pattern
```go
// 1. Create mocks
mockService := &testing.MockProviderService{}
mockMessages := &testing.MockMessagesService{}

// 2. Configure expectations
mockService.On("GetMessagesService").Return(mockMessages)
mockMessages.On("List", ctx, opts).Return(expectedMessages, nil)

// 3. Inject
client.service = mockService

// 4. Execute & assert
result, err := client.ListMessages(ctx, opts)
assert.NoError(t, err)
mockService.AssertExpectations(t)
```

---

## 7. Validation Checklist

Before submitting:

**Code:**
- [ ] `core/` package NOT modified
- [ ] All types convert to `core.*` types
- [ ] All SDK calls through interfaces
- [ ] Error wrapping with context
- [ ] No logging in library

**Testing:**
- [ ] Unit tests (*_test.go)
- [ ] Integration tests (*_integration_test.go)
- [ ] Coverage ≥74%
- [ ] `make pre-commit` passes

**Documentation:**
- [ ] Provider guide in docs/<provider>/
- [ ] Runnable example in examples/<provider>/
- [ ] README.md updated

**Files:**
- [ ] client.go, config.go, messages.go
- [ ] internal/interfaces.go, internal/service_wrapper.go
- [ ] testing/mocks.go

---

## 8. Anti-Patterns

```go
// ❌ Provider logic in core
package core
func ConvertProviderMessage(msg *provider.Message) {...}

// ✅ Provider handles conversions
package provider
func (c *Client) convertMessage(msg *provider.Message) *core.Email {...}

// ❌ Concrete dependencies
type Client struct {
    sdk *provider.SDK  // Hard to mock
}

// ✅ Interface dependencies
type Client struct {
    service internal.ProviderService  // Easy to mock
}

// ❌ Testing against real API
client := setupRealClient()  // Slow, requires credentials

// ✅ Mock-based testing
client := &Client{service: mockService}  // Fast, deterministic

// ❌ Logging
log.Printf("Fetching %s", id)

// ✅ Return errors
return fmt.Errorf("failed to get message: %w", err)
```

---

## 9. Required Methods

```go
// Client
New(config *Config) (*Client, error)
IsConnected() bool
GetAuthURL(state string) string
ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error)
ConnectWithToken(ctx context.Context, token *oauth2.Token) error
RefreshToken(ctx context.Context) (*oauth2.Token, error)
Close() error

// Messages
ListMessages(ctx context.Context, opts *core.ListOptions) (*core.ListResponse, error)
GetMessage(ctx context.Context, messageID string) (*core.Email, error)
GetAttachment(ctx context.Context, messageID, attachmentID string) (*core.Attachment, error)
MarkAsRead(ctx context.Context, messageID string) error
MarkAsUnread(ctx context.Context, messageID string) error
DeleteMessage(ctx context.Context, messageID string) error
```

---

## 10. Reference Implementations

- **Gmail:** See `gmail/` package
- **Outlook:** See `outlook/` package

**Commands:**
```bash
make test              # Run all tests
make test-coverage     # Generate coverage report
make pre-commit        # Lint + test
```

---

**Document Version:** 2.0
**Last Updated:** February 2025
