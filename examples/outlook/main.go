package main

// Outlook MailBridge Example
//
// This example demonstrates how to authenticate and interact with Outlook using Microsoft Entra ID.
//
// SETUP:
// 1. Register an app in Microsoft Entra ID (https://portal.azure.com or https://entra.microsoft.com)
//    - Navigate to: Microsoft Entra ID → Applications → App registrations
//    - Create a new registration with redirect URI: http://localhost:8080/callback
//    - Add API permissions: Mail.Read, Mail.ReadWrite, offline_access
//    - Create a client secret and copy it immediately
//
// 2. Set environment variables with your credentials:
//
//    export OUTLOOK_CLIENT_ID="your-application-client-id"
//    export OUTLOOK_CLIENT_SECRET="your-client-secret-value"
//    export OUTLOOK_TENANT_ID="consumers"  # Use "consumers" for personal accounts (Outlook.com, Hotmail)
//                                          # Use "organizations" for work/school accounts
//                                          # Use "common" for both types
//                                          # Or use your specific tenant ID
//
//    Or create a .env file and source it:
//    source .env
//
// 3. Run the example:
//    go run main.go
//
// The example will:
// - Open a browser for OAuth2 authorization (first time only)
// - Save the token to token.json for future use
// - List recent messages from your inbox
// - List your mail folders
// - Search for messages

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/danielrivera/mailbridge-go/core"
	"github.com/danielrivera/mailbridge-go/outlook"
	"golang.org/x/oauth2"
)

const (
	tokenFile = "token.json"
	port      = ":8080"
)

var (
	client        *outlook.Client
	expectedState = "random-state-string" // In production, generate random state
)

