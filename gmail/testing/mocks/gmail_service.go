package mocks

import (
	"github.com/edaniel30/mailbridge-go/gmail/internal"
	"github.com/stretchr/testify/mock"
)

// MockGmailService is a mock for GmailService
type MockGmailService struct {
	mock.Mock
}

func (m *MockGmailService) GetUsersService() internal.UsersService {
	args := m.Called()
	return args.Get(0).(internal.UsersService)
}
