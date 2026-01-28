package messages

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	gmailtest "github.com/danielrivera/mailbridge-go/gmail/testing"
)

func TestDeleteMessage(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		messageID   string
		setupMock   func(*gmailtest.MockGmailService)
		wantErr     bool
		expectedErr string
	}{
		{
			name:      "successful delete",
			messageID: "msg-123",
			setupMock: func(mockService *gmailtest.MockGmailService) {
				mockUsersService := &gmailtest.MockUsersService{}
				mockMessagesService := &gmailtest.MockMessagesService{}
				mockDeleteCall := &gmailtest.MockMessagesDeleteCall{}

				mockService.On("GetUsersService").Return(mockUsersService)
				mockUsersService.On("GetMessagesService").Return(mockMessagesService)
				mockMessagesService.On("Delete", "me", "msg-123").Return(mockDeleteCall)
				mockDeleteCall.On("Context", ctx).Return(mockDeleteCall)
				mockDeleteCall.On("Do").Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "delete fails",
			messageID: "msg-456",
			setupMock: func(mockService *gmailtest.MockGmailService) {
				mockUsersService := &gmailtest.MockUsersService{}
				mockMessagesService := &gmailtest.MockMessagesService{}
				mockDeleteCall := &gmailtest.MockMessagesDeleteCall{}

				mockService.On("GetUsersService").Return(mockUsersService)
				mockUsersService.On("GetMessagesService").Return(mockMessagesService)
				mockMessagesService.On("Delete", "me", "msg-456").Return(mockDeleteCall)
				mockDeleteCall.On("Context", ctx).Return(mockDeleteCall)
				mockDeleteCall.On("Do").Return(fmt.Errorf("API error"))
			},
			wantErr:     true,
			expectedErr: "failed to delete message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &gmailtest.MockGmailService{}
			tt.setupMock(mockService)

			err := DeleteMessage(ctx, mockService, tt.messageID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestBatchDeleteMessages(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		messageIDs  []string
		setupMock   func(*gmailtest.MockGmailService)
		wantErr     bool
		expectedErr string
	}{
		{
			name:       "empty slice",
			messageIDs: []string{},
			setupMock:  func(mockService *gmailtest.MockGmailService) {},
			wantErr:    false,
		},
		{
			name:       "successful batch delete",
			messageIDs: []string{"msg-1", "msg-2", "msg-3"},
			setupMock: func(mockService *gmailtest.MockGmailService) {
				mockUsersService := &gmailtest.MockUsersService{}
				mockMessagesService := &gmailtest.MockMessagesService{}

				mockService.On("GetUsersService").Return(mockUsersService).Times(3)
				mockUsersService.On("GetMessagesService").Return(mockMessagesService).Times(3)

				for _, msgID := range []string{"msg-1", "msg-2", "msg-3"} {
					mockDeleteCall := &gmailtest.MockMessagesDeleteCall{}
					mockMessagesService.On("Delete", "me", msgID).Return(mockDeleteCall).Once()
					mockDeleteCall.On("Context", ctx).Return(mockDeleteCall).Once()
					mockDeleteCall.On("Do").Return(nil).Once()
				}
			},
			wantErr: false,
		},
		{
			name:       "partial failure",
			messageIDs: []string{"msg-1", "msg-2", "msg-3"},
			setupMock: func(mockService *gmailtest.MockGmailService) {
				mockUsersService := &gmailtest.MockUsersService{}
				mockMessagesService := &gmailtest.MockMessagesService{}

				mockService.On("GetUsersService").Return(mockUsersService).Times(3)
				mockUsersService.On("GetMessagesService").Return(mockMessagesService).Times(3)

				// First succeeds
				mockDeleteCall1 := &gmailtest.MockMessagesDeleteCall{}
				mockMessagesService.On("Delete", "me", "msg-1").Return(mockDeleteCall1).Once()
				mockDeleteCall1.On("Context", ctx).Return(mockDeleteCall1).Once()
				mockDeleteCall1.On("Do").Return(nil).Once()

				// Second fails
				mockDeleteCall2 := &gmailtest.MockMessagesDeleteCall{}
				mockMessagesService.On("Delete", "me", "msg-2").Return(mockDeleteCall2).Once()
				mockDeleteCall2.On("Context", ctx).Return(mockDeleteCall2).Once()
				mockDeleteCall2.On("Do").Return(fmt.Errorf("API error")).Once()

				// Third succeeds
				mockDeleteCall3 := &gmailtest.MockMessagesDeleteCall{}
				mockMessagesService.On("Delete", "me", "msg-3").Return(mockDeleteCall3).Once()
				mockDeleteCall3.On("Context", ctx).Return(mockDeleteCall3).Once()
				mockDeleteCall3.On("Do").Return(nil).Once()
			},
			wantErr:     true,
			expectedErr: "failed to delete 1 messages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &gmailtest.MockGmailService{}
			tt.setupMock(mockService)

			err := BatchDeleteMessages(ctx, mockService, tt.messageIDs)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}

			mockService.AssertExpectations(t)
		})
	}
}
