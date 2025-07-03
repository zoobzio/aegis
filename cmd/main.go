package main

import (
	"os"

	"github.com/spf13/cobra"
	
	"aegis/cmd/internal/data"
	"aegis/cmd/internal/extensibility"
	"aegis/cmd/internal/integration"
	"aegis/cmd/internal/observability"
	"aegis/cmd/internal/security"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "aegis",
		Short: "🚀 Aegis Framework CLI",
		Long: `🚀 Aegis Framework CLI
==================================

The Aegis framework demonstrates advanced Go patterns for
building secure, observable, and extensible applications.`,
	}

	// Create test command that holds all test subcommands
	testCmd := &cobra.Command{
		Use:   "test",
		Short: "Run framework capability tests",
		Long:  "Test various capabilities of the Aegis framework",
	}

	// Add domain test commands under 'test'
	testCmd.AddCommand(&cobra.Command{
		Use:   "all",
		Short: "Run all capability tests",
		Run:   integration.RunAllTests,
	})
	
	testCmd.AddCommand(data.NewDataTestCmd())
	testCmd.AddCommand(observability.NewObservabilityTestCmd())
	testCmd.AddCommand(extensibility.NewExtensibilityTestCmd())
	testCmd.AddCommand(security.NewSecurityTestCmd())

	// Add test command to root
	rootCmd.AddCommand(testCmd)

	// Future commands can be added here:
	// rootCmd.AddCommand(generateCmd)
	// rootCmd.AddCommand(serveCmd)
	// rootCmd.AddCommand(deployCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}