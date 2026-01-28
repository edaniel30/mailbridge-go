package labels

import (
	"context"
	"errors"
	"testing"

	"github.com/danielrivera/mailbridge-go/gmail/internal"
	gmailtest "github.com/danielrivera/mailbridge-go/gmail/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/gmail/v1"
)

func TestListLabels_Success(t *testing.T) {
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

	labels, err := ListLabels(context.Background(), mockGmailService)

	require.NoError(t, err)
	assert.Len(t, labels, 2)
	assert.Equal(t, "INBOX", labels[0].ID)
	assert.Equal(t, "Custom", labels[1].Name)
}

func TestGetLabel_Success(t *testing.T) {
	mockGmailService, mockLabelsService := setupMockLabelsService()
	mockLabelsGetCall := &gmailtest.MockLabelsGetCall{}

	mockLabelsService.On("Get", "me", "label-1").Return(mockLabelsGetCall)
	mockLabelsGetCall.On("Context", context.Background()).Return(mockLabelsGetCall)
	mockLabelsGetCall.On("Do").Return(&gmail.Label{
		Id:   "label-1",
		Name: "Work",
		Type: "user",
	}, nil)

	label, err := GetLabel(context.Background(), mockGmailService, "label-1")

	require.NoError(t, err)
	assert.Equal(t, "label-1", label.ID)
	assert.Equal(t, "Work", label.Name)
	assert.Equal(t, "user", label.Type)
}

func TestCreateLabel_Success(t *testing.T) {
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

	label, err := CreateLabel(context.Background(), mockGmailService, "Projects")

	require.NoError(t, err)
	assert.Equal(t, "label-new", label.ID)
	assert.Equal(t, "Projects", label.Name)
}

func TestDeleteLabel_Success(t *testing.T) {
	mockGmailService, mockLabelsService := setupMockLabelsService()
	mockLabelsDeleteCall := &gmailtest.MockLabelsDeleteCall{}

	mockLabelsService.On("Delete", "me", "label-1").Return(mockLabelsDeleteCall)
	mockLabelsDeleteCall.On("Context", context.Background()).Return(mockLabelsDeleteCall)
	mockLabelsDeleteCall.On("Do").Return(nil)

	err := DeleteLabel(context.Background(), mockGmailService, "label-1")

	require.NoError(t, err)
}

