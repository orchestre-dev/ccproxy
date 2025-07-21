package router

import (
	"strings"
	"testing"

	"github.com/orchestre-dev/ccproxy/internal/config"
)

func TestNew(t *testing.T) {
	cfg := &config.Config{
		Routes: map[string]config.Route{
			"default": {
				Provider: "openai",
				Model:    "gpt-4",
			},
		},
	}
	
	router := New(cfg)
	
	if router == nil {
		t.Error("Expected non-nil router")
	}
	
	if router.config != cfg {
		t.Error("Router config not set correctly")
	}
}

func TestRouter_Route(t *testing.T) {
	cfg := &config.Config{
		Routes: map[string]config.Route{
			"default": {
				Provider: "openai",
				Model:    "gpt-4",
			},
			"longContext": {
				Provider: "anthropic",
				Model:    "claude-3-opus",
			},
			"background": {
				Provider: "groq",
				Model:    "llama-3-8b",
			},
			"think": {
				Provider: "anthropic",
				Model:    "claude-3-opus",
			},
			"gpt-3.5-turbo": {
				Provider: "openai",
				Model:    "gpt-3.5-turbo",
			},
		},
	}
	
	router := New(cfg)
	
	t.Run("ExplicitModelSelection", func(t *testing.T) {
		req := Request{
			Model: "anthropic,claude-3-haiku",
		}
		
		decision := router.Route(req, 1000)
		
		if decision.Provider != "anthropic" {
			t.Errorf("Expected provider 'anthropic', got %s", decision.Provider)
		}
		
		if decision.Model != "claude-3-haiku" {
			t.Errorf("Expected model 'claude-3-haiku', got %s", decision.Model)
		}
		
		if decision.Reason != "explicit model selection" {
			t.Errorf("Expected reason 'explicit model selection', got %s", decision.Reason)
		}
	})
	
	t.Run("DirectModelRoute", func(t *testing.T) {
		req := Request{
			Model: "gpt-3.5-turbo",
		}
		
		decision := router.Route(req, 1000)
		
		if decision.Provider != "openai" {
			t.Errorf("Expected provider 'openai', got %s", decision.Provider)
		}
		
		if decision.Model != "gpt-3.5-turbo" {
			t.Errorf("Expected model 'gpt-3.5-turbo', got %s", decision.Model)
		}
		
		if decision.Reason != "direct model route" {
			t.Errorf("Expected reason 'direct model route', got %s", decision.Reason)
		}
	})
	
	t.Run("LongContextRouting", func(t *testing.T) {
		req := Request{
			Model: "gpt-4", // Would normally go to default
		}
		
		decision := router.Route(req, 80000) // Over 60k token threshold
		
		if decision.Provider != "anthropic" {
			t.Errorf("Expected provider 'anthropic', got %s", decision.Provider)
		}
		
		if decision.Model != "claude-3-opus" {
			t.Errorf("Expected model 'claude-3-opus', got %s", decision.Model)
		}
		
		if !strings.Contains(decision.Reason, "token count") {
			t.Errorf("Expected reason to mention token count, got %s", decision.Reason)
		}
	})
	
	t.Run("BackgroundRouting", func(t *testing.T) {
		req := Request{
			Model: "claude-3-5-haiku-20241022",
		}
		
		decision := router.Route(req, 1000)
		
		if decision.Provider != "groq" {
			t.Errorf("Expected provider 'groq', got %s", decision.Provider)
		}
		
		if decision.Model != "llama-3-8b" {
			t.Errorf("Expected model 'llama-3-8b', got %s", decision.Model)
		}
		
		if decision.Reason != "haiku model routed to background" {
			t.Errorf("Expected background routing reason, got %s", decision.Reason)
		}
	})
	
	t.Run("ThinkingRouting", func(t *testing.T) {
		req := Request{
			Model:    "gpt-4",
			Thinking: true,
		}
		
		decision := router.Route(req, 1000)
		
		if decision.Provider != "anthropic" {
			t.Errorf("Expected provider 'anthropic', got %s", decision.Provider)
		}
		
		if decision.Model != "claude-3-opus" {
			t.Errorf("Expected model 'claude-3-opus', got %s", decision.Model)
		}
		
		if decision.Reason != "thinking parameter enabled" {
			t.Errorf("Expected thinking routing reason, got %s", decision.Reason)
		}
	})
	
	t.Run("DefaultRouting", func(t *testing.T) {
		req := Request{
			Model: "unknown-model",
		}
		
		decision := router.Route(req, 1000)
		
		if decision.Provider != "openai" {
			t.Errorf("Expected provider 'openai', got %s", decision.Provider)
		}
		
		if decision.Model != "gpt-4" {
			t.Errorf("Expected model 'gpt-4', got %s", decision.Model)
		}
		
		if decision.Reason != "default model" {
			t.Errorf("Expected reason 'default model', got %s", decision.Reason)
		}
	})
	
	t.Run("RoutingPriority", func(t *testing.T) {
		// Test that explicit model selection takes priority over everything
		req := Request{
			Model:    "anthropic,claude-3-haiku", // Explicit
			Thinking: true,                      // Would trigger thinking route
		}
		
		decision := router.Route(req, 80000) // Would trigger long context
		
		// Should still use explicit model selection
		if decision.Provider != "anthropic" {
			t.Errorf("Expected provider 'anthropic', got %s", decision.Provider)
		}
		
		if decision.Model != "claude-3-haiku" {
			t.Errorf("Expected model 'claude-3-haiku', got %s", decision.Model)
		}
		
		if decision.Reason != "explicit model selection" {
			t.Errorf("Expected explicit selection to take priority, got %s", decision.Reason)
		}
	})
}

