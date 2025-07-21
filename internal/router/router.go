package router

import (
	"fmt"
	"strings"

	"github.com/orchestre-dev/ccproxy/internal/config"
	"github.com/orchestre-dev/ccproxy/internal/utils"
)

// Request represents the incoming request with model and parameters
type Request struct {
	Model    string `json:"model"`
	Thinking bool   `json:"thinking,omitempty"`
}

// RouteDecision represents the result of routing logic
type RouteDecision struct {
	Provider string
	Model    string
	Reason   string
}

// Router handles intelligent model routing based on various criteria
type Router struct {
	config *config.Config
}

// New creates a new Router instance
func New(cfg *config.Config) *Router {
	return &Router{
		config: cfg,
	}
}

// Route determines which model to use based on request parameters and token count
func (r *Router) Route(req Request, tokenCount int) RouteDecision {
	logger := utils.GetLogger()

	// 1. Check for explicit provider,model format
	if strings.Contains(req.Model, ",") {
		parts := strings.SplitN(req.Model, ",", 2)
		if len(parts) == 2 {
			logger.Debugf("Using explicit model selection: %s", req.Model)
			return RouteDecision{
				Provider: parts[0],
				Model:    parts[1],
				Reason:   "explicit model selection",
			}
		}
	}

	// 2. Check if there's a direct route for this model
	if route, exists := r.config.Routes[req.Model]; exists && route.Provider != "" {
		logger.Debugf("Using direct route for model: %s", req.Model)
		return RouteDecision{
			Provider: route.Provider,
			Model:    route.Model,
			Reason:   "direct model route",
		}
	}

	// 3. Check for long context routing based on token count
	if longContext, exists := r.config.Routes["longContext"]; exists && tokenCount > 60000 && longContext.Provider != "" {
		logger.Infof("Using long context model due to token count: %d", tokenCount)
		return RouteDecision{
			Provider: longContext.Provider,
			Model:    longContext.Model,
			Reason:   fmt.Sprintf("token count (%d) exceeds threshold", tokenCount),
		}
	}

	// 4. Check for background routing for haiku models
	if background, exists := r.config.Routes["background"]; exists && strings.HasPrefix(req.Model, "claude-3-5-haiku") && background.Provider != "" {
		logger.Info("Using background model for claude-3-5-haiku")
		return RouteDecision{
			Provider: background.Provider,
			Model:    background.Model,
			Reason:   "haiku model routed to background",
		}
	}

	// 5. Check for thinking routing based on parameter
	if think, exists := r.config.Routes["think"]; exists && req.Thinking && think.Provider != "" {
		logger.Info("Using think model due to thinking parameter")
		return RouteDecision{
			Provider: think.Provider,
			Model:    think.Model,
			Reason:   "thinking parameter enabled",
		}
	}

	// 6. Fall back to default model
	defaultRoute := r.config.Routes["default"]
	logger.Debug("Using default model")
	return RouteDecision{
		Provider: defaultRoute.Provider,
		Model:    defaultRoute.Model,
		Reason:   "default model",
	}
}

// ParseModelString parses a model string which can be either "model" or "provider,model"
func ParseModelString(modelStr string) (provider, model string) {
	if strings.Contains(modelStr, ",") {
		parts := strings.SplitN(modelStr, ",", 2)
		if len(parts) == 2 {
			return parts[0], parts[1]
		}
	}
	// If no comma, assume it's just a model name
	return "", modelStr
}

// FormatModelString formats provider and model into a comma-separated string
func FormatModelString(provider, model string) string {
	if provider == "" {
		return model
	}
	return fmt.Sprintf("%s,%s", provider, model)
}

// GetProviderForModel returns the provider for a given model based on configuration
func (r *Router) GetProviderForModel(model string) (string, error) {
	// Check if model already has provider specified
	if strings.Contains(model, ",") {
		provider, _ := ParseModelString(model)
		return provider, nil
	}

	// Look for model in provider configurations
	for _, provider := range r.config.Providers {
		for _, m := range provider.Models {
			if m == model {
				return provider.Name, nil
			}
		}
	}

	// If not found, use default provider
	if defaultRoute, exists := r.config.Routes["default"]; exists && defaultRoute.Provider != "" {
		return defaultRoute.Provider, nil
	}

	return "", fmt.Errorf("no provider found for model: %s", model)
}
