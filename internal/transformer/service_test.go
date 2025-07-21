package transformer

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/orchestre-dev/ccproxy/internal/config"
	testutil "github.com/orchestre-dev/ccproxy/internal/testing"
)

// MockTransformer for testing
type MockTransformer struct {
	*BaseTransformer
	transformRequestInFunc   func(ctx context.Context, request interface{}, provider string) (interface{}, error)
	transformRequestOutFunc  func(ctx context.Context, request interface{}) (interface{}, error)
	transformResponseInFunc  func(ctx context.Context, response *http.Response) (*http.Response, error)
	transformResponseOutFunc func(ctx context.Context, response *http.Response) (*http.Response, error)
}

func NewMockTransformer(name, endpoint string) *MockTransformer {
	return &MockTransformer{
		BaseTransformer: NewBaseTransformer(name, endpoint),
	}
}

func (m *MockTransformer) TransformRequestIn(ctx context.Context, request interface{}, provider string) (interface{}, error) {
	if m.transformRequestInFunc != nil {
		return m.transformRequestInFunc(ctx, request, provider)
	}
	return m.BaseTransformer.TransformRequestIn(ctx, request, provider)
}

func (m *MockTransformer) TransformRequestOut(ctx context.Context, request interface{}) (interface{}, error) {
	if m.transformRequestOutFunc != nil {
		return m.transformRequestOutFunc(ctx, request)
	}
	return m.BaseTransformer.TransformRequestOut(ctx, request)
}

func (m *MockTransformer) TransformResponseIn(ctx context.Context, response *http.Response) (*http.Response, error) {
	if m.transformResponseInFunc != nil {
		return m.transformResponseInFunc(ctx, response)
	}
	return m.BaseTransformer.TransformResponseIn(ctx, response)
}

func (m *MockTransformer) TransformResponseOut(ctx context.Context, response *http.Response) (*http.Response, error) {
	if m.transformResponseOutFunc != nil {
		return m.transformResponseOutFunc(ctx, response)
	}
	return m.BaseTransformer.TransformResponseOut(ctx, response)
}

func TestNewService(t *testing.T) {
	service := NewService()

	testutil.AssertNotEqual(t, nil, service)
	testutil.AssertNotEqual(t, nil, service.transformers)
	testutil.AssertNotEqual(t, nil, service.chains)
	testutil.AssertEqual(t, 0, len(service.transformers))
	testutil.AssertEqual(t, 0, len(service.chains))
	testutil.AssertEqual(t, 100, service.maxCacheSize)
}

func TestService_Register(t *testing.T) {
	service := NewService()

	t.Run("SuccessfulRegistration", func(t *testing.T) {
		transformer := NewMockTransformer("test-transformer", "/v1/test")

		err := service.Register(transformer)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, 1, len(service.transformers))
	})

	t.Run("DuplicateRegistration", func(t *testing.T) {
		transformer1 := NewMockTransformer("duplicate", "/v1/duplicate")
		transformer2 := NewMockTransformer("duplicate", "/v1/different")

		err := service.Register(transformer1)
		testutil.AssertNoError(t, err)

		err = service.Register(transformer2)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "transformer already registered")
		testutil.AssertContains(t, err.Error(), "duplicate")
	})

	t.Run("MultipleTransformers", func(t *testing.T) {
		freshService := NewService()
		transformers := []*MockTransformer{
			NewMockTransformer("transformer1", "/v1/endpoint1"),
			NewMockTransformer("transformer2", "/v1/endpoint2"),
			NewMockTransformer("transformer3", "/v1/endpoint3"),
		}

		for _, transformer := range transformers {
			err := freshService.Register(transformer)
			testutil.AssertNoError(t, err)
		}

		testutil.AssertEqual(t, 3, len(freshService.transformers))
	})
}

