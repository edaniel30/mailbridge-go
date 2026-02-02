package outlook

import (
	"context"
	"testing"
	"time"

	"github.com/danielrivera/mailbridge-go/core"
	outlooktest "github.com/danielrivera/mailbridge-go/outlook/testing"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Helper to create test client with mocked service
func createTestClient() (*Client, *outlooktest.MockGraphService, *outlooktest.MockMessagesService) {
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
	mockMessagesService := &outlooktest.MockMessagesService{}

	// Setup mock chain
	mockGraphService.On("GetMeService").Return(mockMeService)
	mockMeService.On("GetMessagesService").Return(mockMessagesService)

	// Inject mocked service
	client.service = mockGraphService

	return client, mockGraphService, mockMessagesService
}

// Helper to create a test message
func createTestMessage() models.Messageable {
	msg := models.NewMessage()

	id := "msg-123"
	subject := "Test Subject"
	bodyContent := "<p>Test body content</p>"
	bodyPreview := "Test body content"
	isRead := false
	hasAttachments := true
	parentFolderID := "folder-inbox"

	msg.SetId(&id)
	msg.SetSubject(&subject)
	msg.SetBodyPreview(&bodyPreview)
	msg.SetIsRead(&isRead)
	msg.SetHasAttachments(&hasAttachments)
	msg.SetParentFolderId(&parentFolderID)

	// Body
	body := models.NewItemBody()
	contentType := models.HTML_BODYTYPE
	body.SetContentType(&contentType)
	body.SetContent(&bodyContent)
	msg.SetBody(body)

	// From
	from := models.NewRecipient()
	fromEmail := models.NewEmailAddress()
	fromName := "John Doe"
	fromAddress := "john@example.com"
	fromEmail.SetName(&fromName)
	fromEmail.SetAddress(&fromAddress)
	from.SetEmailAddress(fromEmail)
	msg.SetFrom(from)

	// To recipients
	toRecipient := models.NewRecipient()
	toEmail := models.NewEmailAddress()
	toName := "Jane Smith"
	toAddress := "jane@example.com"
	toEmail.SetName(&toName)
	toEmail.SetAddress(&toAddress)
	toRecipient.SetEmailAddress(toEmail)
	msg.SetToRecipients([]models.Recipientable{toRecipient})

	// Dates
	now := time.Now()
	msg.SetReceivedDateTime(&now)
	msg.SetSentDateTime(&now)

	return msg
}

func TestClient_ListMessages(t *testing.T) {
	client, mockGraphService, mockMessagesService := createTestClient()
	ctx := context.Background()

	// Create mock response
	mockResponse := models.NewMessageCollectionResponse()
	messages := []models.Messageable{
		createTestMessage(),
	}
	mockResponse.SetValue(messages)

	// Setup expectations
	mockMessagesService.On("List", ctx, mock.AnythingOfType("*users.ItemMessagesRequestBuilderGetRequestConfiguration")).Return(mockResponse, nil)

	// Test
	result, err := client.ListMessages(ctx, &core.ListOptions{
		MaxResults: 10,
	})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Emails, 1)

	email := result.Emails[0]
	assert.Equal(t, "msg-123", email.ID)
	assert.Equal(t, "Test Subject", email.Subject)
	assert.Equal(t, "John Doe", email.From.Name)
	assert.Equal(t, "john@example.com", email.From.Email)
	assert.Len(t, email.To, 1)
	assert.Equal(t, "Jane Smith", email.To[0].Name)
	assert.False(t, email.IsRead)

	mockGraphService.AssertExpectations(t)
	mockMessagesService.AssertExpectations(t)
}

