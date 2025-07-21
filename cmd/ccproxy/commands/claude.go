package commands

import (
	"fmt"

	"github.com/orchestre-dev/ccproxy/internal/claudeconfig"
	"github.com/spf13/cobra"
)

// ClaudeCmd returns the claude command
func ClaudeCmd() *cobra.Command {
	var initialize bool

	cmd := &cobra.Command{
		Use:   "claude",
		Short: "Manage ~/.claude.json configuration",
		Long:  "Read or initialize the Claude Code configuration file at ~/.claude.json",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := claudeconfig.NewManager()
			if err != nil {
				return fmt.Errorf("failed to create config manager: %w", err)
			}

			if initialize {
				// Initialize mode
				if err := manager.Initialize(); err != nil {
					return fmt.Errorf("failed to initialize config: %w", err)
				}
				fmt.Printf("✅ Initialized ~/.claude.json\n")
				return nil
			}

			// Read mode
			if !claudeconfig.Exists() {
				fmt.Println("~/.claude.json not found (run with --init to create)")
				return nil
			}

			if err := manager.Load(); err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			config := manager.Get()
			fmt.Printf("📄 ~/.claude.json\n")
			fmt.Printf("User ID: %s\n", config.UserID)
			fmt.Printf("Startups: %d\n", config.NumStartups)
			fmt.Printf("Onboarding: %t\n", config.HasCompletedOnboarding)
			fmt.Printf("Telemetry: %t\n", config.TelemetryEnabled)
			fmt.Printf("Theme: %s\n", config.Theme)

			if config.PreferredModel != "" {
				fmt.Printf("Preferred Model: %s\n", config.PreferredModel)
			}

			if config.LastLaunchDate != nil {
				fmt.Printf("Last Launch: %s\n", config.LastLaunchDate.Format("2006-01-02 15:04:05"))
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&initialize, "init", "i", false, "Initialize ~/.claude.json with defaults")

	return cmd
}

