package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/kouleen/portable-chat/pkg/logger"
	utils "github.com/kouleen/portable-chat/pkg/util"

	"go.uber.org/zap"
)

func RouteRecover(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			// 2. 打印完整调用堆栈（[]byte）
			stack := debug.Stack()
			logger.Logger.Error("request panic", zap.Any("error", err), zap.Stack(string(stack)))
			fmt.Printf("error: %v\n", err)
			fmt.Printf("debug stack:\n%s\n", stack)
			utils.Error(c, http.StatusInternalServerError, "系统内部错误")
			c.Abort()
		}
	}()
	c.Next()
}
