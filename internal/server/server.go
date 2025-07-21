package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orchestre-dev/ccproxy/internal/config"
	"github.com/orchestre-dev/ccproxy/internal/performance"
	"github.com/orchestre-dev/ccproxy/internal/pipeline"
	"github.com/orchestre-dev/ccproxy/internal/providers"
	modelrouter "github.com/orchestre-dev/ccproxy/internal/router"
	"github.com/orchestre-dev/ccproxy/internal/state"
	"github.com/orchestre-dev/ccproxy/internal/transformer"
	"github.com/orchestre-dev/ccproxy/internal/utils"
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
	stateManager    *state.Manager
	readiness       *state.ReadinessProbe
	performance     *performance.Monitor
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
	
	// Add request size limit middleware
	router.Use(requestSizeLimitMiddleware(cfg.Performance.MaxRequestBodySize))
	
	// Add authentication middleware
	router.Use(authMiddleware(cfg.APIKey, true))
	
	// Add router middleware for intelligent model routing
	router.Use(modelrouter.RouterMiddleware(cfg))
	
	// Create state manager
	stateManager := state.NewManager()
	
	// Create performance monitor with config
	perfConfig := &performance.PerformanceConfig{
		ResourceLimits: performance.ResourceLimits{
			MaxMemoryMB:       2048,
			MaxGoroutines:     10000,
			MaxCPUPercent:     80.0,
			RequestTimeout:    5 * time.Minute,
			MaxRequestBodyMB:  10,
			MaxResponseBodyMB: 100,
		},
		RateLimit: performance.RateLimitConfig{
			Enabled:         cfg.Performance.RateLimitEnabled,
			RequestsPerMin:  cfg.Performance.RateLimitRequestsPerMin,
			BurstSize:       100,
			PerProvider:     true,
			PerAPIKey:       false,
			CleanupInterval: 5 * time.Minute,
		},
		CircuitBreaker: performance.CircuitBreakerConfig{
			Enabled:             cfg.Performance.CircuitBreakerEnabled,
			ErrorThreshold:      0.5,
			ConsecutiveFailures: 5,
			OpenDuration:        30 * time.Second,
			HalfOpenMaxRequests: 3,
		},
		MetricsEnabled:  cfg.Performance.MetricsEnabled,
		MetricsInterval: 1 * time.Minute,
		ProfilerEnabled: false,
	}
	perfMonitor := performance.NewMonitor(perfConfig)
	
	// Create server
	s := &Server{
		config:          cfg,
		configPath:      configPath,
		router:          router,
		providerService: providerService,
		pipeline:        pipelineService,
		startTime:       time.Now(),
		stateManager:    stateManager,
		performance:     perfMonitor,
		server: &http.Server{
			Addr:    fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Handler: router,
			// Add timeouts to prevent hanging connections
			ReadTimeout:    30 * time.Second,
			WriteTimeout:   30 * time.Second,
			IdleTimeout:    120 * time.Second,
		},
	}
	
	// Create readiness probe
	s.readiness = state.NewReadinessProbe(stateManager, 10*time.Second, 5*time.Second)
	
	// Register readiness checks
	s.setupReadinessChecks()
	
	// Register state change handlers
	s.setupStateHandlers()
	
	// Setup routes
	s.setupRoutes()
	
	// Add performance monitoring middleware if enabled
	if cfg.Performance.MetricsEnabled {
		router.Use(s.performanceMiddleware())
		// Add resource limit enforcement middleware
		router.Use(performance.Middleware(s.performance))
	}
	
	return s, nil
}

// Run starts the server and blocks until shutdown
func (s *Server) Run() error {
	// Start readiness probe
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	s.readiness.Start(ctx)
	defer s.readiness.Stop()
	
	// Wait for readiness
	utils.GetLogger().Info("Waiting for server components to be ready...")
	if err := s.readiness.WaitForReady(ctx, 30*time.Second); err != nil {
		s.stateManager.SetError(err)
		return fmt.Errorf("failed to initialize server components: %w", err)
	}
	
	// Mark server as ready
	s.stateManager.SetReady()
	utils.GetLogger().Info("Server components ready")
	
	// Setup graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	
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
		s.stateManager.SetError(err)
		return fmt.Errorf("server error: %w", err)
	case <-stop:
		s.stateManager.SetStopping()
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
	// Update state
	s.stateManager.SetStopping()
	
	// Stop readiness probe
	if s.readiness != nil {
		s.readiness.Stop()
	}
	
	// Stop provider service
	if s.providerService != nil {
		s.providerService.Stop()
	}
	
	// Stop performance monitor
	if s.performance != nil {
		s.performance.Stop()
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}
	
	// Update state to stopped
	s.stateManager.SetComponentState("server", state.StateStopped, nil)
	
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
	// Check if server is healthy
	if !s.stateManager.IsHealthy() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "unhealthy",
			"timestamp": time.Now().Format(time.RFC3339),
		})
		return
	}
	
	// Check if request is authenticated for detailed information
	isAuthenticated := s.isHealthRequestAuthenticated(c)
	
	// Basic health status (always available)
	healthyProviders := s.providerService.GetHealthyProviders()
	
	response := gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"providers": gin.H{
			"healthy": len(healthyProviders),
			"total":   len(s.providerService.GetAllProviders()),
		},
	}
	
	// Add detailed information only if authenticated
	if isAuthenticated {
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
		
		response["state"] = string(s.stateManager.GetState())
		response["providers"].(gin.H)["details"] = providerHealth
		
		// Get component health
		components := s.stateManager.GetComponents()
		componentHealth := make(map[string]interface{})
		for name, comp := range components {
			componentHealth[name] = gin.H{
				"state":        string(comp.State),
				"last_changed": comp.LastChanged.Format(time.RFC3339),
			}
			if comp.Error != nil {
				componentHealth[name].(gin.H)["error"] = comp.Error.Error()
			}
		}
		response["components"] = componentHealth
	}
	
	c.JSON(http.StatusOK, response)
}

