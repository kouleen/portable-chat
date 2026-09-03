package logger

import (
	"context"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm/logger"
)

type ZapLogger struct{}

func NewZapLogger() logger.Interface {
	return &ZapLogger{}
}

func (z *ZapLogger) LogMode(level logger.LogLevel) logger.Interface {
	return z
}

func (z *ZapLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	Logger.Info(msg, zap.Any("data", data))
}

func (z *ZapLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	Logger.Warn(msg, zap.Any("data", data))
}

func (z *ZapLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	Logger.Error(msg, zap.Any("data", data))
}

func (z *ZapLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rows int64), err error) {
	cost := time.Since(begin)
	sql, rows := fc()
	// 从上下文读取trace_id，和HTTP请求链路打通
	traceID, _ := ctx.Value("trace_id").(string)

	// 日志放到goroutine异步打印，不占用请求耗时
	go func(biz time.Duration, traceID string) {
		fields := []zap.Field{
			zap.String("sql", sql),
			zap.Float64("cost_ms", cost.Seconds()*1000),
			zap.Int64("rows_affected", rows),
			zap.String("trace_id", traceID),
		}
		if err != nil {
			Logger.Error("gorm sql error", append(fields, zap.Error(err))...)
		} else {
			Logger.Info("gorm sql exec", fields...)
		}
	}(cost, traceID)
}