func TestMarkReadStatus(t *testing.T) {
	tests := []struct {
		name    string
		method  func(context.Context, internal.GmailService, string) error
		request *gmail.ModifyMessageRequest
	}{
		{
			name:   "MarkAsRead",
			method: MarkAsRead,
			request: &gmail.ModifyMessageRequest{
				RemoveLabelIds: []string{"UNREAD"},
			},
		},
		{
			name:   "MarkAsUnread",
			method: MarkAsUnread,
			request: &gmail.ModifyMessageRequest{
				AddLabelIds: []string{"UNREAD"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGmailService, mockMessagesService := setupMockMessagesService()
			mockMessagesModifyCall := &gmailtest.MockMessagesModifyCall{}

			mockMessagesService.On("Modify", "me", "msg-123", tt.request).Return(mockMessagesModifyCall)
			mockMessagesModifyCall.On("Context", context.Background()).Return(mockMessagesModifyCall)
			mockMessagesModifyCall.On("Do").Return(&gmail.Message{Id: "msg-123"}, nil)

			err := tt.method(context.Background(), mockGmailService, "msg-123")

			require.NoError(t, err)
		})
	}
}

func TestMoveMessageToFolder_ExistingLabel(t *testing.T) {
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

	err := MoveMessageToFolder(context.Background(), mockGmailService, "msg-123", "Work")

	require.NoError(t, err)
}

func TestMoveMessageToFolder_CreateNewLabel(t *testing.T) {
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

	err := MoveMessageToFolder(context.Background(), mockGmailService, "msg-123", "NewFolder")

	require.NoError(t, err)
}

func TestFindLabelByName_NotFound(t *testing.T) {
	mockGmailService, mockLabelsService := setupMockLabelsService()
	mockLabelsListCall := &gmailtest.MockLabelsListCall{}

	mockLabelsService.On("List", "me").Return(mockLabelsListCall)
	mockLabelsListCall.On("Context", context.Background()).Return(mockLabelsListCall)
	mockLabelsListCall.On("Do").Return(&gmail.ListLabelsResponse{
		Labels: []*gmail.Label{
			{Id: "label-1", Name: "Other", Type: "user"},
		},
	}, nil)

	_, err := FindLabelByName(context.Background(), mockGmailService, "NonExistent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "label not found")
}

func TestModifyMessageLabels(t *testing.T) {
	tests := []struct {
		name    string
		method  func(context.Context, internal.GmailService, string, string) error
		request *gmail.ModifyMessageRequest
	}{
		{
			name:   "AddLabelToMessage",
			method: AddLabelToMessage,
			request: &gmail.ModifyMessageRequest{
				AddLabelIds: []string{"label-1"},
			},
		},
		{
			name:   "RemoveLabelFromMessage",
			method: RemoveLabelFromMessage,
			request: &gmail.ModifyMessageRequest{
				RemoveLabelIds: []string{"label-1"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGmailService, mockMessagesService := setupMockMessagesService()
			mockMessagesModifyCall := &gmailtest.MockMessagesModifyCall{}

			mockMessagesService.On("Modify", "me", "msg-123", tt.request).Return(mockMessagesModifyCall)
			mockMessagesModifyCall.On("Context", context.Background()).Return(mockMessagesModifyCall)
			mockMessagesModifyCall.On("Do").Return(&gmail.Message{Id: "msg-123"}, nil)

			err := tt.method(context.Background(), mockGmailService, "msg-123", "label-1")

			require.NoError(t, err)
		})
	}
}

func TestLabelsAPIErrors(t *testing.T) {
	t.Run("ListLabels error", func(t *testing.T) {
		mockGmailService, mockLabelsService := setupMockLabelsService()
		mockLabelsListCall := &gmailtest.MockLabelsListCall{}

		mockLabelsService.On("List", "me").Return(mockLabelsListCall)
		mockLabelsListCall.On("Context", context.Background()).Return(mockLabelsListCall)
		mockLabelsListCall.On("Do").Return(nil, errors.New("API error"))

		_, err := ListLabels(context.Background(), mockGmailService)
		assert.Error(t, err)
	})

	t.Run("GetLabel error", func(t *testing.T) {
		mockGmailService, mockLabelsService := setupMockLabelsService()
		mockLabelsGetCall := &gmailtest.MockLabelsGetCall{}

		mockLabelsService.On("Get", "me", "label-1").Return(mockLabelsGetCall)
		mockLabelsGetCall.On("Context", context.Background()).Return(mockLabelsGetCall)
		mockLabelsGetCall.On("Do").Return(nil, errors.New("not found"))

		_, err := GetLabel(context.Background(), mockGmailService, "label-1")
		assert.Error(t, err)
	})

	t.Run("CreateLabel error", func(t *testing.T) {
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

		_, err := CreateLabel(context.Background(), mockGmailService, "Test")
		assert.Error(t, err)
	})
}

// Helper functions to setup mock services
func setupMockLabelsService() (internal.GmailService, *gmailtest.MockLabelsService) {
	mockGmailService := &gmailtest.MockGmailService{}
	mockUsersService := &gmailtest.MockUsersService{}
	mockLabelsService := &gmailtest.MockLabelsService{}

	mockGmailService.On("GetUsersService").Return(mockUsersService)
	mockUsersService.On("GetLabelsService").Return(mockLabelsService)

	return mockGmailService, mockLabelsService
}

func setupMockMessagesService() (internal.GmailService, *gmailtest.MockMessagesService) {
	mockGmailService := &gmailtest.MockGmailService{}
	mockUsersService := &gmailtest.MockUsersService{}
	mockMessagesService := &gmailtest.MockMessagesService{}

	mockGmailService.On("GetUsersService").Return(mockUsersService)
	mockUsersService.On("GetMessagesService").Return(mockMessagesService)

	return mockGmailService, mockMessagesService
}

func setupMockLabelsAndMessagesService() (internal.GmailService, *gmailtest.MockLabelsService, *gmailtest.MockMessagesService) {
	mockGmailService := &gmailtest.MockGmailService{}
	mockUsersService := &gmailtest.MockUsersService{}
	mockLabelsService := &gmailtest.MockLabelsService{}
	mockMessagesService := &gmailtest.MockMessagesService{}

	mockGmailService.On("GetUsersService").Return(mockUsersService)
	mockUsersService.On("GetLabelsService").Return(mockLabelsService)
	mockUsersService.On("GetMessagesService").Return(mockMessagesService)

	return mockGmailService, mockLabelsService, mockMessagesService
}

func TestTrashMessage(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		messageID   string
		setupMock   func(*gmailtest.MockMessagesService)
		wantErr     bool
		expectedErr string
	}{
		{
			name:      "successful trash",
			messageID: "msg-123",
			setupMock: func(mockMessagesService *gmailtest.MockMessagesService) {
				mockTrashCall := &gmailtest.MockMessagesTrashCall{}
				mockMessagesService.On("Trash", "me", "msg-123").Return(mockTrashCall)
				mockTrashCall.On("Context", ctx).Return(mockTrashCall)
				mockTrashCall.On("Do").Return(&gmail.Message{Id: "msg-123"}, nil)
			},
			wantErr: false,
		},
		{
			name:      "trash fails",
			messageID: "msg-456",
			setupMock: func(mockMessagesService *gmailtest.MockMessagesService) {
				mockTrashCall := &gmailtest.MockMessagesTrashCall{}
				mockMessagesService.On("Trash", "me", "msg-456").Return(mockTrashCall)
				mockTrashCall.On("Context", ctx).Return(mockTrashCall)
				mockTrashCall.On("Do").Return(nil, errors.New("API error"))
			},
			wantErr:     true,
			expectedErr: "failed to trash message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGmailService, mockMessagesService := setupMockMessagesService()
			tt.setupMock(mockMessagesService)

			err := TrashMessage(ctx, mockGmailService, tt.messageID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}

			mockMessagesService.AssertExpectations(t)
		})
	}
}

func TestUntrashMessage(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		messageID   string
		setupMock   func(*gmailtest.MockMessagesService)
		wantErr     bool
		expectedErr string
	}{
		{
			name:      "successful untrash",
			messageID: "msg-123",
			setupMock: func(mockMessagesService *gmailtest.MockMessagesService) {
				mockUntrashCall := &gmailtest.MockMessagesUntrashCall{}
				mockMessagesService.On("Untrash", "me", "msg-123").Return(mockUntrashCall)
				mockUntrashCall.On("Context", ctx).Return(mockUntrashCall)
				mockUntrashCall.On("Do").Return(&gmail.Message{Id: "msg-123"}, nil)
			},
			wantErr: false,
		},
		{
			name:      "untrash fails",
			messageID: "msg-456",
			setupMock: func(mockMessagesService *gmailtest.MockMessagesService) {
				mockUntrashCall := &gmailtest.MockMessagesUntrashCall{}
				mockMessagesService.On("Untrash", "me", "msg-456").Return(mockUntrashCall)
				mockUntrashCall.On("Context", ctx).Return(mockUntrashCall)
				mockUntrashCall.On("Do").Return(nil, errors.New("API error"))
			},
			wantErr:     true,
			expectedErr: "failed to untrash message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGmailService, mockMessagesService := setupMockMessagesService()
			tt.setupMock(mockMessagesService)

			err := UntrashMessage(ctx, mockGmailService, tt.messageID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}

			mockMessagesService.AssertExpectations(t)
		})
	}
}

