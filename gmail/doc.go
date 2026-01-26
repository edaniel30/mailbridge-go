// Package gmail provides a client for interacting with the Gmail API.
//
// This package offers a type-safe, idiomatic Go interface for Gmail operations
// including OAuth2 authentication, message management, and label operations.
//
// # Quick Start
//
// Create a client with your OAuth2 credentials:
//
//	cfg := &gmail.Config{
//		ClientID:     "your-client-id",
//		ClientSecret: "your-client-secret",
//		RedirectURL:  "http://localhost",
//		Scopes:       gmail.DefaultScopes(),
//	}
//
//	client, err := gmail.New(cfg)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close()
//
// # OAuth2 Flow
//
// Get an authorization URL for the user:
//
//	authURL := client.GetAuthURL("state-token")
//	fmt.Printf("Visit: %s\n", authURL)
//
// After the user authorizes, exchange the code for a token:
//
//	token, err := client.ExchangeCode(ctx, "authorization-code")
//	if err != nil {
//		log.Fatal(err)
//	}
//
// Connect using the token:
//
//	err = client.ConnectWithToken(ctx, token)
//	if err != nil {
//		log.Fatal(err)
//	}
//
// # Listing Messages
//
// List messages with optional filters:
//
//	opts := &core.ListOptions{
//		MaxResults: 10,
//		Query:      "is:unread",
//		Labels:     []string{"INBOX"},
//	}
//
//	response, err := client.ListMessages(ctx, opts)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	for _, email := range response.Emails {
//		fmt.Printf("Subject: %s\n", email.Subject)
//	}
//
// # Message Operations
//
// Get a specific message:
//
//	email, err := client.GetMessage(ctx, "message-id")
//
// Mark messages as read or unread:
//
//	err = client.MarkAsRead(ctx, "message-id")
//	err = client.MarkAsUnread(ctx, "message-id")
//
// Move messages to folders/labels:
//
//	err = client.MoveMessageToFolder(ctx, "message-id", "Archive")
//
// # Label Operations
//
// List all labels:
//
//	labels, err := client.ListLabels(ctx)
//
// Create a new label:
//
//	label, err := client.CreateLabel(ctx, "MyLabel")
//
// Add or remove labels from messages:
//
//	err = client.AddLabelToMessage(ctx, "message-id", "label-id")
//	err = client.RemoveLabelFromMessage(ctx, "message-id", "label-id")
//
// # Gmail Search Queries
//
// The Query field in ListOptions supports Gmail's search syntax:
//
//   - is:unread - Unread messages
//   - is:starred - Starred messages
//   - from:user@example.com - From specific sender
//   - subject:keyword - Subject contains keyword
//   - has:attachment - Has attachments
//   - after:2024/01/01 - After specific date
//
// Combine multiple criteria:
//
//	opts := &core.ListOptions{
//		Query: "is:unread is:important from:boss@company.com",
//	}
//
// # Token Management
//
// Refresh an expired token:
//
//	newToken, err := client.RefreshToken(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
//
// Check connection status:
//
//	if !client.IsConnected() {
//		log.Fatal("Not connected to Gmail")
//	}
//
// # Error Handling
//
// All methods return standard Go errors. Configuration errors implement
// the core.ConfigError type:
//
//	err := cfg.Validate()
//	if err != nil {
//		var configErr *core.ConfigError
//		if errors.As(err, &configErr) {
//			fmt.Printf("Configuration error in field %s: %s\n",
//				configErr.Field, configErr.Message)
//		}
//	}
//
// # For More Information
//
// See the examples directory for complete working examples and the
// docs/GMAIL.md file for setup instructions including obtaining OAuth2
// credentials from Google Cloud Console.
package gmail