func TestClient_ListMessages_WithPagination(t *testing.T) {
	client, mockGraphService, mockMessagesService := createTestClient()
	ctx := context.Background()

	// Create mock response with multiple messages
	mockResponse := models.NewMessageCollectionResponse()
	messages := []models.Messageable{
		createTestMessage(),
		createTestMessage(),
	}
	mockResponse.SetValue(messages)

	mockMessagesService.On("List", ctx, mock.AnythingOfType("*users.ItemMessagesRequestBuilderGetRequestConfiguration")).Return(mockResponse, nil)

	// Test with page token
	result, err := client.ListMessages(ctx, &core.ListOptions{
		MaxResults: 2,
		PageToken:  "10",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Emails, 2)
	assert.Equal(t, "12", result.NextPageToken) // 10 + 2 = 12

	mockGraphService.AssertExpectations(t)
	mockMessagesService.AssertExpectations(t)
}

func TestClient_ListMessages_WithQuery(t *testing.T) {
	client, mockGraphService, mockMessagesService := createTestClient()
	ctx := context.Background()

	mockResponse := models.NewMessageCollectionResponse()
	mockResponse.SetValue([]models.Messageable{createTestMessage()})

	var capturedConfig *users.ItemMessagesRequestBuilderGetRequestConfiguration
	mockMessagesService.On("List", ctx, mock.AnythingOfType("*users.ItemMessagesRequestBuilderGetRequestConfiguration")).
		Run(func(args mock.Arguments) {
			capturedConfig = args.Get(1).(*users.ItemMessagesRequestBuilderGetRequestConfiguration)
		}).
		Return(mockResponse, nil)

	// Test with search query
	result, err := client.ListMessages(ctx, &core.ListOptions{
		MaxResults: 10,
		Query:      "from:john@example.com",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, capturedConfig.QueryParameters)
	assert.NotNil(t, capturedConfig.QueryParameters.Search)
	assert.Equal(t, "from:john@example.com", *capturedConfig.QueryParameters.Search)

	mockGraphService.AssertExpectations(t)
	mockMessagesService.AssertExpectations(t)
}

func TestClient_ListMessages_NotConnected(t *testing.T) {
	client := &Client{}
	ctx := context.Background()

	result, err := client.ListMessages(ctx, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "client not connected")
}

func TestClient_GetMessage(t *testing.T) {
	client, mockGraphService, mockMessagesService := createTestClient()
	ctx := context.Background()

	mockMessage := createTestMessage()
	mockMessagesService.On("Get", ctx, "msg-123").Return(mockMessage, nil)

	// Test
	result, err := client.GetMessage(ctx, "msg-123")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "msg-123", result.ID)
	assert.Equal(t, "Test Subject", result.Subject)
	assert.Equal(t, "<p>Test body content</p>", result.Body.HTML)

	mockGraphService.AssertExpectations(t)
	mockMessagesService.AssertExpectations(t)
}

func TestClient_GetMessage_NotConnected(t *testing.T) {
	client := &Client{}
	ctx := context.Background()

	result, err := client.GetMessage(ctx, "msg-123")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "client not connected")
}

func TestClient_GetAttachment(t *testing.T) {
	client, mockGraphService, mockMessagesService := createTestClient()
	ctx := context.Background()

	// Create mock attachment
	mockAttachment := models.NewFileAttachment()
	attID := "att-123"
	attName := "test.pdf"
	attContentType := "application/pdf"
	attSize := int32(1024)
	attData := []byte("test data")

	mockAttachment.SetId(&attID)
	mockAttachment.SetName(&attName)
	mockAttachment.SetContentType(&attContentType)
	mockAttachment.SetSize(&attSize)
	mockAttachment.SetContentBytes(attData)

	mockMessagesService.On("GetAttachment", ctx, "msg-123", "att-123").Return(mockAttachment, nil)

	// Test
	result, err := client.GetAttachment(ctx, "msg-123", "att-123")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "att-123", result.ID)
	assert.Equal(t, "test.pdf", result.Filename)
	assert.Equal(t, "application/pdf", result.MimeType)
	assert.Equal(t, int64(1024), result.Size)
	assert.Equal(t, attData, result.Data)

	mockGraphService.AssertExpectations(t)
	mockMessagesService.AssertExpectations(t)
}

func TestClient_GetAttachment_NotConnected(t *testing.T) {
	client := &Client{}
	ctx := context.Background()

	result, err := client.GetAttachment(ctx, "msg-123", "att-123")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "client not connected")
}

func TestClient_MarkAsRead(t *testing.T) {
	client, mockGraphService, mockMessagesService := createTestClient()
	ctx := context.Background()

	mockMessagesService.On("MarkAsRead", ctx, "msg-123").Return(nil)

	// Test
	err := client.MarkAsRead(ctx, "msg-123")

	// Assert
	assert.NoError(t, err)
	mockGraphService.AssertExpectations(t)
	mockMessagesService.AssertExpectations(t)
}

func TestClient_MarkAsRead_NotConnected(t *testing.T) {
	client := &Client{}
	ctx := context.Background()

	err := client.MarkAsRead(ctx, "msg-123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client not connected")
}

func TestClient_MarkAsUnread(t *testing.T) {
	client, mockGraphService, mockMessagesService := createTestClient()
	ctx := context.Background()

	mockMessagesService.On("MarkAsUnread", ctx, "msg-123").Return(nil)

	// Test
	err := client.MarkAsUnread(ctx, "msg-123")

	// Assert
	assert.NoError(t, err)
	mockGraphService.AssertExpectations(t)
	mockMessagesService.AssertExpectations(t)
}

func TestClient_MarkAsUnread_NotConnected(t *testing.T) {
	client := &Client{}
	ctx := context.Background()

	err := client.MarkAsUnread(ctx, "msg-123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client not connected")
}

func TestClient_MoveMessage(t *testing.T) {
	client, mockGraphService, mockMessagesService := createTestClient()
	ctx := context.Background()

	mockMessagesService.On("Move", ctx, "msg-123", "folder-archive").Return(nil)

	// Test
	err := client.MoveMessage(ctx, "msg-123", "folder-archive")

	// Assert
	assert.NoError(t, err)
	mockGraphService.AssertExpectations(t)
	mockMessagesService.AssertExpectations(t)
}

func TestClient_MoveMessage_NotConnected(t *testing.T) {
	client := &Client{}
	ctx := context.Background()

	err := client.MoveMessage(ctx, "msg-123", "folder-archive")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client not connected")
}

