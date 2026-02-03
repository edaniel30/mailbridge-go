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
	"path/filepath"
	"strings"
	"time"

	"github.com/danielrivera/mailbridge-go/core"
	"github.com/danielrivera/mailbridge-go/gmail"
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

			// Try to refresh if expired
			newToken, err := client.RefreshToken(ctx)
			if err == nil && newToken != nil {
				token = newToken
				if err := saveToken(token); err != nil {
					log.Printf("Warning: failed to save refreshed token: %v", err)
				}
				log.Println("Token refreshed successfully")
			}
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

	// Run examples
	runExamples(ctx)
}

func runExamples(ctx context.Context) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("GMAIL MAILBRIDGE EXAMPLES")
	fmt.Println(strings.Repeat("=", 80) + "\n")

	// Example 1: List recent messages
	fmt.Println("1. Listing recent messages...")
	fmt.Println(strings.Repeat("-", 80))
	listRecentMessages(ctx)

	// Example 2: List unread messages
	fmt.Println("\n2. Listing unread messages...")
	fmt.Println(strings.Repeat("-", 80))
	listUnreadMessages(ctx)

	// Example 3: Search messages
	fmt.Println("\n3. Searching messages...")
	fmt.Println(strings.Repeat("-", 80))
	searchMessages(ctx)

	// Example 4: Get message details
	fmt.Println("\n4. Getting message details...")
	fmt.Println(strings.Repeat("-", 80))
	getMessageDetails(ctx)

	// Example 5: List labels
	fmt.Println("\n5. Listing labels...")
	fmt.Println(strings.Repeat("-", 80))
	listLabels(ctx)

	// Example 6: Download attachments
	// Uncomment the following lines to enable attachment downloading:
	// fmt.Println("\n6. Downloading attachments...")
	// fmt.Println(strings.Repeat("-", 80))
	// downloadAttachments(ctx)

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("Examples completed successfully!")
	fmt.Println(strings.Repeat("=", 80))
}

func listRecentMessages(ctx context.Context) {
	response, err := client.ListMessages(ctx, &core.ListOptions{
		MaxResults: 5,
	})
	if err != nil {
		log.Printf("Failed to list messages: %v\n", err)
		return
	}

	fmt.Printf("Found %d messages (total: %d):\n\n", len(response.Emails), response.TotalCount)
	for i, email := range response.Emails {
		fmt.Printf("[%d] From: %s <%s>\n", i+1, email.From.Name, email.From.Email)
		fmt.Printf("    Subject: %s\n", email.Subject)
		fmt.Printf("    Date: %s\n", email.Date.Format("2006-01-02 15:04"))
		fmt.Printf("    Read: %v | Starred: %v\n", email.IsRead, email.IsStarred)
		if len(email.Snippet) > 80 {
			fmt.Printf("    Preview: %s...\n", email.Snippet[:80])
		} else {
			fmt.Printf("    Preview: %s\n", email.Snippet)
		}
		if len(email.Attachments) > 0 {
			fmt.Printf("    Attachments: %d\n", len(email.Attachments))
		}
		fmt.Println()
	}
}

func listUnreadMessages(ctx context.Context) {
	// Using query builder for readable search
	query := gmail.NewQueryBuilder().
		IsUnread().
		InInbox().
		Build()

	response, err := client.ListMessages(ctx, &core.ListOptions{
		MaxResults: 5,
		Query:      query,
	})
	if err != nil {
		log.Printf("Failed to list unread messages: %v\n", err)
		return
	}

	fmt.Printf("Found %d unread messages:\n\n", len(response.Emails))
	for i, email := range response.Emails {
		fmt.Printf("[%d] %s\n", i+1, email.Subject)
		fmt.Printf("    From: %s\n", email.From.Email)
		fmt.Printf("    Date: %s\n", email.Date.Format("2006-01-02 15:04"))
		fmt.Println()
	}
}

func searchMessages(ctx context.Context) {
	// Example: Search for messages with attachments from last 30 days
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	query := gmail.NewQueryBuilder().
		HasAttachment().
		After(thirtyDaysAgo).
		Build()

	response, err := client.ListMessages(ctx, &core.ListOptions{
		MaxResults: 3,
		Query:      query,
	})
	if err != nil {
		log.Printf("Failed to search messages: %v\n", err)
		return
	}

	fmt.Printf("Query: %s\n", query)
	fmt.Printf("Found %d messages with attachments:\n\n", len(response.Emails))
	for i, email := range response.Emails {
		fmt.Printf("[%d] %s\n", i+1, email.Subject)
		fmt.Printf("    From: %s\n", email.From.Email)
		fmt.Printf("    Date: %s\n", email.Date.Format("2006-01-02"))
		fmt.Printf("    Attachments: %d\n", len(email.Attachments))
		for j, att := range email.Attachments {
			fmt.Printf("      [%d] %s (%s, %s)\n", j+1, att.Filename, att.MimeType, formatBytes(att.Size))
		}
		fmt.Println()
	}
}

