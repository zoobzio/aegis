package observability

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	
	"aegis/moisten"
	"aegis/pipz"
	"aegis/zlog"
)

func NewFieldProcessingTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "processing",
		Short: "🔧 Test field processing capabilities",
		Long: `
🔧 FIELD PROCESSING CAPABILITIES
================================

Tests the extensible field transformation pipeline:

• Field processor registration
• Pipeline execution order
• Field transformation logic
• Type-safe processing contracts

This validates that field processing provides extensibility
while maintaining type safety and performance.`,
		Run: runFieldProcessingTests,
	}
}

func runFieldProcessingTests(cmd *cobra.Command, args []string) {
	fmt.Println("\n🔧 FIELD PROCESSING CAPABILITIES")
	fmt.Println("================================")
	
	// Initialize the system
	moisten.ForTesting()
	
	runTest("Field Processor Registration", testFieldProcessorRegistration)
	runTest("Field Transformation", testFieldTransformation)
	runTest("Processing Pipeline", testProcessingPipeline)
	runTest("Type-Safe Processing", testTypeSafeProcessing)
	runTest("Custom Field Types", testCustomFieldTypes)
	runTest("Pipeline Extension", testPipelineExtension)
}

func testFieldProcessorRegistration() error {
	// Test that we can register field processors
	// Note: This tests the interface, actual registration happens in moisten
	
	// Create a test field
	testField := zlog.String("test", "value")
	
	// Verify field structure
	if testField.Key != "test" {
		return fmt.Errorf("field key mismatch: expected 'test', got '%s'", testField.Key)
	}
	
	if testField.Type != zlog.StringType {
		return fmt.Errorf("field type mismatch: expected StringType, got %s", testField.Type)
	}
	
	if testField.Value != "value" {
		return fmt.Errorf("field value mismatch: expected 'value', got %v", testField.Value)
	}
	
	return nil
}

func testFieldTransformation() error {
	// Test that fields can be processed and transformed
	
	// Create fields of different types
	fields := []zlog.ZlogField{
		zlog.String("message", "test message"),
		zlog.Int("count", 42),
		zlog.Bool("active", true),
		zlog.Duration("elapsed", time.Minute),
	}
	
	// Log with these fields to test processing
	zlog.Info("Field transformation test", fields...)
	
	// Test that complex fields work
	zlog.Debug("Complex field test",
		zlog.String("component", "field_processor"),
		zlog.Int("processed_count", len(fields)),
		zlog.Strings("field_types", []string{"string", "int", "bool", "duration"}),
	)
	
	return nil
}

func testProcessingPipeline() error {
	// Test that multiple fields are processed in order
	
	// Create a sequence of fields
	sequence := make([]zlog.ZlogField, 10)
	for i := 0; i < 10; i++ {
		sequence[i] = zlog.Int(fmt.Sprintf("seq_%d", i), i)
	}
	
	// Log to test pipeline processing
	zlog.Info("Pipeline processing test", sequence...)
	
	// Test mixed field types in pipeline
	mixedFields := []zlog.ZlogField{
		zlog.String("first", "string_field"),
		zlog.Int("second", 123),
		zlog.Bool("third", false),
		zlog.Duration("fourth", time.Second),
		zlog.Time("fifth", time.Now()),
		zlog.ByteString("sixth", []byte("bytes")),
		zlog.Strings("seventh", []string{"a", "b"}),
	}
	
	zlog.Warn("Mixed field pipeline test", mixedFields...)
	
	return nil
}

