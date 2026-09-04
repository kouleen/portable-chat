package model

import (
	"time"
)

type CharHistoryQuery struct {
	ContactID int64 `form:"contactId" binding:"required"`
	Days      int   `form:"days"`
	Limit     int   `form:"limit"`
}

type CharHistoryResponse struct {
	Contact  *AuthorizationProfile `json:"contact"`
	Messages []*ChatMessage        `json:"messages"`
}

type ChatMessage struct {
	ID             int64      `json:"id,string" gorm:"column:id;primary_key;not null;auto_increment"`
	RoomID         string     `json:"roomId" gorm:"column:room_id;not null;index"`
	SenderID       int64      `json:"senderId,string" gorm:"column:sender_id;not null;index"`
	SenderName     string     `json:"senderName" gorm:"column:sender_name;not null"`
	SenderAvatar   string     `json:"senderAvatar" gorm:"column:sender_avatar;default:''"`
	ReceiverID     int64      `json:"receiverId,string" gorm:"column:receiver_id;not null;index"`
	ReceiverName   string     `json:"receiverName" gorm:"column:receiver_name;not null"`
	ReceiverAvatar string     `json:"receiverAvatar" gorm:"column:receiver_avatar;default:''"`
	Content        string     `json:"content" gorm:"column:content;not null"`
	IsRead         bool       `json:"isRead" gorm:"column:is_read;not null;default:false;index"`
	ReadAt         *time.Time `json:"readAt" gorm:"column:read_at"`
	CreateTime     *time.Time `json:"createTime" gorm:"column:create_time;default:CURRENT_TIMESTAMP;index"`
	UpdateTime     *time.Time `json:"updateTime" gorm:"column:update_time;default:CURRENT_TIMESTAMP"`
}

func (ChatMessage) TableName() string {
	return "chat_messages"
}

type WsIncomingMessage struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

type WsEvent struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type ReadReceipt struct {
	ContactID  int64      `json:"contactId,string"`
	MessageIDs []int64    `json:"messageIds"`
	ReadAt     *time.Time `json:"readAt"`
}
