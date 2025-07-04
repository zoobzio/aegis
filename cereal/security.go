package cereal

import (
	"aegis/catalog"
)

// Secure a struct via registered transformer
func Secure[T any](v *T) {
	// Get transformer from catalog (optimized by security adapter)
	transformer, exists := catalog.GetTransformer[T]()
	if !exists {
		// No transformer registered
		return // No transformer, leave as-is
	}

	// Track security metrics
	scopesApplied := []string{}

	// Apply transformation in-place
	if tf, ok := transformer.(catalog.StructTransformer[T]); ok {
		// Get metadata to track what we're doing
		metadata := catalog.Select[T]()
		for _, field := range metadata.Fields {
			if scope := field.Tags["scope"]; scope != "" {
				scopesApplied = append(scopesApplied, scope)
			}
		}

		tf(*v, v) // Source and dest are the same for in-place
	}
}
