package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/orchestre-dev/ccproxy/internal/process"
)

// StopCmd returns the stop command
func StopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the CCProxy service",
		Long:  "Stop the running CCProxy service and clean up resources",
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
			
			if runningPID == 0 {
				fmt.Println("Service is not running")
				return nil
			}
			
			fmt.Printf("Stopping CCProxy service (PID: %d)...\n", runningPID)
			
			// Stop the process
			if err := pidManager.StopProcess(); err != nil {
				// Try to clean up PID file anyway
				pidManager.Cleanup()
				return fmt.Errorf("failed to stop service: %w", err)
			}
			
			fmt.Println("✅ Service stopped successfully")
			return nil
		},
	}
}