package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version is set at build time using ldflags
	Version   = "dev"
	BuildTime = "unknown"
	Commit    = "unknown"
	
	rootCmd = &cobra.Command{
		Use:   "ccproxy",
		Short: "CCProxy - Intelligent LLM proxy for Claude Code",
		Long: `CCProxy is a Go-based proxy server that acts as an intelligent intermediary 
between Claude Code and various Large Language Model (LLM) providers.`,
		Version: Version,
	}
)

func init() {
	// Disable default completion command
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	
	// Add version command
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("ccproxy version %s\n", Version)
			fmt.Printf("Build Time: %s\n", BuildTime)
			fmt.Printf("Commit: %s\n", Commit)
		},
	}
	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}