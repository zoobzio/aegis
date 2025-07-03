package security

import (
	"aegis/catalog"
	"aegis/cereal"
	"aegis/zlog"
)

// RegisterSerializationSecurity registers security transformations for a type
// This hooks into cereal's serialization pipeline
func RegisterSerializationSecurity[T any]() {
	// Check if type has security tags
	metadata := catalog.Select[T]()
	hasSecurityTags := false
	
	for _, field := range metadata.Fields {
		if field.Tags["scope"] != "" || field.Tags["validate"] != "" || 
		   field.Tags["encrypt"] != "" || field.Tags["security"] != "" {
			hasSecurityTags = true
			break
		}
	}
	
	zlog.Debug("RegisterSerializationSecurity called",
		zlog.String("type", catalog.GetTypeName[T]()),
		zlog.Bool("has_security_tags", hasSecurityTags),
		zlog.Int("field_count", len(metadata.Fields)),
	)
	
	if !hasSecurityTags {
		return // No security tags, nothing to do
	}
	
	// Get the serialization pipeline for this type
	pipeline := cereal.GetSerializationPipeline[T]()
	
	// Register our security transformer
	pipeline.Register(cereal.Transform, createSerializationSecurityProcessor[T]())
	
	zlog.Debug("Security processor registered",
		zlog.String("type", catalog.GetTypeName[T]()),
		zlog.String("stage", string(cereal.Transform)),
	)
}

// createSerializationSecurityProcessor creates a processor for cereal's pipeline
func createSerializationSecurityProcessor[T any]() func(cereal.SerializationInput[T]) cereal.SerializationOutput[T] {
	return func(input cereal.SerializationInput[T]) cereal.SerializationOutput[T] {
		// Get metadata and manipulators once
		metadata := catalog.Select[T]()
		manipulators := catalog.GetFieldManipulators[T]()
		
		// Work with the data
		result := input.Data
		redactedFields := []string{}
		
		// Debug logging
		zlog.Debug("Security processor executing",
			zlog.String("type", catalog.GetTypeName[T]()),
			zlog.Int("field_count", len(metadata.Fields)),
		)
		
		// Process each field based on its tags
		for _, field := range metadata.Fields {
			manipulator, exists := manipulators[field.Name]
			if !exists {
				continue
			}
			
			// Check tags
			validateTag := field.Tags["validate"]
			scope := field.Tags["scope"]
			
			// Check scope-based access control first
			if scope != "" && !input.Context.HasPermission(scope) {
				zlog.Debug("No permission for field",
					zlog.String("field", field.Name),
					zlog.String("scope", scope),
					zlog.Strings("user_permissions", input.Context.Permissions),
				)
				
				// No permission - redact the field entirely
				manipulator.Redact(&result)
				redactedFields = append(redactedFields, field.Name)
				continue
			}
			
			// If user has permission (or no scope required), check if we should mask
			if validateTag != "" {
				if maskFunc, hasMask := catalog.GetMaskFunction(validateTag); hasMask {
					if currentVal, err := manipulator.GetString(result); err == nil {
						masked := maskFunc(currentVal)
						manipulator.SetString(&result, masked)
						redactedFields = append(redactedFields, field.Name)
					}
				}
			}
		}
		
		// Return with metadata about what we did
		return cereal.SerializationOutput[T]{
			Data:  result,
			Error: nil,
			ProcessingMetadata: map[string]any{
				"security.redacted_fields": redactedFields,
				"security.processor": "default",
			},
		}
	}
}