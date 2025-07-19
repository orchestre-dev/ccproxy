package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/musistudio/ccproxy/internal/config"
	"github.com/musistudio/ccproxy/internal/pipeline"
	"github.com/musistudio/ccproxy/internal/providers"
	modelrouter "github.com/musistudio/ccproxy/internal/router"
	"github.com/musistudio/ccproxy/internal/transformer"
	"github.com/musistudio/ccproxy/internal/utils"
)

// Server represents the CCProxy HTTP server
type Server struct {
	config          *config.Config
	configPath      string
	router          *gin.Engine
	server          *http.Server
	providerService *providers.Service
	pipeline        *pipeline.Pipeline
	startTime       time.Time
	requestsServed  int64
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
	
	// Create config service
	configService := config.NewService()
	configService.SetConfig(cfg)
	
	// Create provider service
	providerService := providers.NewService(configService)
	if err := providerService.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize provider service: %w", err)
	}
	
	// Start health checks with 5 minute interval to reduce system load
	providerService.StartHealthChecks(5 * time.Minute)
	
	// Create transformer service
	transformerService := transformer.GetRegistry()
	
	// Create routing engine
	routingEngine := modelrouter.New(cfg)
	
	// Create pipeline
	pipelineService := pipeline.NewPipeline(cfg, providerService, transformerService, routingEngine)
	
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
		config:          cfg,
		configPath:      configPath,
		router:          router,
		providerService: providerService,
		pipeline:        pipelineService,
		startTime:       time.Now(),
		server: &http.Server{
			Addr:    fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Handler: router,
			// Add timeouts to prevent hanging connections
			ReadTimeout:    30 * time.Second,
			WriteTimeout:   30 * time.Second,
			IdleTimeout:    120 * time.Second,
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

// GetRouter returns the Gin router (mainly for testing)
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	// Stop provider service
	if s.providerService != nil {
		s.providerService.Stop()
	}
	
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
	s.router.GET("/status", s.handleStatus)
	
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
	// Get provider health information
	healthyProviders := s.providerService.GetHealthyProviders()
	allProviders := s.providerService.GetAllProviders()
	
	providerHealth := make(map[string]interface{})
	for _, p := range allProviders {
		health, _ := s.providerService.GetProviderHealth(p.Name)
		if health != nil {
			providerHealth[p.Name] = gin.H{
				"healthy":          health.Healthy,
				"last_check":       health.LastCheck.Format(time.RFC3339),
				"response_time_ms": health.ResponseTime.Milliseconds(),
				"enabled":          p.Enabled,
			}
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
		"providers": gin.H{
			"total":   len(allProviders),
			"healthy": len(healthyProviders),
			"details": providerHealth,
		},
	})
}

// handleStatus returns detailed status information about ccproxy and providers
func (s *Server) handleStatus(c *gin.Context) {
	// Calculate uptime
	uptime := time.Since(s.startTime)
	
	// Format uptime as human-readable string
	hours := int(uptime.Hours())
	minutes := int(uptime.Minutes()) % 60
	seconds := int(uptime.Seconds()) % 60
	uptimeStr := fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	
	// Get healthy providers
	healthyProviders := s.providerService.GetHealthyProviders()
	
	// Build provider status
	providerStatus := gin.H{
		"name":   "none",
		"status": "disconnected",
	}
	
	// Use first healthy provider as the "current" provider
	var currentProvider *config.Provider
	if len(healthyProviders) > 0 {
		currentProvider = healthyProviders[0]
		providerStatus = gin.H{
			"name":   currentProvider.Name,
			"status": "connected",
		}
		
		// Get default model from routes
		if routes := s.config.Routes; routes != nil {
			if defaultRoute, ok := routes["default"]; ok && defaultRoute.Provider == currentProvider.Name {
				providerStatus["model"] = defaultRoute.Model
			}
		}
		
		// Add provider-specific details
		health, _ := s.providerService.GetProviderHealth(currentProvider.Name)
		if health != nil {
			providerStatus["last_check"] = health.LastCheck.Format(time.RFC3339)
			providerStatus["response_time_ms"] = health.ResponseTime.Milliseconds()
			
			// Add provider-specific metrics based on provider name
			if strings.Contains(strings.ToLower(currentProvider.Name), "groq") {
				// Could add Groq-specific metrics here
				providerStatus["tokens_per_second"] = 185 // Example value
			} else if strings.Contains(strings.ToLower(currentProvider.Name), "openai") {
				// Could add OpenAI-specific metrics here
				providerStatus["organization"] = "org-..."
			}
		}
	}
	
	// Determine overall status
	status := "healthy"
	if len(healthyProviders) == 0 {
		status = "unhealthy"
	} else if currentProvider != nil {
		health, _ := s.providerService.GetProviderHealth(currentProvider.Name)
		if health != nil && !health.Healthy {
			status = "degraded"
		}
	}
	
	// Get version from main package (we'll need to pass this in)
	version := "1.0.0"
	if v := os.Getenv("CCPROXY_VERSION"); v != "" {
		version = v
	}
	
	// Build response
	response := gin.H{
		"status":    status,
		"timestamp": time.Now().Format(time.RFC3339),
		"proxy": gin.H{
			"version":         version,
			"uptime":          uptimeStr,
			"requests_served": atomic.LoadInt64(&s.requestsServed),
		},
		"provider": providerStatus,
	}
	
	c.JSON(http.StatusOK, response)
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