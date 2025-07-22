package providers

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/orchestre-dev/ccproxy/internal/config"
	"github.com/orchestre-dev/ccproxy/internal/proxy"
	"github.com/orchestre-dev/ccproxy/internal/utils"
)

// HealthStatus represents the health status of a provider
type HealthStatus struct {
	Healthy          bool          `json:"healthy"`
	LastCheck        time.Time     `json:"last_check"`
	ResponseTime     time.Duration `json:"response_time_ms"`
	ErrorMessage     string        `json:"error_message,omitempty"`
	ConsecutiveFails int           `json:"consecutive_fails"`
}

// ProviderStats represents usage statistics for a provider
type ProviderStats struct {
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`
	AverageLatency     time.Duration `json:"average_latency_ms"`
	LastUsed           time.Time     `json:"last_used"`
}

// SelectionCriteria defines criteria for selecting a provider
type SelectionCriteria struct {
	Model          string
	RequiresTool   bool
	MinHealthScore float64
}

// Service manages provider lifecycle, health, and selection
type Service struct {
	config       *config.Service
	providers    map[string]*config.Provider
	health       map[string]*HealthStatus
	stats        map[string]*ProviderStats
	mu           sync.RWMutex
	healthCtx    context.Context
	healthCancel context.CancelFunc
	httpClient   *http.Client
	wg           sync.WaitGroup
}

// NewService creates a new provider management service
func NewService(configService *config.Service) *Service {
	ctx, cancel := context.WithCancel(context.Background())

	// Create HTTP client with proxy support
	cfg := configService.Get()
	var proxyConfig *proxy.Config
	if cfg.ProxyURL != "" {
		proxyConfig, _ = proxy.NewConfig(cfg.ProxyURL)
	} else {
		proxyConfig = proxy.GetProxyFromEnvironment()
	}

	httpClient, err := proxy.CreateHTTPClient(proxyConfig, 10*time.Second)
	if err != nil {
		// Fallback to simple client
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	return &Service{
		config:       configService,
		providers:    make(map[string]*config.Provider),
		health:       make(map[string]*HealthStatus),
		stats:        make(map[string]*ProviderStats),
		healthCtx:    ctx,
		healthCancel: cancel,
		httpClient:   httpClient,
	}
}

// Initialize loads providers from configuration
func (s *Service) Initialize() error {
	cfg := s.config.Get()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Load providers into memory
	for i := range cfg.Providers {
		provider := &cfg.Providers[i]
		s.providers[provider.Name] = provider
		s.health[provider.Name] = &HealthStatus{
			Healthy:   true, // Assume healthy initially
			LastCheck: time.Now(),
		}
		s.stats[provider.Name] = &ProviderStats{}
	}

	return nil
}

// StartHealthChecks begins periodic health monitoring
func (s *Service) StartHealthChecks(interval time.Duration) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Initial health check
		s.checkAllProviders()

		for {
			select {
			case <-ticker.C:
				s.checkAllProviders()
			case <-s.healthCtx.Done():
				return
			}
		}
	}()
}

// Stop gracefully shuts down the service
func (s *Service) Stop() {
	if s.healthCancel != nil {
		s.healthCancel()
	}
	s.wg.Wait()
}

// GetProvider returns a provider by name
func (s *Service) GetProvider(name string) (*config.Provider, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	provider, exists := s.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider not found: %s", name)
	}

	return provider, nil
}

// GetAllProviders returns all providers
func (s *Service) GetAllProviders() []*config.Provider {
	s.mu.RLock()
	defer s.mu.RUnlock()

	providers := make([]*config.Provider, 0, len(s.providers))
	for _, p := range s.providers {
		providers = append(providers, p)
	}

	return providers
}

// GetHealthyProviders returns only healthy and enabled providers
func (s *Service) GetHealthyProviders() []*config.Provider {
	s.mu.RLock()
	defer s.mu.RUnlock()

	providers := make([]*config.Provider, 0)
	for name, provider := range s.providers {
		health := s.health[name]
		if provider.Enabled && health.Healthy {
			providers = append(providers, provider)
		}
	}

	return providers
}

// SelectProvider selects the best provider based on criteria
func (s *Service) SelectProvider(criteria SelectionCriteria) (*config.Provider, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var candidates []*config.Provider

	// Find providers that support the requested model
	for name, provider := range s.providers {
		if !provider.Enabled {
			continue
		}

		health := s.health[name]
		if !health.Healthy {
			continue
		}

		// Check if provider supports the model
		modelSupported := false
		for _, m := range provider.Models {
			if m == criteria.Model {
				modelSupported = true
				break
			}
		}

		if modelSupported {
			candidates = append(candidates, provider)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no healthy provider found for model: %s", criteria.Model)
	}

	// For now, return the first candidate
	// TODO: Implement more sophisticated selection logic (load balancing, latency-based)
	return candidates[0], nil
}

// GetProviderHealth returns health status for a provider
func (s *Service) GetProviderHealth(name string) (*HealthStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	health, exists := s.health[name]
	if !exists {
		return nil, fmt.Errorf("provider not found: %s", name)
	}

	return health, nil
}

// GetProviderStats returns usage statistics for a provider
func (s *Service) GetProviderStats(name string) (*ProviderStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats, exists := s.stats[name]
	if !exists {
		return nil, fmt.Errorf("provider not found: %s", name)
	}

	return stats, nil
}

// RecordRequest records a request attempt for metrics
func (s *Service) RecordRequest(provider string, success bool, latency time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stats, exists := s.stats[provider]
	if !exists {
		return
	}

	// Use atomic operations for counters to prevent races
	stats.TotalRequests++
	if success {
		stats.SuccessfulRequests++
	} else {
		stats.FailedRequests++
	}

	// Update average latency (simple moving average)
	if stats.AverageLatency == 0 {
		stats.AverageLatency = latency
	} else {
		stats.AverageLatency = (stats.AverageLatency + latency) / 2
	}

	stats.LastUsed = time.Now()
}

// checkAllProviders performs health checks on all providers
func (s *Service) checkAllProviders() {
	providers := s.GetAllProviders()

	var wg sync.WaitGroup
	for _, provider := range providers {
		if !provider.Enabled {
			continue
		}

		wg.Add(1)
		go func(p *config.Provider) {
			defer wg.Done()
			s.checkProviderHealth(p)
		}(provider)
	}

	wg.Wait()
}

// checkProviderHealth performs a health check on a single provider
func (s *Service) checkProviderHealth(provider *config.Provider) {
	logger := utils.GetLogger()

	// Skip health check if API key is empty or provider is disabled
	if provider.APIKey == "" || !provider.Enabled {
		s.mu.Lock()
		s.health[provider.Name] = &HealthStatus{
			Healthy:      false,
			LastCheck:    time.Now(),
			ErrorMessage: "provider disabled or missing API key",
		}
		s.mu.Unlock()
		return
	}

	start := time.Now()
	healthy := true
	var errorMsg string

	// Create a context with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Perform simple HTTP health check
	// In a real implementation, this would be provider-specific
	req, err := http.NewRequestWithContext(ctx, "GET", provider.APIBaseURL, nil)
	if err != nil {
		healthy = false
		errorMsg = fmt.Sprintf("failed to create request: %v", err)
	} else {
		// Add API key header
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", provider.APIKey))

		resp, err := s.httpClient.Do(req)
		if err != nil {
			healthy = false
			errorMsg = fmt.Sprintf("request failed: %v", err)
		} else {
			_ = resp.Body.Close() // Safe to ignore: just checking health status

			// Check response status
			if resp.StatusCode >= 500 {
				healthy = false
				errorMsg = fmt.Sprintf("server error: %d", resp.StatusCode)
			} else if resp.StatusCode == 401 || resp.StatusCode == 403 {
				healthy = false
				errorMsg = "authentication failed"
			}
		}
	}

	responseTime := time.Since(start)

	// Update health status
	s.mu.Lock()
	defer s.mu.Unlock()

	health := s.health[provider.Name]
	health.LastCheck = time.Now()
	health.ResponseTime = responseTime

	if healthy {
		health.Healthy = true
		health.ConsecutiveFails = 0
		health.ErrorMessage = ""
		logger.Debugf("Provider %s health check passed (response time: %v)", provider.Name, responseTime)
	} else {
		health.ConsecutiveFails++
		health.ErrorMessage = errorMsg

		// Mark as unhealthy after 3 consecutive failures
		if health.ConsecutiveFails >= 3 {
			health.Healthy = false
			logger.Warnf("Provider %s marked unhealthy after %d failures: %s",
				provider.Name, health.ConsecutiveFails, errorMsg)
		}
	}
}

// RefreshProvider reloads a provider from configuration
func (s *Service) RefreshProvider(name string) error {
	cfg := s.config.Get()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Find provider in config
	var found *config.Provider
	for i := range cfg.Providers {
		if cfg.Providers[i].Name == name {
			found = &cfg.Providers[i]
			break
		}
	}

	if found == nil {
		// Provider was removed from config
		delete(s.providers, name)
		delete(s.health, name)
		delete(s.stats, name)
		return nil
	}

	// Update provider
	s.providers[name] = found

	// Reset health status if provider was previously unhealthy
	if health := s.health[name]; health != nil && !health.Healthy {
		health.Healthy = true
		health.ConsecutiveFails = 0
		health.ErrorMessage = ""
	}

	return nil
}
