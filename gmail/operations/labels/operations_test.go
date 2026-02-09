package labels

import (
	"context"
	"errors"
	"testing"

	"github.com/edaniel30/mailbridge-go/gmail/testing/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/gmail/v1"
)

func TestListLabels_Success(t *testing.T) {
	mockGmailService, mockLabelsService := mocks.SetupMockLabelsService()
	mockLabelsListCall := &mocks.MockLabelsListCall{}

	mockLabelsService.On("List", "me").Return(mockLabelsListCall)
	mockLabelsListCall.On("Context", context.Background()).Return(mockLabelsListCall)
	mockLabelsListCall.On("Do").Return(&gmail.ListLabelsResponse{
		Labels: []*gmail.Label{
			{Id: "INBOX", Name: "INBOX", Type: "system"},
			{Id: "label-1", Name: "Custom", Type: "user"},
		},
	}, nil)

	labels, err := ListLabels(context.Background(), mockGmailService)

	require.NoError(t, err)
	assert.Len(t, labels, 2)
	assert.Equal(t, "INBOX", labels[0].ID)
	assert.Equal(t, "Custom", labels[1].Name)
}

func TestListLabels_Error(t *testing.T) {
	mockGmailService, mockLabelsService := mocks.SetupMockLabelsService()
	mockLabelsListCall := &mocks.MockLabelsListCall{}

	mockLabelsService.On("List", "me").Return(mockLabelsListCall)
	mockLabelsListCall.On("Context", context.Background()).Return(mockLabelsListCall)
	mockLabelsListCall.On("Do").Return(nil, errors.New("API error"))

	_, err := ListLabels(context.Background(), mockGmailService)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list labels")
}

func TestFindLabelByName_Success(t *testing.T) {
	mockGmailService, mockLabelsService := mocks.SetupMockLabelsService()
	mockLabelsListCall := &mocks.MockLabelsListCall{}

	mockLabelsService.On("List", "me").Return(mockLabelsListCall)
	mockLabelsListCall.On("Context", context.Background()).Return(mockLabelsListCall)
	mockLabelsListCall.On("Do").Return(&gmail.ListLabelsResponse{
		Labels: []*gmail.Label{
			{Id: "INBOX", Name: "INBOX", Type: "system"},
			{Id: "label-1", Name: "Work", Type: "user"},
		},
	}, nil)

	label, err := FindLabelByName(context.Background(), mockGmailService, "Work")

	require.NoError(t, err)
	assert.Equal(t, "label-1", label.ID)
	assert.Equal(t, "Work", label.Name)
}

func TestFindLabelByName_NotFound(t *testing.T) {
	mockGmailService, mockLabelsService := mocks.SetupMockLabelsService()
	mockLabelsListCall := &mocks.MockLabelsListCall{}

	mockLabelsService.On("List", "me").Return(mockLabelsListCall)
	mockLabelsListCall.On("Context", context.Background()).Return(mockLabelsListCall)
	mockLabelsListCall.On("Do").Return(&gmail.ListLabelsResponse{
		Labels: []*gmail.Label{
			{Id: "INBOX", Name: "INBOX", Type: "system"},
		},
	}, nil)

	_, err := FindLabelByName(context.Background(), mockGmailService, "NonExistent")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "label not found")
}

func TestCreateLabel_Success(t *testing.T) {
	mockGmailService, mockLabelsService := mocks.SetupMockLabelsService()
	mockLabelsCreateCall := &mocks.MockLabelsCreateCall{}

	mockLabelsService.On("Create", "me", &gmail.Label{
		Name:                  "NewLabel",
		LabelListVisibility:   labelListVisibilityShow,
		MessageListVisibility: messageListVisibilityShow,
		Type:                  labelTypeUser,
	}).Return(mockLabelsCreateCall)
	mockLabelsCreateCall.On("Context", context.Background()).Return(mockLabelsCreateCall)
	mockLabelsCreateCall.On("Do").Return(&gmail.Label{
		Id:   "label-new",
		Name: "NewLabel",
		Type: "user",
	}, nil)

	label, err := CreateLabel(context.Background(), mockGmailService, "NewLabel")

	require.NoError(t, err)
	assert.Equal(t, "label-new", label.ID)
	assert.Equal(t, "NewLabel", label.Name)
}

func TestCreateLabel_Error(t *testing.T) {
	mockGmailService, mockLabelsService := mocks.SetupMockLabelsService()
	mockLabelsCreateCall := &mocks.MockLabelsCreateCall{}

	mockLabelsService.On("Create", "me", &gmail.Label{
		Name:                  "NewLabel",
		LabelListVisibility:   labelListVisibilityShow,
		MessageListVisibility: messageListVisibilityShow,
		Type:                  labelTypeUser,
	}).Return(mockLabelsCreateCall)
	mockLabelsCreateCall.On("Context", context.Background()).Return(mockLabelsCreateCall)
	mockLabelsCreateCall.On("Do").Return(nil, errors.New("API error"))

	_, err := CreateLabel(context.Background(), mockGmailService, "NewLabel")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create label")
}

