package data

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	
	"aegis/moisten"
)

// NewDataTestCmd creates the data domain test command
func NewDataTestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "data",
		Short: "📊 Test data handling capabilities",
		Long: `
📊 DATA HANDLING CAPABILITIES
=============================

Tests the framework's data handling capabilities including:
• Type metadata extraction and introspection
• Multi-format serialization (JSON, YAML, TOML)
• Performance characteristics
• Type intelligence features

Run all data tests or specific capabilities.`,
		Run: runAllDataTests,
	}

	// Add subcommands for each test file
	cmd.AddCommand(&cobra.Command{
		Use:   "metadata",
		Short: "Type metadata extraction tests",
		Run:   runMetadataTests,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "formats",
		Short: "Multi-format serialization tests",
		Run:   runFormatTests,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "performance",
		Short: "Performance characteristic tests",
		Run:   runPerformanceTests,
	})

	return cmd
}

// runAllDataTests runs all tests in the data domain
func runAllDataTests(cmd *cobra.Command, args []string) {
	fmt.Println("\n📊 DATA HANDLING CAPABILITIES")
	fmt.Println("=============================")
	
	// Initialize with data-specific behaviors
	moisten.ForTesting(RegisterDataBehaviors)
	
	// Run all test suites
	fmt.Println("\n▶ Type Metadata Tests")
	runMetadataTests(cmd, args)
	
	fmt.Println("\n▶ Serialization Format Tests")
	runFormatTests(cmd, args)
	
	fmt.Println("\n▶ Performance Tests")
	runPerformanceTests(cmd, args)
}

// Helper function to run individual tests
func runTest(name string, testFunc func() error) {
	start := time.Now()
	err := testFunc()
	duration := time.Since(start)
	
	if err != nil {
		fmt.Printf("❌ %s - FAILED: %v (%v)\n", name, err, duration)
	} else {
		fmt.Printf("✅ %s - PASSED (%v)\n", name, duration)
	}
}