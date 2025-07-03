package security

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	
	"aegis/sctx"
)

// runContextIdentityTests tests identity management capabilities
func runIdentityTests(cmd *cobra.Command, args []string) {
	runTest("User Context Creation", testUserContextCreation)
	runTest("System Context Creation", testSystemContextCreation)
	runTest("Context Type Detection", testContextTypeDetection)
	runTest("Identity Preservation", testIdentityPreservation)
	runTest("Public Context Behavior", testPublicContextBehavior)
	runTest("Internal Context Privileges", testInternalContextPrivileges)
}

// testUserContextCreation verifies user identity contexts work correctly
func testUserContextCreation() error {
	// Create a context for a regular user
	userCtx := sctx.NewUserContext("user-123", []string{"profile:read", "profile:write"})
	
	// Verify identity is preserved
	if userCtx.UserID != "user-123" {
		return fmt.Errorf("user identity not preserved")
	}
	
	// Verify permissions are set
	if len(userCtx.Permissions) != 2 {
		return fmt.Errorf("expected 2 permissions, got %d", len(userCtx.Permissions))
	}
	
	// Verify it's not a system context
	if userCtx.IsSystem() {
		return fmt.Errorf("user context incorrectly identified as system")
	}
	
	return nil
}

// testSystemContextCreation verifies service/package contexts work correctly
func testSystemContextCreation() error {
	// Create a context for an internal service
	svcCtx := sctx.NewPackageContext("auth-service", []string{"user:*", "session:*"})
	
	// Verify system prefix is added
	if !strings.HasPrefix(svcCtx.UserID, "system:") {
		return fmt.Errorf("system context missing 'system:' prefix")
	}
	
	// Verify full identity
	if svcCtx.UserID != "system:auth-service" {
		return fmt.Errorf("expected 'system:auth-service', got %s", svcCtx.UserID)
	}
	
	// Verify it's identified as system
	if !svcCtx.IsSystem() {
		return fmt.Errorf("system context not identified correctly")
	}
	
	return nil
}

// testContextTypeDetection verifies we can distinguish context types
func testContextTypeDetection() error {
	contexts := []struct {
		name     string
		ctx      sctx.SecurityContext
		isSystem bool
	}{
		{
			name:     "regular user",
			ctx:      sctx.NewUserContext("user-456", nil),
			isSystem: false,
		},
		{
			name:     "package context",
			ctx:      sctx.NewPackageContext("logger", nil),
			isSystem: true,
		},
		{
			name:     "internal context",
			ctx:      sctx.Internal,
			isSystem: true,
		},
		{
			name:     "public context",
			ctx:      sctx.Public,
			isSystem: false,
		},
	}
	
	for _, tc := range contexts {
		if tc.ctx.IsSystem() != tc.isSystem {
			return fmt.Errorf("%s: expected IsSystem=%v, got %v", 
				tc.name, tc.isSystem, tc.ctx.IsSystem())
		}
		
	}
	
	return nil
}

// testIdentityPreservation verifies identity survives operations
func testIdentityPreservation() error {
	// Create a user context
	original := sctx.NewUserContext("user-789", []string{"read", "write"})
	
	// Add an extension (should return new context)
	extended := original.WithExtension("tenant", "acme-corp")
	
	// Verify original is unchanged
	if len(original.Extensions) > 0 {
		return fmt.Errorf("original context was modified")
	}
	
	// Verify identity is preserved in new context
	if extended.UserID != original.UserID {
		return fmt.Errorf("user identity changed during extension")
	}
	
	// Verify permissions are preserved
	if len(extended.Permissions) != len(original.Permissions) {
		return fmt.Errorf("permissions changed during extension")
	}
	
	// Verify extension was added
	if extended.Extensions["tenant"] != "acme-corp" {
		return fmt.Errorf("extension not added correctly")
	}
	
	return nil
}

// testPublicContextBehavior verifies unauthenticated access patterns
func testPublicContextBehavior() error {
	ctx := sctx.Public
	
	// Should have public identity
	if ctx.UserID != "public" {
		return fmt.Errorf("public context should have public identity, got %s", ctx.UserID)
	}
	
	// Should have public:read permission
	if len(ctx.Permissions) != 1 || ctx.Permissions[0] != "public:read" {
		return fmt.Errorf("public context should have public:read permission, got %v", ctx.Permissions)
	}
	
	// Should not be system
	if ctx.IsSystem() {
		return fmt.Errorf("public context incorrectly identified as system")
	}
	
	// Should have public:read permission
	if !ctx.HasPermission("public:read") {
		return fmt.Errorf("public context should have public:read permission")
	}
	
	// Should not have other permissions
	if ctx.HasPermission("private:read") || ctx.HasPermission("admin:read") {
		return fmt.Errorf("public context should only have public:read permission")
	}
	
	return nil
}

// testInternalContextPrivileges verifies trusted service behavior
func testInternalContextPrivileges() error {
	ctx := sctx.Internal
	
	// Should have special identity
	if ctx.UserID != "system:internal" {
		return fmt.Errorf("internal context has wrong identity")
	}
	
	// Should have wildcard permission
	if len(ctx.Permissions) != 1 || ctx.Permissions[0] != "*" {
		return fmt.Errorf("internal context should have wildcard permission")
	}
	
	// Should be system
	if !ctx.IsSystem() {
		return fmt.Errorf("internal context not properly identified as system")
	}
	
	// Should have wildcard permission
	if !ctx.HasPermission("*") {
		return fmt.Errorf("internal context should have wildcard permission")
	}
	
	// Note: Current implementation doesn't do wildcard matching,
	// it only checks for exact permission matches.
	// So we can only verify it has the "*" permission itself.
	
	return nil
}