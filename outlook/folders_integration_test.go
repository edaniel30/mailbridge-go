package outlook

import (
	"context"
	"testing"

	"github.com/danielrivera/mailbridge-go/core"
	outlooktest "github.com/danielrivera/mailbridge-go/outlook/testing"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Helper to create test folder
func createTestFolder(id, name string, totalCount, unreadCount int32) models.MailFolderable {
	folder := models.NewMailFolder()
	folder.SetId(&id)
	folder.SetDisplayName(&name)
	folder.SetTotalItemCount(&totalCount)
	folder.SetUnreadItemCount(&unreadCount)
	return folder
}

// Helper to create test client for folder operations
func createTestClientForFolders() (*Client, *outlooktest.MockGraphService, *outlooktest.MockMeService, *outlooktest.MockMailFoldersService) {
	client := &Client{
		config: &Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-secret",
			TenantID:     "consumers",
			RedirectURL:  "http://localhost:8080/callback",
		},
	}

	mockGraphService := &outlooktest.MockGraphService{}
	mockMeService := &outlooktest.MockMeService{}
	mockFoldersService := &outlooktest.MockMailFoldersService{}

	// Setup mock chain
	mockGraphService.On("GetMeService").Return(mockMeService)
	mockMeService.On("GetMailFoldersService").Return(mockFoldersService)

	// Inject mocked service
	client.service = mockGraphService

	return client, mockGraphService, mockMeService, mockFoldersService
}

func TestClient_ListFolders(t *testing.T) {
	client, mockGraphService, _, mockFoldersService := createTestClientForFolders()
	ctx := context.Background()

	// Create mock response
	mockResponse := models.NewMailFolderCollectionResponse()
	folders := []models.MailFolderable{
		createTestFolder("folder-1", "Inbox", 100, 10),
		createTestFolder("folder-2", "Sent Items", 50, 0),
		createTestFolder("folder-3", "My Custom Folder", 25, 5),
	}
	mockResponse.SetValue(folders)

	mockFoldersService.On("List", ctx).Return(mockResponse, nil)

	// Test
	result, err := client.ListFolders(ctx)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 3)

	// Check first folder (Inbox - system)
	assert.Equal(t, "folder-1", result[0].ID)
	assert.Equal(t, "Inbox", result[0].Name)
	assert.Equal(t, "system", result[0].Type)
	assert.Equal(t, 100, result[0].TotalMessages)
	assert.Equal(t, 10, result[0].UnreadMessages)

	// Check second folder (Sent Items - system)
	assert.Equal(t, "folder-2", result[1].ID)
	assert.Equal(t, "Sent Items", result[1].Name)
	assert.Equal(t, "system", result[1].Type)

	// Check third folder (custom - user)
	assert.Equal(t, "folder-3", result[2].ID)
	assert.Equal(t, "My Custom Folder", result[2].Name)
	assert.Equal(t, "user", result[2].Type)
	assert.Equal(t, 25, result[2].TotalMessages)
	assert.Equal(t, 5, result[2].UnreadMessages)

	mockGraphService.AssertExpectations(t)
	mockFoldersService.AssertExpectations(t)
}

func TestClient_ListFolders_NotConnected(t *testing.T) {
	client := &Client{}
	ctx := context.Background()

	result, err := client.ListFolders(ctx)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "client not connected")
}

func TestClient_GetFolder(t *testing.T) {
	client, mockGraphService, _, mockFoldersService := createTestClientForFolders()
	ctx := context.Background()

	mockFolder := createTestFolder("folder-inbox", "Inbox", 150, 20)
	mockFoldersService.On("Get", ctx, "folder-inbox").Return(mockFolder, nil)

	// Test
	result, err := client.GetFolder(ctx, "folder-inbox")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "folder-inbox", result.ID)
	assert.Equal(t, "Inbox", result.Name)
	assert.Equal(t, "system", result.Type)
	assert.Equal(t, 150, result.TotalMessages)
	assert.Equal(t, 20, result.UnreadMessages)

	mockGraphService.AssertExpectations(t)
	mockFoldersService.AssertExpectations(t)
}

func TestClient_GetFolder_NotConnected(t *testing.T) {
	client := &Client{}
	ctx := context.Background()

	result, err := client.GetFolder(ctx, "folder-inbox")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "client not connected")
}

