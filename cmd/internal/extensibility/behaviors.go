package extensibility

import (
	"fmt"

	"github.com/spf13/cobra"
	
	"aegis/catalog"
	"aegis/sctx"
)

// runBehaviorTests runs all behavior pipeline tests
func runBehaviorTests(cmd *cobra.Command, args []string) {
	runTest("Pipeline Creation", TestPipelineCreation)
	runTest("Registration Order", TestRegistrationOrder)
	runTest("Type-Safe Contracts", TestTypeSafety)
	runTest("Behavior Isolation", TestBehaviorIsolation)
	runTest("Behavior Replacement", TestBehaviorReplacement)
	runTest("Multiple Behavior Tags", TestMultipleBehaviorTags)
}

// Test type for behavior registration
type BehaviorTestUser struct {
	ID       string `json:"id"`
	Email    string `json:"email" scope:"user:read"`
	Password string `json:"password" scope:"admin"`
	SSN      string `json:"ssn" validate:"ssn" scope:"admin"`
	Salary   int    `json:"salary" scope:"hr:read"`
}

// TestPipelineCreation verifies pipeline initialization per type
func TestPipelineCreation() error {
	// Get security pipeline for TestUser
	pipeline := catalog.GetSecurityPipeline[BehaviorTestUser]()
	
	if pipeline == nil {
		return fmt.Errorf("pipeline is nil")
	}
	
	// Should start empty
	keys := pipeline.ListKeys()
	if len(keys) != 0 {
		return fmt.Errorf("expected empty pipeline, got %d keys", len(keys))
	}
	
	return nil
}

// TestRegistrationOrder verifies behaviors maintain registration order
func TestRegistrationOrder() error {
	pipeline := catalog.GetSecurityPipeline[BehaviorTestUser]()
	
	// Register behaviors in specific order
	pipeline.Register(catalog.AccessControlBehavior, func(input catalog.SecurityInput[BehaviorTestUser]) catalog.SecurityOutput[BehaviorTestUser] {
		return catalog.SecurityOutput[BehaviorTestUser]{Data: input.Data, Error: nil}
	})
	
	pipeline.Register(catalog.RedactionBehavior, func(input catalog.SecurityInput[BehaviorTestUser]) catalog.SecurityOutput[BehaviorTestUser] {
		return catalog.SecurityOutput[BehaviorTestUser]{Data: input.Data, Error: nil}
	})
	
	pipeline.Register(catalog.AuditBehavior, func(input catalog.SecurityInput[BehaviorTestUser]) catalog.SecurityOutput[BehaviorTestUser] {
		return catalog.SecurityOutput[BehaviorTestUser]{Data: input.Data, Error: nil}
	})
	
	// Verify registration order is maintained
	keys := pipeline.ListKeys()
	if len(keys) != 3 {
		return fmt.Errorf("expected 3 behaviors, got %d", len(keys))
	}
	
	expectedOrder := []catalog.SecurityBehaviorKey{
		catalog.AccessControlBehavior,
		catalog.RedactionBehavior,
		catalog.AuditBehavior,
	}
	
	for i, expected := range expectedOrder {
		if keys[i] != expected {
			return fmt.Errorf("order mismatch at position %d: expected %s, got %s", i, expected, keys[i])
		}
	}
	
	return nil
}

// TestTypeSafety verifies compile-time type safety in pipelines
func TestTypeSafety() error {
	// This test verifies compile-time type safety
	pipeline := catalog.GetSecurityPipeline[BehaviorTestUser]()
	
	// Register a behavior with correct types
	pipeline.Register(catalog.AccessControlBehavior, func(input catalog.SecurityInput[BehaviorTestUser]) catalog.SecurityOutput[BehaviorTestUser] {
		// Type-safe access to TestUser fields
		_ = input.Data.Email
		_ = input.Data.SSN
		
		return catalog.SecurityOutput[BehaviorTestUser]{Data: input.Data, Error: nil}
	})
	
	// Process with type safety
	ctx := sctx.NewUserContext("test", []string{"user:read"})
	input := catalog.SecurityInput[BehaviorTestUser]{
		Data: BehaviorTestUser{ID: "123", Email: "test@example.com"},
		Context: ctx,
	}
	
	output, exists := pipeline.Process(catalog.AccessControlBehavior, input)
	if !exists {
		return fmt.Errorf("processor not found")
	}
	
	// Type-safe output access
	if output.Data.Email != "test@example.com" {
		return fmt.Errorf("type safety broken in output")
	}
	
	return nil
}

// TestBehaviorIsolation verifies behavior pipelines are isolated per type
func TestBehaviorIsolation() error {
	// The type signature creates the isolation - these are different pipelines!
	type User1 struct {
		ID string
	}
	
	type User2 struct {
		Name string
	}
	
	// Get pipelines for different types - each type gets its own pipeline
	pipeline1 := catalog.GetSecurityPipeline[User1]()
	pipeline2 := catalog.GetSecurityPipeline[User2]()
	
	// Register behavior on first pipeline
	pipeline1.Register(catalog.AccessControlBehavior, func(input catalog.SecurityInput[User1]) catalog.SecurityOutput[User1] {
		return catalog.SecurityOutput[User1]{Data: input.Data, Error: nil}
	})
	
	// Verify second pipeline is still empty
	keys1 := pipeline1.ListKeys()
	keys2 := pipeline2.ListKeys()
	
	if len(keys1) != 1 {
		return fmt.Errorf("expected 1 behavior in pipeline1, got %d", len(keys1))
	}
	
	if len(keys2) != 0 {
		return fmt.Errorf("expected 0 behaviors in pipeline2, got %d", len(keys2))
	}
	
	return nil
}

