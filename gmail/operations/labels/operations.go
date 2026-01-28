package labels

import (
	"context"
	"fmt"
	"strings"

	"github.com/danielrivera/mailbridge-go/gmail/internal"
	"github.com/danielrivera/mailbridge-go/gmail/operations"
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

// GetLabel gets a specific label by ID
func GetLabel(ctx context.Context, service internal.GmailService, labelID string) (*Label, error) {
	labelsService := service.GetUsersService().GetLabelsService()
	call := labelsService.Get(operations.UserIDMe, labelID)
	l, err := call.Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get label: %w", err)
	}

	return &Label{
		ID:   l.Id,
		Name: l.Name,
		Type: l.Type,
	}, nil
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
// This is a helper function to avoid duplication in MoveMessageToFolder and BatchMoveToFolder.
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

// DeleteLabel deletes a label
func DeleteLabel(ctx context.Context, service internal.GmailService, labelID string) error {
	labelsService := service.GetUsersService().GetLabelsService()
	call := labelsService.Delete(operations.UserIDMe, labelID)
	err := call.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to delete label: %w", err)
	}

	return nil
}

// AddLabelToMessage adds a label to a message
func AddLabelToMessage(ctx context.Context, service internal.GmailService, messageID string, labelID string) error {
	req := &gmail.ModifyMessageRequest{
		AddLabelIds: []string{labelID},
	}

	messagesService := service.GetUsersService().GetMessagesService()
	call := messagesService.Modify(operations.UserIDMe, messageID, req)
	_, err := call.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to add label to message: %w", err)
	}

	return nil
}

// RemoveLabelFromMessage removes a label from a message
func RemoveLabelFromMessage(ctx context.Context, service internal.GmailService, messageID string, labelID string) error {
	req := &gmail.ModifyMessageRequest{
		RemoveLabelIds: []string{labelID},
	}

	messagesService := service.GetUsersService().GetMessagesService()
	call := messagesService.Modify(operations.UserIDMe, messageID, req)
	_, err := call.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to remove label from message: %w", err)
	}

	return nil
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

// MarkAsUnread marks a message as unread
func MarkAsUnread(ctx context.Context, service internal.GmailService, messageID string) error {
	req := &gmail.ModifyMessageRequest{
		AddLabelIds: []string{"UNREAD"},
	}

	messagesService := service.GetUsersService().GetMessagesService()
	call := messagesService.Modify(operations.UserIDMe, messageID, req)
	_, err := call.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to mark message as unread: %w", err)
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

// TrashMessage moves a message to trash (reversible)
func TrashMessage(ctx context.Context, service internal.GmailService, messageID string) error {
	messagesService := service.GetUsersService().GetMessagesService()
	call := messagesService.Trash(operations.UserIDMe, messageID)
	_, err := call.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to trash message: %w", err)
	}

	return nil
}

// UntrashMessage removes a message from trash
func UntrashMessage(ctx context.Context, service internal.GmailService, messageID string) error {
	messagesService := service.GetUsersService().GetMessagesService()
	call := messagesService.Untrash(operations.UserIDMe, messageID)
	_, err := call.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to untrash message: %w", err)
	}

	return nil
}

// BatchTrashMessages moves multiple messages to trash
func BatchTrashMessages(ctx context.Context, service internal.GmailService, messageIDs []string) error {
	return operations.BatchOperation(ctx, messageIDs, func(ctx context.Context, messageID string) error {
		return TrashMessage(ctx, service, messageID)
	}, "trash")
}

// BatchModifyMessages modifies labels on multiple messages
func BatchModifyMessages(ctx context.Context, service internal.GmailService, messageIDs []string, addLabelIDs []string, removeLabelIDs []string) error {
	if len(messageIDs) == 0 {
		return nil
	}

	// Gmail does not have a native batch modify API, so we modify each message individually
	var errors []string
	for _, messageID := range messageIDs {
		req := &gmail.ModifyMessageRequest{
			AddLabelIds:    addLabelIDs,
			RemoveLabelIds: removeLabelIDs,
		}

		messagesService := service.GetUsersService().GetMessagesService()
		call := messagesService.Modify(operations.UserIDMe, messageID, req)
		_, err := call.Context(ctx).Do()
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", messageID, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to modify %d messages: %s", len(errors), strings.Join(errors, "; "))
	}

	return nil
}

// BatchMarkAsRead marks multiple messages as read
func BatchMarkAsRead(ctx context.Context, service internal.GmailService, messageIDs []string) error {
	return operations.BatchOperation(ctx, messageIDs, func(ctx context.Context, messageID string) error {
		return MarkAsRead(ctx, service, messageID)
	}, "mark as read")
}

// BatchMarkAsUnread marks multiple messages as unread
func BatchMarkAsUnread(ctx context.Context, service internal.GmailService, messageIDs []string) error {
	return operations.BatchOperation(ctx, messageIDs, func(ctx context.Context, messageID string) error {
		return MarkAsUnread(ctx, service, messageID)
	}, "mark as unread")
}

// BatchMoveToFolder moves multiple messages to a specific folder/label
func BatchMoveToFolder(ctx context.Context, service internal.GmailService, messageIDs []string, folderName string) error {
	if len(messageIDs) == 0 {
		return nil
	}

	// Find or create the label
	label, err := findOrCreateLabel(ctx, service, folderName)
	if err != nil {
		return err
	}

	// Move all messages to the folder
	var errors []string
	for _, messageID := range messageIDs {
		req := &gmail.ModifyMessageRequest{
			AddLabelIds:    []string{label.ID},
			RemoveLabelIds: []string{"INBOX"},
		}

		messagesService := service.GetUsersService().GetMessagesService()
		call := messagesService.Modify(operations.UserIDMe, messageID, req)
		_, err := call.Context(ctx).Do()
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", messageID, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to move %d messages: %s", len(errors), strings.Join(errors, "; "))
	}

	return nil
}
