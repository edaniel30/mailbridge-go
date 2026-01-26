package gmail

import (
	"context"
	"errors"
	"testing"

	gmailtest "github.com/danielrivera/mailbridge-go/gmail/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/gmail/v1"
)

func TestClient_ListLabels_Success(t *testing.T) {
	client := newTestClient(t)

	mockGmailService, mockLabelsService := setupMockLabelsService()
	mockLabelsListCall := &gmailtest.MockLabelsListCall{}

	mockLabelsService.On("List", "me").Return(mockLabelsListCall)
	mockLabelsListCall.On("Context", context.Background()).Return(mockLabelsListCall)
	mockLabelsListCall.On("Do").Return(&gmail.ListLabelsResponse{
		Labels: []*gmail.Label{
			{Id: "INBOX", Name: "INBOX", Type: "system"},
			{Id: "label-1", Name: "Custom", Type: "user"},
		},
	}, nil)

	client.SetService(mockGmailService)

	labels, err := client.ListLabels(context.Background())

	require.NoError(t, err)
	assert.Len(t, labels, 2)
	assert.Equal(t, "INBOX", labels[0].ID)
	assert.Equal(t, "Custom", labels[1].Name)
}

func TestClient_ListLabels_NotConnected(t *testing.T) {
	client := newTestClient(t)

	_, err := client.ListLabels(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestClient_GetLabel_Success(t *testing.T) {
	client := newTestClient(t)

	mockGmailService, mockLabelsService := setupMockLabelsService()
	mockLabelsGetCall := &gmailtest.MockLabelsGetCall{}

	mockLabelsService.On("Get", "me", "label-1").Return(mockLabelsGetCall)
	mockLabelsGetCall.On("Context", context.Background()).Return(mockLabelsGetCall)
	mockLabelsGetCall.On("Do").Return(&gmail.Label{
		Id:   "label-1",
		Name: "Work",
		Type: "user",
	}, nil)

	client.SetService(mockGmailService)

	label, err := client.GetLabel(context.Background(), "label-1")

	require.NoError(t, err)
	assert.Equal(t, "label-1", label.ID)
	assert.Equal(t, "Work", label.Name)
	assert.Equal(t, "user", label.Type)
}

func TestClient_CreateLabel_Success(t *testing.T) {
	client := newTestClient(t)

	mockGmailService, mockLabelsService := setupMockLabelsService()
	mockLabelsCreateCall := &gmailtest.MockLabelsCreateCall{}

	mockLabelsService.On("Create", "me", &gmail.Label{
		Name:                  "Projects",
		LabelListVisibility:   "labelShow",
		MessageListVisibility: "show",
		Type:                  "user",
	}).Return(mockLabelsCreateCall)
	mockLabelsCreateCall.On("Context", context.Background()).Return(mockLabelsCreateCall)
	mockLabelsCreateCall.On("Do").Return(&gmail.Label{
		Id:   "label-new",
		Name: "Projects",
		Type: "user",
	}, nil)

	client.SetService(mockGmailService)

	label, err := client.CreateLabel(context.Background(), "Projects")

	require.NoError(t, err)
	assert.Equal(t, "label-new", label.ID)
	assert.Equal(t, "Projects", label.Name)
}

func TestClient_DeleteLabel_Success(t *testing.T) {
	client := newTestClient(t)

	mockGmailService, mockLabelsService := setupMockLabelsService()
	mockLabelsDeleteCall := &gmailtest.MockLabelsDeleteCall{}

	mockLabelsService.On("Delete", "me", "label-1").Return(mockLabelsDeleteCall)
	mockLabelsDeleteCall.On("Context", context.Background()).Return(mockLabelsDeleteCall)
	mockLabelsDeleteCall.On("Do").Return(nil)

	client.SetService(mockGmailService)

	err := client.DeleteLabel(context.Background(), "label-1")

	require.NoError(t, err)
}

func TestClient_MarkReadStatus(t *testing.T) {
	tests := []struct {
		name    string
		method  func(*Client, context.Context, string) error
		request *gmail.ModifyMessageRequest
	}{
		{
			name:   "MarkAsRead",
			method: (*Client).MarkAsRead,
			request: &gmail.ModifyMessageRequest{
				RemoveLabelIds: []string{"UNREAD"},
			},
		},
		{
			name:   "MarkAsUnread",
			method: (*Client).MarkAsUnread,
			request: &gmail.ModifyMessageRequest{
				AddLabelIds: []string{"UNREAD"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(t)

			mockGmailService, mockMessagesService := setupMockMessagesService()
			mockMessagesModifyCall := &gmailtest.MockMessagesModifyCall{}

			mockMessagesService.On("Modify", "me", "msg-123", tt.request).Return(mockMessagesModifyCall)
			mockMessagesModifyCall.On("Context", context.Background()).Return(mockMessagesModifyCall)
			mockMessagesModifyCall.On("Do").Return(&gmail.Message{Id: "msg-123"}, nil)

			client.SetService(mockGmailService)

			err := tt.method(client, context.Background(), "msg-123")

			require.NoError(t, err)
		})
	}
}

func TestClient_MoveMessageToFolder_ExistingLabel(t *testing.T) {
	client := newTestClient(t)

	mockGmailService, mockLabelsService, mockMessagesService := setupMockLabelsAndMessagesService()
	mockLabelsListCall := &gmailtest.MockLabelsListCall{}
	mockMessagesModifyCall := &gmailtest.MockMessagesModifyCall{}

	// Setup FindLabelByName expectations
	mockLabelsService.On("List", "me").Return(mockLabelsListCall)
	mockLabelsListCall.On("Context", context.Background()).Return(mockLabelsListCall)
	mockLabelsListCall.On("Do").Return(&gmail.ListLabelsResponse{
		Labels: []*gmail.Label{
			{Id: "label-work", Name: "Work", Type: "user"},
		},
	}, nil)

	// Setup MoveMessage expectations
	mockMessagesService.On("Modify", "me", "msg-123", &gmail.ModifyMessageRequest{
		AddLabelIds:    []string{"label-work"},
		RemoveLabelIds: []string{"INBOX"},
	}).Return(mockMessagesModifyCall)
	mockMessagesModifyCall.On("Context", context.Background()).Return(mockMessagesModifyCall)
	mockMessagesModifyCall.On("Do").Return(&gmail.Message{Id: "msg-123"}, nil)

	client.SetService(mockGmailService)

	err := client.MoveMessageToFolder(context.Background(), "msg-123", "Work")

	require.NoError(t, err)
}

func TestClient_MoveMessageToFolder_CreateNewLabel(t *testing.T) {
	client := newTestClient(t)

	mockGmailService, mockLabelsService, mockMessagesService := setupMockLabelsAndMessagesService()
	mockLabelsListCall := &gmailtest.MockLabelsListCall{}
	mockLabelsCreateCall := &gmailtest.MockLabelsCreateCall{}
	mockMessagesModifyCall := &gmailtest.MockMessagesModifyCall{}

	// Setup FindLabelByName expectations (label not found)
	mockLabelsService.On("List", "me").Return(mockLabelsListCall)
	mockLabelsListCall.On("Context", context.Background()).Return(mockLabelsListCall)
	mockLabelsListCall.On("Do").Return(&gmail.ListLabelsResponse{
		Labels: []*gmail.Label{},
	}, nil)

	// Setup CreateLabel expectations
	mockLabelsService.On("Create", "me", &gmail.Label{
		Name:                  "NewFolder",
		LabelListVisibility:   "labelShow",
		MessageListVisibility: "show",
		Type:                  "user",
	}).Return(mockLabelsCreateCall)
	mockLabelsCreateCall.On("Context", context.Background()).Return(mockLabelsCreateCall)
	mockLabelsCreateCall.On("Do").Return(&gmail.Label{
		Id:   "label-new",
		Name: "NewFolder",
		Type: "user",
	}, nil)

	// Setup MoveMessage expectations
	mockMessagesService.On("Modify", "me", "msg-123", &gmail.ModifyMessageRequest{
		AddLabelIds:    []string{"label-new"},
		RemoveLabelIds: []string{"INBOX"},
	}).Return(mockMessagesModifyCall)
	mockMessagesModifyCall.On("Context", context.Background()).Return(mockMessagesModifyCall)
	mockMessagesModifyCall.On("Do").Return(&gmail.Message{Id: "msg-123"}, nil)

	client.SetService(mockGmailService)

	err := client.MoveMessageToFolder(context.Background(), "msg-123", "NewFolder")

	require.NoError(t, err)
}

func TestClient_FindLabelByName_NotFound(t *testing.T) {
	client := newTestClient(t)

	mockGmailService, mockLabelsService := setupMockLabelsService()
	mockLabelsListCall := &gmailtest.MockLabelsListCall{}

	mockLabelsService.On("List", "me").Return(mockLabelsListCall)
	mockLabelsListCall.On("Context", context.Background()).Return(mockLabelsListCall)
	mockLabelsListCall.On("Do").Return(&gmail.ListLabelsResponse{
		Labels: []*gmail.Label{
			{Id: "label-1", Name: "Other", Type: "user"},
		},
	}, nil)

	client.SetService(mockGmailService)

	_, err := client.FindLabelByName(context.Background(), "NonExistent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "label not found")
}

func TestClient_ModifyMessageLabels(t *testing.T) {
	tests := []struct {
		name    string
		method  func(*Client, context.Context, string, string) error
		request *gmail.ModifyMessageRequest
	}{
		{
			name:   "AddLabelToMessage",
			method: (*Client).AddLabelToMessage,
			request: &gmail.ModifyMessageRequest{
				AddLabelIds: []string{"label-1"},
			},
		},
		{
			name:   "RemoveLabelFromMessage",
			method: (*Client).RemoveLabelFromMessage,
			request: &gmail.ModifyMessageRequest{
				RemoveLabelIds: []string{"label-1"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(t)

			mockGmailService, mockMessagesService := setupMockMessagesService()
			mockMessagesModifyCall := &gmailtest.MockMessagesModifyCall{}

			mockMessagesService.On("Modify", "me", "msg-123", tt.request).Return(mockMessagesModifyCall)
			mockMessagesModifyCall.On("Context", context.Background()).Return(mockMessagesModifyCall)
			mockMessagesModifyCall.On("Do").Return(&gmail.Message{Id: "msg-123"}, nil)

			client.SetService(mockGmailService)

			err := tt.method(client, context.Background(), "msg-123", "label-1")

			require.NoError(t, err)
		})
	}
}

func TestClient_LabelsOperations_NotConnected(t *testing.T) {
	client := newTestClient(t)

	t.Run("GetLabel", func(t *testing.T) {
		_, err := client.GetLabel(context.Background(), "label-1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})

	t.Run("CreateLabel", func(t *testing.T) {
		_, err := client.CreateLabel(context.Background(), "Test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})

	t.Run("DeleteLabel", func(t *testing.T) {
		err := client.DeleteLabel(context.Background(), "label-1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})

	t.Run("AddLabelToMessage", func(t *testing.T) {
		err := client.AddLabelToMessage(context.Background(), "msg-1", "label-1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})

	t.Run("RemoveLabelFromMessage", func(t *testing.T) {
		err := client.RemoveLabelFromMessage(context.Background(), "msg-1", "label-1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})

	t.Run("MarkAsRead", func(t *testing.T) {
		err := client.MarkAsRead(context.Background(), "msg-1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})

	t.Run("MarkAsUnread", func(t *testing.T) {
		err := client.MarkAsUnread(context.Background(), "msg-1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})

	t.Run("MoveMessageToFolder", func(t *testing.T) {
		err := client.MoveMessageToFolder(context.Background(), "msg-1", "Folder")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})
}

func TestClient_LabelsAPIErrors(t *testing.T) {
	t.Run("ListLabels error", func(t *testing.T) {
		client := newTestClient(t)
		mockGmailService, mockLabelsService := setupMockLabelsService()
		mockLabelsListCall := &gmailtest.MockLabelsListCall{}

		mockLabelsService.On("List", "me").Return(mockLabelsListCall)
		mockLabelsListCall.On("Context", context.Background()).Return(mockLabelsListCall)
		mockLabelsListCall.On("Do").Return(nil, errors.New("API error"))

		client.SetService(mockGmailService)

		_, err := client.ListLabels(context.Background())
		assert.Error(t, err)
	})

	t.Run("GetLabel error", func(t *testing.T) {
		client := newTestClient(t)
		mockGmailService, mockLabelsService := setupMockLabelsService()
		mockLabelsGetCall := &gmailtest.MockLabelsGetCall{}

		mockLabelsService.On("Get", "me", "label-1").Return(mockLabelsGetCall)
		mockLabelsGetCall.On("Context", context.Background()).Return(mockLabelsGetCall)
		mockLabelsGetCall.On("Do").Return(nil, errors.New("not found"))

		client.SetService(mockGmailService)

		_, err := client.GetLabel(context.Background(), "label-1")
		assert.Error(t, err)
	})

	t.Run("CreateLabel error", func(t *testing.T) {
		client := newTestClient(t)
		mockGmailService, mockLabelsService := setupMockLabelsService()
		mockLabelsCreateCall := &gmailtest.MockLabelsCreateCall{}

		mockLabelsService.On("Create", "me", &gmail.Label{
			Name:                  "Test",
			LabelListVisibility:   "labelShow",
			MessageListVisibility: "show",
			Type:                  "user",
		}).Return(mockLabelsCreateCall)
		mockLabelsCreateCall.On("Context", context.Background()).Return(mockLabelsCreateCall)
		mockLabelsCreateCall.On("Do").Return(nil, errors.New("creation failed"))

		client.SetService(mockGmailService)

		_, err := client.CreateLabel(context.Background(), "Test")
		assert.Error(t, err)
	})
}
