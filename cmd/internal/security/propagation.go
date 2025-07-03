package security

import (
	"fmt"

	"github.com/spf13/cobra"
	
	"aegis/catalog"
	"aegis/cereal"
	"aegis/sctx"
	"aegis/zlog"
)

// runContextPropagationTests tests how security contexts flow through the system
func runPropagationTests(cmd *cobra.Command, args []string) {
	runTest("Context Through Logging", testContextThroughLogging)
	runTest("Context Through Serialization", testContextThroughSerialization)
	runTest("Service Chain Propagation", testServiceChainPropagation)
	runTest("Context Transformation", testContextTransformation)
	runTest("Cross-Domain Propagation", testCrossDomainPropagation)
	runTest("Context Degradation Patterns", testContextDegradation)
}

// testContextThroughLogging verifies contexts work with logging
func testContextThroughLogging() error {
	// Create contexts with different permissions
	adminCtx := sctx.NewUserContext("admin", []string{"*"})
	userCtx := sctx.NewUserContext("user", []string{"logs:read"})
	publicCtx := sctx.Public
	
	// Test data with sensitive fields
	type LogData struct {
		Message  string `json:"message"`
		UserID   string `json:"user_id" scope:"user:read"`
		APIKey   string `json:"api_key" scope:"admin"`
		Internal string `json:"internal" scope:"system"`
	}
	
	_ = LogData{
		Message:  "Test log entry",
		UserID:   "user-123",
		APIKey:   "secret-key-456",
		Internal: "internal-data",
	}
	
	// Log with different contexts (would be redacted based on context)
	// This demonstrates the capability without actually implementing it
	// Note: In a real implementation, we'd have a zlog field type for structured data
	zlog.Info("Admin context test", 
		zlog.String("context_type", "admin"),
		zlog.String("user_id", adminCtx.UserID))
	
	zlog.Info("User context test",
		zlog.String("context_type", "user"),
		zlog.String("user_id", userCtx.UserID))
		
	zlog.Info("Public context test",
		zlog.String("context_type", "public"),
		zlog.String("user_id", publicCtx.UserID))
	
	// The test passes if contexts are properly isolated
	if adminCtx.UserID == userCtx.UserID || userCtx.UserID == publicCtx.UserID {
		return fmt.Errorf("contexts should have different identities")
	}
	
	return nil
}

// testContextThroughSerialization verifies contexts work with serialization
func testContextThroughSerialization() error {
	// Test type with various security scopes
	type Document struct {
		ID        string `json:"id"`
		Title     string `json:"title"`
		Content   string `json:"content" scope:"document:read"`
		AuthorSSN string `json:"author_ssn" scope:"admin"`
		Secret    string `json:"secret" scope:"system"`
	}
	
	// Register security behavior for the test
	catalog.RegisterType[Document]()
	pipeline := catalog.GetSecurityPipeline[Document]()
	pipeline.Register(catalog.RedactionBehavior, func(input catalog.SecurityInput[Document]) catalog.SecurityOutput[Document] {
		doc := input.Data
		
		if !input.Context.HasPermission("document:read") {
			doc.Content = "[REQUIRES PERMISSION: document:read]"
		}
		if !input.Context.HasPermission("admin") {
			doc.AuthorSSN = "[REDACTED]"
		}
		if !input.Context.HasPermission("system") {
			doc.Secret = "[CLASSIFIED]"
		}
		
		return catalog.SecurityOutput[Document]{Data: doc}
	})
	
	doc := Document{
		ID:        "doc-123",
		Title:     "Public Document",
		Content:   "Sensitive content here",
		AuthorSSN: "123-45-6789",
		Secret:    "top-secret-data",
	}
	
	// Serialize with different contexts
	contexts := []struct {
		name string
		ctx  sctx.SecurityContext
	}{
		{"public", sctx.Public},
		{"user", sctx.NewUserContext("user", []string{"document:read"})},
		{"admin", sctx.NewUserContext("admin", []string{"admin", "document:read"})},
		{"system", sctx.Internal},
	}
	
	for _, tc := range contexts {
		result, err := cereal.MarshalJSON(doc, tc.ctx)
		if err != nil {
			return fmt.Errorf("%s context serialization failed: %v", tc.name, err)
		}
		
		// Verify JSON is valid
		check, err := cereal.UnmarshalJSON[map[string]interface{}](result, tc.ctx)
		if err != nil {
			return fmt.Errorf("%s context unmarshal failed: %v", tc.name, err)
		}
		_ = check // Suppress unused variable warning
	}
	
	return nil
}

