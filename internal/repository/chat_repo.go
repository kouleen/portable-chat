package repository

import (
	"context"
	"time"

	"github.com/kouleen/portable-chat/internal/model"
	"github.com/kouleen/portable-chat/pkg/sqlitecli"
)

func CreateChatMessage(ctx context.Context, message *model.ChatMessage) error {
	return sqlitecli.GetSqliteDB().WithContext(ctx).Create(message).Error
}

func ListChatMessages(ctx context.Context, userID, contactID int64, since time.Time, limit int) ([]*model.ChatMessage, error) {
	var messages []*model.ChatMessage
	err := sqlitecli.GetSqliteDB().WithContext(ctx).
		Where(
			"((sender_id = ? and receiver_id = ?) or (sender_id = ? and receiver_id = ?)) and create_time >= ?",
			userID, contactID, contactID, userID, since,
		).
		Order("create_time asc").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}

func MarkMessagesRead(ctx context.Context, receiverID, senderID int64, readAt time.Time) ([]int64, error) {
	db := sqlitecli.GetSqliteDB().WithContext(ctx)
	var messages []*model.ChatMessage
	if err := db.
		Where("sender_id = ? and receiver_id = ? and is_read = ?", senderID, receiverID, false).
		Order("create_time asc").
		Find(&messages).Error; err != nil {
		return nil, err
	}
	if len(messages) == 0 {
		return nil, nil
	}

	messageIDs := make([]int64, 0, len(messages))
	for _, message := range messages {
		messageIDs = append(messageIDs, message.ID)
	}

	if err := db.Model(&model.ChatMessage{}).
		Where("id in ?", messageIDs).
		Updates(map[string]any{
			"is_read":     true,
			"read_at":     readAt,
			"update_time": readAt,
		}).Error; err != nil {
		return nil, err
	}

	return messageIDs, nil
}
