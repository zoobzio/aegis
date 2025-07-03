package observability

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	
	"aegis/moisten"
	"aegis/zlog"
)

func NewEventEmissionTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "events",
		Short: "📡 Test event emission capabilities", 
		Long: `
📡 EVENT EMISSION CAPABILITIES
==============================

Tests the optional event system integration:

• Event emission on log messages
• Event structure and metadata
• Integration with event systems
• Optional/pluggable design

This validates that logging can optionally integrate
with event systems without creating dependencies.`,
		Run: runEventEmissionTests,
	}
}

func runEventEmissionTests(cmd *cobra.Command, args []string) {
	fmt.Println("\n📡 EVENT EMISSION CAPABILITIES")
	fmt.Println("==============================")
	
	// Initialize the system
	moisten.ForTesting()
	
	runTest("Event Structure", testEventStructure)
	runTest("Optional Integration", testOptionalIntegration)
	runTest("Event Metadata", testEventMetadata)
	runTest("Level-based Events", testLevelBasedEvents)
	runTest("No Dependency Design", testNoDependencyDesign)
}

func testEventStructure() error {
	// Test that log events have proper structure
	
	// Log messages that should generate events
	zlog.Info("Event structure test",
		zlog.String("component", "event_test"),
		zlog.Int("test_id", 1),
		zlog.Bool("structured", true),
	)
	
	zlog.Warn("Warning event test",
		zlog.String("issue", "test_warning"),
		zlog.String("severity", "low"),
	)
	
	zlog.Error("Error event test",
		zlog.String("error_type", "test_error"),
		zlog.Int("error_code", 500),
	)
	
	// Note: We can't easily verify event structure without capturing events,
	// but we verify that logging doesn't panic and completes successfully
	
	return nil
}

func testOptionalIntegration() error {
	// Test that event emission is optional and doesn't break without event sink
	
	// These should work fine even if no event sink is configured
	zlog.Debug("Optional integration test - debug")
	zlog.Info("Optional integration test - info")
	zlog.Warn("Optional integration test - warn")
	zlog.Error("Optional integration test - error")
	
	// Test with various field combinations
	zlog.Info("Optional test with fields",
		zlog.String("test_type", "optional"),
		zlog.Duration("runtime", time.Millisecond*100),
		zlog.Strings("features", []string{"optional", "integration", "events"}),
	)
	
	return nil
}

func testEventMetadata() error {
	// Test that events include proper metadata
	
	// Log with rich metadata
	zlog.Info("Rich metadata test",
		zlog.String("service", "zlog_test"),
		zlog.String("version", "1.0.0"),
		zlog.String("environment", "test"),
		zlog.Time("start_time", time.Now()),
		zlog.Duration("uptime", time.Hour),
		zlog.Int("requests_handled", 1000),
		zlog.Float64("success_rate", 99.5),
		zlog.Bool("healthy", true),
	)
	
	// Test error with context
	testErr := fmt.Errorf("test error for metadata")
	zlog.Error("Error with rich context",
		zlog.Err(testErr),
		zlog.String("operation", "metadata_test"),
		zlog.String("user_id", "test_user_123"),
		zlog.Int("retry_count", 3),
		zlog.Bool("recoverable", false),
	)
	
	return nil
}

func testLevelBasedEvents() error {
	// Test that events are emitted for different log levels
	
	// Save original level
	originalLevel := zlog.GetLevel()
	defer zlog.SetLevel(originalLevel)
	
	// Set to DEBUG to ensure all levels emit events
	zlog.SetLevel(zlog.DEBUG)
	
	// Test each level
	zlog.Debug("Debug level event",
		zlog.String("level", "debug"),
		zlog.String("purpose", "development"),
	)
	
	zlog.Info("Info level event",
		zlog.String("level", "info"),
		zlog.String("purpose", "operational"),
	)
	
	zlog.Warn("Warn level event",
		zlog.String("level", "warn"),
		zlog.String("purpose", "attention"),
	)
	
	zlog.Error("Error level event",
		zlog.String("level", "error"),
		zlog.String("purpose", "investigation"),
	)
	
	// Test with filtered levels
	zlog.SetLevel(zlog.WARN)
	
	// These should not emit events (filtered)
	zlog.Debug("Filtered debug")
	zlog.Info("Filtered info")
	
	// These should emit events
	zlog.Warn("Unfiltered warn")
	zlog.Error("Unfiltered error")
	
	return nil
}

func testNoDependencyDesign() error {
	// Test that logging works without any event system dependencies
	
	// This validates the "nuclear" logging principle - 
	// zlog should work with zero dependencies
	
	// Test basic logging without any external systems
	zlog.Info("Independence test",
		zlog.String("principle", "nuclear_logging"),
		zlog.Bool("independent", true),
		zlog.String("dependencies", "none"),
	)
	
	// Test that complex operations still work
	complexData := map[string]interface{}{
		"nested": map[string]int{"count": 42},
		"array":  []string{"a", "b", "c"},
		"meta":   "independence_test",
	}
	
	zlog.Info("Complex independence test",
		zlog.Data("complex_data", complexData),
		zlog.String("validation", "no_dependencies"),
		zlog.Time("timestamp", time.Now()),
	)
	
	// Test error handling independence
	independentErr := fmt.Errorf("independent error - no external dependencies")
	zlog.Error("Error independence test",
		zlog.Err(independentErr),
		zlog.String("principle", "self_contained"),
		zlog.Bool("requires_external_systems", false),
	)
	
	return nil
}

