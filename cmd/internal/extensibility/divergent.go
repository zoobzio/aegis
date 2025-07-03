package extensibility

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	
	"aegis/catalog"
	"aegis/moisten"
	"aegis/pipz"
)

// Define different behavior key types that will use THE SAME STRING VALUES
type SecurityBehaviorKey string
type ValidationBehaviorKey string
type AuditBehaviorKey string
type TransformBehaviorKey string

// All use "required" but mean completely different things!
const (
	// Security: "required" means authentication required
	SecurityRequired SecurityBehaviorKey = "required"
	
	// Validation: "required" means field must not be empty
	ValidationRequired ValidationBehaviorKey = "required"
	
	// Audit: "required" means must be logged
	AuditRequired AuditBehaviorKey = "required"
	
	// Transform: "required" means must be normalized
	TransformRequired TransformBehaviorKey = "required"
)

// Test data showing a field with multiple "required" behaviors
type DivergentTestData struct {
	ID       string `json:"id"`
	Email    string `json:"email" security:"required" validate:"required" audit:"required" transform:"required"`
	Password string `json:"password" security:"required"`
	Name     string `json:"name" validate:"required"`
}

func runDivergentBehaviorTests(cmd *cobra.Command, args []string) {
	moisten.ForTesting()
	
	runTest("Same String Different Types", testSameStringDifferentTypes)
	runTest("Type Signature Isolation", testTypeSignatureIsolation)
	runTest("Multiple Behaviors Same Tag", testMultipleBehaviorsSameTag)
	runTest("Domain-Specific Universes", testDomainSpecificUniverses)
}

// testSameStringDifferentTypes proves that "required" means different things based on type
func testSameStringDifferentTypes() error {
	// Create four completely independent contracts using the same "required" string
	securityContract := pipz.GetContract[SecurityBehaviorKey, string, bool]()
	validationContract := pipz.GetContract[ValidationBehaviorKey, string, bool]()
	auditContract := pipz.GetContract[AuditBehaviorKey, string, bool]()
	transformContract := pipz.GetContract[TransformBehaviorKey, string, bool]()
	
	// Register completely different behaviors for the same "required" string
	securityContract.Register(SecurityRequired, func(input string) bool {
		// Security: check if user is authenticated
		return strings.Contains(input, "@authenticated")
	})
	
	validationContract.Register(ValidationRequired, func(input string) bool {
		// Validation: check if field is not empty
		return len(strings.TrimSpace(input)) > 0
	})
	
	auditContract.Register(AuditRequired, func(input string) bool {
		// Audit: always return true (everything gets logged)
		return true
	})
	
	transformContract.Register(TransformRequired, func(input string) bool {
		// Transform: check if needs normalization
		return strings.ToLower(input) != input
	})
	
	// Test that each contract is completely isolated
	testInput := "TEST@authenticated"
	
	// Security sees authentication
	secResult, _ := securityContract.Process(SecurityRequired, testInput)
	if !secResult {
		return fmt.Errorf("security should see @authenticated")
	}
	
	// Validation sees non-empty
	valResult, _ := validationContract.Process(ValidationRequired, testInput)
	if !valResult {
		return fmt.Errorf("validation should see non-empty string")
	}
	
	// Audit always passes
	auditResult, _ := auditContract.Process(AuditRequired, "")
	if !auditResult {
		return fmt.Errorf("audit should always pass")
	}
	
	// Transform sees uppercase
	transResult, _ := transformContract.Process(TransformRequired, testInput)
	if !transResult {
		return fmt.Errorf("transform should detect uppercase")
	}
	
	// Prove complete isolation - registering more behaviors doesn't affect others
	securityContract.Register(SecurityRequired, func(input string) bool {
		// Changed security behavior
		return strings.Contains(input, "@admin")
	})
	
	// Original validation still works the same
	valResult2, _ := validationContract.Process(ValidationRequired, testInput)
	if !valResult2 {
		return fmt.Errorf("validation behavior should be unchanged")
	}
	
	return nil
}

