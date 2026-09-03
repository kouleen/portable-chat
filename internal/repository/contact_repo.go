package repository

import (
	"context"

	"github.com/kouleen/portable-chat/internal/model"
	"github.com/kouleen/portable-chat/pkg/sqlitecli"
)

func ListContact(ctx context.Context, req *model.CharContactReq) ([]*model.CharContact, error) {
	query := sqlitecli.GetSqliteDB().WithContext(ctx).Where("user_id = ?", *req.UserID)
	if req.ContactID != nil {
		query = query.Where("contact_id = ? ", *req.ContactID)
	}
	if req.ContactName != "" {
		query = query.Where("contact_name like ?", "%"+req.ContactName+"%")
	}
	var contacts []*model.CharContact
	if err := sqlitecli.GetSqliteDB().WithContext(ctx).Find(&contacts).Error; err != nil {
		return nil, err
	}
	return contacts, nil
}
