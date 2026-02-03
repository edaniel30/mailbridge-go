package operations

import (
	"context"
	"fmt"
	"strings"

	"github.com/danielrivera/mailbridge-go/gmail/internal"
)

// UserIDMe is the special user ID that represents the authenticated user in Gmail API.
const UserIDMe = "me"

// BatchOperation executes a batch operation on multiple messages with error aggregation.
// It continues processing all messages even if some fail, and returns an aggregated error.
func BatchOperation(
	ctx context.Context,
	messageIDs []string,
	operation func(ctx context.Context, messageID string) error,
	operationName string,
) error {
	if len(messageIDs) == 0 {
		return nil
	}

	var errors []string
	for _, messageID := range messageIDs {
		if err := operation(ctx, messageID); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", messageID, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to %s %d messages: %s", operationName, len(errors), strings.Join(errors, "; "))
	}

	return nil
}

// GetMessagesService is a helper to get the MessagesService from a GmailService.
// This centralizes the common pattern of accessing the messages service.
func GetMessagesService(service internal.GmailService) internal.MessagesService {
	return service.GetUsersService().GetMessagesService()
}
