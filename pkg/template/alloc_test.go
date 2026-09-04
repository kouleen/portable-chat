package template

import (
	"context"
	"fmt"
	"testing"

	"github.com/kouleen/portable-chat/pkg/sqlitecli"
)

func TestNext(t *testing.T) {
	ctx := context.Background()
	segmentAlloc, err := NewSegmentAlloc(ctx, sqlitecli.GetSqliteDB(), "portable-chat")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 100; i++ {
		next, err := segmentAlloc.Next(ctx)
		if err != nil {
			t.Fatal(err)
		}
		nextUin, err := Obfuscate(next)
		if err != nil {
			t.Fatal(err)
		}
		inner, err := DeObfuscate(nextUin)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("原始seq:%d → 对外乱序UIN:%d, 还原:%d\n", next, nextUin, inner)
	}
}