func TestClient_DeleteMessage(t *testing.T) {
	client, mockGraphService, mockMessagesService := createTestClient()
	ctx := context.Background()

	mockMessagesService.On("Delete", ctx, "msg-123").Return(nil)

	// Test
	err := client.DeleteMessage(ctx, "msg-123")

	// Assert
	assert.NoError(t, err)
	mockGraphService.AssertExpectations(t)
	mockMessagesService.AssertExpectations(t)
}

func TestClient_DeleteMessage_NotConnected(t *testing.T) {
	client := &Client{}
	ctx := context.Background()

	err := client.DeleteMessage(ctx, "msg-123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client not connected")
}

func TestClient_ConvertMessage_CompleteConversion(t *testing.T) {
	client := &Client{service: &outlooktest.MockGraphService{}}

	// Create a fully populated message
	msg := models.NewMessage()

	id := "msg-456"
	subject := "Complete Test"
	bodyContent := "Plain text body"
	bodyPreview := "Preview text"
	isRead := true
	hasAttachments := false
	parentFolderID := "folder-sent"

	msg.SetId(&id)
	msg.SetSubject(&subject)
	msg.SetBodyPreview(&bodyPreview)
	msg.SetIsRead(&isRead)
	msg.SetHasAttachments(&hasAttachments)
	msg.SetParentFolderId(&parentFolderID)

	// Plain text body
	body := models.NewItemBody()
	contentType := models.TEXT_BODYTYPE
	body.SetContentType(&contentType)
	body.SetContent(&bodyContent)
	msg.SetBody(body)

	// CC recipients
	ccRecipient := models.NewRecipient()
	ccEmail := models.NewEmailAddress()
	ccName := "CC User"
	ccAddress := "cc@example.com"
	ccEmail.SetName(&ccName)
	ccEmail.SetAddress(&ccAddress)
	ccRecipient.SetEmailAddress(ccEmail)
	msg.SetCcRecipients([]models.Recipientable{ccRecipient})

	// BCC recipients
	bccRecipient := models.NewRecipient()
	bccEmail := models.NewEmailAddress()
	bccName := "BCC User"
	bccAddress := "bcc@example.com"
	bccEmail.SetName(&bccName)
	bccEmail.SetAddress(&bccAddress)
	bccRecipient.SetEmailAddress(bccEmail)
	msg.SetBccRecipients([]models.Recipientable{bccRecipient})

	// Dates
	receivedTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	sentTime := time.Date(2024, 1, 15, 9, 30, 0, 0, time.UTC)
	msg.SetReceivedDateTime(&receivedTime)
	msg.SetSentDateTime(&sentTime)

	// Convert
	email := client.convertMessage(msg)

	// Assert complete conversion
	assert.Equal(t, "msg-456", email.ID)
	assert.Equal(t, "Complete Test", email.Subject)
	assert.Equal(t, "Plain text body", email.Body.Text)
	assert.Equal(t, "", email.Body.HTML) // Should be empty for text body
	assert.Equal(t, "Preview text", email.Snippet)
	assert.True(t, email.IsRead)
	assert.Equal(t, sentTime, email.Date) // Should use sent time
	assert.Len(t, email.Cc, 1)
	assert.Equal(t, "CC User", email.Cc[0].Name)
	assert.Len(t, email.Bcc, 1)
	assert.Equal(t, "BCC User", email.Bcc[0].Name)
	assert.Equal(t, []string{"folder-sent"}, email.Labels)
}

func TestClient_ConvertAttachment_FileAttachment(t *testing.T) {
	mockAttachment := models.NewFileAttachment()

	attID := "att-789"
	attName := "document.docx"
	attContentType := "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	attSize := int32(2048)
	attData := []byte("document content")

	mockAttachment.SetId(&attID)
	mockAttachment.SetName(&attName)
	mockAttachment.SetContentType(&attContentType)
	mockAttachment.SetSize(&attSize)
	mockAttachment.SetContentBytes(attData)

	// Convert
	attachment := convertAttachment(mockAttachment)

	// Assert
	assert.Equal(t, "att-789", attachment.ID)
	assert.Equal(t, "document.docx", attachment.Filename)
	assert.Equal(t, attContentType, attachment.MimeType)
	assert.Equal(t, int64(2048), attachment.Size)
	assert.Equal(t, attData, attachment.Data)
}

func TestClient_ConvertMessage_EmptyFields(t *testing.T) {
	client := &Client{service: &outlooktest.MockGraphService{}}

	// Create minimal message with nil/empty fields
	msg := models.NewMessage()

	// Convert
	email := client.convertMessage(msg)

	// Assert defaults
	assert.Equal(t, "", email.ID)
	assert.Equal(t, "", email.Subject)
	assert.Equal(t, "", email.From.Name)
	assert.Equal(t, "", email.From.Email)
	assert.Empty(t, email.To)
	assert.Empty(t, email.Cc)
	assert.Empty(t, email.Bcc)
	assert.False(t, email.IsRead)
	assert.Empty(t, email.Labels)
}
