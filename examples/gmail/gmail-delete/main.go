package main

// Usage:
// export GMAIL_CLIENT_ID="your-client-id"
// export GMAIL_CLIENT_SECRET="your-client-secret"
// go run main.go

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/danielrivera/mailbridge-go/core"
	"github.com/danielrivera/mailbridge-go/gmail"
	"golang.org/x/oauth2"
)

func main() {
	ctx := context.Background()

	// Load configuration from environment
	config := &gmail.Config{
		ClientID:     os.Getenv("GMAIL_CLIENT_ID"),
		ClientSecret: os.Getenv("GMAIL_CLIENT_SECRET"),
		RedirectURL:  "http://localhost",
		Scopes:       gmail.DefaultScopes(),
	}

	// Create client
	client, err := gmail.New(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Error closing client: %v", err)
		}
	}()

	// Load or create token
	token, err := loadToken()
	if err != nil {
		log.Printf("No existing token found. Starting OAuth flow...")
		token, err = performOAuthFlow(ctx, client)
		if err != nil {
			log.Printf("OAuth flow failed: %v", err)
			return
		}
		if err := saveToken(token); err != nil {
			log.Printf("Warning: Failed to save token: %v", err)
		}
	}

	// Connect with token
	if err := client.ConnectWithToken(ctx, token); err != nil {
		log.Printf("Failed to connect: %v", err)
		return
	}

	fmt.Println("✅ Connected to Gmail successfully!")
	fmt.Println()

	// First, let's list some messages to work with
	fmt.Println("📬 Listing recent messages in TRASH (if any)...")
	trashMessages, err := client.ListMessages(ctx, &core.ListOptions{
		MaxResults: 5,
		Labels:     []string{"TRASH"},
	})
	if err != nil {
		log.Printf("Failed to list trash messages: %v", err)
	} else {
		fmt.Printf("Found %d messages in trash\n", len(trashMessages.Emails))
		for i, msg := range trashMessages.Emails {
			fmt.Printf("  %d. %s - %s\n", i+1, msg.ID, msg.Subject)
		}
	}
	fmt.Println()

	// Example 1: Trash a message (reversible)
	fmt.Println("🗑️  Example 1: Moving a message to trash (reversible)")
	demonstrateTrashMessage(ctx, client)

	// Example 2: Untrash a message
	fmt.Println("\n♻️  Example 2: Restoring a message from trash")
	demonstrateUntrashMessage(ctx, client)

	// Example 3: Batch trash multiple messages
	fmt.Println("\n🗑️  Example 3: Batch trash multiple messages")
	demonstrateBatchTrash(ctx, client)

	// Example 4: Batch modify messages (add/remove labels)
	fmt.Println("\n🏷️  Example 4: Batch modify messages (add/remove labels)")
	demonstrateBatchModify(ctx, client)

	// Example 5: Batch mark as read
	fmt.Println("\n✉️  Example 5: Batch mark messages as read")
	demonstrateBatchMarkAsRead(ctx, client)

	// Example 6: Batch move to folder
	fmt.Println("\n📁 Example 6: Batch move messages to folder")
	demonstrateBatchMoveToFolder(ctx, client)

	fmt.Println("\n✅ All examples completed!")
}

func demonstrateTrashMessage(ctx context.Context, client *gmail.Client) {
	// List messages in INBOX to find one to trash
	messages, err := client.ListMessages(ctx, &core.ListOptions{
		MaxResults: 1,
		Labels:     []string{"INBOX"},
	})
	if err != nil {
		log.Printf("❌ Failed to list messages: %v", err)
		return
	}

	if len(messages.Emails) == 0 {
		fmt.Println("ℹ️  No messages in INBOX to trash")
		return
	}

	messageID := messages.Emails[0].ID
	subject := messages.Emails[0].Subject

	fmt.Printf("Moving message to trash:\n")
	fmt.Printf("  ID: %s\n", messageID)
	fmt.Printf("  Subject: %s\n", subject)

	err = client.TrashMessage(ctx, messageID)
	if err != nil {
		log.Printf("❌ Failed to trash message: %v", err)
		return
	}

	fmt.Println("✅ Message moved to trash successfully!")
	fmt.Println("   (Can be restored with UntrashMessage)")
}