func TestBatchTrashMessages(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		messageIDs  []string
		setupMock   func(*gmailtest.MockMessagesService)
		wantErr     bool
		expectedErr string
	}{
		{
			name:       "empty slice",
			messageIDs: []string{},
			setupMock:  func(mockMessagesService *gmailtest.MockMessagesService) {},
			wantErr:    false,
		},
		{
			name:       "successful batch trash",
			messageIDs: []string{"msg-1", "msg-2", "msg-3"},
			setupMock: func(mockMessagesService *gmailtest.MockMessagesService) {
				for _, msgID := range []string{"msg-1", "msg-2", "msg-3"} {
					mockTrashCall := &gmailtest.MockMessagesTrashCall{}
					mockMessagesService.On("Trash", "me", msgID).Return(mockTrashCall).Once()
					mockTrashCall.On("Context", ctx).Return(mockTrashCall).Once()
					mockTrashCall.On("Do").Return(&gmail.Message{Id: msgID}, nil).Once()
				}
			},
			wantErr: false,
		},
		{
			name:       "partial failure",
			messageIDs: []string{"msg-1", "msg-2", "msg-3"},
			setupMock: func(mockMessagesService *gmailtest.MockMessagesService) {
				// First succeeds
				mockTrashCall1 := &gmailtest.MockMessagesTrashCall{}
				mockMessagesService.On("Trash", "me", "msg-1").Return(mockTrashCall1).Once()
				mockTrashCall1.On("Context", ctx).Return(mockTrashCall1).Once()
				mockTrashCall1.On("Do").Return(&gmail.Message{Id: "msg-1"}, nil).Once()

				// Second fails
				mockTrashCall2 := &gmailtest.MockMessagesTrashCall{}
				mockMessagesService.On("Trash", "me", "msg-2").Return(mockTrashCall2).Once()
				mockTrashCall2.On("Context", ctx).Return(mockTrashCall2).Once()
				mockTrashCall2.On("Do").Return(nil, errors.New("API error")).Once()

				// Third succeeds
				mockTrashCall3 := &gmailtest.MockMessagesTrashCall{}
				mockMessagesService.On("Trash", "me", "msg-3").Return(mockTrashCall3).Once()
				mockTrashCall3.On("Context", ctx).Return(mockTrashCall3).Once()
				mockTrashCall3.On("Do").Return(&gmail.Message{Id: "msg-3"}, nil).Once()
			},
			wantErr:     true,
			expectedErr: "failed to trash 1 messages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGmailService, mockMessagesService := setupMockMessagesService()
			tt.setupMock(mockMessagesService)

			err := BatchTrashMessages(ctx, mockGmailService, tt.messageIDs)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}

			mockMessagesService.AssertExpectations(t)
		})
	}
}


