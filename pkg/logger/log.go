package logger

import (
	"log"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger
var Sugar *zap.SugaredLogger

func init() {
	level, err := parseLogLevel("info")
	if err != nil {
		log.Fatal(err)
	}

	// 2. 统一encoder配置
	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, // 短文件路径
	}

	var writeSyncer zapcore.WriteSyncer
	switch strings.ToLower("console") {
	case "stdout":
		fallthrough
	default:
		writeSyncer = zapcore.Lock(os.Stdout) // 加锁防止并发日志穿插
	}

	// 4. json编码器
	encoder := zapcore.NewJSONEncoder(encoderCfg)
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// 5. 构建logger，开启调用者、堆栈
	opts := []zap.Option{
		zap.AddCaller(),                       // 打印调用文件行号
		zap.AddCallerSkip(1),                  // 跳过一层封装
		zap.AddStacktrace(zapcore.ErrorLevel), // error及以上打印堆栈
	}
	Logger = zap.New(core, opts...)
	Sugar = Logger.Sugar()

	// 替换zap全局
	zap.ReplaceGlobals(Logger)
}

// parseLogLevel 字符串转zap日志级别
func parseLogLevel(lvl string) (zapcore.Level, error) {
	switch strings.ToLower(lvl) {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	case "panic":
		return zapcore.PanicLevel, nil
	case "fatal":
		return zapcore.FatalLevel, nil
	default:
		return zapcore.InfoLevel, nil
	}
}