func demonstrateUntrashMessage(ctx context.Context, client *gmail.Client) {
	// List messages in TRASH to find one to restore
	messages, err := client.ListMessages(ctx, &core.ListOptions{
		MaxResults: 1,
		Labels:     []string{"TRASH"},
	})
	if err != nil {
		log.Printf("❌ Failed to list trash messages: %v", err)
		return
	}

	if len(messages.Emails) == 0 {
		fmt.Println("ℹ️  No messages in trash to restore")
		return
	}

	messageID := messages.Emails[0].ID
	subject := messages.Emails[0].Subject

	fmt.Printf("Restoring message from trash:\n")
	fmt.Printf("  ID: %s\n", messageID)
	fmt.Printf("  Subject: %s\n", subject)

	err = client.UntrashMessage(ctx, messageID)
	if err != nil {
		log.Printf("❌ Failed to untrash message: %v", err)
		return
	}

	fmt.Println("✅ Message restored from trash successfully!")
	fmt.Println("   (Message is back in INBOX)")
}

func demonstrateBatchTrash(ctx context.Context, client *gmail.Client) {
	// List multiple messages to trash
	messages, err := client.ListMessages(ctx, &core.ListOptions{
		MaxResults: 3,
		Query:      "is:read", // Only trash read messages
	})
	if err != nil {
		log.Printf("❌ Failed to list messages: %v", err)
		return
	}

	if len(messages.Emails) == 0 {
		fmt.Println("ℹ️  No messages found to batch trash")
		return
	}

	messageIDs := make([]string, 0, len(messages.Emails))
	fmt.Printf("Moving %d messages to trash:\n", len(messages.Emails))
	for i, msg := range messages.Emails {
		messageIDs = append(messageIDs, msg.ID)
		fmt.Printf("  %d. %s - %s\n", i+1, msg.ID, msg.Subject)
	}

	err = client.BatchTrashMessages(ctx, messageIDs)
	if err != nil {
		log.Printf("❌ Failed to batch trash: %v", err)
		return
	}

	fmt.Printf("✅ Successfully moved %d messages to trash!\n", len(messageIDs))
}

func performOAuthFlow(ctx context.Context, client *gmail.Client) (*oauth2.Token, error) {
	authURL := client.GetAuthURL("state")
	fmt.Printf("\nVisit this URL to authorize the application:\n%s\n\n", authURL)
	fmt.Print("Enter the authorization code: ")

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		return nil, fmt.Errorf("failed to read code: %w", err)
	}

	token, err := client.ExchangeCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	return token, nil
}

func loadToken() (*oauth2.Token, error) {
	data, err := os.ReadFile("../token.json")
	if err != nil {
		return nil, err
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}

	return &token, nil
}

func saveToken(token *oauth2.Token) error {
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}

	return os.WriteFile("../token.json", data, 0600)
}