func TestBatchModifyMessages(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		messageIDs     []string
		addLabelIDs    []string
		removeLabelIDs []string
		setupMock      func(*gmailtest.MockMessagesService)
		wantErr        bool
		expectedErr    string
	}{
		{
			name:           "empty slice",
			messageIDs:     []string{},
			addLabelIDs:    []string{"label-1"},
			removeLabelIDs: []string{"UNREAD"},
			setupMock:      func(mockMessagesService *gmailtest.MockMessagesService) {},
			wantErr:        false,
		},
		{
			name:           "successful batch modify",
			messageIDs:     []string{"msg-1", "msg-2", "msg-3"},
			addLabelIDs:    []string{"label-1"},
			removeLabelIDs: []string{"UNREAD"},
			setupMock: func(mockMessagesService *gmailtest.MockMessagesService) {
				for _, msgID := range []string{"msg-1", "msg-2", "msg-3"} {
					mockModifyCall := &gmailtest.MockMessagesModifyCall{}
					mockMessagesService.On("Modify", "me", msgID, &gmail.ModifyMessageRequest{
						AddLabelIds:    []string{"label-1"},
						RemoveLabelIds: []string{"UNREAD"},
					}).Return(mockModifyCall).Once()
					mockModifyCall.On("Context", ctx).Return(mockModifyCall).Once()
					mockModifyCall.On("Do").Return(&gmail.Message{Id: msgID}, nil).Once()
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGmailService, mockMessagesService := setupMockMessagesService()
			tt.setupMock(mockMessagesService)

			err := BatchModifyMessages(ctx, mockGmailService, tt.messageIDs, tt.addLabelIDs, tt.removeLabelIDs)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}

			mockMessagesService.AssertExpectations(t)
		})
	}
}