func main() {
	ctx := context.Background()

	// Load configuration from environment variables
	// Make sure to export these before running:
	//   export OUTLOOK_CLIENT_ID="..."
	//   export OUTLOOK_CLIENT_SECRET="..."
	//   export OUTLOOK_TENANT_ID="consumers"  # for personal accounts
	config := &outlook.Config{
		ClientID:     os.Getenv("OUTLOOK_CLIENT_ID"),
		ClientSecret: os.Getenv("OUTLOOK_CLIENT_SECRET"),
		TenantID:     os.Getenv("OUTLOOK_TENANT_ID"),
		RedirectURL:  fmt.Sprintf("http://localhost%s/callback", port),
	}

	// Validate configuration
	if config.ClientID == "" || config.ClientSecret == "" || config.TenantID == "" {
		log.Println("\n❌ ERROR: Missing required environment variables!")
		log.Println("\nPlease set the following environment variables:")
		log.Println("\n  export OUTLOOK_CLIENT_ID=\"your-application-client-id\"")
		log.Println("  export OUTLOOK_CLIENT_SECRET=\"your-client-secret-value\"")
		log.Println("  export OUTLOOK_TENANT_ID=\"consumers\"")
		log.Println("\nTenant ID options:")
		log.Println("  - \"consumers\"      → Personal accounts (Outlook.com, Hotmail, Xbox, Skype)")
		log.Println("  - \"organizations\"  → Work/school accounts only")
		log.Println("  - \"common\"         → Both personal and work/school accounts")
		log.Println("  - Your tenant ID    → Specific organization only")
		log.Println("\nGet these credentials from:")
		log.Println("  https://portal.azure.com → Microsoft Entra ID → App registrations")
		log.Println("  or https://entra.microsoft.com")
		log.Fatal("\nExiting...")
	}

	// Create client
	var err error
	client, err = outlook.New(config)
	if err != nil {
		log.Fatal("Failed to create client:", err)
	}

	// Try to load existing token
	token, err := loadToken()
	if err == nil {
		log.Println("Loading saved token...")
		err = client.ConnectWithToken(ctx, token)
		if err != nil {
			log.Println("Failed to connect with saved token:", err)
			token = nil
		} else {
			log.Println("Successfully connected with saved token")

			// Check if token is expired
			if token.Expiry.Before(time.Now()) {
				log.Println("Token expired, refreshing...")
				newToken, err := client.RefreshToken(ctx)
				if err != nil {
					log.Println("Failed to refresh token:", err)
					token = nil
				} else {
					token = newToken
					if err := saveToken(token); err != nil {
						log.Printf("Warning: failed to save refreshed token: %v", err)
					}
					log.Println("Token refreshed successfully")
				}
			}
		}
	}

	// If no valid token, start OAuth flow
	if token == nil {
		log.Println("No valid token found, starting OAuth flow...")
		authURL := client.GetAuthURL(expectedState)

		fmt.Println("\n" + strings.Repeat("=", 80))
		fmt.Println("AUTHORIZATION REQUIRED")
		fmt.Println(strings.Repeat("=", 80))
		fmt.Printf("\nPlease visit this URL to authorize:\n\n%s\n\n", authURL)
		fmt.Println(strings.Repeat("=", 80))

		// Start HTTP server for callback
		http.HandleFunc("/callback", handleCallback)
		http.HandleFunc("/", handleRoot)

		fmt.Printf("\nStarting callback server on http://localhost%s\n", port)
		fmt.Println("Waiting for authorization...")

		if err := http.ListenAndServe(port, nil); err != nil {
			log.Fatal("Failed to start server:", err)
		}
	} else {
		// Token is valid, run examples
		runExamples(ctx)
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintf(w, `
		<html>
		<head><title>Outlook Example</title></head>
		<body>
			<h1>Outlook Integration Example</h1>
			<p>Waiting for authorization callback...</p>
			<p>If you've already authorized, this page will close automatically.</p>
		</body>
		</html>
	`)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Verify state
	state := r.URL.Query().Get("state")
	if state != expectedState {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Get authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		errorMsg := r.URL.Query().Get("error")
		errorDesc := r.URL.Query().Get("error_description")
		http.Error(w, fmt.Sprintf("Authorization failed: %s - %s", errorMsg, errorDesc), http.StatusBadRequest)
		return
	}

	// Exchange code for token
	log.Println("Received authorization code, exchanging for token...")
	err := client.ConnectWithAuthCode(ctx, code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to exchange code for token: %v", err), http.StatusInternalServerError)
		return
	}

	// Save token
	token := client.GetToken()
	if err := saveToken(token); err != nil {
		log.Println("Failed to save token:", err)
	} else {
		log.Println("Token saved successfully")
	}

	// Success page
	_, _ = fmt.Fprintf(w, `
		<html>
		<head><title>Authorization Successful</title></head>
		<body>
			<h1>Authorization Successful!</h1>
			<p>You can close this window and return to the terminal.</p>
			<script>setTimeout(() => window.close(), 3000);</script>
		</body>
		</html>
	`)

	// Run examples in background
	go func() {
		time.Sleep(2 * time.Second)
		runExamples(context.Background())
		os.Exit(0)
	}()
}

func runExamples(ctx context.Context) {
	if !client.IsConnected() {
		log.Fatal("Client not connected")
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("OUTLOOK MAILBRIDGE EXAMPLES")
	fmt.Println(strings.Repeat("=", 80) + "\n")

	// Example 1: List recent messages
	fmt.Println("1. Listing recent messages...")
	fmt.Println(strings.Repeat("-", 80))
	listMessages(ctx)

	// Example 2: List folders
	fmt.Println("\n2. Listing folders...")
	fmt.Println(strings.Repeat("-", 80))
	listFolders(ctx)

	// Example 3: Search messages
	fmt.Println("\n3. Searching messages...")
	fmt.Println(strings.Repeat("-", 80))
	searchMessages(ctx)

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("Examples completed successfully!")
	fmt.Println(strings.Repeat("=", 80))
}

func listMessages(ctx context.Context) {
	response, err := client.ListMessages(ctx, &core.ListOptions{
		MaxResults: 5,
	})
	if err != nil {
		log.Printf("Failed to list messages: %v\n", err)
		return
	}

	fmt.Printf("Found %d messages:\n\n", len(response.Emails))
	for i, email := range response.Emails {
		fmt.Printf("[%d] From: %s <%s>\n", i+1, email.From.Name, email.From.Email)
		fmt.Printf("    Subject: %s\n", email.Subject)
		fmt.Printf("    Date: %s\n", email.Date.Format("2006-01-02 15:04"))
		fmt.Printf("    Read: %v\n", email.IsRead)
		if len(email.Snippet) > 100 {
			fmt.Printf("    Preview: %s...\n", email.Snippet[:100])
		} else {
			fmt.Printf("    Preview: %s\n", email.Snippet)
		}
		fmt.Println()
	}
}

func listFolders(ctx context.Context) {
	folders, err := client.ListFolders(ctx)
	if err != nil {
		log.Printf("Failed to list folders: %v\n", err)
		return
	}

	fmt.Printf("Found %d folders:\n\n", len(folders))
	for i, folder := range folders {
		fmt.Printf("[%d] Name: %s\n", i+1, folder.Name)
		fmt.Printf("    ID: %s\n", folder.ID)
		fmt.Printf("    Type: %s\n", folder.Type)
		fmt.Printf("    Total messages: %d\n", folder.TotalMessages)
		fmt.Printf("    Unread messages: %d\n", folder.UnreadMessages)
		fmt.Println()
	}
}

func searchMessages(ctx context.Context) {
	// Search for messages with "invoice" in subject
	response, err := client.ListMessages(ctx, &core.ListOptions{
		MaxResults: 3,
		Query:      "subject:meeting",
	})
	if err != nil {
		log.Printf("Failed to search messages: %v\n", err)
		return
	}

	fmt.Printf("Found %d messages matching 'meeting':\n\n", len(response.Emails))
	for i, email := range response.Emails {
		fmt.Printf("[%d] %s\n", i+1, email.Subject)
		fmt.Printf("    From: %s\n", email.From.Email)
		fmt.Printf("    Date: %s\n", email.Date.Format("2006-01-02 15:04"))
		fmt.Println()
	}
}

// Token persistence functions

func saveToken(token *oauth2.Token) error {
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	err = os.WriteFile(tokenFile, data, 0600)
	if err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}

func loadToken() (*oauth2.Token, error) {
	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read token file: %w", err)
	}

	var token oauth2.Token
	err = json.Unmarshal(data, &token)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %w", err)
	}

	return &token, nil
}