func TestClient_CreateFolder(t *testing.T) {
	client, mockGraphService, _, mockFoldersService := createTestClientForFolders()
	ctx := context.Background()

	mockFolder := createTestFolder("folder-new", "My New Folder", 0, 0)
	mockFoldersService.On("Create", ctx, "My New Folder").Return(mockFolder, nil)

	// Test
	result, err := client.CreateFolder(ctx, "My New Folder")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "folder-new", result.ID)
	assert.Equal(t, "My New Folder", result.Name)
	assert.Equal(t, "user", result.Type)

	mockGraphService.AssertExpectations(t)
	mockFoldersService.AssertExpectations(t)
}

func TestClient_CreateFolder_EmptyName(t *testing.T) {
	client, _, _, _ := createTestClientForFolders()
	ctx := context.Background()

	result, err := client.CreateFolder(ctx, "")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "folder name cannot be empty")
}

func TestClient_CreateFolder_NotConnected(t *testing.T) {
	client := &Client{}
	ctx := context.Background()

	result, err := client.CreateFolder(ctx, "Test Folder")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "client not connected")
}

func TestClient_UpdateFolder(t *testing.T) {
	client, mockGraphService, _, mockFoldersService := createTestClientForFolders()
	ctx := context.Background()

	mockFolder := createTestFolder("folder-123", "Updated Folder Name", 10, 2)
	mockFoldersService.On("Update", ctx, "folder-123", "Updated Folder Name").Return(mockFolder, nil)

	// Test
	result, err := client.UpdateFolder(ctx, "folder-123", "Updated Folder Name")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "folder-123", result.ID)
	assert.Equal(t, "Updated Folder Name", result.Name)

	mockGraphService.AssertExpectations(t)
	mockFoldersService.AssertExpectations(t)
}

func TestClient_UpdateFolder_EmptyName(t *testing.T) {
	client, _, _, _ := createTestClientForFolders()
	ctx := context.Background()

	result, err := client.UpdateFolder(ctx, "folder-123", "")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "folder name cannot be empty")
}

func TestClient_UpdateFolder_NotConnected(t *testing.T) {
	client := &Client{}
	ctx := context.Background()

	result, err := client.UpdateFolder(ctx, "folder-123", "New Name")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "client not connected")
}

func TestClient_DeleteFolder(t *testing.T) {
	client, mockGraphService, _, mockFoldersService := createTestClientForFolders()
	ctx := context.Background()

	mockFoldersService.On("Delete", ctx, "folder-123").Return(nil)

	// Test
	err := client.DeleteFolder(ctx, "folder-123")

	// Assert
	assert.NoError(t, err)
	mockGraphService.AssertExpectations(t)
	mockFoldersService.AssertExpectations(t)
}

func TestClient_DeleteFolder_NotConnected(t *testing.T) {
	client := &Client{}
	ctx := context.Background()

	err := client.DeleteFolder(ctx, "folder-123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client not connected")
}

func TestClient_ListMessagesInFolder(t *testing.T) {
	client, mockGraphService, mockMeService, mockFoldersService := createTestClientForFolders()
	ctx := context.Background()

	// Create mock response with messages
	mockResponse := models.NewMessageCollectionResponse()
	messages := []models.Messageable{
		createTestMessage(),
	}
	mockResponse.SetValue(messages)

	mockFoldersService.On("GetMessages", ctx, "folder-inbox", mock.AnythingOfType("*users.ItemMailFoldersItemMessagesRequestBuilderGetRequestConfiguration")).Return(mockResponse, nil)

	// Test
	result, err := client.ListMessagesInFolder(ctx, "folder-inbox", &core.ListOptions{
		MaxResults: 10,
	})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Emails, 1)

	email := result.Emails[0]
	assert.Equal(t, "msg-123", email.ID)
	assert.Equal(t, "Test Subject", email.Subject)

	mockGraphService.AssertExpectations(t)
	mockMeService.AssertExpectations(t)
	mockFoldersService.AssertExpectations(t)
}

