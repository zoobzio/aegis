// Package sctx provides security context definitions for the entire aegis framework.
// Security contexts are used for both external users and internal package boundaries.
package sctx

import (
	"slices"
	"time"
)

// SecurityContext represents security information for access control.
// It can represent either:
// - External user context (from authentication)
// - Internal package context (for service boundaries)
type SecurityContext struct {
	// Identity of the context owner
	// For users: "user-123"
	// For packages: "system:billing-service"
	UserID string

	// List of permission scopes
	// Examples: "user:read", "order:create", "payment:write"
	Permissions []string

	// Service-specific security data
	// Allows packages to extend context without modifying core
	Extensions map[string]any
}

// NewUserContext creates a security context for an authenticated user
func NewUserContext(userID string, permissions []string) SecurityContext {
	return SecurityContext{
		UserID:      userID,
		Permissions: permissions,
		Extensions:  make(map[string]any),
	}
}

// NewPackageContext creates a security context for an internal package
func NewPackageContext(packageName string, permissions []string) SecurityContext {
	return SecurityContext{
		UserID:      "system:" + packageName,
		Permissions: permissions,
		Extensions:  make(map[string]any),
	}
}

// HasPermission checks if the context includes a specific permission scope
func (ctx SecurityContext) HasPermission(scope string) bool {
	return slices.Contains(ctx.Permissions, scope)
}

// IsSystem checks if this is a system/package context
func (ctx SecurityContext) IsSystem() bool {
	return len(ctx.UserID) > 7 && ctx.UserID[:7] == "system:"
}

// WithExtension returns a new context with an additional extension
func (ctx SecurityContext) WithExtension(key string, value any) SecurityContext {
	// Create a proper copy with new map
	newCtx := SecurityContext{
		UserID:      ctx.UserID,
		Permissions: ctx.Permissions, // slices are ok to share for read-only
		Extensions:  make(map[string]any),
	}

	// Copy existing extensions
	for k, v := range ctx.Extensions {
		newCtx.Extensions[k] = v
	}

	// Add new extension
	newCtx.Extensions[key] = value
	return newCtx
}

// Common package contexts for core services
var (
	// Public context for unauthenticated access
	Public = SecurityContext{
		UserID:      "public",
		Permissions: []string{"public:read"},
	}

	// Internal context for trusted service-to-service communication
	Internal = SecurityContext{
		UserID:      "system:internal",
		Permissions: []string{"*"}, // All permissions
	}
)

// SecurityError represents a security violation that requires action
type SecurityError struct {
	Type     string // "unauthorized", "forbidden", "suspicious", "breach"
	Resource string // What resource was being accessed
	Action   string // What action was attempted
	Message  string // Human-readable error message
}

func (e SecurityError) Error() string {
	return e.Message
}

// SecurityViolationEvent is emitted when security rules are violated
type SecurityViolationEvent struct {
	Timestamp time.Time
	UserID    string
	Resource  string
	Action    string
	Severity  string         // "info", "warning", "critical", "breach"
	Context   map[string]any // Additional context about the violation
}

// SecurityLevel represents the current system security posture
type SecurityLevel int

const (
	GreenLevel  SecurityLevel = iota // All clear, normal operations
	YellowLevel                      // Elevated: suspicious activity detected
	OrangeLevel                      // High: active threat detected
	RedLevel                         // Critical: system under attack
)

// SecurityLevelChange is emitted when the system security level changes
type SecurityLevelChange struct {
	Previous SecurityLevel
	Current  SecurityLevel
	Reason   string
	Time     time.Time
}
