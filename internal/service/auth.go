package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/kouleen/portable-chat/internal/model"
	"github.com/kouleen/portable-chat/internal/repository"
	"github.com/kouleen/portable-chat/pkg/logger"
	"github.com/kouleen/portable-chat/pkg/storecli"
	tmp "github.com/kouleen/portable-chat/pkg/template"
	"github.com/kouleen/portable-chat/static"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const chars = "0123456789"

var tpl *template.Template

func init() {
	emailTpl, err := template.ParseFS(static.FS, "code.html")
	if err != nil {
		log.Fatal("读取模板失败: %w", err)
		return
	}
	log.Println("Email template initialized successfully")
	tpl = emailTpl
}

func SendCode(ctx context.Context, username string) (any, error) {
	durationTtl := storecli.Ttl(username + "_ttl")
	if durationTtl > 0 {
		return nil, errors.New("操作太快了，休息会再试吧！")
	}
	// 记录60秒
	if err := storecli.Set(username+"_ttl", "1", time.Duration(60)*time.Second); err != nil {
		return nil, err
	}
	var sb strings.Builder
	charLen := big.NewInt(int64(len(chars)))

	for i := 0; i < 6; i++ {
		n, _ := rand.Int(rand.Reader, charLen)
		sb.WriteByte(chars[n.Int64()])
	}
	code := sb.String()
	go func() {
		var buf bytes.Buffer
		if err := tpl.Execute(&buf, map[string]any{"Captcha": code}); err != nil {
			logger.Logger.Error("SendCode Execute err", zap.Error(err))
			return
		}
		tmp.SendMail(ctx, username, buf.String())
		if err := storecli.Set(username, code, time.Duration(600)*time.Second); err != nil {
			logger.Logger.Error("SendCode Set err", zap.Error(err))
		}
	}()
	return true, nil
}

func Login(ctx context.Context, req *model.AuthorizationLogin) (any, error) {
	resp, err := repository.GetAuthorizationByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if err = bcrypt.CompareHashAndPassword([]byte(resp.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}
	// 清空密码
	resp.Password = ""
	marshal, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	token := fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	if err = storecli.Set(token, string(marshal), time.Duration(24)*time.Hour); err != nil {
		return nil, err
	}
	return token, nil
}

func Exist(ctx context.Context, username string) (any, error) {
	byUsername, err := repository.GetAuthorizationByUsername(ctx, username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return byUsername != nil, nil
}

func CreateAuthorization(ctx context.Context, req *model.AuthorizationRegister) (any, error) {
	code, err := storecli.Get(req.Email)
	if err != nil {
		return nil, err
	}
	if req.Code != code {
		return false, errors.New("验证码不正确！")
	}
	if err := storecli.Del(req.Username); err != nil {
		return nil, err
	}
	resp, err := repository.GetAuthorizationByUsername(ctx, req.Username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if resp != nil {
		return nil, errors.New("该用户名已存在")
	}
	hashPwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	password := string(hashPwd)
	authorization := &model.Authorization{
		Username: req.Username,
		Password: password,
		Email:    req.Email,
	}
	if err = repository.CreateAuthorization(ctx, authorization); err != nil {
		return nil, err
	}
	return true, nil
}
