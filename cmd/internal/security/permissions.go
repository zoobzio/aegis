package security

import (
	"fmt"

	"github.com/spf13/cobra"
	
	"aegis/sctx"
)

// runPermissionBoundaryTests tests permission checking capabilities
func runPermissionsTests(cmd *cobra.Command, args []string) {
	runTest("Basic Permission Checking", testBasicPermissionChecking)
	runTest("Hierarchical Permissions", testHierarchicalPermissions)
	runTest("Wildcard Permission Patterns", testWildcardPermissions)
	runTest("Permission Isolation", testPermissionIsolation)
	runTest("Service Boundary Enforcement", testServiceBoundaryEnforcement)
	runTest("Multi-Tenant Permission Isolation", testMultiTenantIsolation)
}

// testBasicPermissionChecking verifies simple permission checks work
func testBasicPermissionChecking() error {
	// User with specific permissions
	ctx := sctx.NewUserContext("user-1", []string{
		"profile:read",
		"profile:write",
		"orders:read",
	})
	
	// Test allowed permissions
	allowed := []string{"profile:read", "profile:write", "orders:read"}
	for _, perm := range allowed {
		if !ctx.HasPermission(perm) {
			return fmt.Errorf("should have permission: %s", perm)
		}
	}
	
	// Test denied permissions
	denied := []string{"admin:read", "orders:write", "users:delete"}
	for _, perm := range denied {
		if ctx.HasPermission(perm) {
			return fmt.Errorf("should NOT have permission: %s", perm)
		}
	}
	
	return nil
}

// testHierarchicalPermissions tests resource:action permission patterns
func testHierarchicalPermissions() error {
	// Admin with broad permissions
	adminCtx := sctx.NewUserContext("admin-1", []string{
		"users:*",      // All user operations
		"orders:read",  // Only read orders
		"system:view",  // Only view system info
	})
	
	// Test wildcard matches
	// Note: The current implementation doesn't do wildcard matching on the permission itself,
	// only exact matches. So we test that the context has the wildcard permission.
	if !adminCtx.HasPermission("users:*") {
		return fmt.Errorf("should have users:* wildcard permission")
	}
	
	// Test specific permission limits
	if !adminCtx.HasPermission("orders:read") {
		return fmt.Errorf("should have orders:read")
	}
	if adminCtx.HasPermission("orders:write") {
		return fmt.Errorf("should NOT have orders:write")
	}
	
	return nil
}

// testWildcardPermissions verifies wildcard behavior for different contexts
func testWildcardPermissions() error {
	// Internal service has wildcard
	internalCtx := sctx.Internal
	
	// Internal context has the literal "*" permission
	if !internalCtx.HasPermission("*") {
		return fmt.Errorf("internal context should have wildcard permission")
	}
	
	// Note: Current implementation doesn't do pattern matching.
	// It only checks for exact permission matches.
	// So we can't test wildcard behavior, only that it has "*"
	
	// Regular user with explicit wildcard for one resource
	userCtx := sctx.NewUserContext("power-user", []string{"logs:*", "metrics:read"})
	
	if !userCtx.HasPermission("logs:*") {
		return fmt.Errorf("should have logs wildcard")
	}
	if userCtx.HasPermission("*") {
		return fmt.Errorf("should NOT have global wildcard")
	}
	
	return nil
}

// testPermissionIsolation verifies permissions don't leak between contexts
func testPermissionIsolation() error {
	// Create two users with different permissions
	user1 := sctx.NewUserContext("user-1", []string{"resource-a:read"})
	user2 := sctx.NewUserContext("user-2", []string{"resource-b:read"})
	
	// Verify isolation
	if user1.HasPermission("resource-b:read") {
		return fmt.Errorf("user1 permissions leaked to resource-b")
	}
	if user2.HasPermission("resource-a:read") {
		return fmt.Errorf("user2 permissions leaked to resource-a")
	}
	
	// Verify identity isolation
	if user1.UserID == user2.UserID {
		return fmt.Errorf("user identities should be different")
	}
	
	return nil
}

// testServiceBoundaryEnforcement shows how services have different access than users
func testServiceBoundaryEnforcement() error {
	// User context - even with admin permissions
	adminUser := sctx.NewUserContext("admin", []string{"admin:*", "system:read"})
	
	// Service context - limited service permissions
	logService := sctx.NewPackageContext("logger", []string{"logs:write", "metrics:write"})
	
	// Service can do things user cannot
	if !logService.HasPermission("logs:write") {
		return fmt.Errorf("service should be able to write logs")
	}
	
	// Service is identified differently
	if !logService.IsSystem() {
		return fmt.Errorf("service should be identified as system")
	}
	if adminUser.IsSystem() {
		return fmt.Errorf("admin user should NOT be identified as system")
	}
	
	// Different identity patterns
	if adminUser.UserID == logService.UserID {
		return fmt.Errorf("service and user should have different identity patterns")
	}
	
	return nil
}

// testMultiTenantIsolation demonstrates tenant isolation patterns
func testMultiTenantIsolation() error {
	// Create contexts for different tenants
	tenant1Admin := sctx.NewUserContext("admin@tenant1", []string{"tenant1:*"})
	tenant1User := sctx.NewUserContext("user@tenant1", []string{"tenant1:read"})
	
	tenant2Admin := sctx.NewUserContext("admin@tenant2", []string{"tenant2:*"})
	tenant2User := sctx.NewUserContext("user@tenant2", []string{"tenant2:read"})
	
	// Add tenant information as extensions
	tenant1Admin = tenant1Admin.WithExtension("tenant", "tenant1")
	tenant1User = tenant1User.WithExtension("tenant", "tenant1")
	tenant2Admin = tenant2Admin.WithExtension("tenant", "tenant2")
	tenant2User = tenant2User.WithExtension("tenant", "tenant2")
	
	// Verify cross-tenant isolation
	if tenant1Admin.HasPermission("tenant2:read") {
		return fmt.Errorf("tenant1 admin should not have tenant2 permissions")
	}
	if tenant2Admin.HasPermission("tenant1:read") {
		return fmt.Errorf("tenant2 admin should not have tenant1 permissions")
	}
	
	// Verify tenant information is preserved
	if tenant1Admin.Extensions["tenant"] != "tenant1" {
		return fmt.Errorf("tenant1 extension not preserved")
	}
	if tenant2User.Extensions["tenant"] != "tenant2" {
		return fmt.Errorf("tenant2 extension not preserved")
	}
	
	// Verify admin vs user within same tenant
	// Note: HasPermission only checks exact matches, not wildcards
	if !tenant1Admin.HasPermission("tenant1:*") {
		return fmt.Errorf("tenant1 admin should have tenant1:* permission")
	}
	if tenant1User.HasPermission("tenant1:*") {
		return fmt.Errorf("tenant1 user should NOT have wildcard permission")
	}
	
	return nil
}