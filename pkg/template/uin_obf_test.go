package template

import (
	"fmt"
	"testing"
)

func TestObfuscate(t *testing.T) {
	// 模拟底层号段分配出来有序seq
	for i := int64(100_000_000); i < 100_000_010; i++ {
		outer, err := Obfuscate(i)
		if err != nil {
			t.Fatalf("Obfuscate fail i=%d err=%v", i, err)
		}
		inner, err := DeObfuscate(outer)
		if err != nil {
			t.Fatalf("DeObfuscate fail outer=%d err=%v", outer, err)
		}
		// 核心断言：还原必须等于输入，不对直接终止测试
		if inner != i {
			t.Fatalf("MISMATCH! input=%d outer=%d got inner=%d", i, outer, inner)
		}
		fmt.Printf("原始seq:%d → 对外乱序UIN:%d, 还原:%d\n", i, outer, inner)
	}
}