// isHealthRequestAuthenticated checks if the health request is authenticated
func (s *Server) isHealthRequestAuthenticated(c *gin.Context) bool {
	// If no API key is configured, allow detailed access from localhost only
	if s.config.APIKey == "" {
		return isLocalhost(c)
	}
	
	// Check Authorization header (Bearer token)
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		const bearerPrefix = "Bearer "
		if strings.HasPrefix(authHeader, bearerPrefix) {
			token := authHeader[len(bearerPrefix):]
			if token == s.config.APIKey {
				return true
			}
		}
	}
	
	// Check x-api-key header
	if c.GetHeader("x-api-key") == s.config.APIKey {
		return true
	}
	
	return false
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

// setupReadinessChecks registers readiness checks for server components
func (s *Server) setupReadinessChecks() {
	// Provider service check
	s.readiness.RegisterCheck("providers", func(ctx context.Context) error {
		healthyProviders := s.providerService.GetHealthyProviders()
		if len(healthyProviders) == 0 {
			return fmt.Errorf("no healthy providers available")
		}
		return nil
	})
	
	// Config service check
	s.readiness.RegisterCheck("config", func(ctx context.Context) error {
		if s.config == nil {
			return fmt.Errorf("configuration not loaded")
		}
		return nil
	})
	
	// Pipeline check
	s.readiness.RegisterCheck("pipeline", func(ctx context.Context) error {
		if s.pipeline == nil {
			return fmt.Errorf("pipeline not initialized")
		}
		return nil
	})
	
	// Server port check
	s.readiness.RegisterCheck("server", func(ctx context.Context) error {
		// Check if we can bind to the port
		addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			// Port might already be in use by the server itself
			if strings.Contains(err.Error(), "address already in use") {
				return nil // This is OK if we're already running
			}
			return fmt.Errorf("cannot bind to %s: %w", addr, err)
		}
		listener.Close()
		return nil
	})
}

// setupStateHandlers registers state change handlers
func (s *Server) setupStateHandlers() {
	s.stateManager.OnStateChange(func(old, new state.ServiceState, component string) {
		if component == "service" {
			utils.GetLogger().Infof("Service state changed: %s -> %s", old, new)
		} else {
			utils.GetLogger().Debugf("Component %s state changed: %s -> %s", component, old, new)
		}
		
		// Handle specific state transitions
		if new == state.StateError && component == "providers" {
			utils.GetLogger().Warn("All providers are unhealthy")
		}
	})
}

// performanceMiddleware creates a middleware for performance monitoring
func (s *Server) performanceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Process request
		c.Next()
		
		// Record metrics
		latency := time.Since(start)
		provider := c.GetString("provider")
		model := c.GetString("model")
		
		s.performance.RecordRequest(performance.RequestMetrics{
			Provider:   provider,
			Model:      model,
			StartTime:  start,
			EndTime:    time.Now(),
			Latency:    latency,
			Success:    c.Writer.Status() < 400,
			StatusCode: c.Writer.Status(),
		})
		
		// Increment requests served
		atomic.AddInt64(&s.requestsServed, 1)
	}
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

// GetPort returns the port the server is configured to run on
func (s *Server) GetPort() int {
	return s.config.Port
}

// requestSizeLimitMiddleware limits the size of request bodies
func requestSizeLimitMiddleware(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip size check if maxSize is 0 (disabled)
		if maxSize <= 0 {
			c.Next()
			return
		}
		
		// Check Content-Length header
		contentLength := c.Request.ContentLength
		if contentLength > maxSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "Request body too large",
				"limit": maxSize,
			})
			c.Abort()
			return
		}
		
		// Wrap the body with a limited reader to enforce the limit at read time
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		}
		
		c.Next()
	}
}

// corsMiddleware adds CORS headers
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set CORS headers
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		
		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		c.Next()
	}
}