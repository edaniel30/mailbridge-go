package messages

import (
	"context"
	"fmt"

	"github.com/danielrivera/mailbridge-go/core"
	"github.com/danielrivera/mailbridge-go/gmail/internal"
	"github.com/danielrivera/mailbridge-go/gmail/operations"
)

// ListMessages lists messages from Gmail
func ListMessages(ctx context.Context, service internal.GmailService, opts *core.ListOptions) (*core.ListResponse, error) {
	messagesService := operations.GetMessagesService(service)
	call := messagesService.List(operations.UserIDMe)

	if opts != nil {
		if opts.MaxResults > 0 {
			call = call.MaxResults(opts.MaxResults)
		}
		if opts.PageToken != "" {
			call = call.PageToken(opts.PageToken)
		}
		if opts.Query != "" {
			call = call.Q(opts.Query)
		}
		if len(opts.Labels) > 0 {
			call = call.LabelIds(opts.Labels...)
		}
	}

	resp, err := call.Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	emails := make([]*core.Email, 0, len(resp.Messages))
	for _, msg := range resp.Messages {
		email, err := GetMessage(ctx, service, msg.Id)
		if err != nil {
			// Skip messages that can't be retrieved
			continue
		}
		emails = append(emails, email)
	}

	return &core.ListResponse{
		Emails:        emails,
		NextPageToken: resp.NextPageToken,
		TotalCount:    resp.ResultSizeEstimate,
	}, nil
}
