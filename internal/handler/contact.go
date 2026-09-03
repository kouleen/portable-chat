package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kouleen/portable-chat/internal/model"
	"github.com/kouleen/portable-chat/internal/service"
)

func ListContact(c *gin.Context) {
	ctx := c.Request.Context()
	var req model.CharContactReq
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": err.Error()})
		return
	}
	list, err := service.ListContact(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "timestamp": time.Now().UnixMilli(), "data": list})
}
