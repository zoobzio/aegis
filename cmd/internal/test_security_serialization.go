package main

import (
	"fmt"
	"strings"
	
	"aegis/catalog"
	"aegis/cereal"
	"aegis/moisten"
	"aegis/sctx"
)

// TestData with security tags
type TestData struct {
	Name        string `json:"name"`
	Description string `json:"description" scope:"admin"`
	SSN         string `json:"ssn" validate:"ssn" scope:"admin"`
	Password    string `json:"password" validate:"password"`
}

func main() {
	// Initialize the framework
	moisten.ForTesting()
	
	// Register the type
	catalog.RegisterType[TestData]()
	
	// Create test data
	data := TestData{
		Name:        "Public Name",
		Description: "Secret Description",
		SSN:         "123-45-6789",
		Password:    "super-secret",
	}
	
	// Test with no permissions
	userCtx := sctx.NewUserContext("user-1", []string{})
	result, err := cereal.MarshalJSON(data, userCtx)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}
	
	resultStr := string(result)
	fmt.Printf("User result: %s\n", resultStr)
	
	// Check expectations
	if !strings.Contains(resultStr, "Public Name") {
		fmt.Println("FAIL: Public name should be visible")
	}
	
	if strings.Contains(resultStr, "Secret Description") {
		fmt.Println("FAIL: Description should be redacted")
	}
	
	if strings.Contains(resultStr, "123-45-6789") {
		fmt.Println("FAIL: SSN should be masked")
	}
	
	if strings.Contains(resultStr, "super-secret") {
		fmt.Println("FAIL: Password should be empty")
	}
	
	// Test with admin permissions
	adminCtx := sctx.NewUserContext("admin-1", []string{"admin"})
	adminResult, _ := cereal.MarshalJSON(data, adminCtx)
	fmt.Printf("\nAdmin result: %s\n", string(adminResult))
}