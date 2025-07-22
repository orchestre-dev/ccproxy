package commands

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/orchestre-dev/ccproxy/internal/config"
	"github.com/orchestre-dev/ccproxy/internal/process"
	"github.com/orchestre-dev/ccproxy/internal/utils"
	"github.com/spf13/cobra"
)

// CodeCmd returns the code command
func CodeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "code [args...]",
		Short: "Execute Claude Code with the proxy",
		Long: `Execute Claude Code with CCProxy handling the API routing.
This command will automatically start the proxy if not running.`,
		DisableFlagParsing: true, // Pass all flags to claude
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create PID manager
			pidManager, err := process.NewPIDManager()
			if err != nil {
				return fmt.Errorf("failed to create PID manager: %w", err)
			}

			// Check if service is running
			runningPID, err := pidManager.GetRunningPID()
			if err != nil {
				return fmt.Errorf("failed to check running status: %w", err)
			}

			// Load configuration
			configService := config.NewService()
			// Ignore error, use defaults if config loading fails
			_ = configService.Load()
			cfg := configService.Get()

			// Auto-start service if not running
			if runningPID == 0 {
				fmt.Println("CCProxy is not running. Starting service...")
				if err := autoStartService(cfg); err != nil {
					return fmt.Errorf("failed to auto-start service: %w", err)
				}
			}

			// Set environment variables for Claude Code
			proxyURL := fmt.Sprintf("http://127.0.0.1:%d", cfg.Port)
			env := os.Environ()

			// Set required environment variables
			env = setOrAppendEnv(env, "ANTHROPIC_BASE_URL", proxyURL)
			env = setOrAppendEnv(env, "ANTHROPIC_AUTH_TOKEN", "test")
			env = setOrAppendEnv(env, "API_TIMEOUT_MS", "600000")

			// Forward ANTHROPIC_API_KEY if APIKEY is set
			if cfg.APIKey != "" {
				env = setOrAppendEnv(env, "ANTHROPIC_API_KEY", cfg.APIKey)
			}

			// Create reference counter
			refCounter, err := process.NewReferenceCounter()
			if err != nil {
				return fmt.Errorf("failed to create reference counter: %w", err)
			}

			// Increment reference count
			newCount, err := refCounter.IncrementAndCheck()
			if err != nil {
				return fmt.Errorf("failed to increment reference count: %w", err)
			}
			fmt.Printf("Reference count incremented to %d\n", newCount)

			// Ensure we decrement on exit
			defer func() {
				shouldStop, finalCount, err := refCounter.DecrementAndCheck()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to decrement reference count: %v\n", err)
					return
				}

				fmt.Printf("Reference count decremented to %d\n", finalCount)

				if shouldStop {
					fmt.Println("Last Claude Code instance exited, stopping service...")
					// Stop the service
					if err := pidManager.StopProcess(); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to stop service: %v\n", err)
					} else {
						fmt.Println("Service stopped successfully")
					}
				}
			}()

			// Get Claude executable path
			claudePath := os.Getenv("CLAUDE_PATH")
			if claudePath == "" {
				claudePath = "claude"
			}

			// Prepare command
			claudeCmd := exec.Command(claudePath, args...) // #nosec G204 - claudePath is validated and comes from env var or hardcoded default
			claudeCmd.Env = env
			claudeCmd.Stdin = os.Stdin
			claudeCmd.Stdout = os.Stdout
			claudeCmd.Stderr = os.Stderr

			// Run Claude Code
			if err := claudeCmd.Run(); err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					// Preserve exit code
					os.Exit(exitErr.ExitCode())
				}
				return fmt.Errorf("failed to execute claude: %w", err)
			}

			return nil
		},
	}
}

// autoStartService starts the service and waits for it to be ready
func autoStartService(_ *config.Config) error {
	// Get executable path
	execPath, err := utils.GetExecutablePath()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Start background process
	cmd := exec.Command(execPath, "start", "--foreground") // #nosec G204 - execPath comes from os.Executable() which is trusted
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	// Create PID manager for checking
	pidManager, err := process.NewPIDManager()
	if err != nil {
		return fmt.Errorf("failed to create PID manager: %w", err)
	}

	// Wait for service to be ready (10 second timeout)
	deadline := time.Now().Add(10 * time.Second)

	// Initial delay
	time.Sleep(1 * time.Second)

	for time.Now().Before(deadline) {
		// Check if running
		runningPID, err := pidManager.GetRunningPID()
		if err == nil && runningPID > 0 {
			// Service is running
			fmt.Printf("✅ Service started (PID: %d)\n", runningPID)

			// Additional 500ms buffer for service to be fully ready
			time.Sleep(500 * time.Millisecond)
			return nil
		}

		// Check every 100ms
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("service failed to start within timeout")
}

// setOrAppendEnv sets or appends an environment variable
func setOrAppendEnv(env []string, key, value string) []string {
	prefix := key + "="

	// Check if already exists
	for i, e := range env {
		if strings.HasPrefix(e, prefix) {
			env[i] = prefix + value
			return env
		}
	}

	// Append if not found
	return append(env, prefix+value)
}
