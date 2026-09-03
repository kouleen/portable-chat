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
		publicGroup.GET("/sendCode", handler.SendCode)
		publicGroup.POST("/login", handler.Login)
		publicGroup.GET("/exist", handler.Exist)
		publicGroup.POST("/register", handler.Register)
		r.GET("/", WebHtml)
	}
	rootGroup := publicGroup.Group("/root")
	rootGroup.Use(middleware.Auth())
	{
		rootGroup.GET("/info", handler.Info)
		rootGroup.GET("/page", handler.ListContact)
		//rootGroup.GET("/refresh", api.RefreshAcme)
		//rootGroup.GET("/download", api.DownloadAcme)
		//rootGroup.POST("/create", api.CreateAcme)
		//rootGroup.POST("/put", api.PutAcme)
		//rootGroup.PUT("/updateAuto", api.UpdateAuto)
		//rootGroup.PUT("/updateNotice", api.UpdateNotice)
		//rootGroup.DELETE("/delete", api.DeleteAcme)
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
