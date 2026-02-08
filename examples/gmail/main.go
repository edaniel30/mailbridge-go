package main

// Gmail MailBridge Example
//
// This example demonstrates how to authenticate and interact with Gmail using Google OAuth2.
//
// SETUP:
// 1. Create a project in Google Cloud Console (https://console.cloud.google.com)
//    - Navigate to: APIs & Services → Credentials
//    - Create OAuth 2.0 Client ID (Application type: Desktop app or Web application)
//    - Add authorized redirect URIs (for desktop: http://localhost)
//    - Download the credentials
//
// 2. Set environment variables with your credentials:
//
//    export GMAIL_CLIENT_ID="your-client-id.apps.googleusercontent.com"
//    export GMAIL_CLIENT_SECRET="your-client-secret"
//
//    Or create a .env file and source it:
//    source .env
//
// 3. Run the example:
//    go run main.go
//
// The example will:
// - Prompt for OAuth2 authorization (first time only)
// - Save the token to token.json for future use
// - List recent messages from your inbox
// - Search for messages with various filters
// - Download attachments
// - Manage labels and folders

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/edaniel30/mailbridge-go/gmail"
	"golang.org/x/oauth2"
)

const (
	tokenFile   = "token.json"
	downloadDir = "attachments"
)

var client *gmail.Client

func main() {
	ctx := context.Background()

	// Load configuration from environment variables
	// Make sure to export these before running:
	//   export GMAIL_CLIENT_ID="..."
	//   export GMAIL_CLIENT_SECRET="..."
	config := &gmail.Config{
		ClientID:     os.Getenv("GMAIL_CLIENT_ID"),
		ClientSecret: os.Getenv("GMAIL_CLIENT_SECRET"),
		RedirectURL:  "http://localhost", // Must match Google Cloud Console
		Scopes:       gmail.DefaultScopes(),
	}

	// Validate configuration
	if config.ClientID == "" || config.ClientSecret == "" {
		log.Println("\n❌ ERROR: Missing required environment variables!")
		log.Println("\nPlease set the following environment variables:")
		log.Println("\n  export GMAIL_CLIENT_ID=\"your-client-id.apps.googleusercontent.com\"")
		log.Println("  export GMAIL_CLIENT_SECRET=\"your-client-secret\"")
		log.Println("\nGet these credentials from:")
		log.Println("  https://console.cloud.google.com → APIs & Services → Credentials")
		log.Fatal("\nExiting...")
	}

	// Create client
	var err error
	client, err = gmail.New(config)
	if err != nil {
		log.Fatal("Failed to create client:", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Warning: failed to close client: %v", err)
		}
	}()

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
		}
	}

	// If no valid token, start OAuth flow
	if token == nil {
		log.Println("No valid token found, starting OAuth flow...")
		authURL := client.GetAuthURL("state-token")

		fmt.Println("\n" + strings.Repeat("=", 80))
		fmt.Println("AUTHORIZATION REQUIRED")
		fmt.Println(strings.Repeat("=", 80))
		fmt.Printf("\nPlease visit this URL to authorize:\n\n%s\n\n", authURL)
		fmt.Println(strings.Repeat("=", 80))
		fmt.Print("\nEnter the authorization code: ")

		var code string
		if _, err := fmt.Scanln(&code); err != nil {
			log.Printf("Failed to read code: %v\n", err)
			return
		}

		token, err = client.ExchangeCode(ctx, code)
		if err != nil {
			log.Printf("Failed to exchange code: %v\n", err)
			return
		}

		// Connect with new token
		if err := client.ConnectWithToken(ctx, token); err != nil {
			log.Printf("Failed to connect with token: %v\n", err)
			return
		}

		// Save token
		if err := saveToken(token); err != nil {
			log.Println("Failed to save token:", err)
		} else {
			log.Println("Token saved successfully")
		}
	}

	if !client.IsConnected() {
		log.Println("Failed to establish connection")
		return
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