func TestParseModelString(t *testing.T) {
	tests := []struct {
		input            string
		expectedProvider string
		expectedModel    string
	}{
		{"gpt-4", "", "gpt-4"},
		{"openai,gpt-4", "openai", "gpt-4"},
		{"anthropic,claude-3-haiku", "anthropic", "claude-3-haiku"},
		{"provider,model,extra", "provider", "model,extra"}, // Only split on first comma
		{"", "", ""},
		{"just-a-model", "", "just-a-model"},
	}
	
	for _, test := range tests {
		provider, model := ParseModelString(test.input)
		
		if provider != test.expectedProvider {
			t.Errorf("Input %q: expected provider %q, got %q", test.input, test.expectedProvider, provider)
		}
		
		if model != test.expectedModel {
			t.Errorf("Input %q: expected model %q, got %q", test.input, test.expectedModel, model)
		}
	}
}

func TestFormatModelString(t *testing.T) {
	tests := []struct {
		provider string
		model    string
		expected string
	}{
		{"", "gpt-4", "gpt-4"},
		{"openai", "gpt-4", "openai,gpt-4"},
		{"anthropic", "claude-3-haiku", "anthropic,claude-3-haiku"},
		{"", "", ""},
		{"provider", "", "provider,"},
	}
	
	for _, test := range tests {
		result := FormatModelString(test.provider, test.model)
		
		if result != test.expected {
			t.Errorf("Provider %q, Model %q: expected %q, got %q", test.provider, test.model, test.expected, result)
		}
	}
}

func TestRouter_GetProviderForModel(t *testing.T) {
	cfg := &config.Config{
		Providers: []config.Provider{
			{
				Name:   "openai",
				Models: []string{"gpt-4", "gpt-3.5-turbo"},
			},
			{
				Name:   "anthropic",
				Models: []string{"claude-3-opus", "claude-3-haiku"},
			},
		},
		Routes: map[string]config.Route{
			"default": {
				Provider: "openai",
				Model:    "gpt-4",
			},
		},
	}
	
	router := New(cfg)
	
	t.Run("ModelWithProviderPrefix", func(t *testing.T) {
		provider, err := router.GetProviderForModel("anthropic,claude-3-opus")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if provider != "anthropic" {
			t.Errorf("Expected provider 'anthropic', got %s", provider)
		}
	})
	
	t.Run("ModelInProviderList", func(t *testing.T) {
		provider, err := router.GetProviderForModel("gpt-4")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if provider != "openai" {
			t.Errorf("Expected provider 'openai', got %s", provider)
		}
	})
	
	t.Run("ModelNotFound_UsesDefault", func(t *testing.T) {
		provider, err := router.GetProviderForModel("unknown-model")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if provider != "openai" {
			t.Errorf("Expected default provider 'openai', got %s", provider)
		}
	})
	
	t.Run("ModelNotFound_NoDefault", func(t *testing.T) {
		// Create config without default route
		cfgNoDefault := &config.Config{
			Providers: []config.Provider{
				{
					Name:   "openai",
					Models: []string{"gpt-4"},
				},
			},
			Routes: map[string]config.Route{},
		}
		
		routerNoDefault := New(cfgNoDefault)
		
		_, err := routerNoDefault.GetProviderForModel("unknown-model")
		if err == nil {
			t.Error("Expected error for unknown model with no default")
		}
		
		if !strings.Contains(err.Error(), "no provider found") {
			t.Errorf("Expected 'no provider found' error, got %v", err)
		}
	})
}

