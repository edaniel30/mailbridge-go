package watch

import (
	"context"
	"errors"
	"testing"

	"github.com/danielrivera/mailbridge-go/core"
	gmailtest "github.com/danielrivera/mailbridge-go/gmail/testing"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/gmail/v1"
)

func TestWatchMailbox_Success(t *testing.T) {
	ctx := context.Background()

	// Create mocks
	mockService := &gmailtest.MockGmailService{}
	mockUsersService := &gmailtest.MockUsersService{}
	mockWatchCall := &gmailtest.MockUsersWatchCall{}

	// Setup expectations
	mockService.On("GetUsersService").Return(mockUsersService)
	mockUsersService.On("Watch", "me", &gmail.WatchRequest{
		TopicName:           "projects/test/topics/gmail",
		LabelIds:            []string{"INBOX"},
		LabelFilterBehavior: "",
	}).Return(mockWatchCall)
	mockWatchCall.On("Context", ctx).Return(mockWatchCall)
	mockWatchCall.On("Do").Return(&gmail.WatchResponse{
		HistoryId:  uint64(12345),
		Expiration: int64(1234567890000),
	}, nil)

	// Call function
	resp, err := WatchMailbox(ctx, mockService, &core.WatchRequest{
		TopicName: "projects/test/topics/gmail",
		LabelIDs:  []string{"INBOX"},
	})

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "12345", resp.HistoryID)
	assert.Equal(t, int64(1234567890000), resp.Expiration)

	mockService.AssertExpectations(t)
	mockUsersService.AssertExpectations(t)
	mockWatchCall.AssertExpectations(t)
}

