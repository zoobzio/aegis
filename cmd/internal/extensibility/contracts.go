package extensibility

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	
	"aegis/pipz"
)

// runPipzContractTests tests the core pipz contract functionality
func runPipzContractTests(cmd *cobra.Command, args []string) {
	runTest("Contract Creation and Caching", testContractCreation)
	runTest("Type-Safe Registration", testTypeSafeRegistration)
	runTest("Processing with Type Safety", testProcessingTypeSafety)
	runTest("Contract Registry Isolation", testContractRegistryIsolation)
	runTest("Processor Management", testProcessorManagement)
	runTest("Concurrent Contract Operations", testConcurrentContracts)
	runTest("Simple Contract API", testSimpleContractAPI)
	runTest("Contract Signatures", testContractSignatures)
}

// Define concrete types for testing - this is KEY!
type EmailValidation string
type PhoneValidation string
type SecurityCheck string

// Define different key values
const (
	ValidateFormat EmailValidation = "format"
	ValidateDomain EmailValidation = "domain"
	
	ValidateLength PhoneValidation = "length"
	ValidateCountry PhoneValidation = "country"
	
	CheckAuth SecurityCheck = "auth"
	CheckPermission SecurityCheck = "permission"
)

// testContractCreation verifies contracts are created and cached by type signature
func testContractCreation() error {
	// Get contracts with same Input/Output but different Keys
	emailContract1 := pipz.GetContract[EmailValidation, string, error]()
	emailContract2 := pipz.GetContract[EmailValidation, string, error]()
	phoneContract := pipz.GetContract[PhoneValidation, string, error]()
	
	// Same type signature should return same instance
	if emailContract1 != emailContract2 {
		return fmt.Errorf("contracts with same signature should be cached")
	}
	
	// Different key type means different contract (can't compare pointers of different types)
	if phoneContract == nil {
		return fmt.Errorf("different key type should create new contract")
	}
	
	// Different Input/Output types also create different contracts
	boolContract := pipz.GetContract[EmailValidation, string, bool]()
	if boolContract == nil {
		return fmt.Errorf("different output type should create new contract")
	}
	
	return nil
}

// testTypeSafeRegistration verifies type safety in registration
func testTypeSafeRegistration() error {
	emailContract := pipz.GetContract[EmailValidation, string, error]()
	phoneContract := pipz.GetContract[PhoneValidation, string, error]()
	
	// Register email validators
	emailContract.Register(ValidateFormat, func(email string) error {
		if !strings.Contains(email, "@") {
			return fmt.Errorf("invalid email format")
		}
		return nil
	})
	
	emailContract.Register(ValidateDomain, func(email string) error {
		if strings.HasSuffix(email, "@example.com") {
			return fmt.Errorf("example.com not allowed")
		}
		return nil
	})
	
	// Register phone validators - completely separate from email!
	phoneContract.Register(ValidateLength, func(phone string) error {
		if len(phone) != 10 {
			return fmt.Errorf("phone must be 10 digits")
		}
		return nil
	})
	
	// This would NOT compile - type safety!
	// emailContract.Register(ValidateLength, ...) // Compile error: ValidateLength is PhoneValidation, not EmailValidation
	// phoneContract.Register(ValidateFormat, ...) // Compile error: ValidateFormat is EmailValidation, not PhoneValidation
	
	// Verify registrations
	if !emailContract.HasProcessor(ValidateFormat) {
		return fmt.Errorf("email format validator should be registered")
	}
	
	if !phoneContract.HasProcessor(ValidateLength) {
		return fmt.Errorf("phone length validator should be registered")
	}
	
	// Verify isolation - email contract doesn't have phone validators
	emailKeys := emailContract.ListKeys()
	if len(emailKeys) != 2 {
		return fmt.Errorf("email contract should have 2 validators, got %d", len(emailKeys))
	}
	
	phoneKeys := phoneContract.ListKeys()
	if len(phoneKeys) != 1 {
		return fmt.Errorf("phone contract should have 1 validator, got %d", len(phoneKeys))
	}
	
	return nil
}

