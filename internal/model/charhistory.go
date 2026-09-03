package model

import (
	"time"
)

// CharHistory 聊天记录
type CharHistory struct {
	ID            int64      `json:"id,string" gorm:"column:id;primary_key;not null;auto_increment"`
	CharID        int64      `json:"charId,string" gorm:"column:char_id;not null"`        // 会话ID
	UserID        int64      `json:"userId,string" gorm:"column:user_id;not null"`        // 用户ID
	UserName      string     `json:"username" gorm:"column:username;not null"`            // 用户名称
	Avatar        string     `json:"avatar" gorm:"column:avatar;default:''"`              // 用户头像
	ContactID     int64      `json:"contactId,string" gorm:"column:contact_id;not null"`  // 联系人ID
	ContactName   string     `json:"contactName" gorm:"column:contact_name;not null"`     // 联系人名称
	ContactAvatar string     `json:"contactAvatar" gorm:"column:contact_avatar;not null"` // 联系人头像
	CharContext   string     `json:"charContact" gorm:"column:char_contact;not null"`     // 消息内容
	CreateTime    *time.Time `json:"createTime" gorm:"column:create_time;default:CURRENT_TIMESTAMP"`
	UpdateTime    *time.Time `json:"updateTime" gorm:"column:update_time;default:CURRENT_TIMESTAMP"`
}
