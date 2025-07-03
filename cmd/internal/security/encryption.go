package security

import (
	"encoding/base64"
	"fmt"
	"strings"

	"aegis/catalog"
	"aegis/cereal"
	"aegis/sctx"
)

// Test types for encryption features
type EncryptionTestData struct {
	ID              string `json:"id"`
	PublicData      string `json:"public_data"`
	OwnerSecret     string `json:"owner_secret" encrypt:"owner"`
	SubscriberData  string `json:"subscriber_data" encrypt:"subscriber"`
	HybridData      string `json:"hybrid_data" encrypt:"owner" scope:"admin"`
	SensitiveNumber int    `json:"sensitive_number" encrypt:"owner" scope:"finance:read"`
}

// TestOwnerEncryption verifies owner-based encryption (org master key)
func TestOwnerEncryption() error {
	catalog.RegisterType[EncryptionTestData]()
	
	data := EncryptionTestData{
		ID:          "enc-001",
		PublicData:  "This is public",
		OwnerSecret: "Organization secret data",
	}
	
	// Serialize with any context - owner encryption should apply
	ctx := sctx.NewUserContext("user-1", []string{"user:read"})
	result, err := cereal.MarshalJSON(data, ctx)
	if err != nil {
		return fmt.Errorf("marshal failed: %v", err)
	}
	
	// Check raw JSON to verify encryption
	resultStr := string(result)
	
	// Public data should be visible
	if !strings.Contains(resultStr, "This is public") {
		return fmt.Errorf("public data should not be encrypted")
	}
	
	// Owner secret should be encrypted (base64 encoded)
	if strings.Contains(resultStr, "Organization secret data") {
		return fmt.Errorf("owner secret should be encrypted")
	}
	
	// Should contain base64 encoded encrypted data
	if !strings.Contains(resultStr, "owner_secret") {
		return fmt.Errorf("owner_secret field should be present")
	}
	
	// Unmarshal should decrypt for authorized contexts
	var decrypted EncryptionTestData
	decrypted, err = cereal.UnmarshalJSON[EncryptionTestData](result, ctx)
	if err != nil {
		return fmt.Errorf("unmarshal failed: %v", err)
	}
	
	// With owner encryption, data should be decrypted
	if decrypted.OwnerSecret != "Organization secret data" {
		return fmt.Errorf("owner secret should be decrypted, got: %s", decrypted.OwnerSecret)
	}
	
	return nil
}

// TestSubscriberEncryption verifies subscriber-specific encryption
func TestSubscriberEncryption() error {
	data := EncryptionTestData{
		ID:             "enc-002",
		PublicData:     "Public info",
		SubscriberData: "User-specific secret",
	}
	
	// Serialize with user context
	userCtx := sctx.NewUserContext("user-123", []string{"user:read"})
	result, err := cereal.MarshalJSON(data, userCtx)
	if err != nil {
		return fmt.Errorf("marshal failed: %v", err)
	}
	
	// Subscriber data should be encrypted with user's public key
	resultStr := string(result)
	if strings.Contains(resultStr, "User-specific secret") {
		return fmt.Errorf("subscriber data should be encrypted")
	}
	
	// Different user shouldn't be able to decrypt
	otherUserCtx := sctx.NewUserContext("user-456", []string{"user:read"})
	var otherUserData EncryptionTestData
	otherUserData, err = cereal.UnmarshalJSON[EncryptionTestData](result, otherUserCtx)
	if err != nil {
		return fmt.Errorf("unmarshal failed: %v", err)
	}
	
	// Other user should see encrypted/empty data
	if otherUserData.SubscriberData == "User-specific secret" {
		return fmt.Errorf("other user should not decrypt subscriber data")
	}
	
	// Original user should decrypt successfully
	var originalUserData EncryptionTestData
	originalUserData, err = cereal.UnmarshalJSON[EncryptionTestData](result, userCtx)
	if err != nil {
		return fmt.Errorf("original user unmarshal failed: %v", err)
	}
	
	if originalUserData.SubscriberData != "User-specific secret" {
		return fmt.Errorf("original user should decrypt subscriber data")
	}
	
	return nil
}

