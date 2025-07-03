package security

import (
	"fmt"
	"strings"

	"aegis/catalog"
	"aegis/cereal"
	"aegis/sctx"
)

// Test types for convention features
type ConventionTestData struct {
	// Basic fields
	ID   string `json:"id"`
	Name string `json:"name"`
	
	// Tag-driven security
	PublicInfo    string `json:"public_info"`
	UserInfo      string `json:"user_info" scope:"user:read"`
	AdminInfo     string `json:"admin_info" scope:"admin"`
	MultiScope    string `json:"multi_scope" scope:"user:read,admin"`
	
	// Validation-driven security
	Email      string `json:"email" validate:"email"`
	Phone      string `json:"phone" validate:"phone"`
	CreditCard string `json:"credit_card" validate:"credit_card" scope:"payment:process"`
	
	// Custom convention tags
	CustomField string `json:"custom_field" security:"high" compliance:"pci"`
	Classified  string `json:"classified" classification:"top-secret" scope:"security:clearance"`
}

// TestTagDrivenSecurity verifies security based on struct tags
func TestTagDrivenSecurity() error {
	catalog.RegisterType[ConventionTestData]()
	
	// Register custom tags if needed
	catalog.RegisterTag("security")
	catalog.RegisterTag("compliance")
	catalog.RegisterTag("classification")
	
	data := ConventionTestData{
		ID:         "conv-001",
		Name:       "Tag Test",
		PublicInfo: "Everyone can see this",
		UserInfo:   "Only users can see this",
		AdminInfo:  "Only admins can see this",
		MultiScope: "Users or admins can see this",
	}
	
	// Test with no permissions
	noPermsCtx := sctx.NewUserContext("guest-1", []string{})
	noPermsResult, err := cereal.MarshalJSON(data, noPermsCtx)
	if err != nil {
		return fmt.Errorf("no perms marshal failed: %v", err)
	}
	
	var noPermsData ConventionTestData
	noPermsData, _ = cereal.UnmarshalJSON[ConventionTestData](noPermsResult, noPermsCtx)
	
	// Should only see public fields
	if noPermsData.PublicInfo != "Everyone can see this" {
		return fmt.Errorf("public info should be visible to all")
	}
	
	if noPermsData.UserInfo != "[REDACTED]" {
		return fmt.Errorf("user info should be redacted without permissions")
	}
	
	if noPermsData.AdminInfo != "[REDACTED]" {
		return fmt.Errorf("admin info should be redacted without permissions")
	}
	
	// Test with user permissions
	userCtx := sctx.NewUserContext("user-1", []string{"user:read"})
	userResult, _ := cereal.MarshalJSON(data, userCtx)
	
	var userData ConventionTestData
	userData, _ = cereal.UnmarshalJSON[ConventionTestData](userResult, userCtx)
	
	if userData.UserInfo != "Only users can see this" {
		return fmt.Errorf("user should see user info")
	}
	
	if userData.MultiScope != "Users or admins can see this" {
		return fmt.Errorf("user should see multi-scope field")
	}
	
	if userData.AdminInfo != "[REDACTED]" {
		return fmt.Errorf("user should not see admin info")
	}
	
	return nil
}

// TestScopeTags verifies scope tag processing
func TestScopeTags() error {
	data := ConventionTestData{
		ID:         "scope-001",
		MultiScope: "Field with multiple allowed scopes",
		Classified: "Top secret information",
	}
	
	// Test multi-scope access (user:read,admin)
	userCtx := sctx.NewUserContext("user-1", []string{"user:read"})
	userResult, _ := cereal.MarshalJSON(data, userCtx)
	
	var userData ConventionTestData
	userData, _ = cereal.UnmarshalJSON[ConventionTestData](userResult, userCtx)
	
	// User has one of the required scopes, should see field
	if userData.MultiScope != "Field with multiple allowed scopes" {
		return fmt.Errorf("multi-scope should be visible with any matching scope")
	}
	
	// Test classified field with custom scope
	securityCtx := sctx.NewUserContext("agent-1", []string{"security:clearance"})
	securityResult, _ := cereal.MarshalJSON(data, securityCtx)
	
	var securityData ConventionTestData
	securityData, _ = cereal.UnmarshalJSON[ConventionTestData](securityResult, securityCtx)
	
	if securityData.Classified != "Top secret information" {
		return fmt.Errorf("security clearance should see classified info")
	}
	
	// Without clearance, should be redacted
	if userData.Classified != "[REDACTED]" {
		return fmt.Errorf("classified should be redacted without clearance")
	}
	
	return nil
}

