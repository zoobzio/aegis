package integration

import (
	"fmt"

	"github.com/spf13/cobra"
	
	"aegis/cmd/internal/data"
	"aegis/cmd/internal/extensibility"
	"aegis/cmd/internal/observability"
	"aegis/cmd/internal/security"
	"aegis/moisten"
)

// RunAllTests runs all tests from all domains
func RunAllTests(cmd *cobra.Command, args []string) {
	fmt.Println("\n🧪 ZBZ FRAMEWORK CAPABILITY TESTS")
	fmt.Println("=================================")
	fmt.Println("Running all tests across all domains...")
	
	// Initialize the system once
	moisten.ForTesting()
	
	// Run data domain tests
	fmt.Println("\n\n📊 DATA DOMAIN")
	fmt.Println("==============")
	dataCmd := data.NewDataTestCmd()
	dataCmd.Run(cmd, args)
	
	// Run observability domain tests  
	fmt.Println("\n\n📡 OBSERVABILITY DOMAIN")
	fmt.Println("======================")
	obsCmd := observability.NewObservabilityTestCmd()
	obsCmd.Run(cmd, args)
	
	// Run extensibility domain tests
	fmt.Println("\n\n🔧 EXTENSIBILITY DOMAIN")
	fmt.Println("======================")
	extCmd := extensibility.NewExtensibilityTestCmd()
	extCmd.Run(cmd, args)
	
	// Run security domain tests
	fmt.Println("\n\n🔒 SECURITY DOMAIN")
	fmt.Println("==================")
	secCmd := security.NewSecurityTestCmd()
	secCmd.Run(cmd, args)
	
	// TODO: Add other domains as they're ready:
	// - performance
	
	fmt.Println("\n\n✅ All tests completed!")
}