package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionOptions = cobra.CompletionOptions{DisableDefaultCmd: true}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:               "alpine",
	Short:             "Create, control and connect to Alpine instances.",
	Long:              ``,
	CompletionOptions: completionOptions,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(launchCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(publishCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(execCmd)
	rootCmd.AddCommand(editCmd)
}
