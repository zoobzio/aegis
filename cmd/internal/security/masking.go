package security

import (
	"fmt"
	"strings"

	"aegis/catalog"
	"aegis/cereal"
	"aegis/sctx"
)

// Test types for masking features
type MaskingTestData struct {
	// String fields
	Name        string `json:"name"`
	Description string `json:"description" scope:"admin"`
	Notes       string `json:"notes" scope:"internal"`
	
	// Numeric fields
	Count    int     `json:"count"`
	Price    float64 `json:"price" scope:"finance:read"`
	Quantity int64   `json:"quantity" scope:"inventory:read"`
	
	// Sensitive fields with validation
	SSN         string `json:"ssn" validate:"ssn" scope:"admin"`
	CreditCard  string `json:"credit_card" validate:"credit_card" scope:"payment:read"`
	Email       string `json:"email" validate:"email"`
	Phone       string `json:"phone" validate:"phone"`
	Password    string `json:"password" validate:"password" scope:"never"`
	APIKey      string `json:"api_key" validate:"api_key" scope:"admin"`
}

// TestStringRedaction verifies string field redaction
func TestStringRedaction() error {
	catalog.RegisterType[MaskingTestData]()
	
	data := MaskingTestData{
		Name:        "Public Name",
		Description: "Secret Description",
		Notes:       "Internal Notes Only",
	}
	
	// Test with no permissions
	ctx := sctx.NewUserContext("user-1", []string{})
	result, err := cereal.MarshalJSON(data, ctx)
	if err != nil {
		return fmt.Errorf("marshal failed: %v", err)
	}
	
	// Unmarshal to verify
	var unmarshaled MaskingTestData
	unmarshaled, err = cereal.UnmarshalJSON[MaskingTestData](result, ctx)
	if err != nil {
		return fmt.Errorf("unmarshal failed: %v", err)
	}
	
	// Public field should be visible
	if unmarshaled.Name != "Public Name" {
		return fmt.Errorf("public field should not be redacted")
	}
	
	// Scoped fields should be redacted
	if unmarshaled.Description != "[REDACTED]" {
		return fmt.Errorf("description should be redacted without admin scope")
	}
	
	if unmarshaled.Notes != "[REDACTED]" {
		return fmt.Errorf("notes should be redacted without internal scope")
	}
	
	return nil
}

// TestNumericRedaction verifies numeric field redaction
func TestNumericRedaction() error {
	data := MaskingTestData{
		Count:    42,
		Price:    99.99,
		Quantity: 1000,
	}
	
	// Test with limited permissions
	ctx := sctx.NewUserContext("user-1", []string{"inventory:read"})
	result, err := cereal.MarshalJSON(data, ctx)
	if err != nil {
		return fmt.Errorf("marshal failed: %v", err)
	}
	
	var unmarshaled MaskingTestData
	unmarshaled, err = cereal.UnmarshalJSON[MaskingTestData](result, ctx)
	if err != nil {
		return fmt.Errorf("unmarshal failed: %v", err)
	}
	
	// Unscoped numeric field should be visible
	if unmarshaled.Count != 42 {
		return fmt.Errorf("unscoped count should be visible")
	}
	
	// Scoped price should be redacted to -1
	if unmarshaled.Price != -1 {
		return fmt.Errorf("price should be -1 without finance:read, got %f", unmarshaled.Price)
	}
	
	// Quantity should be visible with inventory:read
	if unmarshaled.Quantity != 1000 {
		return fmt.Errorf("quantity should be visible with inventory:read")
	}
	
	return nil
}

// TestSSNMasking verifies SSN masking shows only last 4 digits
func TestSSNMasking() error {
	data := MaskingTestData{
		SSN: "123-45-6789",
	}
	
	// Without admin permission, SSN should be masked
	ctx := sctx.NewUserContext("user-1", []string{"user:read"})
	result, err := cereal.MarshalJSON(data, ctx)
	if err != nil {
		return fmt.Errorf("marshal failed: %v", err)
	}
	
	var unmarshaled MaskingTestData
	unmarshaled, err = cereal.UnmarshalJSON[MaskingTestData](result, ctx)
	if err != nil {
		return fmt.Errorf("unmarshal failed: %v", err)
	}
	
	// Should show only last 4 digits
	if unmarshaled.SSN != "***-**-6789" {
		return fmt.Errorf("SSN should be masked to show only last 4, got: %s", unmarshaled.SSN)
	}
	
	// With admin permission, full SSN should be visible
	adminCtx := sctx.NewUserContext("admin-1", []string{"admin"})
	adminResult, _ := cereal.MarshalJSON(data, adminCtx)
	
	var adminUnmarshaled MaskingTestData
	adminUnmarshaled, _ = cereal.UnmarshalJSON[MaskingTestData](adminResult, adminCtx)
	
	if adminUnmarshaled.SSN != "123-45-6789" {
		return fmt.Errorf("admin should see full SSN")
	}
	
	return nil
}