// TestValidationTags verifies validation tag integration
func TestValidationTags() error {
	data := ConventionTestData{
		Email:      "test@example.com",
		Phone:      "555-123-4567",
		CreditCard: "4111-1111-1111-1234",
	}
	
	// Without specific permissions, validated fields may be masked
	basicCtx := sctx.NewUserContext("user-1", []string{"user:read"})
	
	// Apply security transformation
	secured := data
	cereal.Secure(&secured)
	
	// Email should be partially masked
	if !strings.Contains(secured.Email, "@example.com") || secured.Email == data.Email {
		return fmt.Errorf("email should be partially masked")
	}
	
	// Phone should be partially masked
	if !strings.HasSuffix(secured.Phone, "4567") || secured.Phone == data.Phone {
		return fmt.Errorf("phone should show only last 4 digits")
	}
	
	// Credit card needs payment:process scope
	paymentCtx := sctx.NewUserContext("payment-1", []string{"payment:process"})
	paymentResult, _ := cereal.MarshalJSON(data, paymentCtx)
	
	var paymentData ConventionTestData
	paymentData, _ = cereal.UnmarshalJSON[ConventionTestData](paymentResult, paymentCtx)
	
	if paymentData.CreditCard != "4111-1111-1111-1234" {
		return fmt.Errorf("payment processor should see full credit card")
	}
	
	// Without payment:process, should be masked
	userResult, _ := cereal.MarshalJSON(data, basicCtx)
	var userData ConventionTestData
	userData, _ = cereal.UnmarshalJSON[ConventionTestData](userResult, basicCtx)
	
	if userData.CreditCard != "****-****-****-1234" {
		return fmt.Errorf("credit card should be masked without payment:process")
	}
	
	return nil
}

// TestCustomConventions verifies custom convention support
func TestCustomConventions() error {
	data := ConventionTestData{
		ID:          "custom-001",
		CustomField: "High security field with PCI compliance",
		Classified:  "Classified with custom tags",
	}
	
	// Custom tags should be extracted by catalog
	metadata := catalog.Select[ConventionTestData]()
	
	// Find custom field
	var customFieldMeta *catalog.FieldMetadata
	for _, field := range metadata.Fields {
		if field.Name == "CustomField" {
			customFieldMeta = &field
			break
		}
	}
	
	if customFieldMeta == nil {
		return fmt.Errorf("custom field metadata not found")
	}
	
	// Verify custom tags were extracted
	if customFieldMeta.Tags["security"] != "high" {
		return fmt.Errorf("security tag not extracted correctly")
	}
	
	if customFieldMeta.Tags["compliance"] != "pci" {
		return fmt.Errorf("compliance tag not extracted correctly")
	}
	
	// Test serialization with custom conventions
	// The actual behavior depends on registered handlers, but tags should be available
	ctx := sctx.NewUserContext("user-1", []string{"user:read"})
	result, err := cereal.MarshalJSON(data, ctx)
	if err != nil {
		return fmt.Errorf("marshal with custom conventions failed: %v", err)
	}
	
	// Basic verification that serialization works
	if len(result) == 0 {
		return fmt.Errorf("serialization produced empty result")
	}
	
	// Custom security adapters could register behaviors based on these tags
	// For example, "security:high" could trigger additional encryption
	// "compliance:pci" could trigger specific masking rules
	
	return nil
}