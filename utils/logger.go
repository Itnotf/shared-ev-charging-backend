package utils

import (
	"io"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// logger 是全局 *zap.SugaredLogger 实例。
var logger *zap.SugaredLogger

// InitLogger 初始化全局 logger。
//   - mode: "prod" | "dev" 决定编码器类型（json or console）。
//   - level: "debug" | "info" | "warn" | "error" | "fatal".
//   - filePath: 日志文件路径；若为空则只输出到 stdout。
func InitLogger(mode, level, filePath string) error {
	// 日志级别解析
	var zapLevel zapcore.Level
	switch strings.ToLower(level) {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	case "fatal":
		zapLevel = zapcore.FatalLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	// 编码器
	encCfg := zap.NewProductionEncoderConfig()
	encCfg.TimeKey = "ts"
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encCfg.EncodeDuration = zapcore.StringDurationEncoder

	var encoder zapcore.Encoder
	if strings.ToLower(mode) == "dev" {
		encoder = zapcore.NewConsoleEncoder(encCfg)
	} else {
		encoder = zapcore.NewJSONEncoder(encCfg)
	}

	// 输出目标
	var sink io.Writer = os.Stdout
	if filePath != "" {
		sink = &lumberjack.Logger{
			Filename:   filePath,
			MaxSize:    128, // MB
			MaxBackups: 10,
			MaxAge:     7, // days
			Compress:   true,
		}
	}

	core := zapcore.NewCore(encoder, zapcore.AddSync(sink), zapLevel)

	// 添加 caller & stacktrace
	z := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	logger = z.Sugar()
	return nil
}

// Sync 刷新缓冲区，应在程序退出前调用。
func Sync() {
	if logger != nil {
		_ = logger.Sync()
	}
}

// GetLogger 返回全局 logger，如果未初始化直接 panic。
func GetLogger() *zap.SugaredLogger {
	if logger == nil {
		panic("Logger not initialized! 请先调用 InitLogger")
	}
	return logger
}

// 以下是便捷的日志记录函数，简化日志记录的使用

// CtxLogger 根据 gin.Context 获取带 trace_id 的 logger
func CtxLogger(c *gin.Context) *zap.SugaredLogger {
	if c == nil {
		return GetLogger()
	}
	if v, ok := c.Get("logger"); ok {
		if lg, ok := v.(*zap.SugaredLogger); ok {
			return lg
		}
	}
	return GetLogger()
}

// 通用日志函数，自动判断是否带 gin.Context
func logWithCtx(ctx interface{}, level string, msg string, args ...interface{}) {
	var l *zap.SugaredLogger
	if c, ok := ctx.(*gin.Context); ok && c != nil {
		l = CtxLogger(c)
	} else {
		l = GetLogger()
	}
	// 统一处理日志级别
	switch strings.ToLower(level) {
	case "debug":
		l.Debugf(msg, args...)
	case "info":
		l.Infof(msg, args...)
	case "warn":
		l.Warnf(msg, args...)
	case "error":
		l.Errorf(msg, args...)
	case "fatal":
		l.Fatalf(msg, args...)
	}
}

// 简化后的便捷日志函数
func Debug(msg string, args ...interface{}) { logWithCtx(nil, "debug", msg, args...) }
func Info(msg string, args ...interface{})  { logWithCtx(nil, "info", msg, args...) }
func Warn(msg string, args ...interface{})  { logWithCtx(nil, "warn", msg, args...) }
func Error(msg string, args ...interface{}) { logWithCtx(nil, "error", msg, args...) }
func Fatal(msg string, args ...interface{}) { logWithCtx(nil, "fatal", msg, args...) }

func DebugCtx(c *gin.Context, msg string, args ...interface{}) { logWithCtx(c, "debug", msg, args...) }
func InfoCtx(c *gin.Context, msg string, args ...interface{})  { logWithCtx(c, "info", msg, args...) }
func WarnCtx(c *gin.Context, msg string, args ...interface{})  { logWithCtx(c, "warn", msg, args...) }
func ErrorCtx(c *gin.Context, msg string, args ...interface{}) { logWithCtx(c, "error", msg, args...) }
func FatalCtx(c *gin.Context, msg string, args ...interface{}) { logWithCtx(c, "fatal", msg, args...) }