// testTypeSignatureIsolation shows how type parameters create additional universes
func testTypeSignatureIsolation() error {
	// Same key type, different input/output types = different universes
	stringToString := pipz.GetContract[SecurityBehaviorKey, string, string]()
	stringToBool := pipz.GetContract[SecurityBehaviorKey, string, bool]()
	intToString := pipz.GetContract[SecurityBehaviorKey, int, string]()
	
	// Register the same key "required" in each universe
	stringToString.Register(SecurityRequired, func(s string) string {
		return "secured: " + s
	})
	
	stringToBool.Register(SecurityRequired, func(s string) bool {
		return len(s) > 5
	})
	
	intToString.Register(SecurityRequired, func(i int) string {
		return fmt.Sprintf("secure-%d", i*100)
	})
	
	// Each lives in its own universe
	r1, _ := stringToString.Process(SecurityRequired, "test")
	if r1 != "secured: test" {
		return fmt.Errorf("string->string universe failed")
	}
	
	r2, _ := stringToBool.Process(SecurityRequired, "test")
	if r2 != false {
		return fmt.Errorf("string->bool universe failed")
	}
	
	r3, _ := intToString.Process(SecurityRequired, 5)
	if r3 != "secure-500" {
		return fmt.Errorf("int->string universe failed")
	}
	
	return nil
}

// testMultipleBehaviorsSameTag demonstrates how one struct tag can trigger multiple behaviors
func testMultipleBehaviorsSameTag() error {
	// First register our custom tags to be extracted
	catalog.RegisterTag("audit")
	catalog.RegisterTag("transform")
	// Note: "security" and "validate" are already registered by default
	
	// Now register the type
	catalog.RegisterType[DivergentTestData]()
	
	// Get metadata to inspect tags
	metadata := catalog.Select[DivergentTestData]()
	emailField := metadata.Fields[1] // Email field
	
	// Count how many different behavior systems can act on the email field
	behaviorCount := 0
	
	if emailField.Tags["security"] == "required" {
		behaviorCount++ // Security behavior
	}
	
	if emailField.Tags["validate"] == "required" {
		behaviorCount++ // Validation behavior
	}
	
	if emailField.Tags["audit"] == "required" {
		behaviorCount++ // Audit behavior
	}
	
	if emailField.Tags["transform"] == "required" {
		behaviorCount++ // Transform behavior
	}
	
	if behaviorCount != 4 {
		return fmt.Errorf("expected 4 behavior systems, got %d", behaviorCount)
	}
	
	// Each system can register its own interpretation of "required"
	// without any coordination or collision!
	
	// Demonstrate that each can have its own pipeline
	securityPipeline := pipz.GetContract[SecurityBehaviorKey, string, error]()
	validationPipeline := pipz.GetContract[ValidationBehaviorKey, string, error]()
	auditPipeline := pipz.GetContract[AuditBehaviorKey, string, error]()
	transformPipeline := pipz.GetContract[TransformBehaviorKey, string, error]()
	
	// Register behaviors for the same "required" string
	securityPipeline.Register(SecurityRequired, func(s string) error {
		return nil // Security check
	})
	
	validationPipeline.Register(ValidationRequired, func(s string) error {
		return nil // Validation check
	})
	
	auditPipeline.Register(AuditRequired, func(s string) error {
		return nil // Audit logging
	})
	
	transformPipeline.Register(TransformRequired, func(s string) error {
		return nil // Data transformation
	})
	
	// All four pipelines can process "required" independently!
	
	return nil
}

// testDomainSpecificUniverses shows how different domains get isolated behavior spaces
func testDomainSpecificUniverses() error {
	// E-commerce domain types
	type EcommerceValidationKey string
	type EcommerceSecurityKey string
	
	// Healthcare domain types  
	type HealthcareValidationKey string
	type HealthcareSecurityKey string
	
	// Both domains use "required" but mean completely different things
	const (
		EcomRequired   EcommerceValidationKey = "required"
		HealthRequired HealthcareValidationKey = "required"
	)
	
	// E-commerce: "required" means must have SKU
	ecomContract := pipz.GetContract[EcommerceValidationKey, string, error]()
	ecomContract.Register(EcomRequired, func(input string) error {
		if !strings.HasPrefix(input, "SKU-") {
			return fmt.Errorf("e-commerce products must have SKU")
		}
		return nil
	})
	
	// Healthcare: "required" means must have patient ID
	healthContract := pipz.GetContract[HealthcareValidationKey, string, error]()
	healthContract.Register(HealthRequired, func(input string) error {
		if !strings.HasPrefix(input, "PT-") {
			return fmt.Errorf("healthcare records must have patient ID")
		}
		return nil
	})
	
	// Test isolation
	ecomErr, _ := ecomContract.Process(EcomRequired, "PT-12345")
	if ecomErr == nil {
		return fmt.Errorf("e-commerce should reject patient IDs")
	}
	
	healthErr, _ := healthContract.Process(HealthRequired, "SKU-12345")
	if healthErr == nil {
		return fmt.Errorf("healthcare should reject SKUs")
	}
	
	// Each domain has its own rules, same string, zero collision!
	return nil
}