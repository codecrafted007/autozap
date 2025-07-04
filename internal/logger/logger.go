package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var globalSugaredLogger *zap.SugaredLogger

func InitLogger() {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.CallerKey = "caller"

	logger, err := config.Build(zap.AddCaller())
	if err != nil {
		panic(err)
	}

	globalSugaredLogger = logger.Sugar()

}

func L() *zap.SugaredLogger {
	if globalSugaredLogger == nil {
		panic("zap logger not initialized, call InitLogger first")
	}
	return globalSugaredLogger
}