// TestBehaviorReplacement verifies behaviors can be replaced
func TestBehaviorReplacement() error {
	pipeline := catalog.GetSecurityPipeline[BehaviorTestUser]()
	
	// Track which behavior executed
	var executedBehavior string
	
	// Register initial behavior
	pipeline.Register(catalog.AccessControlBehavior, func(input catalog.SecurityInput[BehaviorTestUser]) catalog.SecurityOutput[BehaviorTestUser] {
		executedBehavior = "first"
		return catalog.SecurityOutput[BehaviorTestUser]{Data: input.Data, Error: nil}
	})
	
	// Process to verify first behavior
	ctx := sctx.NewUserContext("test", []string{"user:read"})
	input := catalog.SecurityInput[BehaviorTestUser]{
		Data: BehaviorTestUser{ID: "123"},
		Context: ctx,
	}
	
	pipeline.Process(catalog.AccessControlBehavior, input)
	if executedBehavior != "first" {
		return fmt.Errorf("first behavior didn't execute")
	}
	
	// Replace behavior
	pipeline.Register(catalog.AccessControlBehavior, func(input catalog.SecurityInput[BehaviorTestUser]) catalog.SecurityOutput[BehaviorTestUser] {
		executedBehavior = "second"
		return catalog.SecurityOutput[BehaviorTestUser]{Data: input.Data, Error: nil}
	})
	
	// Process again to verify replacement
	pipeline.Process(catalog.AccessControlBehavior, input)
	if executedBehavior != "second" {
		return fmt.Errorf("behavior was not replaced")
	}
	
	return nil
}

// TestMultipleBehaviorTags verifies the tag-based behavior composition pattern
func TestMultipleBehaviorTags() error {
	// This test demonstrates how the pipz contract system allows multiple behaviors
	// to be registered on the same pipeline, creating a composable security system
	
	// Track behavior execution
	var executionOrder []string
	
	// Get security pipeline for BehaviorTestUser - this is isolated by type signature
	pipeline := catalog.GetSecurityPipeline[BehaviorTestUser]()
	
	// Register three distinct behaviors that will execute in order
	// Note: Some behaviors may already be registered by moisten.ForTesting()
	
	// Use behaviors that are less likely to be pre-registered
	pipeline.Register(catalog.MaskingBehavior, func(input catalog.SecurityInput[BehaviorTestUser]) catalog.SecurityOutput[BehaviorTestUser] {
		executionOrder = append(executionOrder, "masking")
		return catalog.SecurityOutput[BehaviorTestUser]{Data: input.Data, Error: nil}
	})
	
	// Register PCI compliance behavior
	pipeline.Register(catalog.PCICompliance, func(input catalog.SecurityInput[BehaviorTestUser]) catalog.SecurityOutput[BehaviorTestUser] {
		executionOrder = append(executionOrder, "pci")
		return catalog.SecurityOutput[BehaviorTestUser]{Data: input.Data, Error: nil}
	})
	
	// Register HIPAA compliance behavior
	pipeline.Register(catalog.HIPAACompliance, func(input catalog.SecurityInput[BehaviorTestUser]) catalog.SecurityOutput[BehaviorTestUser] {
		executionOrder = append(executionOrder, "hipaa")
		return catalog.SecurityOutput[BehaviorTestUser]{Data: input.Data, Error: nil}
	})
	
	// Just verify that we have multiple behaviors registered
	keys := pipeline.ListKeys()
	if len(keys) < 3 {
		return fmt.Errorf("expected at least 3 behaviors registered, got %d", len(keys))
	}
	
	// Execute pipeline with all behaviors
	user := BehaviorTestUser{ID: "123"}
	ctx := sctx.NewUserContext("test", []string{"user:read"})
	
	// Process through pipeline manually (this executes all registered behaviors in order)
	input := catalog.SecurityInput[BehaviorTestUser]{
		Data:    user,
		Context: ctx,
	}
	
	// Process through the security pipeline
	for _, key := range pipeline.ListKeys() {
		output, exists := pipeline.Process(key, input)
		if !exists {
			continue
		}
		if output.Error != nil {
			return fmt.Errorf("pipeline execution failed: %v", output.Error)
		}
		input.Data = output.Data
	}
	
	// Verify result
	if input.Data.ID != "123" {
		return fmt.Errorf("data corrupted during pipeline execution")
	}
	
	// Verify our specific behaviors executed
	// We only check that our registered behaviors ran, not the total count
	// since moisten.ForTesting() may have registered additional behaviors
	expectedBehaviors := map[string]bool{
		"masking": false,
		"pci":     false,
		"hipaa":   false,
	}
	
	for _, behavior := range executionOrder {
		if _, expected := expectedBehaviors[behavior]; expected {
			expectedBehaviors[behavior] = true
		}
	}
	
	// Check all our behaviors executed
	for behavior, executed := range expectedBehaviors {
		if !executed {
			return fmt.Errorf("behavior %s did not execute", behavior)
		}
	}
	
	return nil
}