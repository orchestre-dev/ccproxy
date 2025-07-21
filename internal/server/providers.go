package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/orchestre-dev/ccproxy/internal/config"
	"github.com/orchestre-dev/ccproxy/internal/utils"
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

	// Get current provider from provider service
	provider, err := s.providerService.GetProvider(name)
	if err != nil {
		NotFound(c, fmt.Sprintf("Provider '%s' not found", name))
		return
	}

	// Create a copy to avoid modifying the original
	updatedProvider := *provider

	if req.Name != "" && req.Name != updatedProvider.Name {
		// Check if new name already exists
		providers := s.providerService.GetAllProviders()
		for _, p := range providers {
			if p.Name == req.Name {
				Conflict(c, fmt.Sprintf("Provider '%s' already exists", req.Name))
				return
			}
		}
		updatedProvider.Name = req.Name
	}

	if req.APIBaseURL != "" {
		updatedProvider.APIBaseURL = req.APIBaseURL
	}

	if req.APIKey != "" {
		updatedProvider.APIKey = req.APIKey
	}

	if req.Models != nil {
		updatedProvider.Models = req.Models
	}

	if req.Enabled != nil {
		updatedProvider.Enabled = *req.Enabled
	}

	// Update via config service
	configService := config.NewService()
	cfg := configService.Get()
	configService.SetConfig(cfg)
	if err := configService.UpdateProvider(name, &updatedProvider); err != nil {
		InternalServerError(c, fmt.Sprintf("Failed to update provider: %v", err))
		return
	}

	// Refresh provider service
	if err := s.providerService.RefreshProvider(updatedProvider.Name); err != nil {
		utils.GetLogger().Warnf("Failed to refresh provider in service: %v", err)
	}

	Success(c, &updatedProvider)
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

	// Get current provider from provider service
	provider, err := s.providerService.GetProvider(name)
	if err != nil {
		NotFound(c, fmt.Sprintf("Provider '%s' not found", name))
		return
	}

	// Create a copy to avoid modifying the original
	updatedProvider := *provider
	updatedProvider.Enabled = !updatedProvider.Enabled

	// Update via config service
	configService := config.NewService()
	cfg := configService.Get()
	configService.SetConfig(cfg)
	if err := configService.UpdateProvider(name, &updatedProvider); err != nil {
		InternalServerError(c, fmt.Sprintf("Failed to toggle provider: %v", err))
		return
	}

	// Refresh provider service
	if err := s.providerService.RefreshProvider(name); err != nil {
		utils.GetLogger().Warnf("Failed to refresh provider in service: %v", err)
	}

	message := "Provider disabled successfully"
	if updatedProvider.Enabled {
		message = "Provider enabled successfully"
	}

	Success(c, gin.H{
		"message": message,
	})
}
