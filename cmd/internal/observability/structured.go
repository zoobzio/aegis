package observability

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	
	"aegis/moisten"
	"aegis/zlog"
)

func NewStructuredLoggingTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "structured",
		Short: "📝 Test structured logging capabilities",
		Long: `
📝 STRUCTURED LOGGING CAPABILITIES
==================================

Tests the core structured logging functionality:

• Type-safe field construction
• Message formatting and output
• Field type validation
• Data serialization with security context

This validates that logging provides compile-time safety
while maintaining runtime flexibility.`,
		Run: runStructuredLoggingTests,
	}
}

func runStructuredLoggingTests(cmd *cobra.Command, args []string) {
	fmt.Println("\n📝 STRUCTURED LOGGING CAPABILITIES")
	fmt.Println("==================================")
	
	// Initialize the system
	moisten.ForTesting()
	
	runTest("Field Construction", testFieldConstruction)
	runTest("Type Safety", testTypeSafety)
	runTest("Message Logging", testMessageLogging)
	runTest("Security Context Integration", testStructuredSecurityIntegration)
	runTest("Complex Data Handling", testComplexDataHandling)
}

func testFieldConstruction() error {
	// Test all field constructor functions
	fields := []zlog.ZlogField{
		zlog.String("name", "test"),
		zlog.Int("count", 42),
		zlog.Int64("id", 123456789),
		zlog.Float64("ratio", 3.14),
		zlog.Bool("active", true),
		zlog.Duration("elapsed", time.Minute),
		zlog.Time("timestamp", time.Now()),
		zlog.ByteString("data", []byte("hello")),
		zlog.Strings("tags", []string{"a", "b", "c"}),
	}
	
	// Verify field construction
	if len(fields) != 9 {
		return fmt.Errorf("expected 9 fields, got %d", len(fields))
	}
	
	// Verify field types
	expectedTypes := []zlog.ZlogFieldType{
		zlog.StringType,
		zlog.IntType,
		zlog.Int64Type,
		zlog.Float64Type,
		zlog.BoolType,
		zlog.DurationType,
		zlog.TimeType,
		zlog.ByteStringType,
		zlog.StringsType,
	}
	
	for i, field := range fields {
		if field.Type != expectedTypes[i] {
			return fmt.Errorf("field %d: expected type %s, got %s", i, expectedTypes[i], field.Type)
		}
	}
	
	return nil
}

func testTypeSafety() error {
	// Test that field constructors enforce type safety at compile time
	
	// String field
	stringField := zlog.String("test", "value")
	if stringField.Value != "value" {
		return fmt.Errorf("string value mismatch")
	}
	
	// Int field  
	intField := zlog.Int("test", 42)
	if intField.Value != 42 {
		return fmt.Errorf("int value mismatch")
	}
	
	// Bool field
	boolField := zlog.Bool("test", true)
	if boolField.Value != true {
		return fmt.Errorf("bool value mismatch")
	}
	
	// Error field
	testErr := fmt.Errorf("test error")
	errField := zlog.Err(testErr)
	if errField.Value != testErr {
		return fmt.Errorf("error value mismatch")
	}
	if errField.Key != "error" {
		return fmt.Errorf("error field key should be 'error', got '%s'", errField.Key)
	}
	
	return nil
}

func testMessageLogging() error {
	// Test basic message logging functions
	// Note: We can't easily capture output in tests, so we test that functions don't panic
	
	// Test all log levels
	zlog.Debug("Debug message", zlog.String("level", "debug"))
	zlog.Info("Info message", zlog.String("level", "info"))
	zlog.Warn("Warn message", zlog.String("level", "warn"))
	zlog.Error("Error message", zlog.String("level", "error"))
	
	// Test with multiple fields
	zlog.Info("Multi-field message",
		zlog.String("component", "test"),
		zlog.Int("iteration", 1),
		zlog.Bool("success", true),
	)
	
	// Test with no fields
	zlog.Info("Simple message")
	
	return nil
}

func testStructuredSecurityIntegration() error {
	// Test that Data fields integrate with security context
	
	type TestData struct {
		ID       string `json:"id"`
		Email    string `json:"email" scope:"user:read"`
		Password string `json:"password" scope:"admin"`
	}
	
	data := TestData{
		ID:       "user-123",
		Email:    "test@example.com",
		Password: "secret",
	}
	
	// Test with logger's own security context (should redact sensitive fields)
	userField := zlog.Data("user_data", data)
	
	if userField.Type != zlog.DataType {
		return fmt.Errorf("expected DataType, got %s", userField.Type)
	}
	
	// Test again - logger always uses same security context
	adminField := zlog.Data("admin_data", data)
	
	if adminField.Type != zlog.DataType {
		return fmt.Errorf("expected DataType, got %s", adminField.Type)
	}
	
	// Log both to verify no panics
	zlog.Info("Security context test",
		userField,
		adminField,
	)
	
	return nil
}

func testComplexDataHandling() error {
	// Test logging with complex nested data
	
	type NestedData struct {
		Users []string          `json:"users"`
		Meta  map[string]string `json:"metadata"`
		Count int               `json:"count"`
	}
	
	complexData := NestedData{
		Users: []string{"alice", "bob", "charlie"},
		Meta:  map[string]string{"env": "test", "version": "1.0"},
		Count: 3,
	}
	
	// Test complex data field
	complexField := zlog.Data("complex", complexData)
	if complexField.Type != zlog.DataType {
		return fmt.Errorf("complex data field should have DataType")
	}
	
	// Test logging complex structures
	zlog.Info("Complex data test",
		complexField,
		zlog.Strings("simple_array", []string{"x", "y", "z"}),
		zlog.String("description", "Testing complex nested structures"),
	)
	
	return nil
}