func TestMarkAsRead_Success(t *testing.T) {
	mockGmailService, mockMessagesService := mocks.SetupMockMessagesService()
	mockModifyCall := &mocks.MockMessagesModifyCall{}

	mockMessagesService.On("Modify", "me", "msg-123", &gmail.ModifyMessageRequest{
		RemoveLabelIds: []string{"UNREAD"},
	}).Return(mockModifyCall)
	mockModifyCall.On("Context", context.Background()).Return(mockModifyCall)
	mockModifyCall.On("Do").Return(&gmail.Message{}, nil)

	err := MarkAsRead(context.Background(), mockGmailService, "msg-123")

	require.NoError(t, err)
}

func TestMarkAsRead_Error(t *testing.T) {
	mockGmailService, mockMessagesService := mocks.SetupMockMessagesService()
	mockModifyCall := &mocks.MockMessagesModifyCall{}

	mockMessagesService.On("Modify", "me", "msg-123", &gmail.ModifyMessageRequest{
		RemoveLabelIds: []string{"UNREAD"},
	}).Return(mockModifyCall)
	mockModifyCall.On("Context", context.Background()).Return(mockModifyCall)
	mockModifyCall.On("Do").Return(nil, errors.New("API error"))

	err := MarkAsRead(context.Background(), mockGmailService, "msg-123")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to mark message as read")
}

func TestMoveMessageToFolder_Success(t *testing.T) {
	mockGmailService := &mocks.MockGmailService{}
	mockUsersService := &mocks.MockUsersService{}
	mockLabelsService := &mocks.MockLabelsService{}
	mockMessagesService := &mocks.MockMessagesService{}

	// Setup service hierarchy
	mockGmailService.On("GetUsersService").Return(mockUsersService)
	mockUsersService.On("GetLabelsService").Return(mockLabelsService)
	mockUsersService.On("GetMessagesService").Return(mockMessagesService)

	// Mock ListLabels (label doesn't exist)
	mockLabelsListCall := &mocks.MockLabelsListCall{}
	mockLabelsService.On("List", "me").Return(mockLabelsListCall)
	mockLabelsListCall.On("Context", context.Background()).Return(mockLabelsListCall)
	mockLabelsListCall.On("Do").Return(&gmail.ListLabelsResponse{
		Labels: []*gmail.Label{},
	}, nil)

	// Mock CreateLabel
	mockLabelsCreateCall := &mocks.MockLabelsCreateCall{}
	mockLabelsService.On("Create", "me", &gmail.Label{
		Name:                  "Work",
		LabelListVisibility:   labelListVisibilityShow,
		MessageListVisibility: messageListVisibilityShow,
		Type:                  labelTypeUser,
	}).Return(mockLabelsCreateCall)
	mockLabelsCreateCall.On("Context", context.Background()).Return(mockLabelsCreateCall)
	mockLabelsCreateCall.On("Do").Return(&gmail.Label{
		Id:   "label-work",
		Name: "Work",
		Type: "user",
	}, nil)

	// Mock Modify
	mockModifyCall := &mocks.MockMessagesModifyCall{}
	mockMessagesService.On("Modify", "me", "msg-123", &gmail.ModifyMessageRequest{
		AddLabelIds:    []string{"label-work"},
		RemoveLabelIds: []string{"INBOX"},
	}).Return(mockModifyCall)
	mockModifyCall.On("Context", context.Background()).Return(mockModifyCall)
	mockModifyCall.On("Do").Return(&gmail.Message{}, nil)

	err := MoveMessageToFolder(context.Background(), mockGmailService, "msg-123", "Work")

	require.NoError(t, err)
}

func TestMoveMessageToFolder_Error(t *testing.T) {
	mockGmailService := &mocks.MockGmailService{}
	mockUsersService := &mocks.MockUsersService{}
	mockLabelsService := &mocks.MockLabelsService{}
	mockMessagesService := &mocks.MockMessagesService{}

	// Setup service hierarchy
	mockGmailService.On("GetUsersService").Return(mockUsersService)
	mockUsersService.On("GetLabelsService").Return(mockLabelsService)
	mockUsersService.On("GetMessagesService").Return(mockMessagesService)

	// Mock ListLabels
	mockLabelsListCall := &mocks.MockLabelsListCall{}
	mockLabelsService.On("List", "me").Return(mockLabelsListCall)
	mockLabelsListCall.On("Context", context.Background()).Return(mockLabelsListCall)
	mockLabelsListCall.On("Do").Return(&gmail.ListLabelsResponse{
		Labels: []*gmail.Label{
			{Id: "label-work", Name: "Work", Type: "user"},
		},
	}, nil)

	// Mock Modify (fails)
	mockModifyCall := &mocks.MockMessagesModifyCall{}
	mockMessagesService.On("Modify", "me", "msg-123", &gmail.ModifyMessageRequest{
		AddLabelIds:    []string{"label-work"},
		RemoveLabelIds: []string{"INBOX"},
	}).Return(mockModifyCall)
	mockModifyCall.On("Context", context.Background()).Return(mockModifyCall)
	mockModifyCall.On("Do").Return(nil, errors.New("API error"))

	err := MoveMessageToFolder(context.Background(), mockGmailService, "msg-123", "Work")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to move message")
}
