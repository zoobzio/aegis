package cereal

import (
	"sync"
	"time"

	"aegis/catalog"
	"aegis/zlog"
)

// Secure a struct via registered transformer
func Secure[T any](v *T) {
	// Get type name for logging
	typeName := catalog.GetTypeName[T]()

	// Get transformer from catalog (optimized by security adapter)
	transformer, exists := catalog.GetTransformer[T]()
	if !exists {
		zlog.Debug("no security transformer registered",
			zlog.String("type", typeName))
		return // No transformer, leave as-is
	}

	// Track security metrics
	scopesApplied := []string{}
	fieldsRedacted := 0

	// Apply transformation in-place
	if tf, ok := transformer.(catalog.StructTransformer[T]); ok {
		// Get metadata to track what we're doing
		metadata := catalog.Select[T]()
		for _, field := range metadata.Fields {
			if scope := field.Tags["scope"]; scope != "" {
				scopesApplied = append(scopesApplied, scope)
				if !hasPermission(scope) {
					fieldsRedacted++
				}
			}
		}

		tf(*v, v) // Source and dest are the same for in-place

		// Only log if security actions were taken
		if fieldsRedacted > 0 {
			zlog.Info("security redaction applied",
				zlog.String("type", typeName),
				zlog.String("user", currentContext.UserID),
				zlog.Int("fields_redacted", fieldsRedacted),
				zlog.Strings("scopes_blocked", scopesApplied))

			// Note: Security events are emitted by the security adapter
		}
	}
}

// SecurityContext for permission checking
type SecurityContext struct {
	UserID      string
	Permissions []Permission
}

// Permission represents a scope permission
type Permission struct {
	Scope string
}

// Current context (package-level)
var currentContext SecurityContext

// Access tracking for suspicious pattern detection
var (
	accessHistory = make(map[string][]time.Time)
	accessMutex   sync.RWMutex
)

// SetContext updates the security context
func SetContext(ctx SecurityContext) {
	currentContext = ctx
}

// hasPermission checks if the current context has the required scope
func hasPermission(scope string) bool {
	for _, perm := range currentContext.Permissions {
		if perm.Scope == scope {
			return true
		}
	}
	return false
}
