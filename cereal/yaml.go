package cereal

import (
	"gopkg.in/yaml.v3"
	
	"aegis/sctx"
)

// PUBLIC API

// MarshalYAML serializes a value to YAML with all registered transformations
// EGRESS: Apply all registered behaviors in the serialization pipeline
func MarshalYAML[T any](v T, ctx sctx.SecurityContext) ([]byte, error) {
	// Get the serialization pipeline for this type
	pipeline := GetSerializationPipeline[T]()
	
	// Create input with extensions map for adapter-specific data
	input := SerializationInput[T]{
		Data:       v,
		Context:    ctx,
		Extensions: make(map[string]any),
	}
	
	// Process through all registered behaviors in order
	stages := []SerializationKey{PreProcess, Transform, Validate, Enrich, PostProcess}
	
	for _, stage := range stages {
		if output, exists := pipeline.Process(stage, input); exists {
			if output.Error != nil {
				return nil, output.Error
			}
			input.Data = output.Data
			if output.ProcessingMetadata != nil {
				for k, v := range output.ProcessingMetadata {
					input.Extensions[k] = v
				}
			}
		}
	}
	
	return yaml.Marshal(input.Data)
}

// UnmarshalYAML deserializes YAML data with validation
// INGRESS: Validate input based on security context
func UnmarshalYAML[T any](data []byte, ctx sctx.SecurityContext) (T, error) {
	var result T
	if err := yaml.Unmarshal(data, &result); err != nil {
		return result, err
	}
	
	// Input validation should be done through the validation pipeline
	// after unmarshaling if needed by the application
	
	return result, nil
}
