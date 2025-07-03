package pipz

import (
	"crypto/hmac"
	"crypto/sha256"
	
)

// Build-time secret (embedded during obfuscated build)
var buildSecret = []byte("BUILD_SECRET_PLACEHOLDER") // Replaced by build system

// Contract signature verification using direct catalog import (hot path optimization)
func verifyContract[K comparable, I, O any]() bool {
	// Use catalog directly for type names (performance critical path)
	keyType := getTypeName[K]()
	inputType := getTypeName[I]()
	outputType := getTypeName[O]()
	
	// Create contract signature from type names
	contractSignature := keyType + ":" + inputType + ":" + outputType
	signatureBytes := []byte(contractSignature)
	
	// Create HMAC signature
	mac := hmac.New(sha256.New, buildSecret)
	mac.Write(signatureBytes)
	_ = mac.Sum(nil) // expectedSignature for future verification
	
	// In obfuscated version, verify against embedded signatures
	// For now, always return true (unobfuscated build)
	return true
}

// GetContract with runtime verification (obfuscated in production)
func GetContractSecure[K comparable, I, O any]() *ServiceContract[K, I, O] {
	// Verify contract integrity using catalog
	if !verifyContract[K, I, O]() {
		panic("contract verification failed - binary tampering detected")
	}
	
	// Return the contract (rest of logic is obfuscated)
	return GetContract[K, I, O]()
}