package messages

import (
	"github.com/danielrivera/mailbridge-go/gmail/operations"
	"context"
	"fmt"
	"strings"

	"github.com/danielrivera/mailbridge-go/gmail/internal"
)

// DeleteMessage permanently deletes a message (not reversible)
func DeleteMessage(ctx context.Context, service internal.GmailService, messageID string) error {
	messagesService := service.GetUsersService().GetMessagesService()
	call := messagesService.Delete(operations.UserIDMe, messageID)
	err := call.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}

// BatchDeleteMessages permanently deletes multiple messages
func BatchDeleteMessages(ctx context.Context, service internal.GmailService, messageIDs []string) error {
	if len(messageIDs) == 0 {
		return nil
	}

	// Gmail doesn't have a native batch delete API, so we delete each message individually
	// This could be optimized with goroutines for better performance
	var errors []string
	for _, messageID := range messageIDs {
		if err := DeleteMessage(ctx, service, messageID); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", messageID, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to delete %d messages: %s", len(errors), strings.Join(errors, "; "))
	}

	return nil
}
