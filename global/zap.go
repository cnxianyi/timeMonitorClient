package global

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	Logger *zap.Logger
)

// InitLogger 初始化 zap 日志
func InitLogger() {
	// 获取项目根目录
	rootDir, _ := os.Getwd()

	// 替换日志路径中的 @/ 为项目根目录
	logPath := filepath.Join(rootDir, "log")
	if strings.HasPrefix(logPath, "@/") {
		logPath = filepath.Join(rootDir, strings.TrimPrefix(logPath, "@/"))
	}

	// 确保日志目录存在
	if err := os.MkdirAll(logPath, 0755); err != nil {
		panic("创建日志目录失败: " + err.Error())
	}

	// 配置日志输出
	encoder := getEncoder()

	// 日志文件写入器
	writeSyncer := getLogWriter(filepath.Join(logPath, "app.log"))

	// 创建核心
	core := zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel)

	// 创建 Logger
	Logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	// 替换全局 Logger
	zap.ReplaceGlobals(Logger)
}

// getEncoder 获取日志编码器
func getEncoder() zapcore.Encoder {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     timeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	return zapcore.NewJSONEncoder(encoderConfig)
}

// timeEncoder 自定义时间编码器
func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

// getLogWriter 获取日志写入器
func getLogWriter(filename string) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   filename, // 日志文件路径
		MaxSize:    1,        // 每个日志文件最大尺寸，单位 MB
		MaxBackups: 30,       // 保留旧日志文件的最大个数
		MaxAge:     30,       // 保留旧日志文件的最大天数
		Compress:   false,    // 是否压缩归档的日志文件
	}
	return zapcore.AddSync(lumberJackLogger)
}

// Debug 输出 debug 级别日志
func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

// Info 输出 info 级别日志
func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

// Warn 输出 warn 级别日志
func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

// Error 输出 error 级别日志
func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

// Fatal 输出 fatal 级别日志
func Fatal(msg string, fields ...zap.Field) {
	Logger.Fatal(msg, fields...)
}

// JSON 将任意类型转换为JSON并记录日志
func JSON(level string, msg string, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		Error("JSON序列化失败", zap.Error(err))
		return
	}

	jsonField := zap.String("json_data", string(jsonData))

	switch strings.ToLower(level) {
	case "debug":
		Debug(msg, jsonField)
	case "info":
		Info(msg, jsonField)
	case "warn":
		Warn(msg, jsonField)
	case "error":
		Error(msg, jsonField)
	case "fatal":
		Fatal(msg, jsonField)
	default:
		Info(msg, jsonField)
	}
}

// DebugJSON 输出 debug 级别的JSON日志
func DebugJSON(msg string, data interface{}) {
	JSON("debug", msg, data)
}

// InfoJSON 输出 info 级别的JSON日志
func InfoJSON(msg string, data interface{}) {
	JSON("info", msg, data)
}

// WarnJSON 输出 warn 级别的JSON日志
func WarnJSON(msg string, data interface{}) {
	JSON("warn", msg, data)
}

// ErrorJSON 输出 error 级别的JSON日志
func ErrorJSON(msg string, data interface{}) {
	JSON("error", msg, data)
}

// FatalJSON 输出 fatal 级别的JSON日志
func FatalJSON(msg string, data interface{}) {
	JSON("fatal", msg, data)
}
