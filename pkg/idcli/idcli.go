package idcli

import (
	"context"
	"log"

	"github.com/kouleen/portable-chat/pkg/sqlitecli"
	"github.com/kouleen/portable-chat/pkg/template"
)

var segmentAlloc *template.SegmentAlloc

func init() {
	alloc, err := template.NewSegmentAlloc(context.Background(), sqlitecli.GetSqliteDB(), "portable-chat")
	if err != nil {
		log.Fatal(err)
	}
	segmentAlloc = alloc
}

func Next(ctx context.Context) (id int64, err error) {
	return segmentAlloc.Next(ctx)
}

func NextID(ctx context.Context) (ID, uin int64, err error) {
	ID, err = segmentAlloc.Next(ctx)
	if err != nil {
		return
	}
	uin, err = template.Obfuscate(ID)
	if err != nil {
		return
	}
	return
}

func Obfuscate(id int64) (uin int64, err error) {
	uin, err = template.Obfuscate(id)
	if err != nil {
		return
	}
	return
}

func DeObfuscate(uin int64) (id int64, err error) {
	id, err = template.DeObfuscate(uin)
	if err != nil {
		return
	}
	return
}