func TestClient_ListMessagesInFolder_WithPagination(t *testing.T) {
	client, mockGraphService, mockMeService, mockFoldersService := createTestClientForFolders()
	ctx := context.Background()

	// Create mock response with multiple messages
	mockResponse := models.NewMessageCollectionResponse()
	messages := []models.Messageable{
		createTestMessage(),
		createTestMessage(),
	}
	mockResponse.SetValue(messages)

	mockFoldersService.On("GetMessages", ctx, "folder-inbox", mock.AnythingOfType("*users.ItemMailFoldersItemMessagesRequestBuilderGetRequestConfiguration")).Return(mockResponse, nil)

	// Test with page token
	result, err := client.ListMessagesInFolder(ctx, "folder-inbox", &core.ListOptions{
		MaxResults: 2,
		PageToken:  "5",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Emails, 2)
	assert.Equal(t, "7", result.NextPageToken) // 5 + 2 = 7

	mockGraphService.AssertExpectations(t)
	mockMeService.AssertExpectations(t)
	mockFoldersService.AssertExpectations(t)
}

func TestClient_ListMessagesInFolder_WithQuery(t *testing.T) {
	client, mockGraphService, mockMeService, mockFoldersService := createTestClientForFolders()
	ctx := context.Background()

	mockResponse := models.NewMessageCollectionResponse()
	mockResponse.SetValue([]models.Messageable{createTestMessage()})

	var capturedConfig *users.ItemMailFoldersItemMessagesRequestBuilderGetRequestConfiguration
	mockFoldersService.On("GetMessages", ctx, "folder-inbox", mock.AnythingOfType("*users.ItemMailFoldersItemMessagesRequestBuilderGetRequestConfiguration")).
		Run(func(args mock.Arguments) {
			capturedConfig = args.Get(2).(*users.ItemMailFoldersItemMessagesRequestBuilderGetRequestConfiguration)
		}).
		Return(mockResponse, nil)

	// Test with search query
	result, err := client.ListMessagesInFolder(ctx, "folder-inbox", &core.ListOptions{
		MaxResults: 10,
		Query:      "subject:invoice",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, capturedConfig.QueryParameters)
	assert.NotNil(t, capturedConfig.QueryParameters.Search)
	assert.Equal(t, "subject:invoice", *capturedConfig.QueryParameters.Search)

	mockGraphService.AssertExpectations(t)
	mockMeService.AssertExpectations(t)
	mockFoldersService.AssertExpectations(t)
}

func TestClient_ListMessagesInFolder_NotConnected(t *testing.T) {
	client := &Client{}
	ctx := context.Background()

	result, err := client.ListMessagesInFolder(ctx, "folder-inbox", nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "client not connected")
}

func TestConvertFolder_SystemFolders(t *testing.T) {
	tests := []struct {
		name         string
		folderName   string
		expectedType string
	}{
		{"Inbox", "Inbox", "system"},
		{"Sent Items", "Sent Items", "system"},
		{"Drafts", "Drafts", "system"},
		{"Deleted Items", "Deleted Items", "system"},
		{"Junk Email", "Junk Email", "system"},
		{"Custom Folder", "My Folder", "user"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			folder := createTestFolder("folder-id", tt.folderName, 10, 2)
			label := convertFolder(folder)

			assert.Equal(t, tt.folderName, label.Name)
			assert.Equal(t, tt.expectedType, label.Type)
		})
	}
}

func TestConvertFolder_MessageCounts(t *testing.T) {
	folder := createTestFolder("folder-id", "Test Folder", 150, 25)
	label := convertFolder(folder)

	assert.Equal(t, "folder-id", label.ID)
	assert.Equal(t, "Test Folder", label.Name)
	assert.Equal(t, 150, label.TotalMessages)
	assert.Equal(t, 25, label.UnreadMessages)
}

func TestConvertFolder_EmptyFields(t *testing.T) {
	folder := models.NewMailFolder()

	// Convert with minimal data
	label := convertFolder(folder)

	assert.Equal(t, "", label.ID)
	assert.Equal(t, "", label.Name)
	assert.Equal(t, "user", label.Type) // Default type
	assert.Equal(t, 0, label.TotalMessages)
	assert.Equal(t, 0, label.UnreadMessages)
}

func TestFolderConstants(t *testing.T) {
	// Verify well-known folder IDs
	assert.Equal(t, "inbox", FolderInbox)
	assert.Equal(t, "drafts", FolderDrafts)
	assert.Equal(t, "sentitems", FolderSentItems)
	assert.Equal(t, "deleteditems", FolderDeletedItems)
	assert.Equal(t, "junkemail", FolderJunkEmail)
	assert.Equal(t, "outbox", FolderOutbox)
	assert.Equal(t, "archive", FolderArchive)
}