func getMessageDetails(ctx context.Context) {
	// Get first unread message for details
	response, err := client.ListMessages(ctx, &core.ListOptions{
		MaxResults: 1,
		Labels:     []string{"UNREAD"},
	})
	if err != nil || len(response.Emails) == 0 {
		fmt.Println("No unread messages found")
		return
	}

	messageID := response.Emails[0].ID

	// Get full message details
	email, err := client.GetMessage(ctx, messageID)
	if err != nil {
		log.Printf("Failed to get message: %v\n", err)
		return
	}

	fmt.Printf("Message ID: %s\n", email.ID)
	fmt.Printf("Thread ID: %s\n", email.ThreadID)
	fmt.Printf("Subject: %s\n", email.Subject)
	fmt.Printf("From: %s <%s>\n", email.From.Name, email.From.Email)

	if len(email.To) > 0 {
		fmt.Printf("To: ")
		for i, to := range email.To {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%s <%s>", to.Name, to.Email)
		}
		fmt.Println()
	}

	fmt.Printf("Date: %s\n", email.Date.Format("2006-01-02 15:04:05"))
	fmt.Printf("Labels: %v\n", email.Labels)
	fmt.Printf("\nBody (Text):\n%s\n", truncate(email.Body.Text, 200))

	if email.Body.HTML != "" {
		fmt.Printf("\nBody (HTML):\n%s\n", truncate(email.Body.HTML, 200))
	}
}

func listLabels(ctx context.Context) {
	labels, err := client.ListLabels(ctx)
	if err != nil {
		log.Printf("Failed to list labels: %v\n", err)
		return
	}

	// Group by type
	systemCount := 0
	userCount := 0

	for _, label := range labels {
		if label.Type == "system" {
			systemCount++
		} else {
			userCount++
		}
	}

	fmt.Printf("Found %d labels (%d system, %d user):\n\n", len(labels), systemCount, userCount)

	// Show system labels
	if systemCount > 0 {
		fmt.Println("System Labels:")
		for _, label := range labels {
			if label.Type == "system" {
				fmt.Printf("  - %s (ID: %s)\n", label.Name, label.ID)
			}
		}
		fmt.Println()
	}

	// Show user labels
	if userCount > 0 {
		fmt.Println("User Labels:")
		for _, label := range labels {
			if label.Type != "system" {
				fmt.Printf("  - %s (ID: %s)\n", label.Name, label.ID)
			}
		}
	}
}

// downloadAttachments downloads attachments from messages.
// Uncomment the call in runExamples() to enable this functionality.
//
//nolint:unused // Intentionally unused until user enables it
func downloadAttachments(ctx context.Context) {
	// Create download directory
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		log.Printf("Failed to create download directory: %v\n", err)
		return
	}

	// Search for messages with attachments
	query := gmail.NewQueryBuilder().
		HasAttachment().
		Build()

	response, err := client.ListMessages(ctx, &core.ListOptions{
		MaxResults: 3,
		Query:      query,
	})
	if err != nil {
		log.Printf("Failed to search messages: %v\n", err)
		return
	}

	if len(response.Emails) == 0 {
		fmt.Println("No messages with attachments found")
		return
	}

	fmt.Printf("Downloading attachments from %d messages:\n\n", len(response.Emails))

	totalDownloaded := 0
	totalBytes := int64(0)

	for i, email := range response.Emails {
		if len(email.Attachments) == 0 {
			continue
		}

		fmt.Printf("[%d/%d] %s\n", i+1, len(response.Emails), email.Subject)
		fmt.Printf("        From: %s\n", email.From.Email)
		fmt.Printf("        Attachments: %d\n", len(email.Attachments))

		// Create message subdirectory
		messageDir := filepath.Join(downloadDir, sanitizeFilename(email.ID))
		if err := os.MkdirAll(messageDir, 0755); err != nil {
			log.Printf("Failed to create message directory: %v\n", err)
			continue
		}

		// Download each attachment
		for j, att := range email.Attachments {
			fmt.Printf("  [%d/%d] %s (%s)\n", j+1, len(email.Attachments), att.Filename, formatBytes(att.Size))

			data, err := client.GetAttachment(ctx, email.ID, att.ID)
			if err != nil {
				fmt.Printf("        ❌ Failed: %v\n", err)
				continue
			}

			filename := filepath.Join(messageDir, sanitizeFilename(att.Filename))
			if err := os.WriteFile(filename, data, 0644); err != nil {
				fmt.Printf("        ❌ Failed to save: %v\n", err)
				continue
			}

			fmt.Printf("        ✓ Downloaded: %s\n", filename)
			totalDownloaded++
			totalBytes += int64(len(data))
		}
		fmt.Println()
	}

	fmt.Printf("Summary: Downloaded %d attachments (%s total)\n", totalDownloaded, formatBytes(totalBytes))
	fmt.Printf("Location: %s\n", downloadDir)
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

// Helper functions

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// sanitizeFilename removes invalid characters from filenames.
// Used by downloadAttachments().
//
//nolint:unused // Intentionally unused until downloadAttachments is enabled
func sanitizeFilename(filename string) string {
	// Replace invalid characters with underscores
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := filename

	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}

	// Remove leading/trailing spaces and dots
	result = strings.TrimSpace(result)
	result = strings.Trim(result, ".")

	// Ensure it's not empty
	if result == "" {
		result = "unnamed"
	}

	return result
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
