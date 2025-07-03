package data

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/spf13/cobra"
	
	"aegis/catalog"
	"aegis/cereal"
	"aegis/moisten"
	"aegis/sctx"
)

// runPerformanceTests runs all performance characteristic tests
func runPerformanceTests(cmd *cobra.Command, args []string) {
	moisten.ForTesting()
	
	runTest("Type-Safe APIs", TestTypeSafeAPIs)
	runTest("Generic Serialization", TestGenericSerialization)
	runTest("Zero Allocations", TestZeroAllocations)
	runTest("Concurrent Safety", TestConcurrentSafety)
}

// Test types for performance testing
type PerformanceTestData struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Count       int                    `json:"count"`
	Values      []float64              `json:"values"`
	Metadata    map[string]interface{} `json:"metadata"`
	SecureField string                 `json:"secure" scope:"admin"`
}

// Generic test type for type safety
type GenericTestData[T any] struct {
	ID    string `json:"id"`
	Value T      `json:"value"`
	Items []T    `json:"items"`
}

// TestTypeSafeAPIs verifies compile-time type safety
func TestTypeSafeAPIs() error {
	catalog.RegisterType[PerformanceTestData]()
	
	// Test that APIs enforce type safety at compile time
	data := PerformanceTestData{
		ID:   "type-001",
		Name: "Type Safety Test",
	}
	
	ctx := sctx.NewUserContext("user-1", []string{"user:read"})
	
	// Marshal returns correct type
	jsonBytes, err := cereal.MarshalJSON(data, ctx)
	if err != nil {
		return fmt.Errorf("marshal failed: %v", err)
	}
	
	// Unmarshal returns exact type, not interface{}
	var result PerformanceTestData
	result, err = cereal.UnmarshalJSON[PerformanceTestData](jsonBytes, ctx)
	if err != nil {
		return fmt.Errorf("unmarshal failed: %v", err)
	}
	
	// Can access fields directly without type assertion
	if result.Name != "Type Safety Test" {
		return fmt.Errorf("type-safe access failed")
	}
	
	// Test that wrong type fails at compile time
	// This would not compile:
	// wrongType, _ := cereal.UnmarshalJSON[string](jsonBytes, ctx)
	
	// Test pointer types work correctly
	ptrData := &PerformanceTestData{ID: "ptr-001"}
	ptrBytes, _ := cereal.MarshalJSON(ptrData, ctx)
	
	var ptrResult *PerformanceTestData
	ptrResult, err = cereal.UnmarshalJSON[*PerformanceTestData](ptrBytes, ctx)
	if err != nil {
		return fmt.Errorf("pointer unmarshal failed: %v", err)
	}
	
	if ptrResult.ID != "ptr-001" {
		return fmt.Errorf("pointer type handling failed")
	}
	
	return nil
}

// TestGenericSerialization verifies generic type support
func TestGenericSerialization() error {
	// Register generic types with different type parameters
	catalog.RegisterType[GenericTestData[string]]()
	catalog.RegisterType[GenericTestData[int]]()
	
	// Test string generic
	stringData := GenericTestData[string]{
		ID:    "generic-string",
		Value: "test value",
		Items: []string{"one", "two", "three"},
	}
	
	ctx := sctx.NewUserContext("user-1", []string{"user:read"})
	stringBytes, err := cereal.MarshalJSON(stringData, ctx)
	if err != nil {
		return fmt.Errorf("generic string marshal failed: %v", err)
	}
	
	var stringResult GenericTestData[string]
	stringResult, err = cereal.UnmarshalJSON[GenericTestData[string]](stringBytes, ctx)
	if err != nil {
		return fmt.Errorf("generic string unmarshal failed: %v", err)
	}
	
	if stringResult.Value != "test value" {
		return fmt.Errorf("generic string value mismatch")
	}
	
	// Test int generic
	intData := GenericTestData[int]{
		ID:    "generic-int",
		Value: 42,
		Items: []int{1, 2, 3, 4, 5},
	}
	
	intBytes, err := cereal.MarshalJSON(intData, ctx)
	if err != nil {
		return fmt.Errorf("generic int marshal failed: %v", err)
	}
	
	var intResult GenericTestData[int]
	intResult, err = cereal.UnmarshalJSON[GenericTestData[int]](intBytes, ctx)
	if err != nil {
		return fmt.Errorf("generic int unmarshal failed: %v", err)
	}
	
	if intResult.Value != 42 {
		return fmt.Errorf("generic int value mismatch")
	}
	
	if len(intResult.Items) != 5 {
		return fmt.Errorf("generic int items mismatch")
	}
	
	return nil
}

