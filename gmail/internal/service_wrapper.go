package internal

import (
	"context"

	"google.golang.org/api/gmail/v1"
)

// RealGmailService wraps gmail.Service to implement GmailService interface
type RealGmailService struct {
	service *gmail.Service
}

// NewRealGmailService creates a new RealGmailService
func NewRealGmailService(service *gmail.Service) *RealGmailService {
	return &RealGmailService{service: service}
}

func (r *RealGmailService) GetUsersService() UsersService {
	return &realUsersService{users: r.service.Users}
}

// realUsersService wraps gmail.UsersService
type realUsersService struct {
	users *gmail.UsersService
}

func (r *realUsersService) GetMessagesService() MessagesService {
	return &realMessagesService{messages: r.users.Messages}
}

func (r *realUsersService) GetLabelsService() LabelsService {
	return &realLabelsService{labels: r.users.Labels}
}

// realMessagesService wraps gmail.MessagesService
type realMessagesService struct {
	messages *gmail.UsersMessagesService
}

func (r *realMessagesService) List(userID string) MessagesListCall {
	return &realMessagesListCall{call: r.messages.List(userID)}
}

func (r *realMessagesService) Get(userID, messageID string) MessagesGetCall {
	return &realMessagesGetCall{call: r.messages.Get(userID, messageID)}
}

func (r *realMessagesService) Modify(userID, messageID string, req *gmail.ModifyMessageRequest) MessagesModifyCall {
	return &realMessagesModifyCall{call: r.messages.Modify(userID, messageID, req)}
}

func (r *realMessagesService) GetAttachment(userID, messageID, attachmentID string) MessagesAttachmentGetCall {
	return &realMessagesAttachmentGetCall{call: r.messages.Attachments.Get(userID, messageID, attachmentID)}
}

// realLabelsService wraps gmail.LabelsService
type realLabelsService struct {
	labels *gmail.UsersLabelsService
}

func (r *realLabelsService) List(userID string) LabelsListCall {
	return &realLabelsListCall{call: r.labels.List(userID)}
}

func (r *realLabelsService) Get(userID, labelID string) LabelsGetCall {
	return &realLabelsGetCall{call: r.labels.Get(userID, labelID)}
}

func (r *realLabelsService) Create(userID string, label *gmail.Label) LabelsCreateCall {
	return &realLabelsCreateCall{call: r.labels.Create(userID, label)}
}

func (r *realLabelsService) Delete(userID, labelID string) LabelsDeleteCall {
	return &realLabelsDeleteCall{call: r.labels.Delete(userID, labelID)}
}

// Call wrappers
type realMessagesListCall struct {
	call *gmail.UsersMessagesListCall
}

func (r *realMessagesListCall) MaxResults(maxResults int64) MessagesListCall {
	r.call = r.call.MaxResults(maxResults)
	return r
}

func (r *realMessagesListCall) PageToken(token string) MessagesListCall {
	r.call = r.call.PageToken(token)
	return r
}

func (r *realMessagesListCall) Q(query string) MessagesListCall {
	r.call = r.call.Q(query)
	return r
}

func (r *realMessagesListCall) LabelIds(labelIds ...string) MessagesListCall {
	r.call = r.call.LabelIds(labelIds...)
	return r
}

func (r *realMessagesListCall) Context(ctx context.Context) MessagesListCall {
	r.call = r.call.Context(ctx)
	return r
}

func (r *realMessagesListCall) Do() (*gmail.ListMessagesResponse, error) {
	return r.call.Do()
}

type realMessagesGetCall struct {
	call *gmail.UsersMessagesGetCall
}

func (r *realMessagesGetCall) Format(format string) MessagesGetCall {
	r.call = r.call.Format(format)
	return r
}

func (r *realMessagesGetCall) Context(ctx context.Context) MessagesGetCall {
	r.call = r.call.Context(ctx)
	return r
}

func (r *realMessagesGetCall) Do() (*gmail.Message, error) {
	return r.call.Do()
}

type realMessagesModifyCall struct {
	call *gmail.UsersMessagesModifyCall
}

func (r *realMessagesModifyCall) Context(ctx context.Context) MessagesModifyCall {
	r.call = r.call.Context(ctx)
	return r
}

func (r *realMessagesModifyCall) Do() (*gmail.Message, error) {
	return r.call.Do()
}

type realMessagesAttachmentGetCall struct {
	call *gmail.UsersMessagesAttachmentsGetCall
}

func (r *realMessagesAttachmentGetCall) Context(ctx context.Context) MessagesAttachmentGetCall {
	r.call = r.call.Context(ctx)
	return r
}

func (r *realMessagesAttachmentGetCall) Do() (*gmail.MessagePartBody, error) {
	return r.call.Do()
}

type realLabelsListCall struct {
	call *gmail.UsersLabelsListCall
}

func (r *realLabelsListCall) Context(ctx context.Context) LabelsListCall {
	r.call = r.call.Context(ctx)
	return r
}

func (r *realLabelsListCall) Do() (*gmail.ListLabelsResponse, error) {
	return r.call.Do()
}

type realLabelsGetCall struct {
	call *gmail.UsersLabelsGetCall
}

func (r *realLabelsGetCall) Context(ctx context.Context) LabelsGetCall {
	r.call = r.call.Context(ctx)
	return r
}

func (r *realLabelsGetCall) Do() (*gmail.Label, error) {
	return r.call.Do()
}

type realLabelsCreateCall struct {
	call *gmail.UsersLabelsCreateCall
}

func (r *realLabelsCreateCall) Context(ctx context.Context) LabelsCreateCall {
	r.call = r.call.Context(ctx)
	return r
}

func (r *realLabelsCreateCall) Do() (*gmail.Label, error) {
	return r.call.Do()
}

type realLabelsDeleteCall struct {
	call *gmail.UsersLabelsDeleteCall
}

func (r *realLabelsDeleteCall) Context(ctx context.Context) LabelsDeleteCall {
	r.call = r.call.Context(ctx)
	return r
}

func (r *realLabelsDeleteCall) Do() error {
	return r.call.Do()
}
