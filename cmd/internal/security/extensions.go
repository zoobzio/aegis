package security

import (
	"fmt"

	"github.com/spf13/cobra"
	
	"aegis/sctx"
)

// runContextExtensibilityTests tests security context extension capabilities
func runExtensionsTests(cmd *cobra.Command, args []string) {
	runTest("Basic Extension Addition", testBasicExtensions)
	runTest("Extension Immutability", testExtensionImmutability)
	runTest("Domain-Specific Extensions", testDomainSpecificExtensions)
	runTest("Extension Type Safety", testExtensionTypeSafety)
	runTest("Multi-Service Extensions", testMultiServiceExtensions)
	runTest("Extension Evolution", testExtensionEvolution)
}

// testBasicExtensions verifies basic extension functionality
func testBasicExtensions() error {
	// Start with a basic context
	ctx := sctx.NewUserContext("user-1", []string{"read"})
	
	// Add an extension
	ctxWithTenant := ctx.WithExtension("tenant_id", "tenant-123")
	
	// Verify extension was added
	if ctxWithTenant.Extensions == nil {
		return fmt.Errorf("extensions map not initialized")
	}
	
	tenantID, ok := ctxWithTenant.Extensions["tenant_id"]
	if !ok {
		return fmt.Errorf("tenant_id extension not found")
	}
	
	if tenantID != "tenant-123" {
		return fmt.Errorf("expected tenant-123, got %v", tenantID)
	}
	
	return nil
}

// testExtensionImmutability verifies contexts are immutable
func testExtensionImmutability() error {
	// Create original context
	original := sctx.NewUserContext("user-1", []string{"read"})
	
	// Add multiple extensions
	extended1 := original.WithExtension("key1", "value1")
	extended2 := extended1.WithExtension("key2", "value2")
	extended3 := extended2.WithExtension("key3", "value3")
	
	// Verify original is unchanged
	if len(original.Extensions) > 0 {
		return fmt.Errorf("original context was modified")
	}
	
	// Verify each extension level
	if len(extended1.Extensions) != 1 {
		return fmt.Errorf("extended1 should have 1 extension")
	}
	if len(extended2.Extensions) != 2 {
		return fmt.Errorf("extended2 should have 2 extensions")
	}
	if len(extended3.Extensions) != 3 {
		return fmt.Errorf("extended3 should have 3 extensions")
	}
	
	// Verify all extensions are preserved
	if extended3.Extensions["key1"] != "value1" {
		return fmt.Errorf("key1 not preserved through extensions")
	}
	
	return nil
}

// testDomainSpecificExtensions shows how different domains use extensions
func testDomainSpecificExtensions() error {
	// Base user context
	user := sctx.NewUserContext("user-1", []string{"app:use"})
	
	// E-commerce domain adds shopping cart
	ecommerceCtx := user.
		WithExtension("cart_id", "cart-789").
		WithExtension("currency", "USD").
		WithExtension("tax_exempt", false)
	
	// Healthcare domain adds patient info
	healthcareCtx := user.
		WithExtension("patient_id", "PT-456").
		WithExtension("consent_level", "full").
		WithExtension("emergency_contact", "555-0123")
	
	// Banking domain adds account info
	bankingCtx := user.
		WithExtension("account_id", "ACC-123").
		WithExtension("risk_score", 750).
		WithExtension("2fa_verified", true)
	
	// Verify domain isolation
	if _, ok := ecommerceCtx.Extensions["patient_id"]; ok {
		return fmt.Errorf("e-commerce context should not have healthcare data")
	}
	if _, ok := healthcareCtx.Extensions["account_id"]; ok {
		return fmt.Errorf("healthcare context should not have banking data")
	}
	if _, ok := bankingCtx.Extensions["cart_id"]; ok {
		return fmt.Errorf("banking context should not have e-commerce data")
	}
	
	return nil
}

