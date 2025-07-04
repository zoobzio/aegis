// Package security provides unified security behaviors for the ZBZ framework
// This adapter implements security, validation, masking, and access control
// as cross-cutting concerns using the pipz behavior system
package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"sync"
	
	"aegis/catalog"
	"aegis/sctx"
	"aegis/zlog"
)

// Package-level encryption configuration
var (
	encryptionKey []byte
	gcmCipher     cipher.AEAD
	encryptMutex  sync.RWMutex
)

// Initialize sets up the security adapter
// This is called by moisten during application startup
func Initialize() {
	// Register security-related tags with catalog
	RegisterSecurityTags()
	
	// Register standard mask functions
	RegisterMaskFunctions()
	
	// Register standard validators
	RegisterStandardValidators()
	
	// Register security behaviors for common patterns
	RegisterSecurityBehaviors()
	
	// Register type convention checker
	RegisterSecurityConvention()
}

// InitializeWithKey sets up the security adapter with encryption
func InitializeWithKey(key []byte) {
	SetEncryptionKey(key)
	Initialize()
}

// SetEncryptionKey configures the encryption key for field-level encryption
func SetEncryptionKey(key []byte) {
	encryptMutex.Lock()
	defer encryptMutex.Unlock()
	
	if len(key) != 32 {
		// For AES-256, we need exactly 32 bytes
		panic("security: encryption key must be exactly 32 bytes for AES-256")
	}
	
	encryptionKey = make([]byte, 32)
	copy(encryptionKey, key)
	
	// Initialize AES-GCM cipher
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		panic("security: failed to create cipher: " + err.Error())
	}
	
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic("security: failed to create GCM: " + err.Error())
	}
	
	gcmCipher = gcm
}

// Encrypt encrypts data using AES-256-GCM
func Encrypt(plaintext []byte) (string, error) {
	encryptMutex.RLock()
	cipher := gcmCipher
	encryptMutex.RUnlock()
	
	if cipher == nil {
		return "", nil // No encryption configured, return empty
	}
	
	// Create nonce
	nonce := make([]byte, cipher.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	
	// Encrypt and append nonce
	ciphertext := cipher.Seal(nonce, nonce, plaintext, nil)
	
	// Return base64 encoded
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts data encrypted with Encrypt
func Decrypt(ciphertext string) ([]byte, error) {
	encryptMutex.RLock()
	cipher := gcmCipher
	encryptMutex.RUnlock()
	
	if cipher == nil || ciphertext == "" {
		return nil, nil
	}
	
	// Decode from base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	}
	
	// Extract nonce
	nonceSize := cipher.NonceSize()
	if len(data) < nonceSize {
		return nil, err
	}
	
	nonce, ciphertext := data[:nonceSize], string(data[nonceSize:])
	
	// Decrypt
	plaintext, err := cipher.Open(nil, nonce, []byte(ciphertext), nil)
	if err != nil {
		return nil, err
	}
	
	return plaintext, nil
}

// RegisterSecurityTags registers security-related struct tags with catalog
func RegisterSecurityTags() {
	// Register the 3 core security tags
	catalog.RegisterTag("scope")      // Field-level access control: scope:"admin"
	catalog.RegisterTag("validate")   // Field validation rules: validate:"ssn"
	catalog.RegisterTag("security")   // Security behaviors: security:"encrypt,pii"
}

// RegisterTransformer creates and registers a generic security transformer for any type T.
// This is called by types implementing SecurityConvention in their RegisterSecurity method.
func RegisterTransformer[T any]() {
	transformer := func(source T, dest *T) error {
		*dest = source
		
		// Get metadata and manipulators
		metadata := catalog.Select[T]()
		manipulators := catalog.GetFieldManipulators[T]()
		
		for _, field := range metadata.Fields {
			manipulator, exists := manipulators[field.Name]
			if !exists {
				continue
			}
			
			// Check validate tag for masking functions
			if validateTag := field.Tags["validate"]; validateTag != "" {
				if maskFunc, hasMask := catalog.GetMaskFunction(validateTag); hasMask {
					if val, err := manipulator.GetString(source); err == nil {
						masked := maskFunc(val)
						manipulator.SetString(dest, masked)
					}
				}
			}
			
			// Check for PII/security tags that need generic masking
			if field.Tags["security"] == "pii" || field.Tags["encrypt"] != "" {
				manipulator.Redact(dest)
			}
		}
		
		return nil
	}
	
	// Register the transformer
	catalog.RegisterTransformer[T](transformer)
}

// RegisterSecurityConvention registers the type convention checker with catalog
func RegisterSecurityConvention() {
	catalog.RegisterTypeConvention(func(metadata catalog.ModelMetadata) *catalog.TypeConventionCheck {
		// Check if type needs security
		needsSecurity := false
		for _, field := range metadata.Fields {
			if field.Tags["security"] != "" || field.Tags["validate"] != "" || field.Tags["encrypt"] != "" {
				needsSecurity = true
				break
			}
		}
		
		if !needsSecurity {
			return nil // No security needed for this type
		}
		
		return &catalog.TypeConventionCheck{
			Name: "Setup",
			IsRequired: func(m catalog.ModelMetadata) bool {
				return true // Already checked above
			},
			InterfacePtr:   (*catalog.SetupConvention)(nil),
			FailureMessage: "has security tags but does not implement Setup() method. Use: func (YourType) Setup() { security.RegisterTransformer[YourType]() }",
		}
	})
}

// Re-export convenient security logging functions
var (
	// SecurityContext creates a log field with expanded security context
	SecurityContext = SecurityContextField
	// SensitiveField creates a log field that will be automatically masked
	SensitiveField = SensitiveDataField
	// AuditField creates a log field with compliance metadata
	AuditField = AuditLogField
)

// SecurityContextField creates a security context log field
func SecurityContextField(ctx sctx.SecurityContext) zlog.ZlogField {
	return zlog.ZlogField{
		Key:   "security",
		Type:  zlog.ZlogFieldType("security_context"),
		Value: ctx,
	}
}

// SensitiveDataField creates a sensitive data log field that will be automatically masked
func SensitiveDataField(key string, value interface{}) zlog.ZlogField {
	return zlog.ZlogField{
		Key:   key,
		Type:  zlog.ZlogFieldType("sensitive"),
		Value: value,
	}
}

// AuditLogField creates an audit log field with compliance metadata
func AuditLogField(action string) zlog.ZlogField {
	return zlog.ZlogField{
		Key:   "audit",
		Type:  zlog.ZlogFieldType("audit"),
		Value: action,
	}
}