func TestRouter_EdgeCases(t *testing.T) {
	// Test router with minimal configuration
	cfg := &config.Config{
		Routes: map[string]config.Route{
			"default": {
				Provider: "fallback",
				Model:    "fallback-model",
			},
		},
	}
	
	router := New(cfg)
	
	t.Run("EmptyModelString", func(t *testing.T) {
		req := Request{
			Model: "",
		}
		
		decision := router.Route(req, 1000)
		
		// Should fall back to default
		if decision.Provider != "fallback" {
			t.Errorf("Expected fallback provider, got %s", decision.Provider)
		}
		
		if decision.Reason != "default model" {
			t.Errorf("Expected default routing, got %s", decision.Reason)
		}
	})
	
	t.Run("MalformedExplicitModel", func(t *testing.T) {
		req := Request{
			Model: "provider,", // Malformed - has comma but no model
		}
		
		decision := router.Route(req, 1000)
		
		// Should still use explicit parsing
		if decision.Provider != "provider" {
			t.Errorf("Expected provider 'provider', got %s", decision.Provider)
		}
		
		if decision.Model != "" {
			t.Errorf("Expected empty model, got %s", decision.Model)
		}
	})
	
	t.Run("ZeroTokenCount", func(t *testing.T) {
		req := Request{
			Model: "test-model",
		}
		
		decision := router.Route(req, 0)
		
		// Should not trigger long context routing
		if decision.Reason == "token count (0) exceeds threshold" {
			t.Error("Zero token count should not trigger long context routing")
		}
	})
	
	t.Run("NegativeTokenCount", func(t *testing.T) {
		req := Request{
			Model: "test-model",
		}
		
		decision := router.Route(req, -100)
		
		// Should handle negative token count gracefully
		if decision.Provider != "fallback" {
			t.Errorf("Expected fallback provider for negative token count, got %s", decision.Provider)
		}
	})
}

func TestRouter_ComplexRouting(t *testing.T) {
	// Test complex routing scenarios
	cfg := &config.Config{
		Routes: map[string]config.Route{
			"default": {
				Provider: "openai",
				Model:    "gpt-4",
			},
			"longContext": {
				Provider: "anthropic",
				Model:    "claude-3-opus",
			},
			"background": {
				Provider: "groq",
				Model:    "llama-3-8b",
			},
			"think": {
				Provider: "anthropic",
				Model:    "claude-3-sonnet",
			},
		},
	}
	
	router := New(cfg)
	
	t.Run("ThinkingWithLongContext", func(t *testing.T) {
		// Test priority: long context takes precedence over thinking (per router logic)
		req := Request{
			Model:    "gpt-4",
			Thinking: true,
		}
		
		decision := router.Route(req, 80000) // Long context threshold
		
		// Long context should win (step 3 comes before step 5)
		if !strings.Contains(decision.Reason, "token count") {
			t.Errorf("Expected long context to take precedence, got %s", decision.Reason)
		}
		
		if decision.Model != "claude-3-opus" {
			t.Errorf("Expected long context model, got %s", decision.Model)
		}
	})
	
	t.Run("BackgroundWithLongContext", func(t *testing.T) {
		// Test priority: long context takes precedence over background (step 3 vs step 4)
		req := Request{
			Model: "claude-3-5-haiku-20241022",
		}
		
		decision := router.Route(req, 80000) // Long context threshold
		
		// Long context should win (step 3 comes before step 4)
		if !strings.Contains(decision.Reason, "token count") {
			t.Errorf("Expected long context routing to take precedence, got %s", decision.Reason)
		}
		
		if decision.Model != "claude-3-opus" {
			t.Errorf("Expected long context model, got %s", decision.Model)
		}
	})
	
	t.Run("NonHaikuWithLongContext", func(t *testing.T) {
		// Test that non-haiku models with long context use long context routing
		req := Request{
			Model: "claude-3-sonnet", // Not haiku
		}
		
		decision := router.Route(req, 80000) // Long context threshold
		
		// Should use long context routing
		if !strings.Contains(decision.Reason, "token count") {
			t.Errorf("Expected long context routing, got %s", decision.Reason)
		}
		
		if decision.Model != "claude-3-opus" {
			t.Errorf("Expected long context model, got %s", decision.Model)
		}
	})
}
