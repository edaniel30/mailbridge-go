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
	"strings"
	"time"

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
	fmt.Println("🔍 Gmail Advanced Search Examples")
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Println()

	// Example 1: Unread from specific sender with attachment
	fmt.Println("📧 Example 1: Unread emails from boss with attachments")
	example1(ctx, client)

	// Example 2: Large attachments
	fmt.Println("\n💾 Example 2: Emails with large attachments (> 5MB)")
	example2(ctx, client)

	// Example 3: Recent invoices
	fmt.Println("\n📄 Example 3: Invoice emails from last 30 days")
	example3(ctx, client)

	// Example 4: Starred or important
	fmt.Println("\n⭐ Example 4: Starred OR Important messages")
	example4(ctx, client)

	// Example 5: Exclude specific sender
	fmt.Println("\n🚫 Example 5: Unread messages NOT from specific sender")
	example5(ctx, client)

	// Example 6: Complex query
	fmt.Println("\n🔧 Example 6: Complex search - Unread from client with PDF after date")
	example6(ctx, client)

	// Example 7: Categorized emails
	fmt.Println("\n📂 Example 7: Primary category unread emails")
	example7(ctx, client)

	fmt.Println("\n✅ All search examples completed!")
}

func example1(ctx context.Context, client *gmail.Client) {
	// Using QueryBuilder for readability
	query := gmail.NewQueryBuilder().
		IsUnread().
		From("boss@company.com").
		HasAttachment().
		Build()

	fmt.Printf("Query: %s\n", query)

	messages, err := client.ListMessages(ctx, &core.ListOptions{
		Query:      query,
		MaxResults: 5,
	})
	if err != nil {
		log.Printf("❌ Search failed: %v", err)
		return
	}

	fmt.Printf("Found %d messages\n", len(messages.Emails))
	for i, msg := range messages.Emails {
		fmt.Printf("  %d. %s - %s\n", i+1, msg.Subject, msg.From.Email)
	}
}

func example2(ctx context.Context, client *gmail.Client) {
	// Search for emails larger than 5MB
	query := gmail.NewQueryBuilder().
		LargerThan(gmail.MegaBytes(5)).
		HasAttachment().
		Build()

	fmt.Printf("Query: %s\n", query)

	messages, err := client.ListMessages(ctx, &core.ListOptions{
		Query:      query,
		MaxResults: 5,
	})
	if err != nil {
		log.Printf("❌ Search failed: %v", err)
		return
	}

	fmt.Printf("Found %d messages with large attachments\n", len(messages.Emails))
	for i, msg := range messages.Emails {
		fmt.Printf("  %d. %s (%d attachments)\n", i+1, msg.Subject, len(msg.Attachments))
	}
}

func example3(ctx context.Context, client *gmail.Client) {
	// Invoices from last 30 days
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	query := gmail.NewQueryBuilder().
		Subject("invoice").
		After(thirtyDaysAgo).
		Build()

	fmt.Printf("Query: %s\n", query)

	messages, err := client.ListMessages(ctx, &core.ListOptions{
		Query:      query,
		MaxResults: 10,
	})
	if err != nil {
		log.Printf("❌ Search failed: %v", err)
		return
	}

	fmt.Printf("Found %d invoice emails\n", len(messages.Emails))
	for i, msg := range messages.Emails {
		fmt.Printf("  %d. %s - %s\n", i+1, msg.Subject, msg.Date.Format("2006-01-02"))
	}
}

func example4(ctx context.Context, client *gmail.Client) {
	// Starred OR Important messages
	query := gmail.NewQueryBuilder().
		IsStarred().
		OR().
		IsImportant().
		Build()

	fmt.Printf("Query: %s\n", query)

	messages, err := client.ListMessages(ctx, &core.ListOptions{
		Query:      query,
		MaxResults: 10,
	})
	if err != nil {
		log.Printf("❌ Search failed: %v", err)
		return
	}

	fmt.Printf("Found %d starred or important messages\n", len(messages.Emails))
	for i, msg := range messages.Emails {
		starred := ""
		if msg.IsStarred {
			starred = " ⭐"
		}
		fmt.Printf("  %d. %s%s\n", i+1, msg.Subject, starred)
	}
}

func example5(ctx context.Context, client *gmail.Client) {
	// Exclude specific sender
	query := gmail.NewQueryBuilder().
		IsUnread().
		NOT().
		From("notifications@example.com").
		Build()

	fmt.Printf("Query: %s\n", query)

	messages, err := client.ListMessages(ctx, &core.ListOptions{
		Query:      query,
		MaxResults: 10,
	})
	if err != nil {
		log.Printf("❌ Search failed: %v", err)
		return
	}

	fmt.Printf("Found %d unread messages (excluding notifications)\n", len(messages.Emails))
	for i, msg := range messages.Emails {
		fmt.Printf("  %d. %s - from: %s\n", i+1, msg.Subject, msg.From.Email)
	}
}

func example6(ctx context.Context, client *gmail.Client) {
	// Complex query: Unread from client with PDF attachment after specific date
	lastMonth := time.Now().AddDate(0, -1, 0)

	query := gmail.NewQueryBuilder().
		IsUnread().
		From("client@company.com").
		Filename("pdf").
		After(lastMonth).
		InInbox().
		Build()

	fmt.Printf("Query: %s\n", query)

	messages, err := client.ListMessages(ctx, &core.ListOptions{
		Query:      query,
		MaxResults: 5,
	})
	if err != nil {
		log.Printf("❌ Search failed: %v", err)
		return
	}

	fmt.Printf("Found %d matching messages\n", len(messages.Emails))
	for i, msg := range messages.Emails {
		fmt.Printf("  %d. %s - %s\n", i+1, msg.Subject, msg.Date.Format("2006-01-02"))
		if len(msg.Attachments) > 0 {
			fmt.Printf("      Attachments: ")
			for j, att := range msg.Attachments {
				if j > 0 {
					fmt.Print(", ")
				}
				fmt.Printf("%s", att.Filename)
			}
			fmt.Println()
		}
	}
}

func example7(ctx context.Context, client *gmail.Client) {
	// Primary category unread emails
	query := gmail.NewQueryBuilder().
		Category("primary").
		IsUnread().
		Build()

	fmt.Printf("Query: %s\n", query)

	messages, err := client.ListMessages(ctx, &core.ListOptions{
		Query:      query,
		MaxResults: 10,
	})
	if err != nil {
		log.Printf("❌ Search failed: %v", err)
		return
	}

	fmt.Printf("Found %d unread primary emails\n", len(messages.Emails))
	for i, msg := range messages.Emails {
		fmt.Printf("  %d. %s - %s\n", i+1, msg.Subject, msg.From.Email)
	}
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
