package cereal

import (
	"encoding/json"
	
	"aegis/catalog"
	"aegis/sctx"
	"aegis/zlog"
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
	
	// Process through all registered behaviors in order
	stages := []SerializationKey{PreProcess, Transform, Validate, Enrich, PostProcess}
	
	// Debug: check what's registered
	registeredKeys := pipeline.ListKeys()
	zlog.Debug("MarshalJSON processing",
		zlog.String("type", catalog.GetTypeName[T]()),
		zlog.Int("pipeline_keys", len(registeredKeys)),
	)
	
	// Log all registered keys
	for _, key := range registeredKeys {
		zlog.Debug("Registered key", zlog.String("key", string(key)))
	}
	
	for _, stage := range stages {
		if output, exists := pipeline.Process(stage, input); exists {
			zlog.Debug("Processing stage",
				zlog.String("stage", string(stage)),
				zlog.Bool("exists", exists),
			)
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
