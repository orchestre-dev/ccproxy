package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/musistudio/ccproxy/internal/config"
	"github.com/musistudio/ccproxy/internal/process"
	"github.com/musistudio/ccproxy/internal/utils"
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
			
			// Display status
			fmt.Println("🤖 CCProxy Status")
			fmt.Println("─────────────────")
			
			if runningPID > 0 {
				fmt.Println("📊 Status: ✅ Running")
				fmt.Printf("🔢 PID: %d\n", runningPID)
				fmt.Printf("🌐 Port: %d\n", cfg.Port)
				fmt.Printf("🔗 Endpoint: http://%s:%d\n", cfg.Host, cfg.Port)
			} else {
				fmt.Println("📊 Status: ❌ Not running")
			}
			
			fmt.Printf("📂 PID File: %s\n", homeDir.PIDPath)
			
			// Show available commands based on status
			fmt.Println("\n💡 Available Commands:")
			if runningPID > 0 {
				fmt.Println("   • ccproxy stop    - Stop the service")
				fmt.Println("   • ccproxy code    - Run Claude Code")
			} else {
				fmt.Println("   • ccproxy start   - Start the service")
			}
			
			return nil
		},
	}
}