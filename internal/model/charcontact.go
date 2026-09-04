package model

import "time"

type CharContactQuery struct {
	UserID  int64  `json:"userId,string"`
	Keyword string `form:"keyword" json:"keyword"`
}

// CharContact 联系人列表视图
type CharContact struct {
	ID            int64      `json:"id,string"`
	Uin           int64      `json:"uin,string"`
	ContactName   string     `json:"contactName"`
	ContactAvatar string     `json:"contactAvatar"`
	LastMessage   string     `json:"lastMessage"`
	LastMessageAt *time.Time `json:"lastMessageAt"`
	UnreadCount   int64      `json:"unreadCount"`
}
