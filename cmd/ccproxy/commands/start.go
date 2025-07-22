package commands

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/orchestre-dev/ccproxy/internal/claudeconfig"
	"github.com/orchestre-dev/ccproxy/internal/config"
	"github.com/orchestre-dev/ccproxy/internal/process"
	"github.com/orchestre-dev/ccproxy/internal/server"
	"github.com/orchestre-dev/ccproxy/internal/utils"
	"github.com/spf13/cobra"
)

// StartCmd returns the start command
func StartCmd() *cobra.Command {
	var configPath string
	var foreground bool

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the CCProxy service",
		Long:  "Start the CCProxy service in the background (default) or foreground",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate environment variables
			if err := utils.ValidateEnvironmentVariables(); err != nil {
				return fmt.Errorf("environment variable validation failed: %w", err)
			}

			// Initialize configuration
			configService := config.NewService()
			var cfg *config.Config

			if configPath != "" {
				// Load from specified config file
				loadedCfg, err := config.LoadFromFile(configPath)
				if err != nil {
					return fmt.Errorf("failed to load config from %s: %w", configPath, err)
				}
				cfg = loadedCfg
				configService.SetConfig(cfg)
			} else {
				// Load from default locations
				if err := configService.Load(); err != nil {
					return fmt.Errorf("failed to load configuration: %w", err)
				}
				cfg = configService.Get()
			}

			// Initialize logger
			if err := utils.InitLogger(&utils.LogConfig{
				Enabled:  cfg.Log,
				FilePath: cfg.LogFile,
				Level:    "info",
				Format:   "json",
			}); err != nil {
				return fmt.Errorf("failed to initialize logger: %w", err)
			}

			// Initialize Claude configuration
			claudeManager, err := claudeconfig.NewManager()
			if err != nil {
				utils.GetLogger().Warnf("Failed to create Claude config manager: %v", err)
			} else {
				if err := claudeManager.Initialize(); err != nil {
					utils.GetLogger().Warnf("Failed to initialize Claude config: %v", err)
				} else {
					utils.GetLogger().Info("Claude configuration initialized successfully")
				}
			}

			// Create PID manager
			pidManager, err := process.NewPIDManager()
			if err != nil {
				return fmt.Errorf("failed to create PID manager: %w", err)
			}

			// Check if already running
			runningPID, err := pidManager.GetRunningPID()
			if err != nil {
				return fmt.Errorf("failed to check running status: %w", err)
			}

			if runningPID > 0 {
				fmt.Println("✅ Service is already running in the background")
				fmt.Printf("   PID: %d\n", runningPID)
				fmt.Printf("   Port: %d\n", cfg.Port)
				return nil
			}

			if foreground {
				// Run in foreground
				return runInForeground(cfg, pidManager, configPath)
			}

			// Start in background
			return startInBackground(cfg)
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to config file")
	cmd.Flags().BoolVarP(&foreground, "foreground", "f", false, "Run in foreground")

	return cmd
}

// runInForeground runs the server in the foreground
func runInForeground(cfg *config.Config, pidManager *process.PIDManager, configPath string) error {
	// Acquire lock
	if err := pidManager.AcquireLock(); err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Create cleanup function
	cleanup := func() {
		// Stop signal notification
		signal.Stop(sigChan)
		// Release lock
		// Safe to ignore error during shutdown
		_ = pidManager.ReleaseLock()
	}

	// Ensure cleanup happens
	defer cleanup()

	// Log startup
	utils.LogStartup(cfg.Port, version)

	// Create and start server
	srv, err := server.NewWithPath(cfg, configPath)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	// Run server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := srv.Run(); err != nil {
			errChan <- err
		}
	}()

	// Wait for signal or error
	select {
	case sig := <-sigChan:
		utils.GetLogger().Infof("Received signal: %v, shutting down gracefully...", sig)
		// Shutdown server
		if err := srv.Shutdown(); err != nil {
			utils.GetLogger().Errorf("Error during shutdown: %v", err)
		}
		return nil
	case err := <-errChan:
		return fmt.Errorf("server error: %w", err)
	}
}

