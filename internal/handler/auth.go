package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kouleen/portable-chat/internal/model"
	"github.com/kouleen/portable-chat/internal/service"
	utils "github.com/kouleen/portable-chat/pkg/util"
	"github.com/kouleen/portable-chat/pkg/validator"
)

func SendCode(c *gin.Context) {
	ctx := c.Request.Context()
	email := c.Query("email")
	if !validator.ValidateEmail(email) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": "Invalid email address"})
		return
	}
	resp, err := service.SendCode(ctx, email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "timestamp": time.Now().UnixMilli(), "data": resp})
}

func Login(c *gin.Context) {
	ctx := c.Request.Context()
	var req model.AuthorizationLogin
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": err.Error()})
		return
	}
	if err := validator.ValidateStruct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": err.Error()})
		return
	}
	resp, err := service.Login(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "timestamp": time.Now().UnixMilli(), "data": resp})
}

func Exist(c *gin.Context) {
	ctx := c.Request.Context()
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": "Username is required"})
		return
	}
	resp, err := service.Exist(ctx, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "timestamp": time.Now().UnixMilli(), "data": resp})
}

func Register(c *gin.Context) {
	ctx := c.Request.Context()
	var req model.AuthorizationRegister
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": err.Error()})
		return
	}
	if err := validator.ValidateStruct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": err.Error()})
		return
	}
	resp, err := service.CreateAuthorization(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "timestamp": time.Now().UnixMilli(), "data": resp})
}

func Info(c *gin.Context) {
	ctx := c.Request.Context()
	user := utils.GetAuthorization(ctx)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": "Unauthorized"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "timestamp": time.Now().UnixMilli(), "data": user.Profile()})
}

func Logout(c *gin.Context) {
	ctx := c.Request.Context()
	if err := service.Logout(ctx); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "timestamp": time.Now().UnixMilli(), "data": true})
}
