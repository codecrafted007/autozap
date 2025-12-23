package logger

import (
	"os"
	"path/filepath"

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

// NewWorkflowLogger creates a dedicated logger for a specific workflow
// If logDir is empty, returns the global logger (stdout)
// If logDir is specified, creates a separate log file for the workflow
func NewWorkflowLogger(workflowName, logDir string) (*zap.SugaredLogger, error) {
	// If no log directory specified, use global logger
	if logDir == "" {
		return L().With("workflow_name", workflowName), nil
	}

	// Create log directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	// Create log file path
	logFile := filepath.Join(logDir, workflowName+".log")

	// Open log file
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	// Create encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.CallerKey = "caller"

	// Create core with file output
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(file),
		zapcore.InfoLevel,
	)

	// Create logger with workflow name field
	logger := zap.New(core, zap.AddCaller()).Sugar()
	return logger.With("workflow_name", workflowName), nil
}
