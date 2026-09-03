package utils

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Sign    int64       `json:"sign"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	TraceId string      `json:"trace_id"`
}

func Success(c *gin.Context, data interface{}) {
	traceID := ""
	if tid, exists := c.Get("trace_id"); exists {
		traceID = tid.(string)
	}

	c.JSON(http.StatusOK, Response{
		Sign:    time.Now().UnixMilli(),
		Code:    http.StatusOK,
		Message: http.StatusText(http.StatusOK),
		Data:    data,
		TraceId: traceID,
	})
}

func OsSrsSuccess(c *gin.Context, data interface{}) {
	traceID := ""
	if tid, exists := c.Get("trace_id"); exists {
		traceID = tid.(string)
	}

	c.JSON(http.StatusOK, Response{
		Sign:    time.Now().UnixMilli(),
		Code:    0,
		Message: http.StatusText(http.StatusOK),
		Data:    data,
		TraceId: traceID,
	})
}

func Error(c *gin.Context, statusCode int, message string) {
	traceID := ""
	if tid, exists := c.Get("trace_id"); exists {
		traceID = tid.(string)
	}
	c.JSON(http.StatusOK, Response{
		Sign:    time.Now().UnixMilli(),
		Code:    statusCode,
		Message: message,
		Data:    nil,
		TraceId: traceID,
	})
}
