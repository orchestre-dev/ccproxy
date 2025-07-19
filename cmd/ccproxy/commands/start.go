package commands

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
	"github.com/musistudio/ccproxy/internal/claudeconfig"
	"github.com/musistudio/ccproxy/internal/config"
	"github.com/musistudio/ccproxy/internal/process"
	"github.com/musistudio/ccproxy/internal/server"
	"github.com/musistudio/ccproxy/internal/utils"
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
			// Initialize configuration
			configService := config.NewService()
			if configPath != "" {
				// TODO: Support custom config path
			}
			
			if err := configService.Load(); err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}
			
			cfg := configService.Get()
			
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
	defer pidManager.ReleaseLock()
	
	// Log startup
	utils.LogStartup(cfg.Port, version)
	
	// Create and start server
	srv, err := server.NewWithPath(cfg, configPath)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}
	
	// Run server (blocks until shutdown)
	if err := srv.Run(); err != nil {
		return fmt.Errorf("server error: %w", err)
	}
	
	return nil
}

// startInBackground starts the server in the background
func startInBackground(cfg *config.Config) error {
	// Get executable path
	execPath, err := utils.GetExecutablePath()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	
	// Start background process
	cmd := exec.Command(execPath, "start", "--foreground")
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start background process: %w", err)
	}
	
	// Wait for service to be ready
	fmt.Print("Starting CCProxy service")
	
	// Create PID manager for checking
	pidManager, err := process.NewPIDManager()
	if err != nil {
		return fmt.Errorf("failed to create PID manager: %w", err)
	}
	
	// Poll for up to 10 seconds
	for i := 0; i < 100; i++ {
		time.Sleep(100 * time.Millisecond)
		fmt.Print(".")
		
		// Check if running
		runningPID, err := pidManager.GetRunningPID()
		if err != nil {
			continue
		}
		
		if runningPID > 0 {
			// Service is running
			fmt.Println(" ✅")
			fmt.Println("Service started successfully!")
			fmt.Printf("PID: %d\n", runningPID)
			fmt.Printf("Port: %d\n", cfg.Port)
			fmt.Printf("Endpoint: http://%s:%d\n", cfg.Host, cfg.Port)
			return nil
		}
	}
	
	fmt.Println(" ❌")
	return fmt.Errorf("service failed to start within timeout")
}