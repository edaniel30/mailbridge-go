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

const tokenFile = "token.json"

func main() {
	ctx := context.Background()

	// Create Gmail configuration
	cfg := &gmail.Config{
		ClientID:     os.Getenv("GMAIL_CLIENT_ID"),
		ClientSecret: os.Getenv("GMAIL_CLIENT_SECRET"),
		RedirectURL:  "http://localhost", // Must match Google Cloud Console
		Scopes:       gmail.DefaultScopes(),
	}

	// Create Gmail client
	client, err := gmail.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create Gmail client: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Warning: failed to close client: %v", err)
		}
	}()

	// Check if we have a saved token
	token := loadToken()

	if token == nil {
		// Step 1: Get authorization URL
		authURL := client.GetAuthURL("state-token")
		fmt.Printf("Visit this URL to authorize the application:\n%s\n\n", authURL)

		// Step 2: Exchange code for token
		fmt.Print("Enter the authorization code: ")
		var code string
		if _, err := fmt.Scanln(&code); err != nil {
			log.Printf("Failed to read code: %v", err)
			return
		}

		token, err = client.ExchangeCode(ctx, code)
		if err != nil {
			log.Printf("Failed to exchange code: %v", err)
			return
		}

		// Save token for future use
		saveToken(token)
		fmt.Println("Token saved successfully!")
	} else {
		fmt.Println("Using saved token...")
	}

	// Step 3: Connect to Gmail API
	if err := client.ConnectWithToken(ctx, token); err != nil {
		log.Printf("Failed to connect with token, trying to refresh: %v", err)

		// Try to refresh the token
		newToken, err := client.RefreshToken(ctx)
		if err != nil {
			log.Printf("Failed to refresh token: %v", err)
			return
		}

		token = newToken
		saveToken(token)
		fmt.Println("Token refreshed and saved!")
	}

	if !client.IsConnected() {
		log.Println("Failed to establish connection")
		return
	}

	fmt.Println("✓ Successfully connected to Gmail!")

	// List unread messages
	fmt.Println("\n--- Listing Unread Messages ---")
	listOpts := &core.ListOptions{
		MaxResults: 5,                  // Get last 5 unread messages
		Labels:     []string{"UNREAD"}, // Only unread messages
	}

	// Other useful options:
	// listOpts.Query = "is:starred"           // Only starred
	// listOpts.Query = "from:example@test.com" // From specific sender
	// listOpts.Query = "subject:invoice"      // Subject contains "invoice"
	// listOpts.Query = "is:unread is:important" // Multiple conditions

	response, err := client.ListMessages(ctx, listOpts)
	if err != nil {
		log.Printf("Failed to list messages: %v", err)
	} else {
		fmt.Printf("\nFound %d unread messages (total estimate: %d)\n", len(response.Emails), response.TotalCount)

		for i, email := range response.Emails {
			fmt.Printf("\n[%d] ID: %s\n", i+1, email.ID)
			fmt.Printf("    From: %s <%s>\n", email.From.Name, email.From.Email)
			fmt.Printf("    Subject: %s\n", email.Subject)
			fmt.Printf("    Date: %s\n", email.Date.Format("2006-01-02 15:04:05"))
			fmt.Printf("    Snippet: %s\n", truncate(email.Snippet, 80))
			fmt.Printf("    Labels: %v\n", email.Labels)
			fmt.Printf("    Read: %v | Starred: %v\n", email.IsRead, email.IsStarred)
			if len(email.Attachments) > 0 {
				fmt.Printf("    Attachments: %d\n", len(email.Attachments))
			}
		}

		// Get full details of the first message
		if len(response.Emails) > 0 {
			fmt.Println("\n--- Getting Full Message Details ---")
			firstMessageID := response.Emails[0].ID
			fullEmail, err := client.GetMessage(ctx, firstMessageID)
			if err != nil {
				log.Printf("Failed to get message details: %v", err)
			} else {
				fmt.Printf("\nMessage ID: %s\n", fullEmail.ID)
				fmt.Printf("Thread ID: %s\n", fullEmail.ThreadID)
				fmt.Printf("Subject: %s\n", fullEmail.Subject)
				fmt.Printf("From: %s <%s>\n", fullEmail.From.Name, fullEmail.From.Email)

				if len(fullEmail.To) > 0 {
					fmt.Printf("To: ")
					for i, to := range fullEmail.To {
						if i > 0 {
							fmt.Printf(", ")
						}
						if to.Name != "" {
							fmt.Printf("%s <%s>", to.Name, to.Email)
						} else {
							fmt.Printf("%s", to.Email)
						}
					}
					fmt.Println()
				}

				fmt.Printf("Date: %s\n", fullEmail.Date.Format("2006-01-02 15:04:05"))
				fmt.Printf("\nBody (Text): %s\n", truncate(fullEmail.Body.Text, 200))

				if fullEmail.Body.HTML != "" {
					fmt.Printf("Body (HTML): %s\n", truncate(fullEmail.Body.HTML, 200))
				}

				if len(fullEmail.Attachments) > 0 {
					fmt.Printf("\nAttachments:\n")
					for i, att := range fullEmail.Attachments {
						fmt.Printf("  [%d] %s (%s, %d bytes)\n", i+1, att.Filename, att.MimeType, att.Size)
					}

					// Uncomment to download attachments:
					// fmt.Println("\n--- Downloading Attachments ---")
					// for _, att := range fullEmail.Attachments {
					// 	data, err := client.GetAttachment(ctx, fullEmail.ID, att.ID)
					// 	if err != nil {
					// 		log.Printf("Failed to download %s: %v", att.Filename, err)
					// 		continue
					// 	}
					// 	filename := fmt.Sprintf("download_%s", att.Filename)
					// 	if err := os.WriteFile(filename, data, 0644); err != nil {
					// 		log.Printf("Failed to save %s: %v", filename, err)
					// 		continue
					// 	}
					// 	fmt.Printf("✓ Downloaded: %s (%d bytes)\n", filename, len(data))
					// }
				}
			}

			// Move message to a custom folder and mark as read
			fmt.Println("\n--- Moving Message to Folder ---")
			folderName := "MailBridge/Processed"
			err = client.MoveMessageToFolder(ctx, firstMessageID, folderName)
			if err != nil {
				log.Printf("Failed to move message: %v", err)
			} else {
				fmt.Printf("✓ Message moved to folder: %s\n", folderName)

				// Mark as read
				err = client.MarkAsRead(ctx, firstMessageID)
				if err != nil {
					log.Printf("Failed to mark as read: %v", err)
				} else {
					fmt.Println("✓ Message marked as read")
				}
			}
		}
	}

	fmt.Printf("\nToken info:\n")
	fmt.Printf("- Type: %s\n", token.TokenType)
	fmt.Printf("- Expires: %v\n", token.Expiry)
	fmt.Printf("- Has Refresh Token: %v\n", token.RefreshToken != "")
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

// truncate truncates a string to a maximum length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
