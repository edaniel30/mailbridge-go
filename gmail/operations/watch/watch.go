package watch

import (
	"context"
	"fmt"
	"strconv"

	"github.com/danielrivera/mailbridge-go/core"
	"github.com/danielrivera/mailbridge-go/gmail/internal"
	"github.com/danielrivera/mailbridge-go/gmail/operations"
	"google.golang.org/api/gmail/v1"
)

// WatchMailbox sets up push notifications for the mailbox
func WatchMailbox(ctx context.Context, service internal.GmailService, req *core.WatchRequest) (*core.WatchResponse, error) {
	if req.TopicName == "" {
		return nil, fmt.Errorf("topic name is required")
	}

	gmailReq := &gmail.WatchRequest{
		TopicName:           req.TopicName,
		LabelIds:            req.LabelIDs,
		LabelFilterBehavior: req.LabelFilterBehavior,
	}

	usersService := service.GetUsersService()
	call := usersService.Watch(operations.UserIDMe, gmailReq)
	resp, err := call.Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to watch mailbox: %w", err)
	}

	return &core.WatchResponse{
		HistoryID:  fmt.Sprintf("%d", resp.HistoryId),
		Expiration: resp.Expiration,
	}, nil
}

// StopWatch stops push notifications for the mailbox
func StopWatch(ctx context.Context, service internal.GmailService) error {
	usersService := service.GetUsersService()
	call := usersService.Stop(operations.UserIDMe)
	err := call.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to stop watch: %w", err)
	}

	return nil
}

// GetHistory retrieves mailbox history starting from a history ID
func GetHistory(ctx context.Context, service internal.GmailService, req *core.HistoryRequest) (*core.HistoryResponse, error) {
	if req.StartHistoryID == "" {
		return nil, fmt.Errorf("start history ID is required")
	}

	// Convert StartHistoryID string to uint64
	historyID, err := strconv.ParseUint(req.StartHistoryID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid start history ID: %w", err)
	}

	usersService := service.GetUsersService()
	call := usersService.GetHistory(operations.UserIDMe)
	call = call.StartHistoryId(historyID)

	if req.MaxResults > 0 {
		call = call.MaxResults(req.MaxResults)
	}

	if req.PageToken != "" {
		call = call.PageToken(req.PageToken)
	}

	if req.LabelID != "" {
		call = call.LabelId(req.LabelID)
	}

	if len(req.HistoryTypes) > 0 {
		call = call.HistoryTypes(req.HistoryTypes...)
	}

	resp, err := call.Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get history: %w", err)
	}

	// Convert history records
	historyRecords := make([]*core.HistoryRecord, 0, len(resp.History))
	for _, h := range resp.History {
		record := &core.HistoryRecord{
			ID: fmt.Sprintf("%d", h.Id),
		}

		// Convert messages added
		if len(h.MessagesAdded) > 0 {
			record.MessagesAdded = make([]*core.HistoryMessageAdded, 0, len(h.MessagesAdded))
			for _, ma := range h.MessagesAdded {
				if ma.Message != nil {
					record.MessagesAdded = append(record.MessagesAdded, &core.HistoryMessageAdded{
						Message: convertBasicMessage(ma.Message),
					})
				}
			}
		}

		// Convert messages deleted
		if len(h.MessagesDeleted) > 0 {
			record.MessagesDeleted = make([]*core.HistoryMessageDeleted, 0, len(h.MessagesDeleted))
			for _, md := range h.MessagesDeleted {
				if md.Message != nil {
					record.MessagesDeleted = append(record.MessagesDeleted, &core.HistoryMessageDeleted{
						Message: convertBasicMessage(md.Message),
					})
				}
			}
		}

		// Convert labels added
		if len(h.LabelsAdded) > 0 {
			record.LabelsAdded = make([]*core.HistoryLabelChange, 0, len(h.LabelsAdded))
			for _, la := range h.LabelsAdded {
				if la.Message != nil {
					record.LabelsAdded = append(record.LabelsAdded, &core.HistoryLabelChange{
						Message:  convertBasicMessage(la.Message),
						LabelIDs: la.LabelIds,
					})
				}
			}
		}

		// Convert labels removed
		if len(h.LabelsRemoved) > 0 {
			record.LabelsRemoved = make([]*core.HistoryLabelChange, 0, len(h.LabelsRemoved))
			for _, lr := range h.LabelsRemoved {
				if lr.Message != nil {
					record.LabelsRemoved = append(record.LabelsRemoved, &core.HistoryLabelChange{
						Message:  convertBasicMessage(lr.Message),
						LabelIDs: lr.LabelIds,
					})
				}
			}
		}

		historyRecords = append(historyRecords, record)
	}

	return &core.HistoryResponse{
		History:       historyRecords,
		NextPageToken: resp.NextPageToken,
		HistoryID:     fmt.Sprintf("%d", resp.HistoryId),
	}, nil
}

// convertBasicMessage converts a Gmail message to core.Email with basic info only
// (history responses don't include full message details)
func convertBasicMessage(msg *gmail.Message) *core.Email {
	email := &core.Email{
		ID:       msg.Id,
		ThreadID: msg.ThreadId,
		Snippet:  msg.Snippet,
		Labels:   msg.LabelIds,
	}

	return email
}
