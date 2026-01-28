package operations

import (
	"context"
	"errors"
	"testing"

	"github.com/danielrivera/mailbridge-go/gmail/internal"
	gmailtest "github.com/danielrivera/mailbridge-go/gmail/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserIDMe(t *testing.T) {
	assert.Equal(t, "me", UserIDMe, "UserIDMe constant should be 'me'")
}

func TestBatchOperation_EmptyList(t *testing.T) {
	called := false
	operation := func(ctx context.Context, messageID string) error {
		called = true
		return nil
	}

	err := BatchOperation(context.Background(), []string{}, operation, "test")

	assert.NoError(t, err)
	assert.False(t, called, "operation should not be called with empty list")
}

func TestBatchOperation_AllSuccess(t *testing.T) {
	ctx := context.Background()
	messageIDs := []string{"msg1", "msg2", "msg3"}
	processedIDs := []string{}

	operation := func(ctx context.Context, messageID string) error {
		processedIDs = append(processedIDs, messageID)
		return nil
	}

	err := BatchOperation(ctx, messageIDs, operation, "test")

	assert.NoError(t, err)
	assert.Equal(t, messageIDs, processedIDs, "all messages should be processed")
}

func TestBatchOperation_PartialFailure(t *testing.T) {
	ctx := context.Background()
	messageIDs := []string{"msg1", "msg2", "msg3"}

	operation := func(ctx context.Context, messageID string) error {
		if messageID == "msg2" {
			return errors.New("test error")
		}
		return nil
	}

	err := BatchOperation(ctx, messageIDs, operation, "test operation")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to test operation 1 messages")
	assert.Contains(t, err.Error(), "msg2: test error")
}

func TestBatchOperation_AllFailed(t *testing.T) {
	ctx := context.Background()
	messageIDs := []string{"msg1", "msg2"}

	operation := func(ctx context.Context, messageID string) error {
		return errors.New("error for " + messageID)
	}

	err := BatchOperation(ctx, messageIDs, operation, "delete")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete 2 messages")
	assert.Contains(t, err.Error(), "msg1: error for msg1")
	assert.Contains(t, err.Error(), "msg2: error for msg2")
}

func TestGetMessagesService(t *testing.T) {
	// Create mocks
	mockService := &gmailtest.MockGmailService{}
	mockUsersService := &gmailtest.MockUsersService{}
	mockMessagesService := &gmailtest.MockMessagesService{}

	// Setup expectations
	mockService.On("GetUsersService").Return(mockUsersService)
	mockUsersService.On("GetMessagesService").Return(mockMessagesService)

	// Call helper
	result := GetMessagesService(mockService)

	// Verify
	assert.Equal(t, mockMessagesService, result)
	mockService.AssertExpectations(t)
	mockUsersService.AssertExpectations(t)
}

func TestGetMessagesService_Integration(t *testing.T) {
	// Test that the helper correctly chains the service calls
	mockService := &gmailtest.MockGmailService{}
	mockUsersService := &gmailtest.MockUsersService{}
	mockMessagesService := &gmailtest.MockMessagesService{}

	mockService.On("GetUsersService").Return(mockUsersService).Once()
	mockUsersService.On("GetMessagesService").Return(mockMessagesService).Once()

	// First call
	result1 := GetMessagesService(mockService)
	assert.NotNil(t, result1)

	// Verify mock was called
	mockService.AssertExpectations(t)
	mockUsersService.AssertExpectations(t)
}

// Benchmark tests
func BenchmarkBatchOperation(b *testing.B) {
	ctx := context.Background()
	messageIDs := make([]string, 100)
	for i := 0; i < 100; i++ {
		messageIDs[i] = "msg" + string(rune(i))
	}

	operation := func(ctx context.Context, messageID string) error {
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BatchOperation(ctx, messageIDs, operation, "benchmark")
	}
}

func TestBatchOperation_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	messageIDs := []string{"msg1", "msg2", "msg3"}
	processCount := 0

	operation := func(ctx context.Context, messageID string) error {
		processCount++
		if processCount == 2 {
			cancel()
		}
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	}

	err := BatchOperation(ctx, messageIDs, operation, "test")

	// Should have error from the cancelled context
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

// Helper function to verify mock interactions
func verifyMockCalls(t *testing.T, mocks ...interface{ AssertExpectations(t mock.TestingT) bool }) {
	for _, m := range mocks {
		m.AssertExpectations(t)
	}
}

// Test helper utilities
func TestHelpers_ChainedCalls(t *testing.T) {
	mockService := &gmailtest.MockGmailService{}
	mockUsersService := &gmailtest.MockUsersService{}
	mockMessagesService := &gmailtest.MockMessagesService{}

	// Setup chain
	mockService.On("GetUsersService").Return(mockUsersService)
	mockUsersService.On("GetMessagesService").Return(mockMessagesService)

	// Multiple calls should work
	for i := 0; i < 3; i++ {
		result := GetMessagesService(mockService)
		assert.NotNil(t, result)
	}
}

func TestBatchOperation_ErrorAggregation(t *testing.T) {
	tests := []struct {
		name           string
		messageIDs     []string
		failingIDs     map[string]string // messageID -> error message
		expectedErrors int
	}{
		{
			name:       "single error",
			messageIDs: []string{"msg1", "msg2", "msg3"},
			failingIDs: map[string]string{
				"msg2": "network timeout",
			},
			expectedErrors: 1,
		},
		{
			name:       "multiple errors",
			messageIDs: []string{"msg1", "msg2", "msg3", "msg4"},
			failingIDs: map[string]string{
				"msg1": "auth failed",
				"msg3": "not found",
			},
			expectedErrors: 2,
		},
		{
			name:           "no errors",
			messageIDs:     []string{"msg1", "msg2"},
			failingIDs:     map[string]string{},
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operation := func(ctx context.Context, messageID string) error {
				if errMsg, shouldFail := tt.failingIDs[messageID]; shouldFail {
					return errors.New(errMsg)
				}
				return nil
			}

			err := BatchOperation(context.Background(), tt.messageIDs, operation, "test")

			if tt.expectedErrors == 0 {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "failed to test")
				// Verify all expected errors are in the message
				for msgID, errMsg := range tt.failingIDs {
					assert.Contains(t, err.Error(), msgID)
					assert.Contains(t, err.Error(), errMsg)
				}
			}
		})
	}
}
