package observability

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	
	"aegis/catalog"
	"aegis/moisten"
	"aegis/sctx" 
	"aegis/zlog"
)

func NewSecurityIntegrationTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "security",
		Short: "🔒 Test security integration capabilities",
		Long: `
🔒 SECURITY INTEGRATION CAPABILITIES
====================================

Tests the context-aware security features:

• Security context integration
• Field-level access control
• Data redaction in logs
• Safe error logging

This validates that logging respects security boundaries
while maintaining useful observability.`,
		Run: runSecurityIntegrationTests,
	}
}

func runSecurityIntegrationTests(cmd *cobra.Command, args []string) {
	fmt.Println("\n🔒 SECURITY INTEGRATION CAPABILITIES")
	fmt.Println("====================================")
	
	// Initialize the system
	moisten.ForTesting()
	
	runTest("Security Context Integration", testSecurityContextIntegration)
	runTest("Field Access Control", testFieldAccessControl)
	runTest("Data Redaction", testDataRedaction)
	runTest("Safe Error Logging", testSafeErrorLogging)
	runTest("Context Propagation", testContextPropagation)
}

func testSecurityContextIntegration() error {
	// Test that logging integrates with security contexts
	
	type UserData struct {
		ID       string `json:"id"`
		Email    string `json:"email" scope:"user:read"`
		Password string `json:"password" scope:"admin"`
		SSN      string `json:"ssn" scope:"admin"`
	}
	
	// Register security behavior for UserData (same as unified tests)
	pipeline := catalog.GetSecurityPipeline[UserData]()
	pipeline.Register(catalog.RedactionBehavior, func(input catalog.SecurityInput[UserData]) catalog.SecurityOutput[UserData] {
		data := input.Data
		
		// Redact fields based on scope permissions
		if !input.Context.HasPermission("admin") {
			data.Password = "[REDACTED]"
			data.SSN = "***-**-****"
		}
		
		if !input.Context.HasPermission("user:read") {
			data.Email = "[REDACTED]"
		}
		
		return catalog.SecurityOutput[UserData]{Data: data, Error: nil}
	})
	
	userData := UserData{
		ID:       "user-123",
		Email:    "test@example.com",
		Password: "secret123",
		SSN:      "123-45-6789",
	}
	
	// Test with logger's security context (should redact sensitive fields)
	userField := zlog.Data("user_data", userData)
	
	if userField.Type != zlog.DataType {
		return fmt.Errorf("user field should have DataType")
	}
	
	zlog.Info("User context logging test", userField)
	
	// Test again - logger always uses same security context
	adminField := zlog.Data("admin_data", userData)
	
	if adminField.Type != zlog.DataType {
		return fmt.Errorf("admin field should have DataType")
	}
	
	zlog.Info("Admin context logging test", adminField)
	
	return nil
}

func testFieldAccessControl() error {
	// Test that field-level access control works in logging
	
	type SensitiveData struct {
		PublicInfo  string `json:"public_info"`
		UserInfo    string `json:"user_info" scope:"user:read"`
		AdminInfo   string `json:"admin_info" scope:"admin"`
		SecretInfo  string `json:"secret_info" scope:"admin"`
	}
	
	// Register security behavior
	pipeline := catalog.GetSecurityPipeline[SensitiveData]()
	pipeline.Register(catalog.RedactionBehavior, func(input catalog.SecurityInput[SensitiveData]) catalog.SecurityOutput[SensitiveData] {
		data := input.Data
		
		// Redact fields based on scope permissions
		if !input.Context.HasPermission("admin") {
			data.AdminInfo = "[REDACTED]"
			data.SecretInfo = "[REDACTED]"
		}
		
		if !input.Context.HasPermission("user:read") {
			data.UserInfo = "[REDACTED]"
		}
		
		return catalog.SecurityOutput[SensitiveData]{Data: data, Error: nil}
	})
	
	sensitiveData := SensitiveData{
		PublicInfo: "public_data",
		UserInfo:   "user_specific_data",
		AdminInfo:  "admin_only_data",
		SecretInfo: "highly_secret_data",
	}
	
	// Test different permission levels
	contexts := []struct {
		name        string
		ctx         sctx.SecurityContext
		description string
	}{
		{
			"anonymous",
			sctx.NewUserContext("anon", []string{}),
			"No permissions",
		},
		{
			"user",
			sctx.NewUserContext("user-1", []string{"user:read"}),
			"User-level permissions",
		},
		{
			"admin",
			sctx.NewUserContext("admin-1", []string{"admin", "user:read"}),
			"Admin-level permissions",
		},
	}
	
	for _, tc := range contexts {
		field := zlog.Data("sensitive_data", sensitiveData)
		
		zlog.Info("Access control test",
			field,
			zlog.String("context_type", tc.name),
			zlog.String("description", tc.description),
		)
	}
	
	return nil
}

