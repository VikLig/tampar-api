package utils

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.elastic.co/ecszap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Declare
type Logger struct {
	Logger *zap.SugaredLogger
}

var (
	globalLogger *zap.SugaredLogger
)

func NewLogger(config Config) Logger {
	if gin.IsDebugging() {
		// Zapcore sync lumberjack
		w := zapcore.AddSync(&lumberjack.Logger{
			Filename:   config.LogDirectory,
			MaxSize:    500, // megabytes
			MaxBackups: 3,
			MaxAge:     1,    //days
			Compress:   true, // disabled by default
		})
		encoderConfig := ecszap.ECSCompatibleEncoderConfig(zap.NewDevelopmentEncoderConfig())
		encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
		encoder := zapcore.NewJSONEncoder(encoderConfig)
		// Initialize zap log
		core := zapcore.NewCore(encoder, w, zap.InfoLevel)
		newLogger := zap.New(core)
		defer newLogger.Sync() // flushes buffer, if any
		globalLogger = newLogger.Sugar()
		return Logger{Logger: globalLogger}
	} else {
		// Initialize zap log
		newLogger, err := zap.NewProduction()
		if err != nil {
			//	panic(err)
		}
		defer newLogger.Sync() // flushes buffer, if any
		globalLogger = newLogger.Sugar()
		return Logger{Logger: globalLogger}
	}
}