// testProcessingTypeSafety verifies processing maintains type safety
func testProcessingTypeSafety() error {
	// Create a contract for transformations
	type TransformOp string
	const (
		Upper TransformOp = "upper"
		Lower TransformOp = "lower"
		Trim  TransformOp = "trim"
	)
	
	transformContract := pipz.GetContract[TransformOp, string, string]()
	
	// Register transformers
	transformContract.Register(Upper, strings.ToUpper)
	transformContract.Register(Lower, strings.ToLower)
	transformContract.Register(Trim, strings.TrimSpace)
	
	// Process with type safety
	result1, ok1 := transformContract.Process(Upper, "hello world")
	if !ok1 || result1 != "HELLO WORLD" {
		return fmt.Errorf("upper transform failed")
	}
	
	// Can't accidentally use wrong key type
	// This would not compile:
	// transformContract.Process(ValidateFormat, "test") // Compile error!
	
	// Test missing processor
	type UnregisteredOp string
	const Reverse UnregisteredOp = "reverse"
	// Can't process with unregistered key from different type
	// transformContract.Process(Reverse, "test") // Would not compile!
	
	return nil
}

// testContractRegistryIsolation verifies complete isolation between contracts
func testContractRegistryIsolation() error {
	// Two different validation domains using the same string values
	type UserValidation string
	type ProductValidation string
	
	const (
		// Same string value, different types!
		UserRequired UserValidation = "required"
		ProductRequired ProductValidation = "required"
	)
	
	userContract := pipz.GetContract[UserValidation, map[string]any, error]()
	productContract := pipz.GetContract[ProductValidation, map[string]any, error]()
	
	// Register completely different validation logic for "required"
	userContract.Register(UserRequired, func(data map[string]any) error {
		// User validation: need email and password
		if _, hasEmail := data["email"]; !hasEmail {
			return fmt.Errorf("user requires email")
		}
		if _, hasPass := data["password"]; !hasPass {
			return fmt.Errorf("user requires password")
		}
		return nil
	})
	
	productContract.Register(ProductRequired, func(data map[string]any) error {
		// Product validation: need name and price
		if _, hasName := data["name"]; !hasName {
			return fmt.Errorf("product requires name")
		}
		if _, hasPrice := data["price"]; !hasPrice {
			return fmt.Errorf("product requires price")
		}
		return nil
	})
	
	// Test user validation
	userData := map[string]any{"email": "test@example.com", "password": "secret"}
	err1, ok1 := userContract.Process(UserRequired, userData)
	if !ok1 || err1 != nil {
		return fmt.Errorf("valid user data should pass")
	}
	
	// Same data fails product validation
	err2, ok2 := productContract.Process(ProductRequired, userData)
	if !ok2 || err2 == nil {
		return fmt.Errorf("user data should fail product validation")
	}
	
	return nil
}

// testProcessorManagement tests registration, unregistration, and ordering
func testProcessorManagement() error {
	type Pipeline string
	const (
		Step1 Pipeline = "step1"
		Step2 Pipeline = "step2"
		Step3 Pipeline = "step3"
	)
	
	contract := pipz.GetContract[Pipeline, int, int]()
	
	// Register in specific order
	contract.Register(Step1, func(n int) int { return n + 1 })
	contract.Register(Step2, func(n int) int { return n * 2 })
	contract.Register(Step3, func(n int) int { return n - 3 })
	
	// Verify order is preserved
	keys := contract.ListKeys()
	if len(keys) != 3 || keys[0] != Step1 || keys[1] != Step2 || keys[2] != Step3 {
		return fmt.Errorf("registration order not preserved: %v", keys)
	}
	
	// Test processing
	r1, _ := contract.Process(Step1, 5) // 5 + 1 = 6
	r2, _ := contract.Process(Step2, r1) // 6 * 2 = 12
	r3, _ := contract.Process(Step3, r2) // 12 - 3 = 9
	
	if r3 != 9 {
		return fmt.Errorf("pipeline processing failed: expected 9, got %d", r3)
	}
	
	// Unregister middle step
	contract.Unregister(Step2)
	
	// Verify removal
	keys2 := contract.ListKeys()
	if len(keys2) != 2 || keys2[0] != Step1 || keys2[1] != Step3 {
		return fmt.Errorf("unregister failed: %v", keys2)
	}
	
	// Processing Step2 should now fail
	_, ok := contract.Process(Step2, 10)
	if ok {
		return fmt.Errorf("unregistered processor should not be found")
	}
	
	return nil
}

