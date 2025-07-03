package security

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	
	"aegis/moisten"
)

// NewSecurityTestCmd creates the security domain test command
func NewSecurityTestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "security",
		Short: "🔒 Test security capabilities",
		Long: `
🔒 SECURITY CAPABILITIES
========================

Tests the framework's security capabilities including:
• Context identity management (user vs system)
• Permission boundaries and access control
• Security context extensibility
• Security event generation and auditing
• Cross-service context propagation
• Dynamic security posture management

These tests demonstrate security as a cross-cutting
concern without exposing implementation details.`,
		Run: runAllSecurityTests,
	}

	// Add subcommands for each capability area
	cmd.AddCommand(&cobra.Command{
		Use:   "identity",
		Short: "Context identity management",
		Run:   runIdentityTests,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "permissions",
		Short: "Permission boundaries",
		Run:   runPermissionsTests,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "extensions",
		Short: "Context extensibility",
		Run:   runExtensionsTests,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "events",
		Short: "Security event generation",
		Run:   runEventsTests,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "propagation",
		Short: "Context propagation",
		Run:   runPropagationTests,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "serialization",
		Short: "Serialization security scenarios",
		Run:   runSerializationTests,
	})

	return cmd
}

// runAllSecurityTests runs all tests in the security domain
func runAllSecurityTests(cmd *cobra.Command, args []string) {
	fmt.Println("\n🔒 SECURITY CAPABILITIES")
	fmt.Println("========================")
	
	moisten.ForTesting()
	
	// Run all test suites
	fmt.Println("\n▶ Context Identity Tests")
	runIdentityTests(cmd, args)
	
	fmt.Println("\n▶ Permission Boundary Tests")
	runPermissionsTests(cmd, args)
	
	fmt.Println("\n▶ Context Extensibility Tests")
	runExtensionsTests(cmd, args)
	
	fmt.Println("\n▶ Security Event Tests")
	runEventsTests(cmd, args)
	
	fmt.Println("\n▶ Context Propagation Tests")
	runPropagationTests(cmd, args)
	
	fmt.Println("\n▶ Security Logging Tests")
	runLoggingTests(cmd, args)
	
	fmt.Println("\n▶ Serialization Security Scenarios")
	runSerializationTests(cmd, args)
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