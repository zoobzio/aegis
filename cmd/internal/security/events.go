package security

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	
	"aegis/capitan"
	"aegis/sctx"
)

// Event type definitions for the entire test suite
type SecurityViolationEventType string
const SecurityViolation SecurityViolationEventType = "security.violation"

type SecurityAuditEventType string
const SecurityAudit SecurityAuditEventType = "security.audit"

type ContextEventType string
const ContextEvent ContextEventType = "context.event"

// runSecurityEventTests tests security event generation capabilities
func runEventsTests(cmd *cobra.Command, args []string) {
	runTest("Security Violation Events", testSecurityViolationEvents)
	runTest("Event Categorization", testEventCategorization)
	runTest("Security Level Changes", testSecurityLevelChanges)
	runTest("Audit Trail Generation", testAuditTrailGeneration)
	runTest("Event Context Preservation", testEventContextPreservation)
	runTest("Threat Response Patterns", testThreatResponsePatterns)
}

// testSecurityViolationEvents verifies violation event creation
func testSecurityViolationEvents() error {
	// Track events
	var capturedEvent sctx.SecurityViolationEvent
	eventCaptured := false
	
	// Listen for security violations
	capitan.Listen[SecurityViolationEventType, sctx.SecurityViolationEvent](
		func(event sctx.SecurityViolationEvent) error {
			capturedEvent = event
			eventCaptured = true
			return nil
		},
	)
	
	// Create a violation event
	violation := sctx.SecurityViolationEvent{
		UserID:   "user-123",
		Resource: "admin:panel",
		Action:   "unauthorized_access",
		Severity: "high",
		Context:  map[string]interface{}{
			"ip_address": "192.168.1.100",
			"attempt":    3,
		},
		Timestamp: time.Now(),
	}
	
	// Emit the event
	capitan.Emit[SecurityViolationEventType, sctx.SecurityViolationEvent](violation)
	
	// Allow event processing
	time.Sleep(10 * time.Millisecond)
	
	// Verify event was captured
	if !eventCaptured {
		return fmt.Errorf("security violation event not captured")
	}
	
	// Verify event details
	if capturedEvent.UserID != "user-123" {
		return fmt.Errorf("user ID not preserved in event")
	}
	
	if capturedEvent.Severity != "high" {
		return fmt.Errorf("severity not preserved in event")
	}
	
	return nil
}

// testEventCategorization verifies different security event categories
func testEventCategorization() error {
	events := []sctx.SecurityError{
		{
			Type:     "unauthorized",
			Message:  "Invalid credentials",
			Resource: "/api/login",
			Action:   "login",
		},
		{
			Type:     "forbidden",
			Message:  "Insufficient permissions",
			Resource: "/api/admin",
			Action:   "access",
		},
		{
			Type:     "suspicious",
			Message:  "Multiple failed login attempts",
			Resource: "/api/login",
			Action:   "login",
		},
		{
			Type:     "breach",
			Message:  "Data exfiltration detected",
			Resource: "/api/export",
			Action:   "export",
		},
	}
	
	// Verify each category
	for _, event := range events {
		// SecurityError.Error() just returns the message
		if event.Error() != event.Message {
			return fmt.Errorf("%s event Error() should return message", event.Type)
		}
		
		// Verify all fields are set
		if event.Type == "" || event.Resource == "" || event.Action == "" {
			return fmt.Errorf("security error missing required fields")
		}
	}
	
	return nil
}

// testSecurityLevelChanges verifies security posture changes
func testSecurityLevelChanges() error {
	// Track level changes
	var levelChanges []sctx.SecurityLevelChange
	
	// Listen for level changes
	capitan.Listen[SecurityAuditEventType, sctx.SecurityLevelChange](
		func(change sctx.SecurityLevelChange) error {
			levelChanges = append(levelChanges, change)
			return nil
		},
	)
	
	// Simulate security level progression
	levels := []struct {
		from   sctx.SecurityLevel
		to     sctx.SecurityLevel
		reason string
	}{
		{sctx.GreenLevel, sctx.YellowLevel, "Unusual login pattern detected"},
		{sctx.YellowLevel, sctx.OrangeLevel, "Multiple authentication failures"},
		{sctx.OrangeLevel, sctx.RedLevel, "Potential breach detected"},
		{sctx.RedLevel, sctx.OrangeLevel, "Threat contained"},
		{sctx.OrangeLevel, sctx.GreenLevel, "System recovered"},
	}
	
	// Emit level changes
	for _, change := range levels {
		capitan.Emit[SecurityAuditEventType, sctx.SecurityLevelChange](
			sctx.SecurityLevelChange{
				Previous: change.from,
				Current:  change.to,
				Reason:   change.reason,
				Time:     time.Now(),
			},
		)
	}
	
	// Allow event processing
	time.Sleep(10 * time.Millisecond)
	
	// Verify all changes captured
	if len(levelChanges) != len(levels) {
		return fmt.Errorf("expected %d level changes, got %d", len(levels), len(levelChanges))
	}
	
	// Verify escalation pattern
	if levelChanges[0].Current != sctx.YellowLevel {
		return fmt.Errorf("first escalation should be to Yellow")
	}
	if levelChanges[2].Current != sctx.RedLevel {
		return fmt.Errorf("peak escalation should be Red")
	}
	if levelChanges[4].Current != sctx.GreenLevel {
		return fmt.Errorf("final state should be Green")
	}
	
	return nil
}

