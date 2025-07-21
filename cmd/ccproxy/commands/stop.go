package commands

import (
	"fmt"
	"time"

	"github.com/orchestre-dev/ccproxy/internal/process"
	"github.com/spf13/cobra"
)

// StopCmd returns the stop command
func StopCmd() *cobra.Command {
	var timeout time.Duration
	var force bool

	cmd := &cobra.Command{
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

			// If timeout is not specified, use default
			if timeout == 0 {
				timeout = process.DefaultShutdownTimeout
			}

			fmt.Printf("Stopping CCProxy service (PID: %d)...\n", runningPID)

			if force {
				fmt.Println("Force stopping with immediate SIGKILL")
				// For force stop, use a very short timeout
				timeout = 1 * time.Second
			} else {
				fmt.Printf("Attempting graceful shutdown (timeout: %v)\n", timeout)
			}

			// Stop the process with timeout
			if err := pidManager.StopProcessWithTimeout(timeout); err != nil {
				// Try to clean up PID file anyway
				pidManager.Cleanup()
				return fmt.Errorf("failed to stop service: %w", err)
			}

			fmt.Println("✅ Service stopped successfully")
			return nil
		},
	}

	cmd.Flags().DurationVarP(&timeout, "timeout", "t", 0, "Shutdown timeout (e.g., 30s, 1m)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force immediate shutdown with SIGKILL")

	return cmd
}