// testExtensionTypeSafety verifies different types can be stored
func testExtensionTypeSafety() error {
	ctx := sctx.NewUserContext("user-1", nil)
	
	// Add different types as extensions
	ctx = ctx.
		WithExtension("string_val", "hello").
		WithExtension("int_val", 42).
		WithExtension("bool_val", true).
		WithExtension("float_val", 3.14).
		WithExtension("slice_val", []string{"a", "b", "c"}).
		WithExtension("map_val", map[string]int{"x": 1, "y": 2})
	
	// Type assertions should work
	if str, ok := ctx.Extensions["string_val"].(string); !ok || str != "hello" {
		return fmt.Errorf("string extension failed")
	}
	
	if num, ok := ctx.Extensions["int_val"].(int); !ok || num != 42 {
		return fmt.Errorf("int extension failed")
	}
	
	if b, ok := ctx.Extensions["bool_val"].(bool); !ok || b != true {
		return fmt.Errorf("bool extension failed")
	}
	
	if f, ok := ctx.Extensions["float_val"].(float64); !ok || f != 3.14 {
		return fmt.Errorf("float extension failed")
	}
	
	if slice, ok := ctx.Extensions["slice_val"].([]string); !ok || len(slice) != 3 {
		return fmt.Errorf("slice extension failed")
	}
	
	if m, ok := ctx.Extensions["map_val"].(map[string]int); !ok || m["x"] != 1 {
		return fmt.Errorf("map extension failed")
	}
	
	return nil
}

// testMultiServiceExtensions shows how contexts flow through services
func testMultiServiceExtensions() error {
	// User makes initial request
	userCtx := sctx.NewUserContext("user-1", []string{"app:use"})
	
	// API Gateway adds request metadata
	gatewayCtx := userCtx.
		WithExtension("request_id", "req-123").
		WithExtension("client_ip", "192.168.1.1").
		WithExtension("user_agent", "Mozilla/5.0")
	
	// Auth service adds auth details
	authCtx := gatewayCtx.
		WithExtension("session_id", "sess-456").
		WithExtension("auth_method", "oauth2").
		WithExtension("token_expires", "2024-01-01T00:00:00Z")
	
	// Business service adds domain data
	businessCtx := authCtx.
		WithExtension("tenant_id", "tenant-789").
		WithExtension("feature_flags", []string{"new_ui", "beta_api"})
	
	// Audit service adds tracking
	auditCtx := businessCtx.
		WithExtension("audit_trail", []string{"gateway", "auth", "business"}).
		WithExtension("start_time", "2023-12-01T12:00:00Z")
	
	// Verify all extensions accumulated
	if len(auditCtx.Extensions) < 10 {
		return fmt.Errorf("extensions should accumulate through services")
	}
	
	// Verify early extensions still accessible
	if auditCtx.Extensions["request_id"] != "req-123" {
		return fmt.Errorf("request_id from gateway should be preserved")
	}
	
	return nil
}

// testExtensionEvolution shows how extensions can evolve
func testExtensionEvolution() error {
	// Start with v1 context
	v1Ctx := sctx.NewUserContext("user-1", []string{"app:v1"}).
		WithExtension("api_version", "v1").
		WithExtension("feature", "basic")
	
	// Upgrade to v2 (add new, preserve old)
	v2Ctx := v1Ctx.
		WithExtension("api_version", "v2"). // Override
		WithExtension("feature", "advanced"). // Override
		WithExtension("new_capability", true) // Add new
	
	// Verify evolution
	if v2Ctx.Extensions["api_version"] != "v2" {
		return fmt.Errorf("api_version should be updated to v2")
	}
	
	if v2Ctx.Extensions["feature"] != "advanced" {
		return fmt.Errorf("feature should be upgraded")
	}
	
	if v2Ctx.Extensions["new_capability"] != true {
		return fmt.Errorf("new capability should be added")
	}
	
	// Original context unchanged
	if v1Ctx.Extensions["api_version"] != "v1" {
		return fmt.Errorf("original context should remain v1")
	}
	
	return nil
}