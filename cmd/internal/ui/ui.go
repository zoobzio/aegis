package ui

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/alecthomas/chroma/quick"
)

var (
	// Color definitions for consistent theming
	HeaderColor    = color.New(color.FgCyan, color.Bold)
	SectionColor   = color.New(color.FgYellow, color.Bold)
	SuccessColor   = color.New(color.FgGreen)
	ErrorColor     = color.New(color.FgRed, color.Bold)
	InfoColor      = color.New(color.FgBlue)
	CodeColor      = color.New(color.FgMagenta)
	
	// Speed settings
	TypewriterSpeed = 30 * time.Millisecond  // Per character
	PauseShort      = 800 * time.Millisecond
	PauseMedium     = 1200 * time.Millisecond
	PauseLong       = 2000 * time.Millisecond
)

// TypewriterPrint displays text with a typewriter effect
func TypewriterPrint(text string, slow bool) {
	speed := TypewriterSpeed
	if slow {
		speed = 50 * time.Millisecond // Slower for cinema mode
	}
	
	for _, char := range text {
		fmt.Print(string(char))
		if char != ' ' && char != '\n' {
			time.Sleep(speed)
		}
	}
}

// PrintHeader displays a major section header with formatting
func PrintHeader(title string) {
	fmt.Println()
	HeaderColor.Printf("🚀 %s 🚀\n", title)
	fmt.Println(strings.Repeat("=", len(title)+8))
	time.Sleep(PauseShort)
}

// PrintSection displays a subsection with formatting
func PrintSection(title, subtitle string, slow bool) {
	fmt.Println()
	SectionColor.Printf("🔷 %s\n", title)
	if subtitle != "" {
		InfoColor.Printf("   %s\n", subtitle)
	}
	fmt.Println()
	
	if slow {
		time.Sleep(PauseLong)
	} else {
		time.Sleep(PauseShort)
	}
}

// PrintStep displays a step with progressive reveal
func PrintStep(message string, slow bool) {
	fmt.Print("   💫 ")
	TypewriterPrint(message, slow)
	fmt.Println()
	
	if slow {
		time.Sleep(PauseMedium)
	} else {
		time.Sleep(PauseShort)
	}
}

// PrintSuccess displays success message with color
func PrintSuccess(message string) {
	SuccessColor.Printf("   ✨ %s\n", message)
	time.Sleep(PauseShort)
}

// PrintError displays error message with color
func PrintError(message string) {
	ErrorColor.Printf("   ❌ %s\n", message)
}

// PrintJSON displays JSON with syntax highlighting and pretty printing
func PrintJSON(data []byte, label string, slow bool) {
	// Pretty print the JSON
	var prettyJSON interface{}
	if err := json.Unmarshal(data, &prettyJSON); err != nil {
		PrintError(fmt.Sprintf("Failed to parse JSON: %v", err))
		return
	}
	
	prettyBytes, err := json.MarshalIndent(prettyJSON, "   ", "  ")
	if err != nil {
		PrintError(fmt.Sprintf("Failed to format JSON: %v", err))
		return
	}
	
	fmt.Printf("   📄 %s:\n", label)
	
	// Try syntax highlighting (fallback to plain if highlighting fails)
	highlighted := string(prettyBytes)
	if err := quick.Highlight(os.Stdout, string(prettyBytes), "json", "terminal256", "monokai"); err == nil {
		// Highlighting worked, add some spacing
		fmt.Println()
	} else {
		// Fallback to colored plain text
		CodeColor.Printf("   %s\n", highlighted)
	}
	
	if slow {
		time.Sleep(PauseLong)
	} else {
		time.Sleep(PauseMedium)
	}
}

// WaitForUser pauses for user interaction in interactive mode
func WaitForUser(interactive bool, message string, slow bool) {
	if interactive {
		InfoColor.Print(message)
		reader := bufio.NewReader(os.Stdin)
		reader.ReadLine()
	} else if slow {
		time.Sleep(PauseLong)
	} else {
		time.Sleep(PauseShort)
	}
}

// PrintFooter displays completion message
func PrintFooter() {
	fmt.Println()
	SuccessColor.Println("✨ Demo complete! The type system IS the distributed architecture! ✨")
	fmt.Println()
}

// PrintCode displays code samples with syntax highlighting
func PrintCode(code, language string, slow bool) {
	fmt.Println("   🔧 Code sample:")
	
	if err := quick.Highlight(os.Stdout, code, language, "terminal256", "monokai"); err != nil {
		// Fallback to plain colored text
		CodeColor.Printf("   %s\n", code)
	}
	
	if slow {
		time.Sleep(PauseLong)
	} else {
		time.Sleep(PauseMedium)
	}
}

// PrintComparison shows before/after or different views side by side
func PrintComparison(left, right []byte, leftLabel, rightLabel string, slow bool) {
	fmt.Printf("   📊 Comparison: %s vs %s\n", leftLabel, rightLabel)
	fmt.Println("   " + strings.Repeat("-", 50))
	
	// Print left side
	fmt.Printf("   👈 %s:\n", leftLabel)
	PrintJSON(left, "", false)
	
	// Print right side  
	fmt.Printf("   👉 %s:\n", rightLabel)
	PrintJSON(right, "", false)
	
	if slow {
		time.Sleep(PauseLong)
	} else {
		time.Sleep(PauseMedium)
	}
}