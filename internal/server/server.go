package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/musistudio/ccproxy/internal/config"
	modelrouter "github.com/musistudio/ccproxy/internal/router"
	"github.com/musistudio/ccproxy/internal/utils"
)

// Server represents the CCProxy HTTP server
type Server struct {
	config     *config.Config
	configPath string
	router     *gin.Engine
	server     *http.Server
}

// New creates a new server instance
func New(cfg *config.Config) (*Server, error) {
	return NewWithPath(cfg, "")
}

// NewWithPath creates a new server instance with a specific config path
func NewWithPath(cfg *config.Config, configPath string) (*Server, error) {
	// Set Gin mode based on environment
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}
	
	// Apply security constraint: force localhost when no API key
	if cfg.APIKey == "" && cfg.Host != "" && cfg.Host != "127.0.0.1" && cfg.Host != "localhost" {
		utils.GetLogger().Warn("Forcing host to 127.0.0.1 due to missing API key")
		cfg.Host = "127.0.0.1"
	}
	
	// Create router
	router := gin.New()
	
	// Add middleware
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())
	if cfg.Log {
		router.Use(loggingMiddleware())
	}
	
	// Add authentication middleware
	router.Use(authMiddleware(cfg.APIKey, true))
	
	// Add router middleware for intelligent model routing
	router.Use(modelrouter.RouterMiddleware(cfg))
	
	// Create server
	s := &Server{
		config:     cfg,
		configPath: configPath,
		router:     router,
		server: &http.Server{
			Addr:    fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Handler: router,
		},
	}
	
	// Setup routes
	s.setupRoutes()
	
	return s, nil
}

// Run starts the server and blocks until shutdown
func (s *Server) Run() error {
	// Setup graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	
	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		utils.GetLogger().Infof("Starting server on %s", s.server.Addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()
	
	// Wait for interrupt or error
	select {
	case err := <-errChan:
		return fmt.Errorf("server error: %w", err)
	case <-stop:
		utils.LogShutdown("interrupt signal received")
	}
	
	// Graceful shutdown
	return s.Shutdown()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}
	
	utils.GetLogger().Info("Server stopped gracefully")
	return nil
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// Health check endpoints
	s.router.GET("/", s.handleRoot)
	s.router.GET("/health", s.handleHealth)
	
	// Main API endpoint
	s.router.POST("/v1/messages", s.handleMessages)
	
	// Provider management endpoints
	providers := s.router.Group("/providers")
	{
		providers.GET("", s.handleListProviders)
		providers.POST("", s.handleCreateProvider)
		providers.GET("/:name", s.handleGetProvider)
		providers.PUT("/:name", s.handleUpdateProvider)
		providers.DELETE("/:name", s.handleDeleteProvider)
		providers.PATCH("/:name/toggle", s.handleToggleProvider)
	}
}

// Placeholder handlers - to be implemented
func (s *Server) handleRoot(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "LLMs API",
		"version": "1.0.0",
	})
}

func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}



// loggingMiddleware creates a logging middleware
func loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		
		// Process request
		c.Next()
		
		// Log request
		latency := time.Since(start)
		
		if raw != "" {
			path = path + "?" + raw
		}
		
		utils.LogRequest(c.Request.Method, path, map[string]interface{}{
			"client_ip": c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		})
		
		utils.LogResponse(c.Writer.Status(), latency.Seconds(), map[string]interface{}{
			"size": c.Writer.Size(),
		})
	}
}

// corsMiddleware adds CORS headers
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set CORS headers
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, x-api-key")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		
		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		c.Next()
	}
}