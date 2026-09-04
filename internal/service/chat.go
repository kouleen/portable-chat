package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kouleen/portable-chat/internal/model"
	"github.com/kouleen/portable-chat/internal/repository"
	utils "github.com/kouleen/portable-chat/pkg/util"
)

func ListChatHistory(ctx context.Context, req *model.CharHistoryQuery) (*model.CharHistoryResponse, error) {
	currentUser, err := utils.MustGetAuthorization(ctx)
	if err != nil {
		return nil, err
	}
	contact, err := repository.GetAuthorizationByID(ctx, req.ContactID)
	if err != nil {
		return nil, err
	}

	days := req.Days
	if days <= 0 {
		days = 30
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 200
	}
	if limit > 500 {
		limit = 500
	}

	since := time.Now().AddDate(0, 0, -days)
	messages, err := repository.ListChatMessages(ctx, currentUser.ID, contact.ID, since, limit)
	if err != nil {
		return nil, err
	}

	return &model.CharHistoryResponse{
		Contact:  contact.Profile(),
		Messages: messages,
	}, nil
}

func BuildRoomID(userID, contactID int64) string {
	if userID < contactID {
		return fmt.Sprintf("%d_%d", userID, contactID)
	}
	return fmt.Sprintf("%d_%d", contactID, userID)
}

func SendChatMessage(ctx context.Context, contactID int64, content string) (*model.ChatMessage, error) {
	currentUser, err := utils.MustGetAuthorization(ctx)
	if err != nil {
		return nil, err
	}
	contact, err := repository.GetAuthorizationByID(ctx, contactID)
	if err != nil {
		return nil, err
	}

	content = strings.TrimSpace(content)
	if content == "" {
		return nil, fmt.Errorf("消息内容不能为空")
	}
	if len([]rune(content)) > 2000 {
		return nil, fmt.Errorf("消息内容不能超过2000个字符")
	}

	message := &model.ChatMessage{
		RoomID:         BuildRoomID(currentUser.ID, contact.ID),
		SenderID:       currentUser.ID,
		SenderName:     currentUser.Username,
		SenderAvatar:   currentUser.Avatar,
		ReceiverID:     contact.ID,
		ReceiverName:   contact.Username,
		ReceiverAvatar: contact.Avatar,
		Content:        content,
	}
	if err := repository.CreateChatMessage(ctx, message); err != nil {
		return nil, err
	}
	return message, nil
}

func MarkConversationRead(ctx context.Context, receiverID, senderID int64) (*model.ReadReceipt, error) {
	readAt := time.Now()
	messageIDs, err := repository.MarkMessagesRead(ctx, receiverID, senderID, readAt)
	if err != nil {
		return nil, err
	}
	if len(messageIDs) == 0 {
		return nil, nil
	}
	return &model.ReadReceipt{
		ContactID:  receiverID,
		MessageIDs: messageIDs,
		ReadAt:     &readAt,
	}, nil
}
