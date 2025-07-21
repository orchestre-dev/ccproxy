package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/orchestre-dev/ccproxy/internal/claudeconfig"
	"github.com/spf13/cobra"
)

// ClaudeCmd returns the claude command
func ClaudeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claude",
		Short: "Manage Claude configuration",
		Long:  "Manage the ~/.claude.json configuration file for Claude Code integration",
	}

	cmd.AddCommand(claudeInitCmd())
	cmd.AddCommand(claudeShowCmd())
	cmd.AddCommand(claudeResetCmd())
	cmd.AddCommand(claudeSetCmd())

	return cmd
}

// claudeInitCmd initializes the Claude configuration
func claudeInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize Claude configuration",
		Long:  "Initialize the ~/.claude.json configuration file with default values",
		RunE: func(cmd *cobra.Command, args []string) error {
			force, _ := cmd.Flags().GetBool("force")

			manager, err := claudeconfig.NewManager()
			if err != nil {
				return fmt.Errorf("failed to create manager: %w", err)
			}

			if claudeconfig.Exists() && !force {
				fmt.Println("Claude configuration already exists at ~/.claude.json")
				return nil
			}

			if err := manager.Initialize(); err != nil {
				return fmt.Errorf("failed to initialize: %w", err)
			}

			fmt.Println("✅ Claude configuration initialized at ~/.claude.json")
			return nil
		},
	}

	cmd.Flags().BoolP("force", "f", false, "Force overwrite existing configuration")
	return cmd
}

// claudeShowCmd displays the current Claude configuration
func claudeShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show Claude configuration",
		Long:  "Display the current Claude configuration from ~/.claude.json",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !claudeconfig.Exists() {
				fmt.Println("Claude configuration not found. Run 'ccproxy claude init' to create it.")
				return nil
			}

			manager, err := claudeconfig.NewManager()
			if err != nil {
				return fmt.Errorf("failed to create manager: %w", err)
			}

			if err := manager.Load(); err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			config := manager.Get()
			data, err := json.MarshalIndent(config, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format config: %w", err)
			}

			fmt.Println(string(data))
			return nil
		},
	}
}

// claudeResetCmd resets the Claude configuration
func claudeResetCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset Claude configuration",
		Long:  "Reset the Claude configuration to default values",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !claudeconfig.Exists() {
				fmt.Println("Claude configuration not found.")
				return nil
			}

			if !force {
				fmt.Print("Are you sure you want to reset the Claude configuration? (y/N): ")
				var response string
				// Safe to ignore scan error as we have fallback behavior
				_, _ = fmt.Scanln(&response)
				if response != "y" && response != "Y" {
					fmt.Println("Reset canceled.")
					return nil
				}
			}

			// Get config path and remove it
			configPath, err := claudeconfig.GetConfigPath()
			if err != nil {
				return fmt.Errorf("failed to get config path: %w", err)
			}

			if err := os.Remove(configPath); err != nil {
				return fmt.Errorf("failed to remove config: %w", err)
			}

			// Create new config
			manager, err := claudeconfig.NewManager()
			if err != nil {
				return fmt.Errorf("failed to create manager: %w", err)
			}

			if err := manager.Initialize(); err != nil {
				return fmt.Errorf("failed to initialize new config: %w", err)
			}

			fmt.Println("✅ Claude configuration has been reset to defaults")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force reset without confirmation")
	return cmd
}

// claudeSetCmd sets specific configuration values
func claudeSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set Claude configuration value",
		Long: `Set a specific Claude configuration value.

Available keys:
  - auto-updater (enabled/disabled)
  - telemetry (true/false)
  - theme (system/light/dark)
  - onboarding (complete/reset)`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]

			manager, err := claudeconfig.NewManager()
			if err != nil {
				return fmt.Errorf("failed to create manager: %w", err)
			}

			if !claudeconfig.Exists() {
				if err := manager.Initialize(); err != nil {
					return fmt.Errorf("failed to initialize config: %w", err)
				}
			} else {
				if err := manager.Load(); err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}
			}

			config := manager.Get()

			switch key {
			case "auto-updater":
				if value != "enabled" && value != "disabled" {
					return fmt.Errorf("invalid value for auto-updater: %s (must be 'enabled' or 'disabled')", value)
				}
				config.AutoUpdaterStatus = value

			case "telemetry":
				if value == "true" {
					config.TelemetryEnabled = true
				} else if value == "false" {
					config.TelemetryEnabled = false
				} else {
					return fmt.Errorf("invalid value for telemetry: %s (must be 'true' or 'false')", value)
				}

			case "theme":
				if value != "system" && value != "light" && value != "dark" {
					return fmt.Errorf("invalid value for theme: %s (must be 'system', 'light', or 'dark')", value)
				}
				config.Theme = value

			case "onboarding":
				if value == "complete" {
					if err := manager.CompleteOnboarding("1.0.17"); err != nil {
						return fmt.Errorf("failed to complete onboarding: %w", err)
					}
					fmt.Println("✅ Onboarding marked as completed")
					return nil
				} else if value == "reset" {
					config.HasCompletedOnboarding = false
					config.LastOnboardingVersion = ""
				} else {
					return fmt.Errorf("invalid value for onboarding: %s (must be 'complete' or 'reset')", value)
				}

			default:
				return fmt.Errorf("unknown configuration key: %s", key)
			}

			if err := manager.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("✅ Set %s = %s\n", key, value)
			return nil
		},
	}
}