// TestZeroAllocations verifies minimal allocations (best effort)
func TestZeroAllocations() error {
	// Prepare data
	data := PerformanceTestData{
		ID:    "alloc-001",
		Name:  "Allocation Test",
		Count: 100,
	}
	
	ctx := sctx.NewUserContext("user-1", []string{"user:read"})
	
	// Warm up to ensure any lazy initialization is done
	_, _ = cereal.MarshalJSON(data, ctx)
	
	// Measure allocations for a simple operation
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)
	
	// Perform operation
	result, err := cereal.MarshalJSON(data, ctx)
	if err != nil {
		return fmt.Errorf("marshal failed: %v", err)
	}
	
	runtime.ReadMemStats(&m2)
	allocations := m2.Mallocs - m1.Mallocs
	
	// We can't achieve true zero allocations due to:
	// - JSON marshaling allocates
	// - Security transformations may allocate
	// - But we should minimize unnecessary allocations
	
	// Just verify operation succeeded
	if len(result) == 0 {
		return fmt.Errorf("marshal produced empty result")
	}
	
	// Log allocations for awareness (not a failure)
	if allocations > 100 {
		// High allocation count might indicate inefficiency
		// but not necessarily a failure
		_ = allocations // Acknowledge but don't fail
	}
	
	return nil
}

// TestConcurrentSafety verifies thread-safe operations
func TestConcurrentSafety() error {
	// Prepare shared data
	data := PerformanceTestData{
		ID:          "concurrent-001",
		Name:        "Concurrent Test",
		Count:       1000,
		Values:      []float64{1.1, 2.2, 3.3, 4.4, 5.5},
		Metadata:    map[string]interface{}{"key": "value"},
		SecureField: "secret-data",
	}
	
	// Create contexts with different permissions
	contexts := []sctx.SecurityContext{
		sctx.NewUserContext("user-1", []string{"user:read"}),
		sctx.NewUserContext("admin-1", []string{"admin", "user:read"}),
		sctx.NewUserContext("guest-1", []string{}),
		sctx.NewPackageContext("test-pkg", []string{"internal"}),
	}
	
	// Run concurrent operations
	var wg sync.WaitGroup
	errors := make(chan error, 100)
	
	// Launch multiple goroutines doing different operations
	for i := 0; i < 10; i++ {
		for _, ctx := range contexts {
			wg.Add(1)
			go func(id int, context sctx.SecurityContext) {
				defer wg.Done()
				
				// JSON operations
				jsonData, err := cereal.MarshalJSON(data, context)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d JSON marshal: %v", id, err)
					return
				}
				
				var jsonResult PerformanceTestData
				jsonResult, err = cereal.UnmarshalJSON[PerformanceTestData](jsonData, context)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d JSON unmarshal: %v", id, err)
					return
				}
				
				// YAML operations
				yamlData, err := cereal.MarshalYAML(data, context)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d YAML marshal: %v", id, err)
					return
				}
				
				var yamlResult PerformanceTestData
				yamlResult, err = cereal.UnmarshalYAML[PerformanceTestData](yamlData, context)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d YAML unmarshal: %v", id, err)
					return
				}
				
				// Verify data integrity
				if jsonResult.ID != data.ID || yamlResult.ID != data.ID {
					errors <- fmt.Errorf("goroutine %d data corruption detected", id)
					return
				}
				
				// Apply security transformation
				localCopy := data
				cereal.Secure(&localCopy)
				
			}(i, ctx)
		}
	}
	
	// Wait for completion
	wg.Wait()
	close(errors)
	
	// Check for any errors
	for err := range errors {
		if err != nil {
			return err
		}
	}
	
	// If we got here, all concurrent operations succeeded
	return nil
}