func TestService_Get(t *testing.T) {
	service := NewService()

	t.Run("GetExistingTransformer", func(t *testing.T) {
		original := NewMockTransformer("get-test", "/v1/get")
		err := service.Register(original)
		testutil.AssertNoError(t, err)

		retrieved, err := service.Get("get-test")
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, "get-test", retrieved.GetName())
		testutil.AssertEqual(t, "/v1/get", retrieved.GetEndpoint())
	})

	t.Run("GetNonExistentTransformer", func(t *testing.T) {
		_, err := service.Get("non-existent")
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "transformer not found")
		testutil.AssertContains(t, err.Error(), "non-existent")
	})

	t.Run("GetFromEmptyService", func(t *testing.T) {
		emptyService := NewService()
		_, err := emptyService.Get("any-name")
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "transformer not found")
	})
}

func TestService_GetByEndpoint(t *testing.T) {
	service := NewService()

	t.Run("SingleTransformerForEndpoint", func(t *testing.T) {
		transformer := NewMockTransformer("endpoint-test", "/v1/chat")
		err := service.Register(transformer)
		testutil.AssertNoError(t, err)

		result := service.GetByEndpoint("/v1/chat")
		testutil.AssertEqual(t, 1, len(result))
		testutil.AssertEqual(t, "endpoint-test", result[0].GetName())
	})

	t.Run("MultipleTransformersForEndpoint", func(t *testing.T) {
		freshService := NewService()
		transformers := []*MockTransformer{
			NewMockTransformer("chat-transformer-1", "/v1/chat"),
			NewMockTransformer("chat-transformer-2", "/v1/chat"),
			NewMockTransformer("different-endpoint", "/v1/completions"),
		}

		for _, transformer := range transformers {
			err := freshService.Register(transformer)
			testutil.AssertNoError(t, err)
		}

		chatTransformers := freshService.GetByEndpoint("/v1/chat")
		testutil.AssertEqual(t, 2, len(chatTransformers))

		completionTransformers := freshService.GetByEndpoint("/v1/completions")
		testutil.AssertEqual(t, 1, len(completionTransformers))
	})

	t.Run("NoTransformersForEndpoint", func(t *testing.T) {
		transformer := NewMockTransformer("test", "/v1/test")
		err := service.Register(transformer)
		testutil.AssertNoError(t, err)

		result := service.GetByEndpoint("/v1/nonexistent")
		testutil.AssertEqual(t, 0, len(result))
	})

	t.Run("EmptyEndpoint", func(t *testing.T) {
		transformer := NewMockTransformer("empty-endpoint", "")
		err := service.Register(transformer)
		testutil.AssertNoError(t, err)

		result := service.GetByEndpoint("")
		testutil.AssertEqual(t, 1, len(result))
		testutil.AssertEqual(t, "empty-endpoint", result[0].GetName())
	})
}

func TestService_CreateChain(t *testing.T) {
	service := NewService()

	// Register test transformers
	transformer1 := NewMockTransformer("transformer-1", "/v1/endpoint1")
	transformer2 := NewMockTransformer("transformer-2", "/v1/endpoint2")

	err := service.Register(transformer1)
	testutil.AssertNoError(t, err)
	err = service.Register(transformer2)
	testutil.AssertNoError(t, err)

	t.Run("SuccessfulChainCreation", func(t *testing.T) {
		configs := []config.TransformerConfig{
			{Name: "transformer-1"},
			{Name: "transformer-2"},
		}

		chain, err := service.CreateChain(configs)
		testutil.AssertNoError(t, err)
		testutil.AssertNotEqual(t, nil, chain)
		testutil.AssertEqual(t, 2, len(chain.transformers))
	})

	t.Run("EmptyChain", func(t *testing.T) {
		configs := []config.TransformerConfig{}

		chain, err := service.CreateChain(configs)
		testutil.AssertNoError(t, err)
		testutil.AssertNotEqual(t, nil, chain)
		testutil.AssertEqual(t, 0, len(chain.transformers))
	})

	t.Run("NonExistentTransformer", func(t *testing.T) {
		configs := []config.TransformerConfig{
			{Name: "transformer-1"},
			{Name: "non-existent"},
		}

		_, err := service.CreateChain(configs)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "failed to get transformer")
		testutil.AssertContains(t, err.Error(), "non-existent")
	})
}

