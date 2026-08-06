package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func InitLogger(pwd string) *zap.Logger {
	// 1. Define the file rotation logic via Lumberjack
	writerSync := zapcore.AddSync(&lumberjack.Logger{
		Filename:   pwd,
		MaxSize:    10,   // megabytes
		MaxBackups: 3,    // number of old log files to keep
		MaxAge:     28,   // days to keep old log files
		Compress:   true, // compress old logs (gzip)
	})

	// 2. Define the log format (JSON is best for file logs)
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // Human-readable timestamps
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	// 3. Create the core and the logger
	core := zapcore.NewCore(encoder, writerSync, zap.InfoLevel)
	return zap.New(core, zap.AddCaller()) // AddCaller shows the file/line number
}

func main() {
	logger := InitLogger("app.log")
	defer logger.Sync() // Flushes buffer before app exits

	logger.Info("Application started",
		zap.String("status", "running"),
		zap.Int("process_id", 12345),
	)
}
