package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/orchestre-dev/ccproxy/cmd/ccproxy/commands"
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
	
	// Set version info for commands to use
	commands.SetVersionInfo(Version, BuildTime, Commit)
	
	// Add commands
	rootCmd.AddCommand(commands.StartCmd())
	rootCmd.AddCommand(commands.StopCmd())
	rootCmd.AddCommand(commands.StatusCmd())
	rootCmd.AddCommand(commands.CodeCmd())
	rootCmd.AddCommand(commands.ClaudeCmd())
	rootCmd.AddCommand(commands.VersionCmd())
	rootCmd.AddCommand(commands.EnvCmd())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}