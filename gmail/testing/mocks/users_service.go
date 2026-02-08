package mocks

import (
	"context"

	"github.com/edaniel30/mailbridge-go/gmail/internal"
	"github.com/stretchr/testify/mock"
	gmailapi "google.golang.org/api/gmail/v1"
)

// MockUsersService is a mock for UsersService
type MockUsersService struct {
	mock.Mock
}

func (m *MockUsersService) GetMessagesService() internal.MessagesService {
	args := m.Called()
	return args.Get(0).(internal.MessagesService)
}

func (m *MockUsersService) GetLabelsService() internal.LabelsService {
	args := m.Called()
	return args.Get(0).(internal.LabelsService)
}

func (m *MockUsersService) Watch(userID string, req *gmailapi.WatchRequest) internal.UsersWatchCall {
	args := m.Called(userID, req)
	return args.Get(0).(internal.UsersWatchCall)
}

func (m *MockUsersService) Stop(userID string) internal.UsersStopCall {
	args := m.Called(userID)
	return args.Get(0).(internal.UsersStopCall)
}

func (m *MockUsersService) GetHistory(userID string) internal.UsersHistoryListCall {
	args := m.Called(userID)
	return args.Get(0).(internal.UsersHistoryListCall)
}

func (m *MockUsersService) GetProfile(userID string) internal.UsersGetProfileCall {
	args := m.Called(userID)
	return args.Get(0).(internal.UsersGetProfileCall)
}

// MockUsersWatchCall is a mock for UsersWatchCall
type MockUsersWatchCall struct {
	mock.Mock
}

func (m *MockUsersWatchCall) Context(ctx context.Context) internal.UsersWatchCall {
	m.Called(ctx)
	return m
}

func (m *MockUsersWatchCall) Do() (*gmailapi.WatchResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gmailapi.WatchResponse), args.Error(1)
}

// MockUsersStopCall is a mock for UsersStopCall
type MockUsersStopCall struct {
	mock.Mock
}

func (m *MockUsersStopCall) Context(ctx context.Context) internal.UsersStopCall {
	m.Called(ctx)
	return m
}

func (m *MockUsersStopCall) Do() error {
	args := m.Called()
	return args.Error(0)
}

// MockUsersHistoryListCall is a mock for UsersHistoryListCall
type MockUsersHistoryListCall struct {
	mock.Mock
}

func (m *MockUsersHistoryListCall) MaxResults(maxResults int64) internal.UsersHistoryListCall {
	m.Called(maxResults)
	return m
}

func (m *MockUsersHistoryListCall) PageToken(token string) internal.UsersHistoryListCall {
	m.Called(token)
	return m
}

func (m *MockUsersHistoryListCall) LabelId(labelId string) internal.UsersHistoryListCall {
	m.Called(labelId)
	return m
}

func (m *MockUsersHistoryListCall) StartHistoryId(historyId uint64) internal.UsersHistoryListCall {
	m.Called(historyId)
	return m
}

func (m *MockUsersHistoryListCall) HistoryTypes(types ...string) internal.UsersHistoryListCall {
	m.Called(types)
	return m
}

func (m *MockUsersHistoryListCall) Context(ctx context.Context) internal.UsersHistoryListCall {
	m.Called(ctx)
	return m
}

func (m *MockUsersHistoryListCall) Do() (*gmailapi.ListHistoryResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gmailapi.ListHistoryResponse), args.Error(1)
}

// MockUsersGetProfileCall is a mock for UsersGetProfileCall
type MockUsersGetProfileCall struct {
	mock.Mock
}

func (m *MockUsersGetProfileCall) Context(ctx context.Context) internal.UsersGetProfileCall {
	m.Called(ctx)
	return m
}

func (m *MockUsersGetProfileCall) Do() (*gmailapi.Profile, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gmailapi.Profile), args.Error(1)
}
