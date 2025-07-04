/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "autozap",
	Short: "AutoZap: A self-hosted, event-driven automation engine.",
	Long: `AutoZap is a lightweight, local-first, terminal-friendly automation engine
that allows users to define workflows in YAML that react to events
(like cron schedules or file changes) and perform actions (like running Bash commands).
Think of it as “Zapier for infra and Bash scripts” — without the cloud.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(runCmd)
}
