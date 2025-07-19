package commands

import (
	"fmt"
	"os"
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
			fmt.Println("[DEBUG] Initializing configuration...")
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
			fmt.Printf("[DEBUG] Initializing logger (enabled=%v, file=%s)...\n", cfg.Log, cfg.LogFile)
			if err := utils.InitLogger(&utils.LogConfig{
				Enabled:  cfg.Log,
				FilePath: cfg.LogFile,
				Level:    "info",
				Format:   "json",
			}); err != nil {
				return fmt.Errorf("failed to initialize logger: %w", err)
			}
			fmt.Println("[DEBUG] Logger initialized")
			
			// Initialize Claude configuration
			fmt.Println("[DEBUG] Initializing Claude configuration...")
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
			fmt.Println("[DEBUG] Creating PID manager...")
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
	fmt.Println("[DEBUG] Running in foreground mode...")
	// Acquire lock
	fmt.Println("[DEBUG] Acquiring PID lock...")
	if err := pidManager.AcquireLock(); err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer pidManager.ReleaseLock()
	
	// Log startup
	utils.LogStartup(cfg.Port, version)
	
	// Create and start server
	fmt.Printf("[DEBUG] Creating server on %s:%d...\n", cfg.Host, cfg.Port)
	srv, err := server.NewWithPath(cfg, configPath)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}
	fmt.Println("[DEBUG] Server created successfully")
	
	// Run server (blocks until shutdown)
	fmt.Println("[DEBUG] Starting server...")
	if err := srv.Run(); err != nil {
		return fmt.Errorf("server error: %w", err)
	}
	
	return nil
}

// startInBackground starts the server in the background
func startInBackground(cfg *config.Config) error {
	// Check if we're already running in foreground mode to prevent infinite spawning
	if os.Getenv("CCPROXY_FOREGROUND") == "1" {
		return fmt.Errorf("cannot start background process from foreground mode")
	}
	
	// Get executable path
	execPath, err := utils.GetExecutablePath()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	
	// Start background process
	cmd := exec.Command(execPath, "start", "--foreground")
	cmd.Env = append(os.Environ(), "CCPROXY_FOREGROUND=1")
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