package data

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	
	"aegis/catalog"
	"aegis/moisten"
)

// runMetadataTests runs all metadata extraction tests
func runMetadataTests(cmd *cobra.Command, args []string) {
	// Initialize framework with data behaviors
	// This ensures behaviors are registered BEFORE any metadata extraction
	moisten.ForTesting(RegisterDataBehaviors)
	
	runTest("Type Registration", TestTypeRegistration)
	runTest("Field Metadata", TestFieldMetadata)
	runTest("Tag Processing", TestTagProcessing)
	runTest("Type Name Generation", TestTypeNameGeneration)
	runTest("Metadata Caching", TestMetadataCaching)
	runTest("Adapter Registration", TestAdapterRegistration)
}

// Test type for metadata extraction
type MetadataTestUser struct {
	ID       string `json:"id"`
	Email    string `json:"email" scope:"user:read"`
	Password string `json:"password" scope:"admin"`
	SSN      string `json:"ssn" validate:"ssn" scope:"admin"`
	Salary   int    `json:"salary" scope:"hr:read"`
}

// TestTypeRegistration verifies that types are registered with catalog
func TestTypeRegistration() error {
	// Register our test type
	catalog.RegisterType[MetadataTestUser]()
	
	// Verify type is registered
	metadata := catalog.Select[MetadataTestUser]()
	if metadata.TypeName == "" {
		return fmt.Errorf("type name not extracted")
	}
	
	if !strings.Contains(metadata.TypeName, "MetadataTestUser") {
		return fmt.Errorf("expected MetadataTestUser in type name, got %s", metadata.TypeName)
	}
	
	return nil
}

// TestFieldMetadata verifies field extraction and metadata population
func TestFieldMetadata() error {
	metadata := catalog.Select[MetadataTestUser]()
	
	// Should have extracted all fields
	expectedFields := []string{"ID", "Email", "Password", "SSN", "Salary"}
	if len(metadata.Fields) != len(expectedFields) {
		return fmt.Errorf("expected %d fields, got %d", len(expectedFields), len(metadata.Fields))
	}
	
	// Check field names are extracted
	for _, expectedField := range expectedFields {
		found := false
		for _, field := range metadata.Fields {
			if field.Name == expectedField {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("field %s not found in metadata", expectedField)
		}
	}
	
	return nil
}

// TestTagProcessing verifies struct tag parsing and storage
func TestTagProcessing() error {
	// This test demonstrates emergent behavior:
	// Because we registered behaviors in RegisterDataBehaviors() that care about
	// "scope" and "validate" tags, those tags are automatically extracted
	
	metadata := catalog.Select[MetadataTestUser]()
	
	// Find SSN field and check tags
	var ssnField *catalog.FieldMetadata
	for _, field := range metadata.Fields {
		if field.Name == "SSN" {
			ssnField = &field
			break
		}
	}
	
	if ssnField == nil {
		return fmt.Errorf("SSN field not found")
	}
	
	// Check scope tag
	if ssnField.Tags["scope"] != "admin" {
		return fmt.Errorf("expected scope=admin, got %s", ssnField.Tags["scope"])
	}
	
	// Check validate tag
	if ssnField.Tags["validate"] != "ssn" {
		return fmt.Errorf("expected validate=ssn, got %s", ssnField.Tags["validate"])
	}
	
	return nil
}

// TestTypeNameGeneration verifies type name generation for different types
func TestTypeNameGeneration() error {
	// Test simple type
	catalog.RegisterType[MetadataTestUser]()
	simpleMetadata := catalog.Select[MetadataTestUser]()
	if !strings.Contains(simpleMetadata.TypeName, "MetadataTestUser") {
		return fmt.Errorf("simple type name generation failed: %s", simpleMetadata.TypeName)
	}
	
	// Test generic type
	type GenericContainer[T any] struct {
		Value T
	}
	catalog.RegisterType[GenericContainer[string]]()
	genericMetadata := catalog.Select[GenericContainer[string]]()
	if !strings.Contains(genericMetadata.TypeName, "GenericContainer") {
		return fmt.Errorf("generic type name generation failed: %s", genericMetadata.TypeName)
	}
	
	// Test pointer type
	catalog.RegisterType[*MetadataTestUser]()
	pointerMetadata := catalog.Select[*MetadataTestUser]()
	if !strings.Contains(pointerMetadata.TypeName, "MetadataTestUser") {
		return fmt.Errorf("pointer type name generation failed: %s", pointerMetadata.TypeName)
	}
	
	return nil
}

// TestMetadataCaching verifies that metadata is cached and reused
func TestMetadataCaching() error {
	// First call should create metadata
	metadata1 := catalog.Select[MetadataTestUser]()
	
	// Second call should return cached metadata
	metadata2 := catalog.Select[MetadataTestUser]()
	
	// They should be identical (same memory address would be ideal, but we check content)
	if metadata1.TypeName != metadata2.TypeName {
		return fmt.Errorf("metadata not cached properly")
	}
	
	if len(metadata1.Fields) != len(metadata2.Fields) {
		return fmt.Errorf("cached metadata has different field count")
	}
	
	return nil
}

// TestAdapterRegistration verifies tag adapter registration
func TestAdapterRegistration() error {
	// Register the custom tag so it will be extracted
	catalog.RegisterTag("custom")
	
	// Define type with custom tag
	type CustomTagType struct {
		Field string `custom:"value" json:"field"`
	}
	
	// Register type to extract metadata
	catalog.RegisterType[CustomTagType]()
	
	// Get metadata to verify tag extraction
	metadata := catalog.Select[CustomTagType]()
	
	// Find the field with custom tag
	for _, field := range metadata.Fields {
		if field.Name == "Field" {
			// Verify custom tag was extracted
			if field.Tags["custom"] != "value" {
				return fmt.Errorf("custom tag not extracted correctly, got: %v", field.Tags)
			}
			// Also verify json tag was extracted (always included)
			if field.Tags["json"] != "field" {
				return fmt.Errorf("json tag not extracted correctly")
			}
			return nil
		}
	}
	
	return fmt.Errorf("field with custom tag not found")
}