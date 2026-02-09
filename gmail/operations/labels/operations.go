package labels

import (
	"context"
	"fmt"

	"github.com/edaniel30/mailbridge-go/gmail/internal"
	"github.com/edaniel30/mailbridge-go/gmail/operations"
	"google.golang.org/api/gmail/v1"
)

const (
	// Label visibility constants for creating labels
	labelListVisibilityShow   = "labelShow"
	messageListVisibilityShow = "show"
	labelTypeUser             = "user"
)

// Label represents a Gmail label (folder/tag)
type Label struct {
	ID   string
	Name string
	Type string
}

// ListLabels lists all labels in the user's mailbox
func ListLabels(ctx context.Context, service internal.GmailService) ([]*Label, error) {
	labelsService := service.GetUsersService().GetLabelsService()
	call := labelsService.List(operations.UserIDMe)
	resp, err := call.Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list labels: %w", err)
	}

	labels := make([]*Label, 0, len(resp.Labels))
	for _, l := range resp.Labels {
		labels = append(labels, &Label{
			ID:   l.Id,
			Name: l.Name,
			Type: l.Type,
		})
	}

	return labels, nil
}

// FindLabelByName finds a label by its name
func FindLabelByName(ctx context.Context, service internal.GmailService, name string) (*Label, error) {
	labels, err := ListLabels(ctx, service)
	if err != nil {
		return nil, fmt.Errorf("failed to find label by name %s: %w", name, err)
	}

	for _, label := range labels {
		if label.Name == name {
			return label, nil
		}
	}

	return nil, fmt.Errorf("label not found: %s", name)
}

// findOrCreateLabel finds a label by name, creating it if it doesn't exist.
// This is a helper function used by MoveMessageToFolder.
func findOrCreateLabel(ctx context.Context, service internal.GmailService, name string) (*Label, error) {
	label, err := FindLabelByName(ctx, service, name)
	if err != nil {
		// Label doesn't exist, create it
		label, err = CreateLabel(ctx, service, name)
		if err != nil {
			return nil, fmt.Errorf("failed to create label: %w", err)
		}
	}
	return label, nil
}

// CreateLabel creates a new label (folder)
func CreateLabel(ctx context.Context, service internal.GmailService, name string) (*Label, error) {
	label := &gmail.Label{
		Name:                  name,
		LabelListVisibility:   labelListVisibilityShow,
		MessageListVisibility: messageListVisibilityShow,
		Type:                  labelTypeUser,
	}

	labelsService := service.GetUsersService().GetLabelsService()
	call := labelsService.Create(operations.UserIDMe, label)
	created, err := call.Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create label: %w", err)
	}

	return &Label{
		ID:   created.Id,
		Name: created.Name,
		Type: created.Type,
	}, nil
}

// MarkAsRead marks a message as read
func MarkAsRead(ctx context.Context, service internal.GmailService, messageID string) error {
	req := &gmail.ModifyMessageRequest{
		RemoveLabelIds: []string{"UNREAD"},
	}

	messagesService := service.GetUsersService().GetMessagesService()
	call := messagesService.Modify(operations.UserIDMe, messageID, req)
	_, err := call.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to mark message as read: %w", err)
	}

	return nil
}

// MoveMessageToFolder moves a message to a specific folder/label
// Creates the label if it doesn't exist
func MoveMessageToFolder(ctx context.Context, service internal.GmailService, messageID string, folderName string) error {
	// Find or create the label
	label, err := findOrCreateLabel(ctx, service, folderName)
	if err != nil {
		return err
	}

	// Remove INBOX label and add new label
	req := &gmail.ModifyMessageRequest{
		AddLabelIds:    []string{label.ID},
		RemoveLabelIds: []string{"INBOX"},
	}

	messagesService := service.GetUsersService().GetMessagesService()
	call := messagesService.Modify(operations.UserIDMe, messageID, req)
	_, err = call.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to move message: %w", err)
	}

	return nil
}
