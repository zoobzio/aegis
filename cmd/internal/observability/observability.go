package observability

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	
	"aegis/moisten"
)

// NewObservabilityTestCmd creates the observability domain test command
func NewObservabilityTestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "observability",
		Short: "📡 Test observability capabilities",
		Long: `
📡 OBSERVABILITY CAPABILITIES
=============================

Tests the framework's observability capabilities including:
• Structured logging with type safety
• Event emission and handling
• Log level management
• Field processing and security
• Performance monitoring

Run all observability tests or specific capabilities.`,
		Run: runAllObservabilityTests,
	}

	// Add subcommands for each test file
	cmd.AddCommand(&cobra.Command{
		Use:   "logging",
		Short: "Core logging capabilities",
		Run:   runLoggingTests,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "structured",
		Short: "Structured logging tests",
		Run:   runStructuredLoggingTests,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "events",
		Short: "Event emission tests",
		Run:   runEventEmissionTests,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "levels",
		Short: "Log level management tests",
		Run:   runLevelManagementTests,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "processing",
		Short: "Field processing tests",
		Run:   runFieldProcessingTests,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "security",
		Short: "Security integration tests",
		Run:   runSecurityIntegrationTests,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "events-arch",
		Short: "Event architecture tests",
		Run:   runEventArchitectureTests,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "commands",
		Short: "Command authorization tests",
		Run:   runCommandAuthorizationTests,
	})

	return cmd
}

// runAllObservabilityTests runs all tests in the observability domain
func runAllObservabilityTests(cmd *cobra.Command, args []string) {
	fmt.Println("\n📡 OBSERVABILITY CAPABILITIES")
	fmt.Println("=============================")
	
	moisten.ForTesting()
	
	// Run all test suites
	fmt.Println("\n▶ Core Logging Tests")
	runLoggingTests(cmd, args)
	
	fmt.Println("\n▶ Structured Logging Tests")
	runStructuredLoggingTests(cmd, args)
	
	fmt.Println("\n▶ Event Emission Tests")
	runEventEmissionTests(cmd, args)
	
	fmt.Println("\n▶ Level Management Tests")
	runLevelManagementTests(cmd, args)
	
	fmt.Println("\n▶ Field Processing Tests")
	runFieldProcessingTests(cmd, args)
	
	fmt.Println("\n▶ Security Integration Tests")
	runSecurityIntegrationTests(cmd, args)
	
	fmt.Println("\n▶ Event Architecture Tests")
	runEventArchitectureTests(cmd, args)
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