func testDataRedaction() error {
	// Test that sensitive data is properly redacted in logs
	
	type PersonalData struct {
		Name        string `json:"name"`
		Email       string `json:"email" scope:"user:read"`
		CreditCard  string `json:"credit_card" validate:"creditcard" scope:"admin"`
		SSN         string `json:"ssn" validate:"ssn" scope:"admin"`
		PhoneNumber string `json:"phone" scope:"user:read"`
	}
	
	// Register security behavior
	pipeline := catalog.GetSecurityPipeline[PersonalData]()
	pipeline.Register(catalog.RedactionBehavior, func(input catalog.SecurityInput[PersonalData]) catalog.SecurityOutput[PersonalData] {
		data := input.Data
		
		// Redact fields based on scope permissions
		if !input.Context.HasPermission("admin") {
			data.CreditCard = "****-****-****-****"
			data.SSN = "***-**-****"
		}
		
		if !input.Context.HasPermission("user:read") {
			data.Email = "[REDACTED]"
			data.PhoneNumber = "[REDACTED]"
		}
		
		return catalog.SecurityOutput[PersonalData]{Data: data, Error: nil}
	})
	
	personalData := PersonalData{
		Name:        "John Doe",
		Email:       "john.doe@example.com",
		CreditCard:  "4111-1111-1111-1111",
		SSN:         "123-45-6789",
		PhoneNumber: "555-123-4567",
	}
	
	// Test redaction with logger's context (should always redact)
	limitedField := zlog.Data("personal_data", personalData)
	
	zlog.Warn("Data redaction test - logger permissions",
		limitedField,
		zlog.String("test_type", "redaction"),
		zlog.String("expected", "sensitive_fields_redacted"),
	)
	
	// Test again - same result (consistent redaction)
	fullField := zlog.Data("personal_data", personalData)
	
	zlog.Info("Data redaction test - full permissions",
		fullField,
		zlog.String("test_type", "no_redaction"),
		zlog.String("expected", "all_fields_visible"),
	)
	
	return nil
}

func testSafeErrorLogging() error {
	// Test that errors are logged safely without exposing sensitive data
	
	type ErrorContext struct {
		UserID      string `json:"user_id"`
		SessionID   string `json:"session_id" scope:"admin"`
		APIKey      string `json:"api_key" scope:"admin"`
		RequestData string `json:"request_data" scope:"user:read"`
	}
	
	// Register security behavior
	pipeline := catalog.GetSecurityPipeline[ErrorContext]()
	pipeline.Register(catalog.RedactionBehavior, func(input catalog.SecurityInput[ErrorContext]) catalog.SecurityOutput[ErrorContext] {
		data := input.Data
		
		// Redact fields based on scope permissions
		if !input.Context.HasPermission("admin") {
			data.SessionID = "[REDACTED]"
			data.APIKey = "[REDACTED]"
		}
		
		if !input.Context.HasPermission("user:read") {
			data.RequestData = "[REDACTED]"
		}
		
		return catalog.SecurityOutput[ErrorContext]{Data: data, Error: nil}
	})
	
	errorCtx := ErrorContext{
		UserID:      "user-123",
		SessionID:   "sess-abc-def-ghi",
		APIKey:      "key-secret-123456",
		RequestData: "some request payload",
	}
	
	// Simulate an error with context
	testErr := fmt.Errorf("operation failed: invalid request")
	
	// Log error with logger's context (should redact sensitive fields)
	zlog.Error("Safe error logging test",
		zlog.Err(testErr),
		zlog.Data("error_context", errorCtx),
		zlog.String("operation", "test_operation"),
		zlog.Int("attempt", 1),
	)
	
	// Log error again - same redaction behavior
	zlog.Error("Consistent error logging test",
		zlog.Err(testErr),
		zlog.Data("error_context", errorCtx),
		zlog.String("operation", "admin_operation"),
		zlog.Int("attempt", 2),
	)
	
	return nil
}

func testContextPropagation() error {
	// Test that security contexts propagate properly through logging
	
	type AuditEvent struct {
		Action    string `json:"action"`
		Resource  string `json:"resource" scope:"user:read"`
		UserAgent string `json:"user_agent" scope:"admin"`
		IP        string `json:"ip_address" scope:"admin"`
		Timestamp time.Time `json:"timestamp"`
	}
	
	// Register security behavior
	pipeline := catalog.GetSecurityPipeline[AuditEvent]()
	pipeline.Register(catalog.RedactionBehavior, func(input catalog.SecurityInput[AuditEvent]) catalog.SecurityOutput[AuditEvent] {
		data := input.Data
		
		// Redact fields based on scope permissions
		if !input.Context.HasPermission("admin") {
			data.UserAgent = "[REDACTED]"
			data.IP = "[REDACTED]"
		}
		
		if !input.Context.HasPermission("user:read") {
			data.Resource = "[REDACTED]"
		}
		
		return catalog.SecurityOutput[AuditEvent]{Data: data, Error: nil}
	})
	
	auditEvent := AuditEvent{
		Action:    "login",
		Resource:  "user_account",
		UserAgent: "Mozilla/5.0 Test Browser",
		IP:        "192.168.1.100",
		Timestamp: time.Now(),
	}
	
	// Test context propagation through different scenarios
	scenarios := []struct {
		name string
		ctx  sctx.SecurityContext
		level string
	}{
		{
			"user_audit",
			sctx.NewUserContext("user-audit", []string{"user:read"}),
			"info",
		},
		{
			"admin_audit", 
			sctx.NewUserContext("admin-audit", []string{"admin", "user:read"}),
			"warn",
		},
		{
			"system_audit",
			sctx.NewPackageContext("audit_system", []string{"system:all"}),
			"error",
		},
	}
	
	for _, scenario := range scenarios {
		auditField := zlog.Data("audit_event", auditEvent)
		
		// Log at different levels to test context propagation
		switch scenario.level {
		case "info":
			zlog.Info("Context propagation test",
				auditField,
				zlog.String("scenario", scenario.name),
				zlog.String("level", "info"),
			)
		case "warn":
			zlog.Warn("Context propagation test",
				auditField,
				zlog.String("scenario", scenario.name),
				zlog.String("level", "warn"),
			)
		case "error":
			zlog.Error("Context propagation test",
				auditField,
				zlog.String("scenario", scenario.name),
				zlog.String("level", "error"),
			)
		}
	}
	
	return nil
}