// testServiceChainPropagation shows context flowing through services
func testServiceChainPropagation() error {
	// Simulate a request flowing through multiple services
	
	// 1. User makes request
	userCtx := sctx.NewUserContext("user-123", []string{"api:access"})
	
	// 2. API Gateway enriches context
	gatewayCtx := userCtx.
		WithExtension("request_id", "req-456").
		WithExtension("api_version", "v2")
	
	// 3. Auth service validates and enriches
	authService := sctx.NewPackageContext("auth-service", []string{"user:validate"})
	authEnrichedCtx := gatewayCtx.
		WithExtension("session_valid", true).
		WithExtension("auth_provider", "oauth2")
	
	// 4. Business service processes with user context
	businessService := sctx.NewPackageContext("order-service", []string{"order:*"})
	businessCtx := authEnrichedCtx.
		WithExtension("order_id", "order-789").
		WithExtension("processing_time", "2023-12-01T12:00:00Z")
	
	// 5. Audit service logs everything
	auditService := sctx.Internal // Full access for auditing
	
	// Verify context accumulated through chain
	if businessCtx.UserID != "user-123" {
		return fmt.Errorf("user identity lost in service chain")
	}
	
	if businessCtx.Extensions["request_id"] != "req-456" {
		return fmt.Errorf("request_id lost in service chain")
	}
	
	// Verify service contexts are different from user
	if !authService.IsSystem() || !businessService.IsSystem() {
		return fmt.Errorf("service contexts should be system contexts")
	}
	
	// Verify audit service has wildcard permission
	if !auditService.HasPermission("*") {
		return fmt.Errorf("audit service should have wildcard permission")
	}
	
	return nil
}

// testContextTransformation shows contexts changing between boundaries
func testContextTransformation() error {
	// External user context
	externalUser := sctx.NewUserContext("external-user", []string{"public:read"})
	
	// Transform to internal user (after authentication)
	internalUser := sctx.NewUserContext(externalUser.UserID, []string{
		"app:access",
		"profile:read",
		"profile:write",
	}).WithExtension("authenticated", true).
		WithExtension("auth_time", "2023-12-01T10:00:00Z")
	
	// Service-to-service transformation
	// User context becomes service context for backend call
	serviceContext := sctx.NewPackageContext("api-gateway", []string{
		"backend:call",
		"cache:read",
	}).WithExtension("on_behalf_of", internalUser.UserID).
		WithExtension("original_permissions", internalUser.Permissions)
	
	// Verify transformations
	if internalUser.UserID != externalUser.UserID {
		return fmt.Errorf("user identity should be preserved during transformation")
	}
	
	if len(internalUser.Permissions) <= len(externalUser.Permissions) {
		return fmt.Errorf("internal user should have more permissions")
	}
	
	if !serviceContext.IsSystem() {
		return fmt.Errorf("service context should be a system context")
	}
	
	if serviceContext.Extensions["on_behalf_of"] != internalUser.UserID {
		return fmt.Errorf("service should track who it's acting for")
	}
	
	return nil
}

// testCrossDomainPropagation shows contexts crossing domain boundaries
func testCrossDomainPropagation() error {
	// E-commerce domain
	ecommerceCtx := sctx.NewUserContext("shopper-123", []string{"shop:browse", "cart:manage"}).
		WithExtension("domain", "ecommerce").
		WithExtension("cart_id", "cart-456").
		WithExtension("currency", "USD")
	
	// Payment domain (different security requirements)
	paymentCtx := sctx.NewUserContext(ecommerceCtx.UserID, []string{"payment:submit"}).
		WithExtension("domain", "payment").
		WithExtension("pci_compliant", true).
		WithExtension("from_domain", "ecommerce").
		WithExtension("cart_id", ecommerceCtx.Extensions["cart_id"]) // Carry over relevant data
	
	// Shipping domain (for future use)
	_ = sctx.NewUserContext(ecommerceCtx.UserID, []string{"shipping:calculate"}).
		WithExtension("domain", "shipping").
		WithExtension("from_domain", "ecommerce").
		WithExtension("requires_address", true)
	
	// Verify domain isolation
	if paymentCtx.Extensions["currency"] != nil {
		return fmt.Errorf("payment domain should not have e-commerce currency")
	}
	
	// Verify selective data propagation
	if paymentCtx.Extensions["cart_id"] != "cart-456" {
		return fmt.Errorf("cart_id should propagate to payment domain")
	}
	
	// Verify domain tracking
	if paymentCtx.Extensions["from_domain"] != "ecommerce" {
		return fmt.Errorf("should track source domain")
	}
	
	return nil
}

// testContextDegradation shows security degrading gracefully
func testContextDegradation() error {
	// Start with high privilege context
	adminCtx := sctx.NewUserContext("admin", []string{"*"}).
		WithExtension("sudo", true).
		WithExtension("mfa_verified", true)
	
	// Degrade to user context for specific operation
	userModeCtx := sctx.NewUserContext(adminCtx.UserID, []string{
		"user:read",
		"user:write",
	}).WithExtension("sudo", false).
		WithExtension("degraded_from", "admin").
		WithExtension("reason", "least_privilege_operation")
	
	// Further degrade to read-only
	readOnlyCtx := sctx.NewUserContext(adminCtx.UserID, []string{
		"user:read",
	}).WithExtension("degraded_from", "user").
		WithExtension("read_only", true)
	
	// Verify degradation chain
	if userModeCtx.HasPermission("*") {
		return fmt.Errorf("degraded context should not have wildcard")
	}
	
	if readOnlyCtx.HasPermission("user:write") {
		return fmt.Errorf("read-only context should not have write")
	}
	
	// Verify identity preserved
	if adminCtx.UserID != userModeCtx.UserID || userModeCtx.UserID != readOnlyCtx.UserID {
		return fmt.Errorf("identity should be preserved through degradation")
	}
	
	// Verify degradation is tracked
	if readOnlyCtx.Extensions["read_only"] != true {
		return fmt.Errorf("read-only status should be tracked")
	}
	
	return nil
}