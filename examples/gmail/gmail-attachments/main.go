package main

// Gmail Attachments Download Example
//
// This example demonstrates how to:
// - Search for emails with attachments
// - List attachment metadata
// - Download attachment contents
// - Save attachments to disk
//
// Usage:
//   export GMAIL_CLIENT_ID="your-client-id"
//   export GMAIL_CLIENT_SECRET="your-client-secret"
//   go run main.go

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/danielrivera/mailbridge-go/core"
	"github.com/danielrivera/mailbridge-go/gmail"
	"golang.org/x/oauth2"
)

const (
	tokenFile        = "../token.json"
	downloadDir      = "attachments"
	maxMessagesToScan = 10
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run() error {
	ctx := context.Background()

	// Create Gmail configuration
	cfg := &gmail.Config{
		ClientID:     os.Getenv("GMAIL_CLIENT_ID"),
		ClientSecret: os.Getenv("GMAIL_CLIENT_SECRET"),
		RedirectURL:  "http://localhost",
		Scopes:       gmail.DefaultScopes(),
	}

	// Create Gmail client
	client, err := gmail.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create Gmail client: %w", err)
	}

	// Authenticate
	token := authenticateOrLoad(ctx, client)
	if token == nil {
		return fmt.Errorf("failed to authenticate")
	}

	if err := client.ConnectWithToken(ctx, token); err != nil {
		return fmt.Errorf("failed to connect with token: %w", err)
	}

	if !client.IsConnected() {
		return fmt.Errorf("failed to establish connection")
	}

	// Defer close after successful connection
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Warning: failed to close client: %v", err)
		}
	}()

	fmt.Println("✓ Successfully connected to Gmail!")
	fmt.Println()

	// Create download directory
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		return fmt.Errorf("failed to create download directory: %w", err)
	}

	// Search for messages with attachments
	fmt.Println("🔍 Searching for messages with attachments...")
	listOpts := &core.ListOptions{
		MaxResults: maxMessagesToScan,
		Query:      "has:attachment", // Gmail search: only messages with attachments
	}

	response, err := client.ListMessages(ctx, listOpts)
	if err != nil {
		return fmt.Errorf("failed to list messages: %w", err)
	}

	if len(response.Emails) == 0 {
		fmt.Println("No messages with attachments found.")
		return nil
	}

	fmt.Printf("Found %d messages with attachments\n\n", len(response.Emails))

	// Process each message
	totalAttachments := 0
	totalBytes := int64(0)

	for i, email := range response.Emails {
		if len(email.Attachments) == 0 {
			continue // Skip if no attachments (shouldn't happen with our query)
		}

		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("📧 Message %d/%d\n", i+1, len(response.Emails))
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("From:    %s <%s>\n", email.From.Name, email.From.Email)
		fmt.Printf("Subject: %s\n", email.Subject)
		fmt.Printf("Date:    %s\n", email.Date.Format("2006-01-02 15:04:05"))
		fmt.Printf("Attachments: %d\n\n", len(email.Attachments))

		// Create subdirectory for this message
		messageDir := filepath.Join(downloadDir, sanitizeFilename(email.ID))
		if err := os.MkdirAll(messageDir, 0755); err != nil {
			log.Printf("Failed to create message directory: %v", err)
			continue
		}

		// Download each attachment
		for j, att := range email.Attachments {
			fmt.Printf("  [%d/%d] %s\n", j+1, len(email.Attachments), att.Filename)
			fmt.Printf("        Type: %s\n", att.MimeType)
			fmt.Printf("        Size: %s\n", formatBytes(att.Size))

			// Download attachment data
			data, err := client.GetAttachment(ctx, email.ID, att.ID)
			if err != nil {
				fmt.Printf("        ❌ Failed to download: %v\n\n", err)
				continue
			}

			// Save to file
			filename := filepath.Join(messageDir, sanitizeFilename(att.Filename))
			if err := os.WriteFile(filename, data, 0644); err != nil {
				fmt.Printf("        ❌ Failed to save: %v\n\n", err)
				continue
			}

			fmt.Printf("        ✓ Downloaded: %s\n", filename)
			fmt.Printf("        Actual size: %s\n\n", formatBytes(int64(len(data))))

			totalAttachments++
			totalBytes += int64(len(data))
		}
	}

	// Summary
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("📊 Download Summary\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("Total attachments downloaded: %d\n", totalAttachments)
	fmt.Printf("Total size: %s\n", formatBytes(totalBytes))
	fmt.Printf("Download directory: %s\n", downloadDir)
	fmt.Println()

	return nil
}

// authenticateOrLoad handles OAuth2 authentication
func authenticateOrLoad(ctx context.Context, client *gmail.Client) *oauth2.Token {
	// Try to load existing token
	token := loadToken()

	if token == nil {
		// First time authentication
		authURL := client.GetAuthURL("state-token")
		fmt.Printf("Visit this URL to authorize the application:\n%s\n\n", authURL)

		fmt.Print("Enter the authorization code: ")
		var code string
		if _, err := fmt.Scanln(&code); err != nil {
			log.Printf("Failed to read code: %v", err)
			return nil
		}

		var err error
		token, err = client.ExchangeCode(ctx, code)
		if err != nil {
			log.Printf("Failed to exchange code: %v", err)
			return nil
		}

		saveToken(token)
		fmt.Println("✓ Token saved successfully!")
		fmt.Println()
	} else {
		fmt.Println("Using saved token...")
	}

	return token
}

// loadToken loads a token from file
func loadToken() *oauth2.Token {
	f, err := os.Open(tokenFile)
	if err != nil {
		return nil
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("Warning: failed to close token file: %v", err)
		}
	}()

	token := &oauth2.Token{}
	if err := json.NewDecoder(f).Decode(token); err != nil {
		return nil
	}

	return token
}

// saveToken saves a token to file
func saveToken(token *oauth2.Token) {
	f, err := os.Create(tokenFile)
	if err != nil {
		log.Printf("Warning: failed to save token: %v", err)
		return
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("Warning: failed to close token file: %v", err)
		}
	}()

	if err := json.NewEncoder(f).Encode(token); err != nil {
		log.Printf("Warning: failed to encode token: %v", err)
	}
}

// sanitizeFilename removes/replaces characters that are invalid in filenames
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

// formatBytes formats bytes to human-readable format
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
