package main

// Usage:
// export GMAIL_CLIENT_ID="your-client-id"
// export GMAIL_CLIENT_SECRET="your-client-secret"
// export GOOGLE_CLOUD_PROJECT="your-project-id"
// export PUBSUB_TOPIC="projects/your-project-id/topics/gmail-notifications"
//
// go run main.go start    # Start watching and consuming notifications
// go run main.go stop     # Stop watching

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/pubsub/v2"
	"github.com/danielrivera/mailbridge-go/core"
	"github.com/danielrivera/mailbridge-go/gmail"
	"golang.org/x/oauth2"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go [start|stop]")
		os.Exit(1)
	}

	command := os.Args[1]

	ctx := context.Background()

	// Load configuration
	config := &gmail.Config{
		ClientID:     os.Getenv("GMAIL_CLIENT_ID"),
		ClientSecret: os.Getenv("GMAIL_CLIENT_SECRET"),
		RedirectURL:  "http://localhost",
		Scopes:       gmail.DefaultScopes(),
	}

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
			log.Fatalf("OAuth flow failed: %v", err)
		}
		if err := saveToken(token); err != nil {
			log.Printf("Warning: Failed to save token: %v", err)
		}
	}

	if err := client.ConnectWithToken(ctx, token); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	switch command {
	case "start":
		runNotificationConsumer(ctx, client)
	case "stop":
		stopWatch(ctx, client)
	default:
		fmt.Println("Invalid command. Use: start or stop")
		os.Exit(1)
	}
}

func runNotificationConsumer(ctx context.Context, client *gmail.Client) {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	topicName := os.Getenv("PUBSUB_TOPIC")

	if projectID == "" || topicName == "" {
		log.Fatal("GOOGLE_CLOUD_PROJECT and PUBSUB_TOPIC environment variables are required")
	}

	// Step 1: Setup Gmail watch
	fmt.Println("🔔 Setting up Gmail watch...")
	watchResp, err := client.WatchMailbox(ctx, &core.WatchRequest{
		TopicName: topicName,
		LabelIDs:  []string{"INBOX"},
	})
	if err != nil {
		log.Fatalf("Failed to setup watch: %v", err)
	}

	expirationTime := time.Unix(watchResp.Expiration/1000, 0)
	fmt.Printf("✅ Watch established successfully!\n")
	fmt.Printf("   History ID: %s\n", watchResp.HistoryID)
	fmt.Printf("   Expires: %s\n", expirationTime.Format(time.RFC3339))
	fmt.Println()

	// Step 2: Create Pub/Sub client
	pubsubClient, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create Pub/Sub client: %v", err)
	}
	defer pubsubClient.Close()

	subscriptionID := "gmail-notifications-sub"
	lastHistoryID := watchResp.HistoryID

	fmt.Println("📬 Listening for notifications (Ctrl+C to stop)...")
	fmt.Println()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Create cancellable context
	consumerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Handle shutdown signal
	go func() {
		<-sigChan
		fmt.Println("\n\n⏹️  Shutting down gracefully...")
		cancel()
	}()

	// Step 3: Consume notifications from Pub/Sub
	subscriber := pubsubClient.Subscriber(subscriptionID)
	err = subscriber.Receive(consumerCtx, func(ctx context.Context, msg *pubsub.Message) {
		// Parse notification from Gmail
		var notification struct {
			EmailAddress string `json:"emailAddress"`
			HistoryID    uint64 `json:"historyId"`
		}
		if err := json.Unmarshal(msg.Data, &notification); err != nil {
			log.Printf("❌ Error parsing notification: %v", err)
			msg.Nack()
			return
		}

		fmt.Printf("🔔 Notification received!\n")
		fmt.Printf("   Email: %s\n", notification.EmailAddress)
		fmt.Printf("   History ID: %d\n", notification.HistoryID)

		// Step 4: Fetch changes using history API
		historyID := fmt.Sprintf("%d", notification.HistoryID)
		if historyID != lastHistoryID {
			processHistory(ctx, client, lastHistoryID, historyID)
			lastHistoryID = historyID
		} else {
			fmt.Println("   No new changes (same history ID)")
		}

		fmt.Println()
		msg.Ack()
	})

	if err != nil && err != context.Canceled {
		log.Printf("Error receiving messages: %v", err)
	}

	fmt.Println("\n✅ Consumer stopped successfully")
}

func stopWatch(ctx context.Context, client *gmail.Client) {
	fmt.Println("⏹️  Stopping Gmail watch...")
	if err := client.StopWatch(ctx); err != nil {
		log.Fatalf("Failed to stop watch: %v", err)
	}
	fmt.Println("✅ Watch stopped successfully!")
}

func processHistory(ctx context.Context, client *gmail.Client, startHistoryID, currentHistoryID string) {
	fmt.Printf("\n📜 Fetching history changes (from %s to %s)...\n", startHistoryID, currentHistoryID)

	history, err := client.GetHistory(ctx, &core.HistoryRequest{
		StartHistoryID: startHistoryID,
		MaxResults:     100,
	})
	if err != nil {
		log.Printf("❌ Error fetching history: %v", err)
		return
	}

	if len(history.History) == 0 {
		fmt.Println("   No changes found")
		return
	}

	fmt.Printf("   Found %d history record(s)\n", len(history.History))

	for i, record := range history.History {
		fmt.Printf("\n   Record #%d (ID: %s):\n", i+1, record.ID)

		// Messages added
		if len(record.MessagesAdded) > 0 {
			fmt.Printf("      📨 %d new message(s):\n", len(record.MessagesAdded))
			for _, added := range record.MessagesAdded {
				fmt.Printf("         • ID: %s\n", added.Message.ID)
				if added.Message.Snippet != "" {
					fmt.Printf("           Snippet: %s\n", added.Message.Snippet)
				}
			}
		}

		// Messages deleted
		if len(record.MessagesDeleted) > 0 {
			fmt.Printf("      🗑️  %d message(s) deleted:\n", len(record.MessagesDeleted))
			for _, deleted := range record.MessagesDeleted {
				fmt.Printf("         • ID: %s\n", deleted.Message.ID)
			}
		}

		// Labels added
		if len(record.LabelsAdded) > 0 {
			fmt.Printf("      🏷️  Labels added to %d message(s):\n", len(record.LabelsAdded))
			for _, labelChange := range record.LabelsAdded {
				fmt.Printf("         • Message ID: %s\n", labelChange.Message.ID)
				fmt.Printf("           Labels: %v\n", labelChange.LabelIDs)
			}
		}

		// Labels removed
		if len(record.LabelsRemoved) > 0 {
			fmt.Printf("      🏷️  Labels removed from %d message(s):\n", len(record.LabelsRemoved))
			for _, labelChange := range record.LabelsRemoved {
				fmt.Printf("         • Message ID: %s\n", labelChange.Message.ID)
				fmt.Printf("           Labels: %v\n", labelChange.LabelIDs)
			}
		}
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
