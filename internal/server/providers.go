package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/orchestre-dev/ccproxy/internal/config"
)

// Provider request/response structures
type CreateProviderRequest struct {
	Name       string   `json:"name" binding:"required"`
	APIBaseURL string   `json:"apiBaseUrl" binding:"required"`
	APIKey     string   `json:"apiKey" binding:"required"`
	Models     []string `json:"models"`
	Enabled    bool     `json:"enabled"`
}

type UpdateProviderRequest struct {
	Name       string   `json:"name"`
	APIBaseURL string   `json:"apiBaseUrl"`
	APIKey     string   `json:"apiKey"`
	Models     []string `json:"models"`
	Enabled    *bool    `json:"enabled"`
}

// handleListProviders returns all configured providers
func (s *Server) handleListProviders(c *gin.Context) {
	providers := s.config.Providers
	Success(c, providers)
}

// handleCreateProvider creates a new provider
func (s *Server) handleCreateProvider(c *gin.Context) {
	var req CreateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	// Validate provider name doesn't already exist
	for _, p := range s.config.Providers {
		if p.Name == req.Name {
			Conflict(c, fmt.Sprintf("Provider '%s' already exists", req.Name))
			return
		}
	}

	// Create new provider
	provider := config.Provider{
		Name:       req.Name,
		APIBaseURL: req.APIBaseURL,
		APIKey:     req.APIKey,
		Models:     req.Models,
		Enabled:    req.Enabled,
	}

	// Add to config
	s.config.Providers = append(s.config.Providers, provider)

	// Save config
	configService := config.NewService()
	if err := configService.SaveProvider(&provider); err != nil {
		InternalServerError(c, fmt.Sprintf("Failed to save provider: %v", err))
		return
	}

	Created(c, provider)
}

// handleGetProvider returns a specific provider by name
func (s *Server) handleGetProvider(c *gin.Context) {
	name := c.Param("name")

	// Find provider
	for _, provider := range s.config.Providers {
		if provider.Name == name {
			Success(c, provider)
			return
		}
	}

	NotFound(c, fmt.Sprintf("Provider '%s' not found", name))
}

// handleUpdateProvider updates an existing provider
func (s *Server) handleUpdateProvider(c *gin.Context) {
	name := c.Param("name")

	var req UpdateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	// Find provider index
	providerIndex := -1
	for i, p := range s.config.Providers {
		if p.Name == name {
			providerIndex = i
			break
		}
	}

	if providerIndex == -1 {
		NotFound(c, fmt.Sprintf("Provider '%s' not found", name))
		return
	}

	// Update provider fields
	provider := &s.config.Providers[providerIndex]

	if req.Name != "" && req.Name != provider.Name {
		// Check if new name already exists
		for i, p := range s.config.Providers {
			if i != providerIndex && p.Name == req.Name {
				Conflict(c, fmt.Sprintf("Provider '%s' already exists", req.Name))
				return
			}
		}
		provider.Name = req.Name
	}

	if req.APIBaseURL != "" {
		provider.APIBaseURL = req.APIBaseURL
	}

	if req.APIKey != "" {
		provider.APIKey = req.APIKey
	}

	if req.Models != nil {
		provider.Models = req.Models
	}

	if req.Enabled != nil {
		provider.Enabled = *req.Enabled
	}

	// Save config
	configService := config.NewService()
	configService.SetConfig(s.config)
	if err := configService.UpdateProvider(name, provider); err != nil {
		InternalServerError(c, fmt.Sprintf("Failed to update provider: %v", err))
		return
	}

	Success(c, provider)
}

// handleDeleteProvider deletes a provider
func (s *Server) handleDeleteProvider(c *gin.Context) {
	name := c.Param("name")

	// Check if provider exists in provider service
	if _, err := s.providerService.GetProvider(name); err != nil {
		NotFound(c, fmt.Sprintf("Provider '%s' not found", name))
		return
	}

	// Delete from config service
	configService := config.NewService()
	configService.SetConfig(s.config)
	if err := configService.DeleteProvider(name); err != nil {
		InternalServerError(c, fmt.Sprintf("Failed to delete provider: %v", err))
		return
	}

	// Reload config in provider service
	s.config = configService.Get()
	if err := s.providerService.Initialize(); err != nil {
		InternalServerError(c, fmt.Sprintf("Failed to reinitialize provider service: %v", err))
		return
	}

	Success(c, gin.H{
		"message": "Provider deleted successfully",
	})
}

// handleToggleProvider enables or disables a provider
func (s *Server) handleToggleProvider(c *gin.Context) {
	name := c.Param("name")

	// Find provider
	var provider *config.Provider
	for i := range s.config.Providers {
		if s.config.Providers[i].Name == name {
			provider = &s.config.Providers[i]
			break
		}
	}

	if provider == nil {
		NotFound(c, fmt.Sprintf("Provider '%s' not found", name))
		return
	}

	// Toggle enabled state
	provider.Enabled = !provider.Enabled

	// Save config
	configService := config.NewService()
	configService.SetConfig(s.config)
	if err := configService.UpdateProvider(name, provider); err != nil {
		InternalServerError(c, fmt.Sprintf("Failed to toggle provider: %v", err))
		return
	}

	message := "Provider disabled successfully"
	if provider.Enabled {
		message = "Provider enabled successfully"
	}

	Success(c, gin.H{
		"message": message,
	})
}