// startInBackground starts the server in the background
func startInBackground(cfg *config.Config) error {
	// Check if we're already running in foreground mode to prevent infinite spawning
	if os.Getenv("CCPROXY_FOREGROUND") == "1" {
		return fmt.Errorf("cannot start background process from foreground mode")
	}

	// Check if we're in test mode - tests should use --foreground
	if os.Getenv("CCPROXY_TEST_MODE") == "1" {
		return fmt.Errorf("background mode disabled in tests - use --foreground flag")
	}

	// Additional safety check - if CCPROXY_SPAWN_DEPTH is set, we're in a spawn chain
	spawnDepth := 0
	if depthStr := os.Getenv("CCPROXY_SPAWN_DEPTH"); depthStr != "" {
		depth, err := strconv.Atoi(depthStr)
		if err != nil {
			utils.GetLogger().Warnf("Invalid CCPROXY_SPAWN_DEPTH value '%s': %v, using default 0", depthStr, err)
		} else if depth < 0 {
			utils.GetLogger().Warnf("Negative CCPROXY_SPAWN_DEPTH value %d, using 0", depth)
			spawnDepth = 0
		} else if depth > 10 {
			// Prevent overflow and unreasonable depth values
			return fmt.Errorf("CCPROXY_SPAWN_DEPTH value %d exceeds maximum allowed depth of 10", depth)
		} else {
			spawnDepth = depth
		}
	}
	if spawnDepth > 0 {
		return fmt.Errorf("detected spawn chain (depth: %d), preventing infinite spawn", spawnDepth)
	}

	// Create startup lock to prevent concurrent starts
	startupLock, err := process.NewStartupLock()
	if err != nil {
		return fmt.Errorf("failed to create startup lock: %w", err)
	}

	// Try to acquire exclusive startup lock
	locked, err := startupLock.TryLock()
	if err != nil {
		return fmt.Errorf("failed to check startup lock: %w", err)
	}
	if !locked {
		return fmt.Errorf("another ccproxy startup is already in progress")
	}
	defer func() {
		// Safe to ignore error on defer cleanup
		_ = startupLock.Unlock()
	}()

	// Check if service is already running (while holding startup lock)
	pidManager, err := process.NewPIDManager()
	if err != nil {
		return fmt.Errorf("failed to create PID manager: %w", err)
	}

	if runningPID, _ := pidManager.GetRunningPID(); runningPID > 0 {
		return fmt.Errorf("service is already running with PID %d", runningPID)
	}

	// Get executable path
	execPath, err := utils.GetExecutablePath()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Prepare the background process command
	cmd := exec.Command(execPath, "start", "--foreground")
	cmd.Env = append(os.Environ(),
		"CCPROXY_FOREGROUND=1",
		fmt.Sprintf("CCPROXY_SPAWN_DEPTH=%d", spawnDepth+1),
	)
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	// Set platform-specific attributes
	setPlatformSpecificAttrs(cmd)

	// Start the process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start background process: %w", err)
	}

	// Store the PID immediately
	backgroundPID := cmd.Process.Pid

	// Write PID file with proper locking
	// This will fail if another process is already running
	if err := pidManager.WritePIDForProcess(backgroundPID); err != nil {
		// Kill the process we just started if we can't write PID
		// Safe to ignore error during cleanup
		_ = cmd.Process.Kill()
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	// Wait for service to be ready
	fmt.Print("Starting CCProxy service")

	// Poll for up to 10 seconds
	for i := 0; i < 100; i++ {
		time.Sleep(100 * time.Millisecond)
		fmt.Print(".")

		// Check if process is still running
		if !pidManager.IsProcessRunning(backgroundPID) {
			// Process exited prematurely
			fmt.Println(" ❌")
			// Safe to ignore error during cleanup
			_ = pidManager.Cleanup()
			return fmt.Errorf("background process exited prematurely")
		}

		// Check if running with proper PID
		runningPID, err := pidManager.GetRunningPID()
		if err != nil {
			continue
		}

		if runningPID == backgroundPID {
			// Service is running with correct PID
			fmt.Println(" ✅")
			fmt.Println("Service started successfully!")
			fmt.Printf("PID: %d\n", runningPID)
			fmt.Printf("Port: %d\n", cfg.Port)
			fmt.Printf("Endpoint: http://%s:%d\n", cfg.Host, cfg.Port)
			return nil
		}
	}

	fmt.Println(" ❌")
	// Clean up if startup failed
	// Safe to ignore errors during cleanup
	_ = cmd.Process.Kill()
	_ = pidManager.Cleanup()
	return fmt.Errorf("service failed to start within timeout")
}
