package service

import (
	"context"
	"fmt"

	"github.com/kouleen/portable-chat/internal/model"
	"github.com/kouleen/portable-chat/internal/repository"
	utils "github.com/kouleen/portable-chat/pkg/util"
)

func ListContact(ctx context.Context, req *model.CharContactQuery) ([]*model.CharContact, error) {
	authorization := utils.GetAuthorization(ctx)
	if authorization == nil {
		return nil, fmt.Errorf("authorization is nil")
	}
	req.UserID = authorization.ID
	return repository.ListContact(ctx, req)
}
