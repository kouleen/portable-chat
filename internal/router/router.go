package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kouleen/portable-chat/internal/handler"
	"github.com/kouleen/portable-chat/internal/middleware"
	"github.com/kouleen/portable-chat/static"
)

func Register(r *gin.Engine) {
	r.StaticFS("/static", http.FS(static.FS))

	publicGroup := r.Group("/portable")
	{
		r.GET("/", WebHtml)
		publicGroup.GET("/sendCode", handler.SendCode)
		publicGroup.POST("/login", handler.Login)
		publicGroup.GET("/exist", handler.Exist)
		publicGroup.POST("/register", handler.Register)
	}
	rootGroup := publicGroup.Group("/root")
	rootGroup.Use(middleware.Auth())
	{
		rootGroup.GET("/info", handler.Info)
		rootGroup.POST("/logout", handler.Logout)
		rootGroup.GET("/contacts", handler.ListContact)
		rootGroup.GET("/history", handler.ListHistory)
		rootGroup.GET("/events", handler.StreamEvents)
		rootGroup.GET("/ws", handler.ChatRoom)
	}
}

func WebHtml(c *gin.Context) {
	// 读取内嵌html文件内容
	data, err := static.FS.ReadFile("index.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "页面加载失败: %v", err)
		return
	}
	// 返回html页面，设置Content-Type
	c.Data(http.StatusOK, "text/html; charset=utf-8", data)
}
