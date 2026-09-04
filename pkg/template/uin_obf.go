package template

import "errors"

const (
	minUin int64 = 100_000_000
	maxUin int64 = 999_999_999
	M      int64 = maxUin - minUin + 1 // 900_000_000
)

// a 与 M 互质，上线禁止修改！
const a int64 = 626262629

// a 模 M 的乘法逆元，精确计算值，上线禁止修改！
const aInv int64 = 146863469

// Obfuscate 内部有序seq → 对外乱序9位UIN
func Obfuscate(seq int64) (int64, error) {
	if seq < minUin || seq > maxUin {
		return 0, errors.New("seq out of 9-bit uin range")
	}
	x := seq - minUin
	y := mulMod(x, a, M)
	outer := y + minUin
	return outer, nil
}

// DeObfuscate 对外UIN还原内部原始seq
func DeObfuscate(uin int64) (int64, error) {
	if uin < minUin || uin > maxUin {
		return 0, errors.New("uin out of 9-bit uin range")
	}
	y := uin - minUin
	x := mulMod(y, aInv, M)
	seq := x + minUin
	return seq, nil
}

// mulMod 防止int64乘法溢出 (a*b) mod m
func mulMod(a, b, m int64) int64 {
	var res int64 = 0
	a = a % m
	for b > 0 {
		if b&1 == 1 {
			res = (res + a) % m
		}
		a = (a * 2) % m
		b >>= 1
	}
	return res
}