// TestCreditCardMasking verifies credit card masking
func TestCreditCardMasking() error {
	data := MaskingTestData{
		CreditCard: "4111-1111-1111-1234",
	}
	
	// Without payment:read, card should be masked
	ctx := sctx.NewUserContext("user-1", []string{})
	result, err := cereal.MarshalJSON(data, ctx)
	if err != nil {
		return fmt.Errorf("marshal failed: %v", err)
	}
	
	var unmarshaled MaskingTestData
	unmarshaled, err = cereal.UnmarshalJSON[MaskingTestData](result, ctx)
	if err != nil {
		return fmt.Errorf("unmarshal failed: %v", err)
	}
	
	// Should show only last 4 digits
	if unmarshaled.CreditCard != "****-****-****-1234" {
		return fmt.Errorf("credit card should be masked to show only last 4, got: %s", unmarshaled.CreditCard)
	}
	
	// With payment:read, full number should be visible
	paymentCtx := sctx.NewUserContext("payment-1", []string{"payment:read"})
	paymentResult, _ := cereal.MarshalJSON(data, paymentCtx)
	
	var paymentUnmarshaled MaskingTestData
	paymentUnmarshaled, _ = cereal.UnmarshalJSON[MaskingTestData](paymentResult, paymentCtx)
	
	if paymentUnmarshaled.CreditCard != "4111-1111-1111-1234" {
		return fmt.Errorf("payment:read should see full credit card")
	}
	
	return nil
}

// TestEmailMasking verifies email partial masking
func TestEmailMasking() error {
	data := MaskingTestData{
		Email: "john.doe@example.com",
	}
	
	// Email fields are typically partially masked for privacy
	ctx := sctx.NewUserContext("user-1", []string{})
	
	// Apply security transformation through marshal/unmarshal
	result, err := cereal.MarshalJSON(data, ctx)
	if err != nil {
		return fmt.Errorf("marshal failed: %v", err)
	}
	
	var secured MaskingTestData
	secured, err = cereal.UnmarshalJSON[MaskingTestData](result, ctx)
	if err != nil {
		return fmt.Errorf("unmarshal failed: %v", err)
	}
	
	// Email should be partially masked
	if !strings.HasPrefix(secured.Email, "j") || !strings.Contains(secured.Email, "***@") {
		return fmt.Errorf("email should be partially masked, got: %s", secured.Email)
	}
	
	// Should preserve domain
	if !strings.HasSuffix(secured.Email, "@example.com") {
		return fmt.Errorf("email domain should be preserved")
	}
	
	return nil
}

// TestCustomMaskFunctions verifies custom mask function support
func TestCustomMaskFunctions() error {
	// Test password field - should always be empty
	data := MaskingTestData{
		Password: "super-secret-password",
	}
	
	// Regardless of permissions, password should never be visible
	adminCtx := sctx.NewUserContext("admin-1", []string{"admin", "never"})
	result, err := cereal.MarshalJSON(data, adminCtx)
	if err != nil {
		return fmt.Errorf("marshal failed: %v", err)
	}
	
	var unmarshaled MaskingTestData
	unmarshaled, err = cereal.UnmarshalJSON[MaskingTestData](result, adminCtx)
	if err != nil {
		return fmt.Errorf("unmarshal failed: %v", err)
	}
	
	if unmarshaled.Password != "" {
		return fmt.Errorf("password should always be empty, got: %s", unmarshaled.Password)
	}
	
	// Test API key - should show only first few characters
	data.APIKey = "sk-1234567890abcdef"
	data.Password = "" // Clear for clean test
	
	userCtx := sctx.NewUserContext("user-1", []string{})
	userResult, _ := cereal.MarshalJSON(data, userCtx)
	
	var userUnmarshaled MaskingTestData
	userUnmarshaled, _ = cereal.UnmarshalJSON[MaskingTestData](userResult, userCtx)
	
	// API key should be masked
	if !strings.HasPrefix(userUnmarshaled.APIKey, "sk-1") || !strings.Contains(userUnmarshaled.APIKey, "****") {
		return fmt.Errorf("API key should show only first 4 chars, got: %s", userUnmarshaled.APIKey)
	}
	
	// Admin should see full API key
	adminResult, _ := cereal.MarshalJSON(data, adminCtx)
	var adminUnmarshaled MaskingTestData
	adminUnmarshaled, _ = cereal.UnmarshalJSON[MaskingTestData](adminResult, adminCtx)
	
	if adminUnmarshaled.APIKey != "sk-1234567890abcdef" {
		return fmt.Errorf("admin should see full API key")
	}
	
	return nil
}