// testConcurrentContracts verifies thread safety
func testConcurrentContracts() error {
	type ConcurrentOp string
	contract := pipz.GetContract[ConcurrentOp, string, string]()
	
	// Concurrent registrations
	done := make(chan bool, 20)
	
	// 10 goroutines registering
	for i := 0; i < 10; i++ {
		go func(id int) {
			key := ConcurrentOp(fmt.Sprintf("op%d", id))
			contract.Register(key, func(s string) string {
				return fmt.Sprintf("%s-%d", s, id)
			})
			done <- true
		}(i)
	}
	
	// 10 goroutines processing
	for i := 0; i < 10; i++ {
		go func(id int) {
			key := ConcurrentOp(fmt.Sprintf("op%d", id%5)) // Use first 5 ops
			result, ok := contract.Process(key, "test")
			if ok && !strings.Contains(result, fmt.Sprintf("-%d", id%5)) {
				panic("concurrent processing failed")
			}
			done <- true
		}(i)
	}
	
	// Wait for completion
	for i := 0; i < 20; i++ {
		<-done
	}
	
	// Verify all registered
	keys := contract.ListKeys()
	if len(keys) != 10 {
		return fmt.Errorf("expected 10 processors, got %d", len(keys))
	}
	
	return nil
}

// testSimpleContractAPI tests the simple contract functionality
func testSimpleContractAPI() error {
	// Simple contract for string reversal
	reverser := pipz.GetSimpleContract[string, string]()
	
	// Initially no processor
	if reverser.HasProcessor() {
		return fmt.Errorf("new simple contract should have no processor")
	}
	
	// Process without processor should fail
	_, ok := reverser.Process("hello")
	if ok {
		return fmt.Errorf("processing without processor should fail")
	}
	
	// Set processor
	reverser.SetProcessor(func(s string) string {
		runes := []rune(s)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes)
	})
	
	// Process should work
	result, ok := reverser.Process("hello")
	if !ok || result != "olleh" {
		return fmt.Errorf("string reversal failed")
	}
	
	// Replace processor
	reverser.SetProcessor(strings.ToUpper)
	
	// New processor should be active
	result2, ok2 := reverser.Process("hello")
	if !ok2 || result2 != "HELLO" {
		return fmt.Errorf("processor replacement failed")
	}
	
	return nil
}

// testContractSignatures verifies how type signatures create unique contracts
func testContractSignatures() error {
	// These all create DIFFERENT contracts even with same key strings
	type AppSecurity string
	type NetworkSecurity string
	type DataSecurity string
	
	const Validate = "validate" // Same string used everywhere!
	
	// Each gets its own universe
	appSec := pipz.GetContract[AppSecurity, string, bool]()
	netSec := pipz.GetContract[NetworkSecurity, string, bool]()
	dataSec := pipz.GetContract[DataSecurity, string, bool]()
	
	// Register different validation for each
	appSec.Register(AppSecurity(Validate), func(s string) bool {
		return len(s) > 0 // App: just not empty
	})
	
	netSec.Register(NetworkSecurity(Validate), func(s string) bool {
		return strings.Contains(s, "https://") // Network: must be HTTPS
	})
	
	dataSec.Register(DataSecurity(Validate), func(s string) bool {
		return !strings.Contains(s, "password") // Data: no passwords in plain text
	})
	
	// Test same input, different validations
	testInput := "https://example.com"
	
	appResult, _ := appSec.Process(AppSecurity(Validate), testInput)
	netResult, _ := netSec.Process(NetworkSecurity(Validate), testInput)
	dataResult, _ := dataSec.Process(DataSecurity(Validate), testInput)
	
	if !appResult || !netResult || !dataResult {
		return fmt.Errorf("all should pass for valid input")
	}
	
	// Test input that only passes some
	testInput2 := "password123"
	
	appResult2, _ := appSec.Process(AppSecurity(Validate), testInput2)
	netResult2, _ := netSec.Process(NetworkSecurity(Validate), testInput2)
	dataResult2, _ := dataSec.Process(DataSecurity(Validate), testInput2)
	
	if !appResult2 || netResult2 || dataResult2 {
		return fmt.Errorf("only app validation should pass for password")
	}
	
	return nil
}