package cereal

import (
	"aegis/pipz"
	"aegis/sctx"
)

// SerializationKey identifies stages in the serialization pipeline
type SerializationKey string

const (
	// PreProcess runs before any transformation (auditing, metrics, etc.)
	PreProcess SerializationKey = "pre_process"
	
	// Transform modifies the data (security, normalization, etc.)
	Transform SerializationKey = "transform"
	
	// Validate ensures data meets requirements before serialization
	Validate SerializationKey = "validate"
	
	// Enrich adds computed fields or metadata
	Enrich SerializationKey = "enrich"
	
	// PostProcess runs after all transformations (logging, caching, etc.)
	PostProcess SerializationKey = "post_process"
)

// SerializationInput provides data and context to processors
type SerializationInput[T any] struct {
	Data    T
	Context sctx.SecurityContext
	// Extensions for adapter-specific data
	Extensions map[string]any
}

// SerializationOutput returns processed data or error
type SerializationOutput[T any] struct {
	Data  T
	Error error
	// Metadata about what happened during processing
	ProcessingMetadata map[string]any
}

// SerializationProcessor is a function that processes data during serialization
type SerializationProcessor[T any] pipz.Processor[SerializationInput[T], SerializationOutput[T]]

// GetSerializationPipeline returns the serialization pipeline for type T
func GetSerializationPipeline[T any]() *pipz.ServiceContract[SerializationKey, SerializationInput[T], SerializationOutput[T]] {
	return pipz.GetContract[SerializationKey, SerializationInput[T], SerializationOutput[T]]()
}

// EnsureSerializationPipeline ensures a serialization pipeline exists for type T
// and optionally registers default behaviors if needed
func EnsureSerializationPipeline[T any]() *pipz.ServiceContract[SerializationKey, SerializationInput[T], SerializationOutput[T]] {
	pipeline := GetSerializationPipeline[T]()
	
	// The pipeline is automatically created by pipz.GetContract
	// Adapters can register their own processors to it
	
	return pipeline
}