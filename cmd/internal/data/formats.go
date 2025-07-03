package data

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	
	"aegis/catalog"
	"aegis/cereal"
	"aegis/moisten"
	"aegis/sctx"
)

// runFormatTests runs all serialization format tests
func runFormatTests(cmd *cobra.Command, args []string) {
	moisten.ForTesting()
	
	// Register security behavior for our test type
	registerFormatTestSecurityBehavior()
	
	runTest("JSON Serialization", TestJSONSerialization)
	runTest("YAML Serialization", TestYAMLSerialization)
	runTest("TOML Serialization", TestTOMLSerialization)
	runTest("Format Consistency", TestFormatConsistency)
}

// registerFormatTestSecurityBehavior sets up security behaviors for FormatTestData
func registerFormatTestSecurityBehavior() {
	pipeline := catalog.GetSecurityPipeline[FormatTestData]()
	pipeline.Register(catalog.RedactionBehavior, func(input catalog.SecurityInput[FormatTestData]) catalog.SecurityOutput[FormatTestData] {
		data := input.Data
		
		// Redact fields based on scope permissions
		if !input.Context.HasPermission("admin") {
			data.SecretField = "[REDACTED]"
		}
		
		if !input.Context.HasPermission("finance:read") {
			data.Price = -1
		}
		
		return catalog.SecurityOutput[FormatTestData]{Data: data, Error: nil}
	})
}

// Test type for format consistency
type FormatTestData struct {
	ID          string   `json:"id" yaml:"id" toml:"id"`
	Name        string   `json:"name" yaml:"name" toml:"name"`
	Count       int      `json:"count" yaml:"count" toml:"count"`
	Price       float64  `json:"price" yaml:"price" toml:"price" scope:"finance:read"`
	IsActive    bool     `json:"is_active" yaml:"is_active" toml:"is_active"`
	Tags        []string `json:"tags" yaml:"tags" toml:"tags"`
	SecretField string   `json:"secret" yaml:"secret" toml:"secret" scope:"admin"`
}

// TestJSONSerialization verifies JSON format with security
func TestJSONSerialization() error {
	catalog.RegisterType[FormatTestData]()
	
	data := FormatTestData{
		ID:          "fmt-001",
		Name:        "JSON Test",
		Count:       42,
		Price:       99.99,
		IsActive:    true,
		Tags:        []string{"test", "json", "security"},
		SecretField: "hidden-value",
	}
	
	// Test with limited permissions
	ctx := sctx.NewUserContext("user-1", []string{"user:read"})
	result, err := cereal.MarshalJSON(data, ctx)
	if err != nil {
		return fmt.Errorf("JSON marshal failed: %v", err)
	}
	
	// Verify it's valid JSON
	var jsonCheck map[string]interface{}
	if err := json.Unmarshal(result, &jsonCheck); err != nil {
		return fmt.Errorf("invalid JSON produced: %v", err)
	}
	
	// Verify security was applied
	if jsonCheck["secret"] != "[REDACTED]" {
		return fmt.Errorf("secret field should be redacted")
	}
	
	if jsonCheck["price"] != float64(-1) {
		return fmt.Errorf("price should be -1 without finance:read")
	}
	
	// Test unmarshaling
	var unmarshaled FormatTestData
	unmarshaled, err = cereal.UnmarshalJSON[FormatTestData](result, ctx)
	if err != nil {
		return fmt.Errorf("JSON unmarshal failed: %v", err)
	}
	
	// Verify data integrity for allowed fields
	if unmarshaled.Name != "JSON Test" {
		return fmt.Errorf("name corrupted during JSON round-trip")
	}
	
	if len(unmarshaled.Tags) != 3 {
		return fmt.Errorf("tags corrupted during JSON round-trip")
	}
	
	return nil
}

// TestYAMLSerialization verifies YAML format with security
func TestYAMLSerialization() error {
	data := FormatTestData{
		ID:          "fmt-002",
		Name:        "YAML Test",
		Count:       100,
		Price:       199.99,
		IsActive:    false,
		Tags:        []string{"yaml", "config", "secure"},
		SecretField: "yaml-secret",
	}
	
	// Test with admin permissions
	ctx := sctx.NewUserContext("admin-1", []string{"admin", "finance:read"})
	result, err := cereal.MarshalYAML(data, ctx)
	if err != nil {
		return fmt.Errorf("YAML marshal failed: %v", err)
	}
	
	// Verify it's valid YAML and contains expected content
	yamlStr := string(result)
	if !strings.Contains(yamlStr, "id: fmt-002") {
		return fmt.Errorf("YAML should contain id field")
	}
	
	if !strings.Contains(yamlStr, "yaml-secret") {
		return fmt.Errorf("admin should see secret in YAML")
	}
	
	// Test unmarshaling
	var unmarshaled FormatTestData
	unmarshaled, err = cereal.UnmarshalYAML[FormatTestData](result, ctx)
	if err != nil {
		return fmt.Errorf("YAML unmarshal failed: %v", err)
	}
	
	// Verify full data with admin context
	if unmarshaled.SecretField != "yaml-secret" {
		return fmt.Errorf("admin should unmarshal secret from YAML")
	}
	
	if unmarshaled.Price != 199.99 {
		return fmt.Errorf("price corrupted during YAML round-trip")
	}
	
	// Test with limited permissions
	userCtx := sctx.NewUserContext("user-1", []string{})
	userResult, _ := cereal.MarshalYAML(data, userCtx)
	
	var userUnmarshaled FormatTestData
	userUnmarshaled, _ = cereal.UnmarshalYAML[FormatTestData](userResult, userCtx)
	
	if userUnmarshaled.SecretField != "[REDACTED]" {
		return fmt.Errorf("user should see redacted secret in YAML")
	}
	
	return nil
}