// testAuditTrailGeneration verifies comprehensive audit logging
func testAuditTrailGeneration() error {
	// Simulate a complete user session with audit events
	userCtx := sctx.NewUserContext("user-456", []string{"app:use"})
	
	// Track all security events
	var auditTrail []interface{}
	
	// Use the same event types defined at the top of the file
	
	// Listen for all security event types
	capitan.Listen[SecurityViolationEventType, sctx.SecurityViolationEvent](
		func(event sctx.SecurityViolationEvent) error {
			auditTrail = append(auditTrail, event)
			return nil
		},
	)
	
	// Define concrete audit event type
	type AuditEvent struct {
		Event   string
		UserID  string
		Time    time.Time
		Success bool
		// Additional fields as needed
		Resource  string
		Operation string
		Records   int
	}
	
	capitan.Listen[SecurityAuditEventType, AuditEvent](
		func(event AuditEvent) error {
			auditTrail = append(auditTrail, event)
			return nil
		},
	)
	
	// Generate various audit events
	
	// 1. Login attempt
	capitan.Emit[SecurityAuditEventType, AuditEvent](AuditEvent{
		Event:   "login_attempt",
		UserID:  userCtx.UserID,
		Time:    time.Now(),
		Success: true,
	})
	
	// 2. Permission check
	if !userCtx.HasPermission("admin:read") {
		capitan.Emit[SecurityViolationEventType, sctx.SecurityViolationEvent](
			sctx.SecurityViolationEvent{
				UserID:   userCtx.UserID,
				Resource: "admin:dashboard",
				Action:   "access_denied",
				Severity: "medium",
			},
		)
	}
	
	// 3. Data access
	capitan.Emit[SecurityAuditEventType, AuditEvent](AuditEvent{
		Event:     "data_access",
		UserID:    userCtx.UserID,
		Resource:  "users:list",
		Operation: "read",
		Records:   100,
		Time:      time.Now(),
	})
	
	// Allow event processing
	time.Sleep(10 * time.Millisecond)
	
	// Verify audit trail has multiple events
	if len(auditTrail) < 3 {
		return fmt.Errorf("audit trail should have at least 3 events, got %d", len(auditTrail))
	}
	
	return nil
}

// testEventContextPreservation verifies context flows through events
func testEventContextPreservation() error {
	// Create a rich context
	ctx := sctx.NewUserContext("user-789", []string{"app:use"}).
		WithExtension("tenant_id", "tenant-123").
		WithExtension("session_id", "sess-456").
		WithExtension("request_id", "req-789")
	
	// Define concrete event type with embedded context
	type ContextualEvent struct {
		Event     string
		UserID    string
		TenantID  string
		SessionID string
		RequestID string
		Timestamp time.Time
	}
	
	// Track event with context
	var capturedEvent ContextualEvent
	
	capitan.Listen[ContextEventType, ContextualEvent](
		func(event ContextualEvent) error {
			capturedEvent = event
			return nil
		},
	)
	
	// Emit event with full context
	capitan.Emit[ContextEventType, ContextualEvent](ContextualEvent{
		Event:     "sensitive_operation",
		UserID:    ctx.UserID,
		TenantID:  ctx.Extensions["tenant_id"].(string),
		SessionID: ctx.Extensions["session_id"].(string),
		RequestID: ctx.Extensions["request_id"].(string),
		Timestamp: time.Now(),
	})
	
	// Allow event processing
	time.Sleep(10 * time.Millisecond)
	
	// Verify context preserved
	if capturedEvent.UserID == "" {
		return fmt.Errorf("event not captured")
	}
	
	if capturedEvent.TenantID != "tenant-123" {
		return fmt.Errorf("tenant_id not preserved in event context")
	}
	
	if capturedEvent.RequestID != "req-789" {
		return fmt.Errorf("request_id not preserved in event context")
	}
	
	return nil
}

// testThreatResponsePatterns shows automated threat responses
func testThreatResponsePatterns() error {
	// Track automated responses
	var responses []string
	
	// Use the event type defined at the top
	
	// Set up threat response handlers
	capitan.Listen[SecurityViolationEventType, sctx.SecurityViolationEvent](
		func(event sctx.SecurityViolationEvent) error {
			// Automated response based on severity
			switch event.Severity {
			case "low":
				responses = append(responses, "logged")
			case "medium":
				responses = append(responses, "alerted")
			case "high":
				responses = append(responses, "blocked")
			case "critical":
				responses = append(responses, "lockdown")
			}
			return nil
		},
	)
	
	// Simulate threat progression
	threats := []sctx.SecurityViolationEvent{
		{Severity: "low", Action: "invalid_input", Timestamp: time.Now()},
		{Severity: "medium", Action: "repeated_failures", Timestamp: time.Now()},
		{Severity: "high", Action: "sql_injection_attempt", Timestamp: time.Now()},
		{Severity: "critical", Action: "data_breach_attempt", Timestamp: time.Now()},
	}
	
	// Emit threats
	for _, threat := range threats {
		capitan.Emit[SecurityViolationEventType, sctx.SecurityViolationEvent](threat)
	}
	
	// Allow event processing
	time.Sleep(10 * time.Millisecond)
	
	// Verify response escalation
	expectedResponses := []string{"logged", "alerted", "blocked", "lockdown"}
	if len(responses) != len(expectedResponses) {
		return fmt.Errorf("expected %d responses, got %d", len(expectedResponses), len(responses))
	}
	
	for i, expected := range expectedResponses {
		if responses[i] != expected {
			return fmt.Errorf("expected response %s, got %s", expected, responses[i])
		}
	}
	
	return nil
}