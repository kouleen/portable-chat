package repository

import (
	"context"

	"github.com/kouleen/portable-chat/internal/model"
	"github.com/kouleen/portable-chat/pkg/sqlitecli"
)

func GetAuthorizationByUsername(ctx context.Context, username string) (*model.Authorization, error) {
	authorization := new(model.Authorization)
	if err := sqlitecli.GetSqliteDB().WithContext(ctx).First(authorization, "username = ?", username).First(authorization).Error; err != nil {
		return nil, err
	}
	return authorization, nil
}

func CreateAuthorization(ctx context.Context, authorization *model.Authorization) error {
	if err := sqlitecli.GetSqliteDB().WithContext(ctx).Create(authorization).Error; err != nil {
		return err
	}
	return nil
}
