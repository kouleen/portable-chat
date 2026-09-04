package repository

import (
	"context"
	"errors"
	"slices"

	"github.com/kouleen/portable-chat/internal/model"
	"github.com/kouleen/portable-chat/pkg/sqlitecli"
	"gorm.io/gorm"
)

func ListContact(ctx context.Context, req *model.CharContactQuery) ([]*model.CharContact, error) {
	db := sqlitecli.GetSqliteDB().WithContext(ctx)
	query := db.Model(&model.Authorization{}).Where("id <> ?", req.UserID)
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		query = query.Where("username like ? or uin like ?", like, like)
	}

	var users []*model.Authorization
	if err := query.Find(&users).Error; err != nil {
		return nil, err
	}

	contacts := make([]*model.CharContact, 0, len(users))
	for _, user := range users {
		contact := &model.CharContact{
			ID:            user.ID,
			Uin:           user.Uin,
			ContactName:   user.Username,
			ContactAvatar: user.Avatar,
		}

		var latest model.ChatMessage
		err := db.Where(
			"(sender_id = ? and receiver_id = ?) or (sender_id = ? and receiver_id = ?)",
			req.UserID, user.ID, user.ID, req.UserID,
		).Order("create_time desc").First(&latest).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if err == nil {
			contact.LastMessage = latest.Content
			contact.LastMessageAt = latest.CreateTime
		}

		if err := db.Model(&model.ChatMessage{}).
			Where("sender_id = ? and receiver_id = ? and is_read = ?", user.ID, req.UserID, false).
			Count(&contact.UnreadCount).Error; err != nil {
			return nil, err
		}

		contacts = append(contacts, contact)
	}

	slices.SortFunc(contacts, func(a, b *model.CharContact) int {
		switch {
		case a.LastMessageAt == nil && b.LastMessageAt == nil:
			if a.ContactName < b.ContactName {
				return -1
			}
			if a.ContactName > b.ContactName {
				return 1
			}
			return 0
		case a.LastMessageAt == nil:
			return 1
		case b.LastMessageAt == nil:
			return -1
		case a.LastMessageAt.After(*b.LastMessageAt):
			return -1
		case a.LastMessageAt.Before(*b.LastMessageAt):
			return 1
		default:
			return 0
		}
	})

	return contacts, nil
}
