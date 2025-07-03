package extensibility

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	
	"aegis/moisten"
)

// NewExtensibilityTestCmd creates the extensibility domain test command
func NewExtensibilityTestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "extensibility",
		Short: "🔧 Test extensibility capabilities",
		Long: `
🔧 EXTENSIBILITY CAPABILITIES
=============================

Tests the framework's revolutionary extensibility features:
• Divergent behaviors through type signatures
• Convention-based discovery patterns
• Behavior composition and pipelines
• Type-safe contracts
• Zero-coordination extension

These tests prove the core innovation: the same string
can mean completely different things based on type context.`,
		Run: runAllExtensibilityTests,
	}

	// Add subcommands for each test file
	cmd.AddCommand(&cobra.Command{
		Use:   "divergent",
		Short: "Type signature divergent behaviors",
		Run:   runDivergentBehaviorTests,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "behaviors",
		Short: "Pipeline and behavior tests",
		Run:   runBehaviorTests,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "conventions",
		Short: "Convention discovery tests",
		Run:   runConventionTests,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "pipz",
		Short: "Core pipz contract tests",
		Run:   runPipzContractTests,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "emergent",
		Short: "Emergent behavior tests",
		Run:   runPipzEmergentTests,
	})

	// Temporarily disabled for build
	// cmd.AddCommand(&cobra.Command{
	//	Use:   "multitenancy", 
	//	Short: "Zero-config multi-tenancy tests",
	//	Run:   runZeroConfigMultitenancyTests,
	// })

	// Add the visual showcase
	cmd.AddCommand(ShowcaseInfiniteParallelismCmd())

	return cmd
}

// runAllExtensibilityTests runs all tests in the extensibility domain
func runAllExtensibilityTests(cmd *cobra.Command, args []string) {
	fmt.Println("\n🔧 EXTENSIBILITY CAPABILITIES")
	fmt.Println("=============================")
	
	moisten.ForTesting()
	
	// Run all test suites
	fmt.Println("\n▶ Divergent Behavior Tests (CORE INNOVATION)")
	runDivergentBehaviorTests(cmd, args)
	
	fmt.Println("\n▶ Behavior Pipeline Tests")
	runBehaviorTests(cmd, args)
	
	fmt.Println("\n▶ Convention Discovery Tests")
	runConventionTests(cmd, args)
	
	fmt.Println("\n▶ Core pipz Contract Tests")
	runPipzContractTests(cmd, args)
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