package observability

import (
	"fmt"

	"github.com/spf13/cobra"
	
	"aegis/moisten"
)

func runLoggingTests(cmd *cobra.Command, args []string) {
	fmt.Println("\n📝 LOGGING CAPABILITIES")
	fmt.Println("=======================")
	
	// Initialize the system
	moisten.ForTesting()
	
	fmt.Println("\n📝 CAPABILITY 1: STRUCTURED LOGGING")
	fmt.Println("===================================")
	runStructuredLoggingTests(cmd, args)
	
	fmt.Println("\n🎚️ CAPABILITY 2: LEVEL MANAGEMENT")
	fmt.Println("==================================")
	runLevelManagementTests(cmd, args)
	
	fmt.Println("\n🔧 CAPABILITY 3: FIELD PROCESSING")
	fmt.Println("==================================")
	runFieldProcessingTests(cmd, args)
	
	fmt.Println("\n📡 CAPABILITY 4: EVENT EMISSION")
	fmt.Println("===============================")
	runEventEmissionTests(cmd, args)
	
	fmt.Println("\n🔒 CAPABILITY 5: SECURITY INTEGRATION")
	fmt.Println("=====================================")
	runSecurityIntegrationTests(cmd, args)
	
	fmt.Println("\n✅ All logging capabilities validated!")
}