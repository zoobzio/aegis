package observability

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	
	"aegis/capitan"
	"aegis/sctx"
)

// runCommandAuthorizationTests tests command pattern with security
func runCommandAuthorizationTests(cmd *cobra.Command, args []string) {
	runTest("Command Authorization", testCommandAuthorization)
	runTest("Role-Based Commands", testRoleBasedCommands)
	runTest("Command Audit Trail", testCommandAuditTrail)
}

// Command types with naval hierarchy
type CaptainOrder string
type LieutenantOrder string
type CrewOrder string

const (
	NavigateShip CaptainOrder = "navigate"
	SetCourse CaptainOrder = "set_course"
	
	ManageCrew LieutenantOrder = "manage_crew"
	AssignDuties LieutenantOrder = "assign_duties"
	
	ExecuteTask CrewOrder = "execute"
)

// Command data
type NavigationCommand struct {
	Heading  int
	Speed    int
	IssuedBy sctx.SecurityContext
}

type CrewCommand struct {
	Task     string
	Assignee string
	IssuedBy sctx.SecurityContext
}

// testCommandAuthorization verifies only authorized contexts can issue commands
func testCommandAuthorization() error {
	var lastCommand NavigationCommand
	var commandExecuted bool
	
	// Register handler that checks authorization
	capitan.Listen[CaptainOrder, NavigationCommand](func(cmd NavigationCommand) error {
		// Only captain can navigate
		if !cmd.IssuedBy.HasPermission("captain:navigate") {
			return fmt.Errorf("unauthorized: only captain can navigate")
		}
		lastCommand = cmd
		commandExecuted = true
		return nil
	})
	
	// Try with unauthorized crew member
	crewCtx := sctx.NewUserContext("crew-member-1", []string{"crew:work"})
	err := capitan.Emit[CaptainOrder, NavigationCommand](NavigationCommand{
		Heading:  90,
		Speed:    20,
		IssuedBy: crewCtx,
	})
	// Should get authorization error
	if err == nil {
		return fmt.Errorf("expected authorization error for crew member")
	}
	
	if commandExecuted {
		return fmt.Errorf("unauthorized command should not have been executed")
	}
	
	// Try with authorized captain
	captainCtx := sctx.NewUserContext("captain-smith", []string{"captain:navigate", "captain:command"})
	err = capitan.Emit[CaptainOrder, NavigationCommand](NavigationCommand{
		Heading:  180,
		Speed:    15,
		IssuedBy: captainCtx,
	})
	if err != nil {
		return fmt.Errorf("captain emit failed: %v", err)
	}
	
	time.Sleep(10 * time.Millisecond)
	
	if !commandExecuted || lastCommand.Heading != 180 {
		return fmt.Errorf("captain command should have been executed")
	}
	
	return nil
}

// testRoleBasedCommands shows different roles handle different commands
func testRoleBasedCommands() error {
	var captainCommandReceived bool
	var lieutenantCommandReceived bool
	
	// Only one captain handles navigation
	capitan.Listen[CaptainOrder, NavigationCommand](func(cmd NavigationCommand) error {
		if cmd.IssuedBy.UserID == "captain-1" {
			captainCommandReceived = true
		}
		return nil
	})
	
	// Only one lieutenant handles crew management
	capitan.Listen[LieutenantOrder, CrewCommand](func(cmd CrewCommand) error {
		if cmd.IssuedBy.UserID == "lieutenant-1" {
			lieutenantCommandReceived = true
		}
		return nil
	})
	
	// Captain issues navigation order
	captainCtx := sctx.NewUserContext("captain-1", []string{"captain:all"})
	capitan.Emit[CaptainOrder, NavigationCommand](NavigationCommand{
		Heading:  270,
		Speed:    10,
		IssuedBy: captainCtx,
	})
	
	// Lieutenant issues crew order
	lieutenantCtx := sctx.NewUserContext("lieutenant-1", []string{"lieutenant:crew"})
	capitan.Emit[LieutenantOrder, CrewCommand](CrewCommand{
		Task:     "clean_deck",
		Assignee: "crew-5",
		IssuedBy: lieutenantCtx,
	})
	
	time.Sleep(20 * time.Millisecond)
	
	if !captainCommandReceived {
		return fmt.Errorf("captain command not received")
	}
	
	if !lieutenantCommandReceived {
		return fmt.Errorf("lieutenant command not received")
	}
	
	return nil
}

// testCommandAuditTrail shows commands can trigger audit events
func testCommandAuditTrail() error {
	type AuditEvent struct {
		CommandType string
		IssuedBy    string
		Timestamp   time.Time
		Success     bool
	}
	
	type AuditEventType string
	const CommandAudit AuditEventType = "audit"
	
	var auditCount int
	
	// Hook into audit events (multiple handlers allowed)
	capitan.Hook[AuditEventType, AuditEvent](func(event AuditEvent) error {
		auditCount++
		fmt.Printf("AUDIT: %s issued %s at %v\n", event.IssuedBy, event.CommandType, event.Timestamp)
		return nil
	})
	
	// Command handler that emits audit events
	capitan.Listen[CaptainOrder, NavigationCommand](func(cmd NavigationCommand) error {
		// Process command
		success := cmd.IssuedBy.HasPermission("captain:navigate")
		
		// Emit audit event (using Broadcast for multiple audit systems)
		capitan.Broadcast[AuditEventType, AuditEvent](AuditEvent{
			CommandType: "navigation",
			IssuedBy:    cmd.IssuedBy.UserID,
			Timestamp:   time.Now(),
			Success:     success,
		})
		
		if !success {
			return fmt.Errorf("unauthorized")
		}
		return nil
	})
	
	// Issue a command
	captainCtx := sctx.NewUserContext("captain-jones", []string{"captain:navigate"})
	capitan.Emit[CaptainOrder, NavigationCommand](NavigationCommand{
		Heading:  45,
		Speed:    25,
		IssuedBy: captainCtx,
	})
	
	time.Sleep(50 * time.Millisecond) // Allow async audit
	
	if auditCount != 1 {
		return fmt.Errorf("expected 1 audit event, got %d", auditCount)
	}
	
	return nil
}