// TestSelectiveEncryption verifies encryption respects field selection
func TestSelectiveEncryption() error {
	data := EncryptionTestData{
		ID:              "enc-003",
		PublicData:      "Always visible",
		OwnerSecret:     "Org secret",
		SubscriberData:  "User secret",
		HybridData:      "Admin-only encrypted data",
		SensitiveNumber: 42,
	}
	
	// Test with non-admin user
	userCtx := sctx.NewUserContext("user-1", []string{"user:read"})
	userResult, err := cereal.MarshalJSON(data, userCtx)
	if err != nil {
		return fmt.Errorf("user marshal failed: %v", err)
	}
	
	var userData EncryptionTestData
	userData, err = cereal.UnmarshalJSON[EncryptionTestData](userResult, userCtx)
	if err != nil {
		return fmt.Errorf("user unmarshal failed: %v", err)
	}
	
	// Hybrid data should be redacted (no admin scope)
	if userData.HybridData != "[REDACTED]" {
		return fmt.Errorf("hybrid data should be redacted without admin scope")
	}
	
	// Sensitive number should be -1 (no finance:read)
	if userData.SensitiveNumber != -1 {
		return fmt.Errorf("sensitive number should be -1 without finance:read")
	}
	
	// Test with admin + finance user
	adminCtx := sctx.NewUserContext("admin-1", []string{"admin", "finance:read"})
	adminResult, err := cereal.MarshalJSON(data, adminCtx)
	if err != nil {
		return fmt.Errorf("admin marshal failed: %v", err)
	}
	
	var adminData EncryptionTestData
	adminData, err = cereal.UnmarshalJSON[EncryptionTestData](adminResult, adminCtx)
	if err != nil {
		return fmt.Errorf("admin unmarshal failed: %v", err)
	}
	
	// Admin should see decrypted hybrid data
	if adminData.HybridData != "Admin-only encrypted data" {
		return fmt.Errorf("admin should see hybrid data")
	}
	
	// Admin with finance:read should see number
	if adminData.SensitiveNumber != 42 {
		return fmt.Errorf("admin should see sensitive number")
	}
	
	return nil
}

// TestEncryptionContext verifies encryption context handling
func TestEncryptionContext() error {
	// Test with package context
	pkgCtx := sctx.NewPackageContext("encryption-test", []string{"internal"})
	
	data := EncryptionTestData{
		ID:          "enc-004",
		PublicData:  "Public",
		OwnerSecret: "Package encrypted",
	}
	
	// Package context should handle encryption
	result, err := cereal.MarshalJSON(data, pkgCtx)
	if err != nil {
		return fmt.Errorf("package context marshal failed: %v", err)
	}
	
	// Verify encryption occurred
	resultStr := string(result)
	if strings.Contains(resultStr, "Package encrypted") {
		return fmt.Errorf("data should be encrypted with package context")
	}
	
	// Package context should decrypt
	var decrypted EncryptionTestData
	decrypted, err = cereal.UnmarshalJSON[EncryptionTestData](result, pkgCtx)
	if err != nil {
		return fmt.Errorf("package context unmarshal failed: %v", err)
	}
	
	if decrypted.OwnerSecret != "Package encrypted" {
		return fmt.Errorf("package context should decrypt data")
	}
	
	// Test that encryption works consistently across contexts
	// Note: cereal requires a valid SecurityContext
	anonymousCtx := sctx.NewUserContext("anonymous", []string{})
	anonResult, err := cereal.MarshalJSON(data, anonymousCtx)
	if err != nil {
		return fmt.Errorf("anonymous context marshal failed: %v", err)
	}
	
	if len(anonResult) == 0 {
		return fmt.Errorf("should produce output with anonymous context")
	}
	
	return nil
}

// Helper to check if string is base64 encoded
func isBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}