package service

import (
	"context"

	"github.com/kouleen/portable-chat/internal/model"
	"github.com/kouleen/portable-chat/internal/repository"
)

func ListContact(ctx context.Context, req *model.CharContactReq) ([]*model.CharContact, error) {
	return repository.ListContact(ctx, req)
}