func TestBatchMarkAsRead(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		messageIDs  []string
		setupMock   func(*gmailtest.MockMessagesService)
		wantErr     bool
		expectedErr string
	}{
		{
			name:       "empty slice",
			messageIDs: []string{},
			setupMock:  func(mockMessagesService *gmailtest.MockMessagesService) {},
			wantErr:    false,
		},
		{
			name:       "successful batch mark as read",
			messageIDs: []string{"msg-1", "msg-2"},
			setupMock: func(mockMessagesService *gmailtest.MockMessagesService) {
				for _, msgID := range []string{"msg-1", "msg-2"} {
					mockModifyCall := &gmailtest.MockMessagesModifyCall{}
					mockMessagesService.On("Modify", "me", msgID, &gmail.ModifyMessageRequest{
						RemoveLabelIds: []string{"UNREAD"},
					}).Return(mockModifyCall).Once()
					mockModifyCall.On("Context", ctx).Return(mockModifyCall).Once()
					mockModifyCall.On("Do").Return(&gmail.Message{Id: msgID}, nil).Once()
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGmailService, mockMessagesService := setupMockMessagesService()
			tt.setupMock(mockMessagesService)

			err := BatchMarkAsRead(ctx, mockGmailService, tt.messageIDs)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}

			mockMessagesService.AssertExpectations(t)
		})
	}
}

func TestBatchMoveToFolder(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		messageIDs  []string
		folderName  string
		setupMock   func(*gmailtest.MockLabelsService, *gmailtest.MockMessagesService)
		wantErr     bool
		expectedErr string
	}{
		{
			name:       "empty slice",
			messageIDs: []string{},
			folderName: "Archive",
			setupMock:  func(mockLabelsService *gmailtest.MockLabelsService, mockMessagesService *gmailtest.MockMessagesService) {},
			wantErr:    false,
		},
		{
			name:       "successful move with existing label",
			messageIDs: []string{"msg-1", "msg-2"},
			folderName: "Archive",
			setupMock: func(mockLabelsService *gmailtest.MockLabelsService, mockMessagesService *gmailtest.MockMessagesService) {
				// List labels to find existing label
				mockLabelsListCall := &gmailtest.MockLabelsListCall{}
				mockLabelsService.On("List", "me").Return(mockLabelsListCall).Once()
				mockLabelsListCall.On("Context", ctx).Return(mockLabelsListCall).Once()
				mockLabelsListCall.On("Do").Return(&gmail.ListLabelsResponse{
					Labels: []*gmail.Label{
						{Id: "label-archive", Name: "Archive", Type: "user"},
					},
				}, nil).Once()

				// Modify messages
				for _, msgID := range []string{"msg-1", "msg-2"} {
					mockModifyCall := &gmailtest.MockMessagesModifyCall{}
					mockMessagesService.On("Modify", "me", msgID, &gmail.ModifyMessageRequest{
						AddLabelIds:    []string{"label-archive"},
						RemoveLabelIds: []string{"INBOX"},
					}).Return(mockModifyCall).Once()
					mockModifyCall.On("Context", ctx).Return(mockModifyCall).Once()
					mockModifyCall.On("Do").Return(&gmail.Message{Id: msgID}, nil).Once()
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGmailService, mockLabelsService, mockMessagesService := setupMockLabelsAndMessagesService()
			tt.setupMock(mockLabelsService, mockMessagesService)

			err := BatchMoveToFolder(ctx, mockGmailService, tt.messageIDs, tt.folderName)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}

			mockLabelsService.AssertExpectations(t)
			mockMessagesService.AssertExpectations(t)
		})
	}
}

