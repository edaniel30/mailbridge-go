package mocks

import (
	"context"

	"github.com/edaniel30/mailbridge-go/gmail/internal"
	"github.com/stretchr/testify/mock"
	gmailapi "google.golang.org/api/gmail/v1"
)

// MockLabelsService is a mock for LabelsService
type MockLabelsService struct {
	mock.Mock
}

func (m *MockLabelsService) List(userID string) internal.LabelsListCall {
	args := m.Called(userID)
	return args.Get(0).(internal.LabelsListCall)
}

func (m *MockLabelsService) Get(userID, labelID string) internal.LabelsGetCall {
	args := m.Called(userID, labelID)
	return args.Get(0).(internal.LabelsGetCall)
}

func (m *MockLabelsService) Create(userID string, label *gmailapi.Label) internal.LabelsCreateCall {
	args := m.Called(userID, label)
	return args.Get(0).(internal.LabelsCreateCall)
}

func (m *MockLabelsService) Delete(userID, labelID string) internal.LabelsDeleteCall {
	args := m.Called(userID, labelID)
	return args.Get(0).(internal.LabelsDeleteCall)
}

// MockLabelsListCall is a mock for LabelsListCall
type MockLabelsListCall struct {
	mock.Mock
}

func (m *MockLabelsListCall) Context(ctx context.Context) internal.LabelsListCall {
	m.Called(ctx)
	return m
}

func (m *MockLabelsListCall) Do() (*gmailapi.ListLabelsResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gmailapi.ListLabelsResponse), args.Error(1)
}

// MockLabelsGetCall is a mock for LabelsGetCall
type MockLabelsGetCall struct {
	mock.Mock
}

func (m *MockLabelsGetCall) Context(ctx context.Context) internal.LabelsGetCall {
	m.Called(ctx)
	return m
}

func (m *MockLabelsGetCall) Do() (*gmailapi.Label, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gmailapi.Label), args.Error(1)
}

// MockLabelsCreateCall is a mock for LabelsCreateCall
type MockLabelsCreateCall struct {
	mock.Mock
}

func (m *MockLabelsCreateCall) Context(ctx context.Context) internal.LabelsCreateCall {
	m.Called(ctx)
	return m
}

func (m *MockLabelsCreateCall) Do() (*gmailapi.Label, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gmailapi.Label), args.Error(1)
}

// MockLabelsDeleteCall is a mock for LabelsDeleteCall
type MockLabelsDeleteCall struct {
	mock.Mock
}

func (m *MockLabelsDeleteCall) Context(ctx context.Context) internal.LabelsDeleteCall {
	m.Called(ctx)
	return m
}

func (m *MockLabelsDeleteCall) Do() error {
	args := m.Called()
	return args.Error(0)
}
