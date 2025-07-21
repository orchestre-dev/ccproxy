package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/orchestre-dev/ccproxy/internal/config"
	"github.com/orchestre-dev/ccproxy/internal/process"
	"github.com/orchestre-dev/ccproxy/internal/utils"
)

// StatusCmd returns the status command
func StatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check the status of the CCProxy service",
		Long:  "Display the current status of the CCProxy service with detailed information",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create PID manager
			pidManager, err := process.NewPIDManager()
			if err != nil {
				return fmt.Errorf("failed to create PID manager: %w", err)
			}
			
			// Check if running
			runningPID, err := pidManager.GetRunningPID()
			if err != nil {
				return fmt.Errorf("failed to check running status: %w", err)
			}
			
			// Get home directory info
			homeDir, err := utils.InitializeHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			
			// Load configuration to get port
			configService := config.NewService()
			configService.Load() // Ignore error, use defaults if fails
			cfg := configService.Get()
			
			// Display status with exact formatting from TypeScript version
			fmt.Println("")
			fmt.Println("📊 Claude Code Router Status")
			fmt.Println("════════════════════════════════════════")
			
			if runningPID > 0 {
				fmt.Println("✅ Status: Running")
				fmt.Printf("🆔 Process ID: %d\n", runningPID)
				fmt.Printf("🌐 Port: %d\n", cfg.Port)
				fmt.Printf("📡 API Endpoint: http://%s:%d\n", cfg.Host, cfg.Port)
				fmt.Printf("📄 PID File: %s\n", homeDir.PIDPath)
				fmt.Println("")
				fmt.Println("🚀 Ready to use! Run the following commands:")
				fmt.Println("   ccproxy code    # Start coding with Claude")
				fmt.Println("   ccproxy stop    # Stop the service")
			} else {
				fmt.Println("❌ Status: Not Running")
				fmt.Println("")
				fmt.Println("💡 To start the service:")
				fmt.Println("   ccproxy start")
			}
			
			fmt.Println("")
			
			return nil
		},
	}
}