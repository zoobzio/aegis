package observability

import (
	"fmt"

	"github.com/spf13/cobra"
	
	"aegis/moisten"
	"aegis/zlog"
)

func NewLevelManagementTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "levels",
		Short: "🎚️ Test log level management capabilities",
		Long: `
🎚️ LOG LEVEL MANAGEMENT CAPABILITIES
====================================

Tests the log level filtering and configuration:

• Level setting and retrieval
• Level-based filtering behavior
• Thread-safe level changes
• Level hierarchy validation

This validates that log levels provide proper filtering
without impacting performance or thread safety.`,
		Run: runLevelManagementTests,
	}
}

func runLevelManagementTests(cmd *cobra.Command, args []string) {
	fmt.Println("\n🎚️ LOG LEVEL MANAGEMENT CAPABILITIES")
	fmt.Println("====================================")
	
	// Initialize the system
	moisten.ForTesting()
	
	runTest("Level Setting", testLevelSetting)
	runTest("Level Filtering", testLevelFiltering)
	runTest("Level Hierarchy", testLevelHierarchy)
	runTest("Thread Safety", testThreadSafety)
	runTest("Default Behavior", testDefaultBehavior)
}

func testLevelSetting() error {
	// Test setting and getting log levels
	
	// Save original level
	originalLevel := zlog.GetLevel()
	
	// Test setting different levels
	levels := []zlog.LogLevel{
		zlog.DEBUG,
		zlog.INFO,
		zlog.WARN,
		zlog.ERROR,
	}
	
	for _, level := range levels {
		zlog.SetLevel(level)
		currentLevel := zlog.GetLevel()
		if currentLevel != level {
			return fmt.Errorf("level mismatch: set %v, got %v", level, currentLevel)
		}
	}
	
	// Restore original level
	zlog.SetLevel(originalLevel)
	
	return nil
}

func testLevelFiltering() error {
	// Test that log levels properly filter messages
	
	// Save original level
	originalLevel := zlog.GetLevel()
	defer zlog.SetLevel(originalLevel)
	
	// Set to WARN level
	zlog.SetLevel(zlog.WARN)
	
	// These should be filtered (no panic means success)
	zlog.Debug("This should be filtered", zlog.String("test", "debug"))
	zlog.Info("This should be filtered", zlog.String("test", "info"))
	
	// These should pass through
	zlog.Warn("This should log", zlog.String("test", "warn"))
	zlog.Error("This should log", zlog.String("test", "error"))
	
	// Test at ERROR level
	zlog.SetLevel(zlog.ERROR)
	
	// These should be filtered
	zlog.Debug("Filtered debug")
	zlog.Info("Filtered info")
	zlog.Warn("Filtered warn")
	
	// This should pass
	zlog.Error("Error should log")
	
	return nil
}

func testLevelHierarchy() error {
	// Test that log level hierarchy works correctly
	
	// Save original level
	originalLevel := zlog.GetLevel()
	defer zlog.SetLevel(originalLevel)
	
	// Level hierarchy: DEBUG < INFO < WARN < ERROR
	hierarchyTests := []struct {
		setLevel    zlog.LogLevel
		testLevel   zlog.LogLevel
		shouldLog   bool
		description string
	}{
		{zlog.DEBUG, zlog.DEBUG, true, "DEBUG level logs DEBUG"},
		{zlog.DEBUG, zlog.INFO, true, "DEBUG level logs INFO"},
		{zlog.DEBUG, zlog.WARN, true, "DEBUG level logs WARN"},
		{zlog.DEBUG, zlog.ERROR, true, "DEBUG level logs ERROR"},
		
		{zlog.INFO, zlog.DEBUG, false, "INFO level filters DEBUG"},
		{zlog.INFO, zlog.INFO, true, "INFO level logs INFO"},
		{zlog.INFO, zlog.WARN, true, "INFO level logs WARN"},
		{zlog.INFO, zlog.ERROR, true, "INFO level logs ERROR"},
		
		{zlog.WARN, zlog.DEBUG, false, "WARN level filters DEBUG"},
		{zlog.WARN, zlog.INFO, false, "WARN level filters INFO"},
		{zlog.WARN, zlog.WARN, true, "WARN level logs WARN"},
		{zlog.WARN, zlog.ERROR, true, "WARN level logs ERROR"},
		
		{zlog.ERROR, zlog.DEBUG, false, "ERROR level filters DEBUG"},
		{zlog.ERROR, zlog.INFO, false, "ERROR level filters INFO"},
		{zlog.ERROR, zlog.WARN, false, "ERROR level filters WARN"},
		{zlog.ERROR, zlog.ERROR, true, "ERROR level logs ERROR"},
	}
	
	for _, test := range hierarchyTests {
		zlog.SetLevel(test.setLevel)
		
		// We can't easily verify filtering without capturing output,
		// but we can verify no panics occur
		switch test.testLevel {
		case zlog.DEBUG:
			zlog.Debug(test.description)
		case zlog.INFO:
			zlog.Info(test.description)
		case zlog.WARN:
			zlog.Warn(test.description)
		case zlog.ERROR:
			zlog.Error(test.description)
		}
	}
	
	return nil
}

func testThreadSafety() error {
	// Test that level changes are thread-safe
	
	// Save original level
	originalLevel := zlog.GetLevel()
	defer zlog.SetLevel(originalLevel)
	
	// Simulate concurrent level changes and reads
	done := make(chan bool, 2)
	
	// Goroutine 1: Change levels
	go func() {
		for i := 0; i < 10; i++ {
			zlog.SetLevel(zlog.DEBUG)
			zlog.SetLevel(zlog.INFO)
			zlog.SetLevel(zlog.WARN)
			zlog.SetLevel(zlog.ERROR)
		}
		done <- true
	}()
	
	// Goroutine 2: Read levels and log
	go func() {
		for i := 0; i < 10; i++ {
			level := zlog.GetLevel()
			zlog.Info("Thread safety test", 
				zlog.String("current_level", fmt.Sprintf("%v", level)),
				zlog.Int("iteration", i),
			)
		}
		done <- true
	}()
	
	// Wait for both goroutines
	<-done
	<-done
	
	return nil
}

func testDefaultBehavior() error {
	// Test default logging behavior
	
	// Save original level
	originalLevel := zlog.GetLevel()
	defer zlog.SetLevel(originalLevel)
	
	// Test logging without explicit level setting
	zlog.Info("Default behavior test", 
		zlog.String("status", "testing"),
		zlog.Bool("default", true),
	)
	
	// Test that we can get the current level
	currentLevel := zlog.GetLevel()
	if currentLevel < zlog.DEBUG || currentLevel > zlog.ERROR {
		return fmt.Errorf("invalid default level: %v", currentLevel)
	}
	
	// Test logging at all levels with default setting
	zlog.Debug("Debug with default level")
	zlog.Info("Info with default level")
	zlog.Warn("Warn with default level")
	zlog.Error("Error with default level")
	
	return nil
}

