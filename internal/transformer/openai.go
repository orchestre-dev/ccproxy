package transformer

// OpenAITransformer is a minimal transformer for OpenAI API
// It primarily serves to define the endpoint, as OpenAI's format
// is already the standard that other transformers convert to/from
type OpenAITransformer struct {
	BaseTransformer
}

// NewOpenAITransformer creates a new OpenAI transformer
func NewOpenAITransformer() *OpenAITransformer {
	return &OpenAITransformer{
		BaseTransformer: *NewBaseTransformer("openai", "/v1/chat/completions"),
	}
}
