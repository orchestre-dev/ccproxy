package router

import (
	"strings"
	"testing"

	"github.com/orchestre-dev/ccproxy/internal/config"
)

func TestRouter_Route(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		Routes: map[string]config.Route{
			"default": {
				Provider: "anthropic",
				Model:    "claude-3-sonnet-20240229",
			},
			"longContext": {
				Provider: "openrouter",
				Model:    "google/gemini-2.5-pro-preview",
			},
			"background": {
				Provider: "ollama",
				Model:    "qwen2.5-coder:latest",
			},
			"think": {
				Provider: "deepseek",
				Model:    "deepseek-reasoner",
			},
		},
	}
	
	router := New(cfg)
	
	tests := []struct {
		name          string
		req           Request
		tokenCount    int
		wantProvider  string
		wantModel     string
		wantReasonKey string
	}{
		{
			name: "explicit model selection",
			req: Request{
				Model: "openai,gpt-4",
			},
			tokenCount:    1000,
			wantProvider:  "openai",
			wantModel:     "gpt-4",
			wantReasonKey: "explicit",
		},
		{
			name: "long context routing",
			req: Request{
				Model: "claude-3-opus-20240229",
			},
			tokenCount:    65000,
			wantProvider:  "openrouter",
			wantModel:     "google/gemini-2.5-pro-preview",
			wantReasonKey: "token count",
		},
		{
			name: "background routing for haiku",
			req: Request{
				Model: "claude-3-5-haiku-20241022",
			},
			tokenCount:    1000,
			wantProvider:  "ollama",
			wantModel:     "qwen2.5-coder:latest",
			wantReasonKey: "haiku",
		},
		{
			name: "thinking routing",
			req: Request{
				Model:    "claude-3-opus-20240229",
				Thinking: true,
			},
			tokenCount:    1000,
			wantProvider:  "deepseek",
			wantModel:     "deepseek-reasoner",
			wantReasonKey: "thinking",
		},
		{
			name: "default routing",
			req: Request{
				Model: "claude-3-opus-20240229",
			},
			tokenCount:    1000,
			wantProvider:  "anthropic",
			wantModel:     "claude-3-sonnet-20240229",
			wantReasonKey: "default",
		},
		{
			name: "explicit takes precedence over long context",
			req: Request{
				Model: "openai,gpt-4-turbo",
			},
			tokenCount:    100000,
			wantProvider:  "openai",
			wantModel:     "gpt-4-turbo",
			wantReasonKey: "explicit",
		},
		{
			name: "long context takes precedence over thinking",
			req: Request{
				Model:    "claude-3-opus-20240229",
				Thinking: true,
			},
			tokenCount:    70000,
			wantProvider:  "openrouter",
			wantModel:     "google/gemini-2.5-pro-preview",
			wantReasonKey: "token count",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision := router.Route(tt.req, tt.tokenCount)
			
			if decision.Provider != tt.wantProvider {
				t.Errorf("Provider = %v, want %v", decision.Provider, tt.wantProvider)
			}
			if decision.Model != tt.wantModel {
				t.Errorf("Model = %v, want %v", decision.Model, tt.wantModel)
			}
			if !contains(decision.Reason, tt.wantReasonKey) {
				t.Errorf("Reason = %v, want to contain %v", decision.Reason, tt.wantReasonKey)
			}
		})
	}
}

func TestRouter_Route_EmptyConfig(t *testing.T) {
	// Test with minimal config - only default route
	cfg := &config.Config{
		Routes: map[string]config.Route{
			"default": {
				Provider: "anthropic",
				Model:    "claude-3-sonnet-20240229",
			},
		},
	}
	
	router := New(cfg)
	
	// Should always fall back to default
	tests := []struct {
		name       string
		req        Request
		tokenCount int
	}{
		{
			name:       "high token count with no long context config",
			req:        Request{Model: "claude-3-opus-20240229"},
			tokenCount: 100000,
		},
		{
			name: "haiku model with no background config",
			req:  Request{Model: "claude-3-5-haiku-20241022"},
		},
		{
			name: "thinking with no think config",
			req:  Request{Model: "claude-3-opus-20240229", Thinking: true},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision := router.Route(tt.req, tt.tokenCount)
			
			if decision.Provider != "anthropic" {
				t.Errorf("Provider = %v, want anthropic", decision.Provider)
			}
			if decision.Model != "claude-3-sonnet-20240229" {
				t.Errorf("Model = %v, want claude-3-sonnet-20240229", decision.Model)
			}
			if !contains(decision.Reason, "default") {
				t.Errorf("Reason = %v, want to contain 'default'", decision.Reason)
			}
		})
	}
}

func TestParseModelString(t *testing.T) {
	tests := []struct {
		name         string
		modelStr     string
		wantProvider string
		wantModel    string
	}{
		{
			name:         "provider,model format",
			modelStr:     "openai,gpt-4",
			wantProvider: "openai",
			wantModel:    "gpt-4",
		},
		{
			name:         "model only",
			modelStr:     "claude-3-opus-20240229",
			wantProvider: "",
			wantModel:    "claude-3-opus-20240229",
		},
		{
			name:         "empty string",
			modelStr:     "",
			wantProvider: "",
			wantModel:    "",
		},
		{
			name:         "multiple commas",
			modelStr:     "provider,model,extra",
			wantProvider: "provider",
			wantModel:    "model,extra",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, model := ParseModelString(tt.modelStr)
			if provider != tt.wantProvider {
				t.Errorf("ParseModelString() provider = %v, want %v", provider, tt.wantProvider)
			}
			if model != tt.wantModel {
				t.Errorf("ParseModelString() model = %v, want %v", model, tt.wantModel)
			}
		})
	}
}

func TestFormatModelString(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		model    string
		want     string
	}{
		{
			name:     "with provider",
			provider: "openai",
			model:    "gpt-4",
			want:     "openai,gpt-4",
		},
		{
			name:     "without provider",
			provider: "",
			model:    "claude-3-opus-20240229",
			want:     "claude-3-opus-20240229",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatModelString(tt.provider, tt.model)
			if got != tt.want {
				t.Errorf("FormatModelString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRouter_GetProviderForModel(t *testing.T) {
	cfg := &config.Config{
		Routes: map[string]config.Route{
			"default": {
				Provider: "anthropic",
				Model:    "claude-3-sonnet-20240229",
			},
		},
		Providers: []config.Provider{
			{
				Name:   "openai",
				Models: []string{"gpt-4", "gpt-3.5-turbo"},
			},
			{
				Name:   "anthropic",
				Models: []string{"claude-3-opus-20240229", "claude-3-sonnet-20240229"},
			},
		},
	}
	
	router := New(cfg)
	
	tests := []struct {
		name         string
		model        string
		wantProvider string
		wantErr      bool
	}{
		{
			name:         "explicit provider in model string",
			model:        "openai,gpt-4",
			wantProvider: "openai",
			wantErr:      false,
		},
		{
			name:         "model found in providers",
			model:        "gpt-4",
			wantProvider: "openai",
			wantErr:      false,
		},
		{
			name:         "model not found uses default",
			model:        "unknown-model",
			wantProvider: "anthropic",
			wantErr:      false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := router.GetProviderForModel(tt.model)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetProviderForModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if provider != tt.wantProvider {
				t.Errorf("GetProviderForModel() = %v, want %v", provider, tt.wantProvider)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}