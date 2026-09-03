package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kouleen/portable-chat/pkg/logger"
	"go.uber.org/zap"
)

const (
	TraceIDKey    = "trace_id"
	TraceIDHeader = "X-Trace-ID"
)

func Trace() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(TraceIDHeader)
		if traceID == "" {
			traceID = uuid.New().String()
		}

		c.Set(TraceIDKey, traceID)
		c.Header(TraceIDHeader, traceID)

		c.Next()
	}
}

func GetTraceID(c *gin.Context) string {
	if traceID, exists := c.Get(TraceIDKey); exists {
		return traceID.(string)
	}
	return ""
}

func GetLogger(c *gin.Context) *zap.Logger {
	traceID := GetTraceID(c)
	return logger.Logger.With(zap.String("trace_id", traceID))
}
