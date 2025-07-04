/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"os"

	"github.com/codecrafted007/autozap/cmd"
	"github.com/codecrafted007/autozap/internal/logger"
)

func main() {
	logger.InitLogger()

	defer func() {
		if err := logger.L().Sync(); err != nil {
			os.Stderr.WriteString("Failed to sync logger: " + err.Error() + "\n")
		}
	}()

	if err := cmd.Execute(); err != nil {
		logger.L().Errorf("CLI execution failed: %v", err)
		os.Exit(1)
	}
}