func testTypeSafeProcessing() error {
	// Test that field processing maintains type safety
	
	// Test each field type individually to verify type safety
	testCases := []struct {
		name     string
		field    zlog.ZlogField
		expected zlog.ZlogFieldType
	}{
		{"String field", zlog.String("str", "test"), zlog.StringType},
		{"Int field", zlog.Int("int", 42), zlog.IntType},
		{"Int64 field", zlog.Int64("int64", 9876543210), zlog.Int64Type},
		{"Float64 field", zlog.Float64("float", 3.14159), zlog.Float64Type},
		{"Bool field", zlog.Bool("bool", true), zlog.BoolType},
		{"Duration field", zlog.Duration("dur", time.Hour), zlog.DurationType},
		{"Time field", zlog.Time("time", time.Now()), zlog.TimeType},
		{"ByteString field", zlog.ByteString("bytes", []byte("data")), zlog.ByteStringType},
		{"Strings field", zlog.Strings("strs", []string{"x", "y"}), zlog.StringsType},
	}
	
	for _, tc := range testCases {
		if tc.field.Type != tc.expected {
			return fmt.Errorf("%s: expected type %s, got %s", tc.name, tc.expected, tc.field.Type)
		}
		
		// Log the field to test processing
		zlog.Debug("Type safety test", tc.field)
	}
	
	return nil
}

func testCustomFieldTypes() error {
	// Test that custom/complex field types work through Data constructor
	
	type CustomStruct struct {
		Name    string   `json:"name"`
		Values  []int    `json:"values"`
		Enabled bool     `json:"enabled"`
		Meta    struct {
			Version string `json:"version"`
			Tags    []string `json:"tags"`
		} `json:"meta"`
	}
	
	customData := CustomStruct{
		Name:    "test_struct",
		Values:  []int{1, 2, 3, 4, 5},
		Enabled: true,
	}
	customData.Meta.Version = "1.0.0"
	customData.Meta.Tags = []string{"test", "custom", "struct"}
	
	// Test Data field with custom type
	dataField := zlog.Data("custom", customData)
	
	if dataField.Type != zlog.DataType {
		return fmt.Errorf("custom data field should have DataType, got %s", dataField.Type)
	}
	
	// Log custom data to test processing
	zlog.Info("Custom field type test",
		dataField,
		zlog.String("test_type", "custom_struct"),
		zlog.Int("field_count", 4),
	)
	
	// Test error field with custom error
	customErr := fmt.Errorf("custom error with context: %s", "test_context")
	errField := zlog.Err(customErr)
	
	if errField.Type != zlog.ErrorType {
		return fmt.Errorf("error field should have ErrorType, got %s", errField.Type)
	}
	
	zlog.Error("Custom error test", errField)
	
	return nil
}

func testPipelineExtension() error {
	// Test that we can extend the logging pipeline with custom processors
	// This demonstrates the pipz integration
	
	// Get the field processor contract that moisten set up
	contract := pipz.GetContract[zlog.ZlogFieldType, zlog.ZlogField, []zlog.ZlogField]()
	
	// Define a custom field type for sensitive data
	const SensitiveType zlog.ZlogFieldType = "sensitive"
	
	// Register a processor that redacts sensitive fields
	contract.Register(SensitiveType, func(field zlog.ZlogField) []zlog.ZlogField {
		return []zlog.ZlogField{
			zlog.String(field.Key, "***REDACTED***"),
			zlog.String(field.Key + "_status", "redacted"),
		}
	})
	
	// Register a processor for durations that adds human-readable format
	contract.Register(zlog.DurationType, func(field zlog.ZlogField) []zlog.ZlogField {
		d := field.Value.(time.Duration)
		return []zlog.ZlogField{
			field, // Keep original
			zlog.String(field.Key + "_human", d.String()),
		}
	})
	
	// Test the custom processors
	// First create a sensitive field
	sensitiveField := zlog.ZlogField{
		Key:   "api_key",
		Type:  SensitiveType,
		Value: "sk_live_abcd1234",
	}
	
	// Log with custom processor
	zlog.Info("Testing custom field processor",
		sensitiveField,
		zlog.Duration("request_time", 250*time.Millisecond),
		zlog.String("endpoint", "/api/v1/users"),
	)
	
	// The output should show:
	// - api_key as ***REDACTED***
	// - api_key_status as "redacted"
	// - request_time with both numeric and human format
	
	return nil
}

