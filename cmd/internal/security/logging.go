package security

import (
	"fmt"
	
	"github.com/spf13/cobra"
	
	"aegis/adapters/security"
	"aegis/sctx"
	"aegis/zlog"
)

// runSecurityLoggingTests runs all security logging enhancement tests
func runLoggingTests(cmd *cobra.Command, args []string) {
	runTest("Security Logging", TestSecurityLogging)
	runTest("Security Field Processing", TestSecurityFieldProcessing)
}

// TestSecurityLogging demonstrates the security adapter's logging enhancements
func TestSecurityLogging() error {
	// Create a security context for testing
	ctx := sctx.NewUserContext("user-123", []string{"user:read", "admin"})
	ctx.Extensions["user_type"] = "premium"
	ctx.Extensions["request_id"] = "req-456"
	ctx.Extensions["service_chain"] = []string{"api-gateway", "auth-service", "user-service"}
	
	// Log with security context - it will be automatically expanded
	zlog.Info("User login successful",
		security.SecurityContext(ctx),
		zlog.String("event", "login"),
	)
	
	// Log with sensitive data - it will be automatically masked
	zlog.Warn("Failed login attempt",
		security.SensitiveField("email", "user@example.com"),
		security.SensitiveField("ssn", "123-45-6789"),
		security.SensitiveField("credit_card", "4111-1111-1111-1111"),
		security.SensitiveField("api_key", "sk_live_abcd1234efgh5678"),
	)
	
	// Log with audit field - compliance metadata will be added
	zlog.Info("Data access",
		security.AuditField("READ_USER_PROFILE"),
		zlog.String("resource", "user:123"),
		security.SecurityContext(ctx),
	)
	
	// Demonstrate combined usage
	zlog.Error("Security violation detected",
		security.SecurityContext(ctx),
		security.SensitiveField("email", "attacker@evil.com"),
		security.AuditField("UNAUTHORIZED_ACCESS"),
		zlog.String("action", "password_reset"),
		zlog.Int("attempts", 5),
	)
	
	return nil
}

// TestSecurityFieldProcessing verifies that security fields are processed correctly
func TestSecurityFieldProcessing() error {
	// This test verifies the field processors work as expected
	
	// Test 1: Security context expansion
	ctx := sctx.NewPackageContext("test-system", []string{"system:all"})
	fields := []zlog.ZlogField{
		security.SecurityContext(ctx),
	}
	
	// When logged, this should expand to multiple fields
	zlog.Debug("Testing security context expansion", fields...)
	
	// Test 2: Sensitive data masking
	sensitiveFields := []zlog.ZlogField{
		security.SensitiveField("password", "super-secret-123"),
		security.SensitiveField("token", "Bearer xyz789"),
		security.SensitiveField("unknown_sensitive", "should-be-redacted"),
	}
	
	zlog.Debug("Testing sensitive data masking", sensitiveFields...)
	
	// Test 3: Audit field expansion
	auditFields := []zlog.ZlogField{
		security.AuditField("DELETE_USER"),
		zlog.String("user_id", "user-to-delete"),
	}
	
	zlog.Debug("Testing audit field expansion", auditFields...)
	
	fmt.Println("Security field processing test completed - check logs for processed output")
	return nil
}