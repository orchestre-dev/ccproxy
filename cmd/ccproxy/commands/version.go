package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version   string
	buildTime string
	commit    string
)

// SetVersionInfo sets the version information for all commands
func SetVersionInfo(v, b, c string) {
	version = v
	buildTime = b
	commit = c
}

// VersionCmd returns the version command
func VersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("ccproxy version %s\n", version)
			fmt.Printf("Build Time: %s\n", buildTime)
			fmt.Printf("Commit: %s\n", commit)
		},
	}
}
