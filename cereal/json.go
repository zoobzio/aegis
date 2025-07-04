package cereal

import (
	"encoding/json"
	
	"aegis/sctx"
)

// PUBLIC API

// MarshalJSON serializes a value to JSON with all registered transformations
// EGRESS: Apply all registered behaviors in the serialization pipeline
func MarshalJSON[T any](v T, ctx sctx.SecurityContext) ([]byte, error) {
	// Ensure the serialization pipeline exists for this type
	pipeline := EnsureSerializationPipeline[T]()
	
	// Create input with extensions map for adapter-specific data
	input := SerializationInput[T]{
		Data:       v,
		Context:    ctx,
		Extensions: make(map[string]any),
	}
	
	// Apply security transformations first!
	Secure(&input.Data)
	
	// Process through all registered behaviors in order
	stages := []SerializationKey{PreProcess, Transform, Validate, Enrich, PostProcess}
	
	for _, stage := range stages {
		if output, exists := pipeline.Process(stage, input); exists {
			if output.Error != nil {
				return nil, output.Error
			}
			// Chain the output data to next stage
			input.Data = output.Data
			// Merge any metadata
			if output.ProcessingMetadata != nil {
				for k, v := range output.ProcessingMetadata {
					input.Extensions[k] = v
				}
			}
		}
	}
	
	return json.Marshal(input.Data)
}

// UnmarshalJSON deserializes JSON data with validation
// INGRESS: Validate input based on security context
func UnmarshalJSON[T any](data []byte, ctx sctx.SecurityContext) (T, error) {
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return result, err
	}
	
	// Input validation should be done through the validation pipeline
	// after unmarshaling if needed by the application
	
	return result, nil
}
