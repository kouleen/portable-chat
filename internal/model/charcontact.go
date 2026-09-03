package model

import "time"

type CharContactReq struct {
	UserID      *int64 `json:"userId,string"`      // 用户ID
	ContactID   *int64 `json:"contactId,string"`   // 联系人ID
	ContactName string `json:"contactName,string"` // 联系人名称
}

// CharContact 联系人
type CharContact struct {
	ID            int64      `json:"id,string" gorm:"column:id;primary_key;not null;auto_increment"`
	UserID        int64      `json:"userId,string" gorm:"column:user_id;not null"`        // 用户ID
	ContactID     int64      `json:"contactId,string" gorm:"column:contact_id;not null"`  // 联系人ID
	ContactName   string     `json:"contactName" gorm:"column:contact_name;not null"`     // 联系人名称
	ContactAvatar string     `json:"contactAvatar" gorm:"column:contact_avatar;not null"` // 联系人头像
	Status        uint8      `json:"status" gorm:"column:status;not null;default:1"`      // 关系状态	0 拉黑  1 正常
	CreateTime    *time.Time `json:"createTime" gorm:"column:create_time;default:CURRENT_TIMESTAMP"`
	UpdateTime    *time.Time `json:"updateTime" gorm:"column:update_time;default:CURRENT_TIMESTAMP"`
}