// TestTOMLSerialization verifies TOML format with security
func TestTOMLSerialization() error {
	data := FormatTestData{
		ID:          "fmt-003",
		Name:        "TOML Test",
		Count:       256,
		Price:       49.95,
		IsActive:    true,
		Tags:        []string{"toml", "configuration"},
		SecretField: "toml-classified",
	}
	
	// Test with partial permissions
	ctx := sctx.NewUserContext("user-1", []string{"finance:read"})
	result, err := cereal.MarshalTOML(data, ctx)
	if err != nil {
		return fmt.Errorf("TOML marshal failed: %v", err)
	}
	
	// Verify TOML format
	tomlStr := string(result)
	if !strings.Contains(tomlStr, `id = "fmt-003"`) {
		return fmt.Errorf("TOML should contain quoted id")
	}
	
	// Should see price with finance:read
	if !strings.Contains(tomlStr, "price = 49.95") {
		return fmt.Errorf("should see price with finance:read permission")
	}
	
	// Should not see secret without admin
	if strings.Contains(tomlStr, "toml-classified") {
		return fmt.Errorf("should not see secret without admin permission")
	}
	
	// Test unmarshaling
	var unmarshaled FormatTestData
	unmarshaled, err = cereal.UnmarshalTOML[FormatTestData](result, ctx)
	if err != nil {
		return fmt.Errorf("TOML unmarshal failed: %v", err)
	}
	
	// Verify data integrity
	if unmarshaled.Price != 49.95 {
		return fmt.Errorf("price should be visible with finance:read")
	}
	
	if unmarshaled.SecretField != "[REDACTED]" {
		return fmt.Errorf("secret should be redacted without admin")
	}
	
	if len(unmarshaled.Tags) != 2 {
		return fmt.Errorf("tags corrupted during TOML round-trip")
	}
	
	return nil
}

// TestFormatConsistency verifies all formats produce consistent security
func TestFormatConsistency() error {
	data := FormatTestData{
		ID:          "fmt-consistency",
		Name:        "Consistency Test",
		Count:       777,
		Price:       1234.56,
		IsActive:    true,
		Tags:        []string{"all", "formats", "test"},
		SecretField: "super-secret",
	}
	
	// Use consistent limited context
	ctx := sctx.NewUserContext("user-1", []string{"user:read"})
	
	// Marshal in all formats
	jsonData, err := cereal.MarshalJSON(data, ctx)
	if err != nil {
		return fmt.Errorf("JSON marshal failed: %v", err)
	}
	
	yamlData, err := cereal.MarshalYAML(data, ctx)
	if err != nil {
		return fmt.Errorf("YAML marshal failed: %v", err)
	}
	
	tomlData, err := cereal.MarshalTOML(data, ctx)
	if err != nil {
		return fmt.Errorf("TOML marshal failed: %v", err)
	}
	
	// Unmarshal from all formats
	var fromJSON, fromYAML, fromTOML FormatTestData
	
	fromJSON, err = cereal.UnmarshalJSON[FormatTestData](jsonData, ctx)
	if err != nil {
		return fmt.Errorf("JSON unmarshal failed: %v", err)
	}
	
	fromYAML, err = cereal.UnmarshalYAML[FormatTestData](yamlData, ctx)
	if err != nil {
		return fmt.Errorf("YAML unmarshal failed: %v", err)
	}
	
	fromTOML, err = cereal.UnmarshalTOML[FormatTestData](tomlData, ctx)
	if err != nil {
		return fmt.Errorf("TOML unmarshal failed: %v", err)
	}
	
	// All formats should produce identical security results
	
	// Check redacted fields are consistent
	if fromJSON.SecretField != "[REDACTED]" || fromYAML.SecretField != "[REDACTED]" || fromTOML.SecretField != "[REDACTED]" {
		return fmt.Errorf("secret field redaction inconsistent across formats")
	}
	
	if fromJSON.Price != -1 || fromYAML.Price != -1 || fromTOML.Price != -1 {
		return fmt.Errorf("price redaction inconsistent across formats")
	}
	
	// Check allowed fields are consistent
	if fromJSON.Name != data.Name || fromYAML.Name != data.Name || fromTOML.Name != data.Name {
		return fmt.Errorf("name field inconsistent across formats")
	}
	
	if fromJSON.Count != data.Count || fromYAML.Count != data.Count || fromTOML.Count != data.Count {
		return fmt.Errorf("count field inconsistent across formats")
	}
	
	// Check complex types (arrays) are consistent
	if len(fromJSON.Tags) != len(data.Tags) || len(fromYAML.Tags) != len(data.Tags) || len(fromTOML.Tags) != len(data.Tags) {
		return fmt.Errorf("tags array inconsistent across formats")
	}
	
	return nil
}