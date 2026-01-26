package gmail

import (
	"context"
	"fmt"

	"google.golang.org/api/gmail/v1"
)

// Label represents a Gmail label (folder)
type Label struct {
	ID   string
	Name string
	Type string
}

// ListLabels lists all labels in the user's mailbox
func (c *Client) ListLabels(ctx context.Context) ([]*Label, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("client not connected")
	}

	labelsService := c.service.GetUsersService().GetLabelsService()
	call := labelsService.List("me")
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
func (c *Client) GetLabel(ctx context.Context, labelID string) (*Label, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("client not connected")
	}

	labelsService := c.service.GetUsersService().GetLabelsService()
	call := labelsService.Get("me", labelID)
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
func (c *Client) FindLabelByName(ctx context.Context, name string) (*Label, error) {
	labels, err := c.ListLabels(ctx)
	if err != nil {
		return nil, err
	}

	for _, label := range labels {
		if label.Name == name {
			return label, nil
		}
	}

	return nil, fmt.Errorf("label not found: %s", name)
}

// CreateLabel creates a new label (folder)
func (c *Client) CreateLabel(ctx context.Context, name string) (*Label, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("client not connected")
	}

	label := &gmail.Label{
		Name:                  name,
		LabelListVisibility:   "labelShow",
		MessageListVisibility: "show",
		Type:                  "user",
	}

	labelsService := c.service.GetUsersService().GetLabelsService()
	call := labelsService.Create("me", label)
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
func (c *Client) DeleteLabel(ctx context.Context, labelID string) error {
	if !c.IsConnected() {
		return fmt.Errorf("client not connected")
	}

	labelsService := c.service.GetUsersService().GetLabelsService()
	call := labelsService.Delete("me", labelID)
	err := call.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to delete label: %w", err)
	}

	return nil
}

// AddLabelToMessage adds a label to a message
func (c *Client) AddLabelToMessage(ctx context.Context, messageID string, labelID string) error {
	if !c.IsConnected() {
		return fmt.Errorf("client not connected")
	}

	req := &gmail.ModifyMessageRequest{
		AddLabelIds: []string{labelID},
	}

	messagesService := c.service.GetUsersService().GetMessagesService()
	call := messagesService.Modify("me", messageID, req)
	_, err := call.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to add label to message: %w", err)
	}

	return nil
}

// RemoveLabelFromMessage removes a label from a message
func (c *Client) RemoveLabelFromMessage(ctx context.Context, messageID string, labelID string) error {
	if !c.IsConnected() {
		return fmt.Errorf("client not connected")
	}

	req := &gmail.ModifyMessageRequest{
		RemoveLabelIds: []string{labelID},
	}

	messagesService := c.service.GetUsersService().GetMessagesService()
	call := messagesService.Modify("me", messageID, req)
	_, err := call.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to remove label from message: %w", err)
	}

	return nil
}

// MarkAsRead marks a message as read
func (c *Client) MarkAsRead(ctx context.Context, messageID string) error {
	if !c.IsConnected() {
		return fmt.Errorf("client not connected")
	}

	req := &gmail.ModifyMessageRequest{
		RemoveLabelIds: []string{"UNREAD"},
	}

	messagesService := c.service.GetUsersService().GetMessagesService()
	call := messagesService.Modify("me", messageID, req)
	_, err := call.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to mark message as read: %w", err)
	}

	return nil
}

// MarkAsUnread marks a message as unread
func (c *Client) MarkAsUnread(ctx context.Context, messageID string) error {
	if !c.IsConnected() {
		return fmt.Errorf("client not connected")
	}

	req := &gmail.ModifyMessageRequest{
		AddLabelIds: []string{"UNREAD"},
	}

	messagesService := c.service.GetUsersService().GetMessagesService()
	call := messagesService.Modify("me", messageID, req)
	_, err := call.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to mark message as unread: %w", err)
	}

	return nil
}

// MoveMessageToFolder moves a message to a specific folder/label
// Creates the label if it doesn't exist
func (c *Client) MoveMessageToFolder(ctx context.Context, messageID string, folderName string) error {
	if !c.IsConnected() {
		return fmt.Errorf("client not connected")
	}

	// Try to find existing label
	label, err := c.FindLabelByName(ctx, folderName)
	if err != nil {
		// Label doesn't exist, create it
		label, err = c.CreateLabel(ctx, folderName)
		if err != nil {
			return fmt.Errorf("failed to create label: %w", err)
		}
	}

	// Remove INBOX label and add new label
	req := &gmail.ModifyMessageRequest{
		AddLabelIds:    []string{label.ID},
		RemoveLabelIds: []string{"INBOX"},
	}

	messagesService := c.service.GetUsersService().GetMessagesService()
	call := messagesService.Modify("me", messageID, req)
	_, err = call.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to move message: %w", err)
	}

	return nil
}
