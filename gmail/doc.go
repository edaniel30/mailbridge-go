// Package gmail provides integration with Gmail via Google's Gmail API.
//
// Architecture:
//
// The package follows a modular design with three layers:
// 1. Public API (gmail/*.go) - User-facing operations
// 2. Internal interfaces (gmail/internal/) - Abstraction layer for testability
// 3. External SDK (Google Gmail API) - Actual API implementation
//
// Key Features:
//
//   - OAuth2 authentication with Google
//   - List, get, and download email messages
//   - Manage labels (Gmail's equivalent to folders)
//   - Send, reply, and forward messages
//   - Mark messages as read/unread, star/unstar
//   - Move and delete messages
//   - Advanced search with query builder
//   - Full interface-based mocking for testing
//
// Basic Usage:
//
//	// 1. Create configuration
//	config := &gmail.Config{
//	    ClientID:     "your-google-client-id.apps.googleusercontent.com",
//	    ClientSecret: "your-google-client-secret",
//	    RedirectURL:  "http://localhost",
//	    Scopes:       gmail.DefaultScopes(),
//	}
//
//	// 2. Create client
//	client, err := gmail.New(config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	// 3. Get authorization URL
//	authURL := client.GetAuthURL("random-state-string")
//	fmt.Println("Visit:", authURL)
//
//	// 4. Exchange auth code for token (after user authorizes)
//	token, err := client.ExchangeCode(ctx, authCode)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 5. Connect with token
//	err = client.ConnectWithToken(ctx, token)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 6. List messages
//	response, err := client.ListMessages(ctx, &core.ListOptions{
//	    MaxResults: 10,
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	for _, email := range response.Emails {
//	    fmt.Printf("%s: %s\n", email.From.Email, email.Subject)
//	}
//
// Google Cloud Console Setup:
//
// To use this package, you need to create OAuth2 credentials in Google Cloud Console:
//  1. Go to https://console.cloud.google.com
//  2. Create a new project or select an existing one
//  3. Navigate to "APIs & Services" > "Credentials"
//  4. Click "Create Credentials" > "OAuth client ID"
//  5. Configure:
//     - Application type: Desktop app (for CLI) or Web application
//     - Authorized redirect URIs: http://localhost (for desktop apps)
//  6. Download the credentials JSON or note the Client ID and Client Secret
//  7. Enable Gmail API:
//     - Go to "APIs & Services" > "Library"
//     - Search for "Gmail API"
//     - Click "Enable"
//
// Token Persistence:
//
//	// Save token after connection
//	token := client.GetToken()
//	tokenJSON, _ := json.Marshal(token)
//	os.WriteFile("token.json", tokenJSON, 0600)
//
//	// Load token for future use
//	tokenJSON, _ := os.ReadFile("token.json")
//	var token oauth2.Token
//	json.Unmarshal(tokenJSON, &token)
//	client.ConnectWithToken(ctx, &token)
//
// Query Builder:
//
// Gmail provides a query builder for advanced message filtering:
//
//	query := gmail.NewQueryBuilder().
//	    IsUnread().
//	    InInbox().
//	    From("user@example.com").
//	    HasAttachment().
//	    After(time.Now().AddDate(0, 0, -7)).
//	    Build()
//
//	messages, err := client.ListMessages(ctx, &core.ListOptions{
//	    Query: query,
//	})
//
// Available query methods:
//   - IsUnread() / IsRead()
//   - InInbox() / InSent() / InDrafts() / InSpam() / InTrash()
//   - From(email) / To(email) / Subject(text)
//   - HasAttachment()
//   - After(date) / Before(date)
//   - Label(name)
//   - And more...
//
// Labels Management:
//
//	// List all labels
//	labels, err := client.ListLabels(ctx)
//
//	// Create a new label
//	label, err := client.CreateLabel(ctx, "MyLabel")
//
//	// Apply label to message
//	err = client.ModifyLabels(ctx, messageID, []string{labelID}, nil)
//
//	// Remove label from message
//	err = client.ModifyLabels(ctx, messageID, nil, []string{labelID})
//
// Well-known Label IDs:
//
//   - gmail.LabelInbox: "INBOX"
//   - gmail.LabelSent: "SENT"
//   - gmail.LabelDrafts: "DRAFT"
//   - gmail.LabelSpam: "SPAM"
//   - gmail.LabelTrash: "TRASH"
//   - gmail.LabelUnread: "UNREAD"
//   - gmail.LabelStarred: "STARRED"
//   - gmail.LabelImportant: "IMPORTANT"
//
// Sending Messages:
//
//	// Send a simple message
//	err := client.SendMessage(ctx, &core.Message{
//	    To:      []core.EmailAddress{{Email: "recipient@example.com"}},
//	    Subject: "Hello from MailBridge",
//	    Body:    core.EmailBody{Text: "This is a test message"},
//	})
//
//	// Reply to a message
//	err := client.ReplyToMessage(ctx, messageID, &core.Message{
//	    Body: core.EmailBody{Text: "Thanks for your message!"},
//	})
//
// Attachments:
//
//	// Download attachment
//	attachment, err := client.GetAttachment(ctx, messageID, attachmentID)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	os.WriteFile(attachment.Filename, attachment.Data, 0644)
//
//	// Send message with attachment
//	err := client.SendMessage(ctx, &core.Message{
//	    To:      []core.EmailAddress{{Email: "recipient@example.com"}},
//	    Subject: "File attached",
//	    Body:    core.EmailBody{Text: "Please find the file attached"},
//	    Attachments: []core.Attachment{
//	        {
//	            Filename: "document.pdf",
//	            MimeType: "application/pdf",
//	            Data:     fileData,
//	        },
//	    },
//	})
//
// Error Handling:
//
//	// All errors from Gmail API are wrapped with context
//	_, err := client.GetMessage(ctx, "invalid-id")
//	if err != nil {
//	    // Error will contain descriptive message
//	    fmt.Println(err) // "failed to get message invalid-id: googleapi: Error 404: Not Found"
//	}
//
// Testing:
//
// The package provides mocks in gmail/testing for unit testing:
//
//	import gmailtest "github.com/danielrivera/mailbridge-go/gmail/testing"
//
//	func TestMyCode(t *testing.T) {
//	    mockService := &gmailtest.MockGmailService{}
//	    mockMessages := &gmailtest.MockMessagesService{}
//
//	    // Configure mock expectations
//	    mockService.On("GetUsersService").Return(mockUsers)
//	    mockMessages.On("List", mock.Anything).Return(expectedMessages, nil)
//
//	    // Inject mock
//	    client.SetService(mockService)
//
//	    // Test your code...
//	}
//
// Rate Limiting:
//
// Gmail API has rate limits. The SDK handles retries automatically, but you should:
//   - Batch operations when possible
//   - Use pagination for large result sets
//   - Cache label information
//   - Avoid polling; use push notifications when available
//
// Security Best Practices:
//
//   - Store tokens securely (use file permissions 0600)
//   - Never commit credentials to version control
//   - Use environment variables for client ID and secret
//   - Request minimal OAuth scopes needed
//   - Implement token refresh logic
//   - Handle token revocation gracefully
//
// For complete examples, see the examples/gmail directory.
package gmail
