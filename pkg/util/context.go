package utils

import (
	"context"
	"errors"

	"github.com/kouleen/portable-chat/internal/model"
)

// 自定义key，防止冲突
type contextKey string

const AUTHORIZATION = contextKey("Authorization")

// SetAuthorization 中间件里把用户放进 ctx
func SetAuthorization(ctx context.Context, user *model.Authorization) context.Context {
	return context.WithValue(ctx, AUTHORIZATION, user)
}

// GetAuthorization 全局获取登录用户（你要的核心方法！）
func GetAuthorization(ctx context.Context) *model.Authorization {
	user, ok := ctx.Value(AUTHORIZATION).(*model.Authorization)
	if !ok {
		return nil
	}
	return user
}

// MustGetAuthorization 必须登录，否则 panic / 返回错误
func MustGetAuthorization(ctx context.Context) (*model.Authorization, error) {
	user := GetAuthorization(ctx)
	if user == nil {
		return nil, errors.New("unauthenticated")
	}
	return user, nil
}