func TestWatchMailbox_EmptyTopic(t *testing.T) {
	ctx := context.Background()
	mockService := &gmailtest.MockGmailService{}

	resp, err := WatchMailbox(ctx, mockService, &core.WatchRequest{
		TopicName: "",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "topic name is required")
}

func TestWatchMailbox_APIError(t *testing.T) {
	ctx := context.Background()

	mockService := &gmailtest.MockGmailService{}
	mockUsersService := &gmailtest.MockUsersService{}
	mockWatchCall := &gmailtest.MockUsersWatchCall{}

	mockService.On("GetUsersService").Return(mockUsersService)
	mockUsersService.On("Watch", "me", &gmail.WatchRequest{
		TopicName: "projects/test/topics/gmail",
	}).Return(mockWatchCall)
	mockWatchCall.On("Context", ctx).Return(mockWatchCall)
	mockWatchCall.On("Do").Return(nil, errors.New("API error"))

	resp, err := WatchMailbox(ctx, mockService, &core.WatchRequest{
		TopicName: "projects/test/topics/gmail",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to watch mailbox")
}

func TestStopWatch_Success(t *testing.T) {
	ctx := context.Background()

	mockService := &gmailtest.MockGmailService{}
	mockUsersService := &gmailtest.MockUsersService{}
	mockStopCall := &gmailtest.MockUsersStopCall{}

	mockService.On("GetUsersService").Return(mockUsersService)
	mockUsersService.On("Stop", "me").Return(mockStopCall)
	mockStopCall.On("Context", ctx).Return(mockStopCall)
	mockStopCall.On("Do").Return(nil)

	err := StopWatch(ctx, mockService)

	assert.NoError(t, err)
	mockService.AssertExpectations(t)
	mockUsersService.AssertExpectations(t)
	mockStopCall.AssertExpectations(t)
}

func TestStopWatch_APIError(t *testing.T) {
	ctx := context.Background()

	mockService := &gmailtest.MockGmailService{}
	mockUsersService := &gmailtest.MockUsersService{}
	mockStopCall := &gmailtest.MockUsersStopCall{}

	mockService.On("GetUsersService").Return(mockUsersService)
	mockUsersService.On("Stop", "me").Return(mockStopCall)
	mockStopCall.On("Context", ctx).Return(mockStopCall)
	mockStopCall.On("Do").Return(errors.New("API error"))

	err := StopWatch(ctx, mockService)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to stop watch")
}

func TestGetHistory_Success(t *testing.T) {
	ctx := context.Background()

	mockService := &gmailtest.MockGmailService{}
	mockUsersService := &gmailtest.MockUsersService{}
	mockHistoryCall := &gmailtest.MockUsersHistoryListCall{}

	mockService.On("GetUsersService").Return(mockUsersService)
	mockUsersService.On("GetHistory", "me").Return(mockHistoryCall)
	mockHistoryCall.On("StartHistoryId", uint64(100)).Return(mockHistoryCall)
	mockHistoryCall.On("Context", ctx).Return(mockHistoryCall)
	mockHistoryCall.On("Do").Return(&gmail.ListHistoryResponse{
		History: []*gmail.History{
			{
				Id: uint64(101),
				MessagesAdded: []*gmail.HistoryMessageAdded{
					{
						Message: &gmail.Message{
							Id:       "msg1",
							ThreadId: "thread1",
							Snippet:  "Test message",
						},
					},
				},
			},
		},
		HistoryId: uint64(101),
	}, nil)

	resp, err := GetHistory(ctx, mockService, &core.HistoryRequest{
		StartHistoryID: "100",
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "101", resp.HistoryID)
	assert.Len(t, resp.History, 1)
	assert.Len(t, resp.History[0].MessagesAdded, 1)
	assert.Equal(t, "msg1", resp.History[0].MessagesAdded[0].Message.ID)
}

func TestGetHistory_EmptyStartHistoryID(t *testing.T) {
	ctx := context.Background()
	mockService := &gmailtest.MockGmailService{}

	resp, err := GetHistory(ctx, mockService, &core.HistoryRequest{
		StartHistoryID: "",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "start history ID is required")
}

func TestGetHistory_InvalidHistoryID(t *testing.T) {
	ctx := context.Background()
	mockService := &gmailtest.MockGmailService{}

	resp, err := GetHistory(ctx, mockService, &core.HistoryRequest{
		StartHistoryID: "invalid",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "invalid start history ID")
}

func TestGetHistory_WithOptions(t *testing.T) {
	ctx := context.Background()

	mockService := &gmailtest.MockGmailService{}
	mockUsersService := &gmailtest.MockUsersService{}
	mockHistoryCall := &gmailtest.MockUsersHistoryListCall{}

	mockService.On("GetUsersService").Return(mockUsersService)
	mockUsersService.On("GetHistory", "me").Return(mockHistoryCall)
	mockHistoryCall.On("StartHistoryId", uint64(100)).Return(mockHistoryCall)
	mockHistoryCall.On("MaxResults", int64(50)).Return(mockHistoryCall)
	mockHistoryCall.On("PageToken", "token123").Return(mockHistoryCall)
	mockHistoryCall.On("LabelId", "INBOX").Return(mockHistoryCall)
	mockHistoryCall.On("HistoryTypes", []string{"messageAdded", "messageDeleted"}).Return(mockHistoryCall)
	mockHistoryCall.On("Context", ctx).Return(mockHistoryCall)
	mockHistoryCall.On("Do").Return(&gmail.ListHistoryResponse{
		History:       []*gmail.History{},
		HistoryId:     uint64(100),
		NextPageToken: "nextToken",
	}, nil)

	resp, err := GetHistory(ctx, mockService, &core.HistoryRequest{
		StartHistoryID: "100",
		MaxResults:     50,
		PageToken:      "token123",
		LabelID:        "INBOX",
		HistoryTypes:   []string{"messageAdded", "messageDeleted"},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "nextToken", resp.NextPageToken)
}

func TestConvertBasicMessage(t *testing.T) {
	gmailMsg := &gmail.Message{
		Id:       "msg123",
		ThreadId: "thread456",
		Snippet:  "Test snippet",
		LabelIds: []string{"INBOX", "UNREAD"},
	}

	email := convertBasicMessage(gmailMsg)

	assert.Equal(t, "msg123", email.ID)
	assert.Equal(t, "thread456", email.ThreadID)
	assert.Equal(t, "Test snippet", email.Snippet)
	assert.Equal(t, []string{"INBOX", "UNREAD"}, email.Labels)
}

func TestGetHistory_CompleteHistory(t *testing.T) {
	ctx := context.Background()

	mockService := &gmailtest.MockGmailService{}
	mockUsersService := &gmailtest.MockUsersService{}
	mockHistoryCall := &gmailtest.MockUsersHistoryListCall{}

	mockService.On("GetUsersService").Return(mockUsersService)
	mockUsersService.On("GetHistory", "me").Return(mockHistoryCall)
	mockHistoryCall.On("StartHistoryId", uint64(100)).Return(mockHistoryCall)
	mockHistoryCall.On("Context", ctx).Return(mockHistoryCall)
	mockHistoryCall.On("Do").Return(&gmail.ListHistoryResponse{
		History: []*gmail.History{
			{
				Id: uint64(101),
				MessagesAdded: []*gmail.HistoryMessageAdded{
					{Message: &gmail.Message{Id: "msg1"}},
				},
				MessagesDeleted: []*gmail.HistoryMessageDeleted{
					{Message: &gmail.Message{Id: "msg2"}},
				},
				LabelsAdded: []*gmail.HistoryLabelAdded{
					{
						Message:  &gmail.Message{Id: "msg3"},
						LabelIds: []string{"IMPORTANT"},
					},
				},
				LabelsRemoved: []*gmail.HistoryLabelRemoved{
					{
						Message:  &gmail.Message{Id: "msg4"},
						LabelIds: []string{"UNREAD"},
					},
				},
			},
		},
		HistoryId: uint64(101),
	}, nil)

	resp, err := GetHistory(ctx, mockService, &core.HistoryRequest{
		StartHistoryID: "100",
	})

	assert.NoError(t, err)
	assert.Len(t, resp.History, 1)
	assert.Len(t, resp.History[0].MessagesAdded, 1)
	assert.Len(t, resp.History[0].MessagesDeleted, 1)
	assert.Len(t, resp.History[0].LabelsAdded, 1)
	assert.Len(t, resp.History[0].LabelsRemoved, 1)
}
