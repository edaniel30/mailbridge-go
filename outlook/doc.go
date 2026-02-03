// Package outlook provides integration with Microsoft Outlook/Exchange via Microsoft Graph API.
//
// Architecture:
//
// The package follows the same modular design as the Gmail provider:
// 1. Public API (outlook/*.go) - User-facing operations
// 2. Internal interfaces (outlook/internal/) - Abstraction layer for testability
// 3. External SDK (Microsoft Graph SDK) - Actual API implementation
//
// Key Features:
//
//   - OAuth2 authentication with Microsoft Entra ID (formerly Azure AD)
//   - List, get, and download email messages
//   - Manage mail folders (Outlook's equivalent to Gmail labels)
//   - Mark messages as read/unread
//   - Move and delete messages
//   - Full interface-based mocking for testing
//
// Basic Usage:
//
//	// 1. Create configuration
//	config := &outlook.Config{
//	    ClientID:     "your-azure-app-client-id",
//	    ClientSecret: "your-azure-app-client-secret",
//	    TenantID:     "your-azure-tenant-id",
//	    RedirectURL:  "http://localhost:8080/callback",
//	}
//
//	// 2. Create client
//	client, err := outlook.New(config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 3. Get authorization URL
//	authURL := client.GetAuthURL("random-state-string")
//	fmt.Println("Visit:", authURL)
//
//	// 4. Exchange auth code for token (after user authorizes)
//	err = client.ConnectWithAuthCode(ctx, authCode)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 5. List messages
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
// Microsoft Entra ID App Registration:
//
// To use this package, you need to register an application in Microsoft Entra ID:
//  1. Go to https://portal.azure.com or https://entra.microsoft.com
//  2. Navigate to "Microsoft Entra ID" > "Applications" > "App registrations"
//     (you can also search for "Azure Active Directory" - both names work)
//  3. Click "New registration"
//  4. Configure:
//     - Name: Your app name (e.g., "MailBridge Outlook Integration")
//     - Supported account types: Choose based on your needs
//     - Redirect URI: http://localhost:8080/callback (for development)
//  5. After creation, note the "Application (client) ID" and "Directory (tenant) ID"
//  6. Create a client secret under "Certificates & secrets"
//     IMPORTANT: Copy the secret value immediately - it won't be shown again
//  7. Add API permissions:
//     - Microsoft Graph > Delegated permissions
//     - Add: Mail.Read, Mail.ReadWrite, offline_access
//     - Grant admin consent if you're an administrator
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
// Error Handling:
//
//	// All errors from Microsoft Graph API are converted to readable messages
//	_, err := client.GetMessage(ctx, "invalid-id")
//	if err != nil {
//	    // Error will contain Microsoft Graph error code and message
//	    fmt.Println(err) // "microsoft graph error [ErrorItemNotFound]: The specified object was not found"
//	}
//
// Folders vs Labels:
//
// Outlook uses "folders" while Gmail uses "labels". This package converts Outlook folders
// to core.Label types for consistency:
//
//	// List all folders
//	folders, err := client.ListFolders(ctx)
//
//	// List messages in a specific folder
//	messages, err := client.ListMessagesInFolder(ctx, outlook.FolderInbox, opts)
//
//	// Move message to folder
//	err = client.MoveMessage(ctx, messageID, outlook.FolderArchive)
//
// Well-known Folder IDs:
//
//   - outlook.FolderInbox: "inbox"
//   - outlook.FolderDrafts: "drafts"
//   - outlook.FolderSentItems: "sentitems"
//   - outlook.FolderDeletedItems: "deleteditems"
//   - outlook.FolderJunkEmail: "junkemail"
//   - outlook.FolderArchive: "archive"
//
// Testing:
//
// The package provides mocks in outlook/testing for unit testing:
//
//	import outlooktest "github.com/danielrivera/mailbridge-go/outlook/testing"
//
//	func TestMyCode(t *testing.T) {
//	    mockService := &outlooktest.MockGraphService{}
//	    // Configure mock expectations...
//	    client.SetService(mockService)
//	    // Test your code...
//	}
//
// For complete examples, see the examples/outlook directory.
package outlook
