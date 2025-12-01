package logger

import (
	"testing"
)

func TestInitLogger(t *testing.T) {
	t.Run("Initialize Logger", func(t *testing.T) {
		// Reset the global logger
		globalSugaredLogger = nil

		InitLogger()

		if globalSugaredLogger == nil {
			t.Fatal("Expected logger to be initialized, got nil")
		}
	})
}

func TestL(t *testing.T) {
	t.Run("Get Logger After Init", func(t *testing.T) {
		globalSugaredLogger = nil
		InitLogger()

		logger := L()
		if logger == nil {
			t.Fatal("Expected logger instance, got nil")
		}
	})

	t.Run("Panic When Logger Not Initialized", func(t *testing.T) {
		globalSugaredLogger = nil

		defer func() {
			if r := recover(); r == nil {
				t.Fatal("Expected panic when logger not initialized, got none")
			}
		}()

		L()
	})
}

func TestLoggerUsage(t *testing.T) {
	t.Run("Logger Can Log Messages", func(t *testing.T) {
		globalSugaredLogger = nil
		InitLogger()

		logger := L()

		// These shouldn't panic
		logger.Info("Test info message")
		logger.Infof("Test info message with format: %s", "value")
		logger.Infow("Test info message with fields", "key", "value")
	})
}