func demonstrateBatchModify(ctx context.Context, client *gmail.Client) {
	// List some unread messages
	messages, err := client.ListMessages(ctx, &core.ListOptions{
		MaxResults: 3,
		Query:      "is:unread",
	})
	if err != nil {
		log.Printf("❌ Failed to list messages: %v", err)
		return
	}

	if len(messages.Emails) == 0 {
		fmt.Println("ℹ️  No unread messages found to batch modify")
		return
	}

	messageIDs := make([]string, 0, len(messages.Emails))
	fmt.Printf("Batch modifying %d messages:\n", len(messages.Emails))
	for i, msg := range messages.Emails {
		messageIDs = append(messageIDs, msg.ID)
		fmt.Printf("  %d. %s - %s\n", i+1, msg.ID, msg.Subject)
	}

	// Get or create a label
	fmt.Println("Creating/finding 'Processed' label...")
	labels, err := client.ListLabels(ctx)
	if err != nil {
		log.Printf("❌ Failed to list labels: %v", err)
		return
	}

	var labelID string
	for _, label := range labels {
		if label.Name == "Processed" {
			labelID = label.ID
			break
		}
	}

	if labelID == "" {
		newLabel, err := client.CreateLabel(ctx, "Processed")
		if err != nil {
			log.Printf("❌ Failed to create label: %v", err)
			return
		}
		labelID = newLabel.ID
		fmt.Printf("✅ Created new label 'Processed' with ID: %s\n", labelID)
	} else {
		fmt.Printf("✅ Found existing label 'Processed' with ID: %s\n", labelID)
	}

	// Batch modify: Add "Processed" label and remove "UNREAD" label
	err = client.BatchModifyMessages(ctx, &core.BatchModifyRequest{
		MessageIDs:     messageIDs,
		AddLabelIDs:    []string{labelID},
		RemoveLabelIDs: []string{"UNREAD"},
	})
	if err != nil {
		log.Printf("❌ Failed to batch modify: %v", err)
		return
	}

	fmt.Printf("✅ Successfully modified %d messages!\n", len(messageIDs))
	fmt.Println("   - Added 'Processed' label")
	fmt.Println("   - Removed 'UNREAD' label")
}

func demonstrateBatchMarkAsRead(ctx context.Context, client *gmail.Client) {
	// List unread messages
	messages, err := client.ListMessages(ctx, &core.ListOptions{
		MaxResults: 5,
		Query:      "is:unread",
	})
	if err != nil {
		log.Printf("❌ Failed to list messages: %v", err)
		return
	}

	if len(messages.Emails) == 0 {
		fmt.Println("ℹ️  No unread messages to mark as read")
		return
	}

	messageIDs := make([]string, 0, len(messages.Emails))
	fmt.Printf("Marking %d messages as read:\n", len(messages.Emails))
	for i, msg := range messages.Emails {
		messageIDs = append(messageIDs, msg.ID)
		fmt.Printf("  %d. %s - %s\n", i+1, msg.ID, msg.Subject)
	}

	err = client.BatchMarkAsRead(ctx, messageIDs)
	if err != nil {
		log.Printf("❌ Failed to batch mark as read: %v", err)
		return
	}

	fmt.Printf("✅ Successfully marked %d messages as read!\n", len(messageIDs))
}

func demonstrateBatchMoveToFolder(ctx context.Context, client *gmail.Client) {
	// List some messages from INBOX
	messages, err := client.ListMessages(ctx, &core.ListOptions{
		MaxResults: 3,
		Labels:     []string{"INBOX"},
		Query:      "is:read", // Only move read messages
	})
	if err != nil {
		log.Printf("❌ Failed to list messages: %v", err)
		return
	}

	if len(messages.Emails) == 0 {
		fmt.Println("ℹ️  No messages found to move")
		return
	}

	messageIDs := make([]string, 0, len(messages.Emails))
	folderName := "Archive-Demo"
	fmt.Printf("Moving %d messages to '%s' folder:\n", len(messages.Emails), folderName)
	for i, msg := range messages.Emails {
		messageIDs = append(messageIDs, msg.ID)
		fmt.Printf("  %d. %s - %s\n", i+1, msg.ID, msg.Subject)
	}

	err = client.BatchMoveToFolder(ctx, messageIDs, folderName)
	if err != nil {
		log.Printf("❌ Failed to batch move: %v", err)
		return
	}

	fmt.Printf("✅ Successfully moved %d messages to '%s'!\n", len(messageIDs), folderName)
	fmt.Println("   (Label will be created if it doesn't exist)")
}