func TestService_CreateChainFromNames(t *testing.T) {
	service := NewService()

	// Register test transformers
	transformer1 := NewMockTransformer("name-transformer-1", "/v1/name1")
	transformer2 := NewMockTransformer("name-transformer-2", "/v1/name2")

	err := service.Register(transformer1)
	testutil.AssertNoError(t, err)
	err = service.Register(transformer2)
	testutil.AssertNoError(t, err)

	t.Run("SuccessfulChainFromNames", func(t *testing.T) {
		names := []string{"name-transformer-1", "name-transformer-2"}

		chain, err := service.CreateChainFromNames(names)
		testutil.AssertNoError(t, err)
		testutil.AssertNotEqual(t, nil, chain)
		testutil.AssertEqual(t, 2, len(chain.transformers))
	})

	t.Run("EmptyNames", func(t *testing.T) {
		names := []string{}

		chain, err := service.CreateChainFromNames(names)
		testutil.AssertNoError(t, err)
		testutil.AssertNotEqual(t, nil, chain)
		testutil.AssertEqual(t, 0, len(chain.transformers))
	})

	t.Run("NonExistentTransformerByName", func(t *testing.T) {
		names := []string{"name-transformer-1", "missing-transformer"}

		_, err := service.CreateChainFromNames(names)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "failed to get transformer")
		testutil.AssertContains(t, err.Error(), "missing-transformer")
	})

	t.Run("SingleTransformerChain", func(t *testing.T) {
		names := []string{"name-transformer-1"}

		chain, err := service.CreateChainFromNames(names)
		testutil.AssertNoError(t, err)
		testutil.AssertNotEqual(t, nil, chain)
		testutil.AssertEqual(t, 1, len(chain.transformers))
	})
}

func TestService_GetChainForProvider(t *testing.T) {
	service := NewService()

	// Register test transformers
	providerTransformer := NewMockTransformer("test-provider", "/v1/test")
	maxTokenTransformer := NewMockTransformer("maxtoken", "/v1/maxtoken")

	err := service.Register(providerTransformer)
	testutil.AssertNoError(t, err)
	err = service.Register(maxTokenTransformer)
	testutil.AssertNoError(t, err)

	t.Run("CreateChainForProvider", func(t *testing.T) {
		chain := service.GetChainForProvider("test-provider")
		testutil.AssertNotEqual(t, nil, chain)
		// Should have at least the provider transformer and common transformers
		testutil.AssertTrue(t, len(chain.transformers) >= 1)
	})

	t.Run("CachedChainForProvider", func(t *testing.T) {
		// First call
		chain1 := service.GetChainForProvider("cached-provider")
		testutil.AssertNotEqual(t, nil, chain1)

		// Second call should return cached version
		chain2 := service.GetChainForProvider("cached-provider")
		testutil.AssertNotEqual(t, nil, chain2)

		// Should have same number of transformers
		testutil.AssertEqual(t, len(chain1.transformers), len(chain2.transformers))
	})
}

func TestService_ConcurrentAccess(t *testing.T) {
	service := NewService()

	t.Run("ConcurrentRegistration", func(t *testing.T) {
		// Test concurrent registration doesn't cause race conditions
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func(id int) {
				transformer := NewMockTransformer(
					fmt.Sprintf("concurrent-%d", id),
					fmt.Sprintf("/v1/concurrent-%d", id),
				)
				err := service.Register(transformer)
				testutil.AssertNoError(t, err)
				done <- true
			}(i)
		}

		// Wait for all registrations to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		testutil.AssertEqual(t, 10, len(service.transformers))
	})
}

func TestService_EdgeCases(t *testing.T) {
	service := NewService()

	t.Run("NilTransformer", func(t *testing.T) {
		// This would panic in real implementation, but let's test the service is robust
		// We can't actually pass nil due to interface requirements, so we'll skip this
		// or test with a transformer that has empty name
		emptyTransformer := NewMockTransformer("", "/empty")
		err := service.Register(emptyTransformer)
		testutil.AssertNoError(t, err) // Empty name should be allowed

		retrieved, err := service.Get("")
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, "", retrieved.GetName())
	})
}
