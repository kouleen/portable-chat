package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"
)

// GenerateRandomByNumCode 生成数字
func GenerateRandomByNumCode(length int) string {
	const chars = "0123456789"
	return randomString(chars, length)
}

// GenerateRandomCode 生成数字+字母验证码
func GenerateRandomCode(length int) string {
	const chars = "0123456789ABCDEFGHJKLMNPQRSTWXYZabcdefhijkmnprstwxyz"
	return randomString(chars, length)
}

// 生成随机字符串
func randomString(chars string, length int) string {
	var sb strings.Builder
	charLen := big.NewInt(int64(len(chars)))

	for i := 0; i < length; i++ {
		n, _ := rand.Int(rand.Reader, charLen)
		sb.WriteByte(chars[n.Int64()])
	}
	return sb.String()
}

// GenerateUUID 生成 UUID（和 crypto.randomUUID() 一致）
func GenerateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// 生成随机深色（和你 TS 一模一样）
func randomColor() string {
	r, _ := rand.Int(rand.Reader, big.NewInt(80))
	g, _ := rand.Int(rand.Reader, big.NewInt(80))
	b, _ := rand.Int(rand.Reader, big.NewInt(80))

	return fmt.Sprintf("rgb(%d,%d,%d)", r.Int64()+60, g.Int64()+60, b.Int64()+60)
}

// CreateCaptchaSvg ===================== 核心：生成 SVG 验证码 =====================
func CreateCaptchaSvg(code string) string {
	const width = 110
	const height = 40
	const fontSize = 22

	// 随机背景色
	r, _ := rand.Int(rand.Reader, big.NewInt(10))
	g, _ := rand.Int(rand.Reader, big.NewInt(10))
	b, _ := rand.Int(rand.Reader, big.NewInt(10))
	bgColor := fmt.Sprintf("rgb(%d,%d,%d)", 245+r.Int64(), 245+g.Int64(), 245+b.Int64())

	// 4 条干扰线
	var paths strings.Builder
	for i := 0; i < 4; i++ {
		x1, _ := rand.Int(rand.Reader, big.NewInt(width))
		y1, _ := rand.Int(rand.Reader, big.NewInt(height))
		x2, _ := rand.Int(rand.Reader, big.NewInt(width))
		y2, _ := rand.Int(rand.Reader, big.NewInt(height))
		color := randomColor()

		paths.WriteString(fmt.Sprintf(
			`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s" stroke-width="1" opacity="0.3"/>`,
			x1.Int64(), y1.Int64(), x2.Int64(), y2.Int64(), color,
		))
	}

	// 文字：随机位置、旋转、颜色（和你 TS 完全一致）
	var textElements strings.Builder
	xStep := width / (len(code) + 1)

	for i, c := range code {
		xi, _ := rand.Int(rand.Reader, big.NewInt(6))
		yi, _ := rand.Int(rand.Reader, big.NewInt(6))
		ro, _ := rand.Int(rand.Reader, big.NewInt(20))

		x := xStep*(i+1) + (int(xi.Int64()) - 3)
		y := height/2 + fontSize/3 + (int(yi.Int64()) - 3)
		rotate := int(ro.Int64()) - 10
		color := randomColor()

		textElements.WriteString(fmt.Sprintf(
			`<text x="%d" y="%d" font-size="%d" fill="%s" font-weight="bold" transform="rotate(%d %d %d)">%c</text>`,
			x, y, fontSize, color, rotate, x, y, c,
		))
	}

	// 输出 SVG
	svg := fmt.Sprintf(`
	<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">
	<rect width="100%%" height="100%%" fill="%s" rx="4" ry="4"/>
	%s
	%s
	</svg>
	`, width, height, bgColor, paths.String(), textElements.String())
	base64Str := base64.StdEncoding.EncodeToString([]byte(svg))
	return "data:image/svg+xml;base64," + base64Str
}
