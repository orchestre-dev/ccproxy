package converter

import (
	"encoding/json"
	"fmt"

	"github.com/musistudio/ccproxy/internal/errors"
)

// Converter interface defines methods for converting between message formats
type Converter interface {
	// ConvertRequest converts a request from one format to another
	ConvertRequest(data json.RawMessage, from, to MessageFormat) (json.RawMessage, error)
	
	// ConvertResponse converts a response from one format to another
	ConvertResponse(data json.RawMessage, from, to MessageFormat) (json.RawMessage, error)
	
	// ConvertStreamEvent converts a streaming event from one format to another
	ConvertStreamEvent(data []byte, from, to MessageFormat) ([]byte, error)
}

// MessageConverter implements the Converter interface
type MessageConverter struct {
	converters map[string]FormatConverter
}

// FormatConverter handles conversion for a specific format
type FormatConverter interface {
	// ToGeneric converts from the specific format to generic format
	ToGeneric(data json.RawMessage, isRequest bool) (json.RawMessage, error)
	
	// FromGeneric converts from generic format to the specific format
	FromGeneric(data json.RawMessage, isRequest bool) (json.RawMessage, error)
	
	// ConvertStreamEvent handles streaming event conversion
	ConvertStreamEvent(data []byte, toFormat MessageFormat) ([]byte, error)
}

// NewMessageConverter creates a new message converter
func NewMessageConverter() *MessageConverter {
	return &MessageConverter{
		converters: map[string]FormatConverter{
			string(FormatAnthropic): NewAnthropicConverter(),
			string(FormatOpenAI):    NewOpenAIConverter(),
			string(FormatGoogle):    NewGoogleConverter(),
			string(FormatAWS):       NewAWSConverter(),
		},
	}
}

// ConvertRequest converts a request between formats
func (mc *MessageConverter) ConvertRequest(data json.RawMessage, from, to MessageFormat) (json.RawMessage, error) {
	if from == to {
		return data, nil
	}

	// Get source converter
	fromConverter, ok := mc.converters[string(from)]
	if !ok {
		return nil, errors.New(errors.ErrorTypeBadRequest, fmt.Sprintf("unsupported source format: %s", from))
	}

	// Convert to generic format
	genericData, err := fromConverter.ToGeneric(data, true)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeBadRequest, "failed to convert to generic format")
	}

	// If target is generic, return now
	if to == FormatGeneric {
		return genericData, nil
	}

	// Get target converter
	toConverter, ok := mc.converters[string(to)]
	if !ok {
		return nil, errors.New(errors.ErrorTypeBadRequest, fmt.Sprintf("unsupported target format: %s", to))
	}

	// Convert from generic to target format
	result, err := toConverter.FromGeneric(genericData, true)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeBadRequest, "failed to convert from generic format")
	}

	return result, nil
}

// ConvertResponse converts a response between formats
func (mc *MessageConverter) ConvertResponse(data json.RawMessage, from, to MessageFormat) (json.RawMessage, error) {
	if from == to {
		return data, nil
	}

	// Get source converter
	fromConverter, ok := mc.converters[string(from)]
	if !ok {
		return nil, errors.New(errors.ErrorTypeBadRequest, fmt.Sprintf("unsupported source format: %s", from))
	}

	// Convert to generic format
	genericData, err := fromConverter.ToGeneric(data, false)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeBadRequest, "failed to convert to generic format")
	}

	// If target is generic, return now
	if to == FormatGeneric {
		return genericData, nil
	}

	// Get target converter
	toConverter, ok := mc.converters[string(to)]
	if !ok {
		return nil, errors.New(errors.ErrorTypeBadRequest, fmt.Sprintf("unsupported target format: %s", to))
	}

	// Convert from generic to target format
	result, err := toConverter.FromGeneric(genericData, false)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeBadRequest, "failed to convert from generic format")
	}

	return result, nil
}

// ConvertStreamEvent converts a streaming event between formats
func (mc *MessageConverter) ConvertStreamEvent(data []byte, from, to MessageFormat) ([]byte, error) {
	if from == to {
		return data, nil
	}

	// Get source converter
	fromConverter, ok := mc.converters[string(from)]
	if !ok {
		return nil, errors.New(errors.ErrorTypeBadRequest, fmt.Sprintf("unsupported source format: %s", from))
	}

	// Convert the stream event
	result, err := fromConverter.ConvertStreamEvent(data, to)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeBadRequest, "failed to convert stream event")
	}

	return result, nil
}