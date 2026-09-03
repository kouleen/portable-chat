package model

import "time"

type AuthorizationRegister struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	Email    string `json:"email" validate:"required"`
	Code     string `json:"code" validate:"required"`
}
type AuthorizationLogin struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type Authorization struct {
	ID         int64      `json:"id,string" gorm:"column:id;primary_key;not null;auto_increment"`
	Username   string     `json:"username" gorm:"column:username;not null;uq_auth_username"`
	Password   string     `json:"password" gorm:"column:password;not null"`
	Email      string     `json:"email" gorm:"column:email;not null"`
	Avatar     string     `json:"avatar" gorm:"column:avatar;default:''"`
	CreateTime *time.Time `json:"createTime" gorm:"column:create_time;default:CURRENT_TIMESTAMP"`
	UpdateTime *time.Time `json:"updateTime" gorm:"column:update_time;default:CURRENT_TIMESTAMP"`
}

func (Authorization) TableName() string {
	return "authorizations"
}
