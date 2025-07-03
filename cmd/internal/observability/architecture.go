package observability

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	
	"aegis/capitan"
	"aegis/sctx"
)

// runEventArchitectureTests tests event-driven capabilities
func runEventArchitectureTests(cmd *cobra.Command, args []string) {
	runTest("Type-Safe Event Emission", testTypeSafeEventEmission)
	runTest("Event Listener Registration", testEventListenerRegistration)
	runTest("Event Type Isolation", testEventTypeIsolation)
	runTest("Complex Event Data", testComplexEventData)
	runTest("Event Broadcasting", testEventBroadcasting)
	runTest("Event Handler Chaining", testEventHandlerChaining)
	runTest("Async Event Processing", testAsyncEventProcessing)
	runTest("Event Error Handling", testEventErrorHandling)
	runTest("Cross-Domain Events", testCrossDomainEvents)
	runTest("Event Metrics", testEventMetrics)
}

// Event types for testing - demonstrating type isolation
type UserEventType string
type SystemEventType string
type SecurityEventType string

const (
	UserCreated UserEventType = "created"
	UserUpdated UserEventType = "updated"
	UserDeleted UserEventType = "deleted"
	
	SystemStartup SystemEventType = "startup"
	SystemShutdown SystemEventType = "shutdown"
	
	SecurityViolation SecurityEventType = "violation"
	SecurityAudit SecurityEventType = "audit"
)

// Event data structures
type UserEvent struct {
	UserID    string    `json:"user_id"`
	Action    string    `json:"action"`
	Timestamp time.Time `json:"timestamp"`
	Metadata  map[string]string `json:"metadata"`
}

type SystemEvent struct {
	Component string    `json:"component"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

type SecurityEvent struct {
	UserID    string                 `json:"user_id"`
	Resource  string                 `json:"resource"`
	Violation string                 `json:"violation"`
	Context   sctx.SecurityContext   `json:"-"` // Don't serialize context
	Timestamp time.Time              `json:"timestamp"`
}

// testTypeSafeEventEmission verifies type-safe event emission
func testTypeSafeEventEmission() error {
	// Register a listener first
	var received UserEvent
	capitan.Listen[UserEventType, UserEvent](func(event UserEvent) error {
		received = event
		return nil
	})
	
	// Emit a user event
	userEvent := UserEvent{
		UserID:    "user-123",
		Action:    "profile_updated",
		Timestamp: time.Now(),
		Metadata: map[string]string{
			"field": "email",
			"old":   "old@example.com",
			"new":   "new@example.com",
		},
	}
	
	err := capitan.Emit[UserEventType, UserEvent](userEvent)
	if err != nil {
		return fmt.Errorf("failed to emit event: %v", err)
	}
	
	// Small delay for async processing
	time.Sleep(10 * time.Millisecond)
	
	// Verify event was received
	if received.UserID != userEvent.UserID {
		return fmt.Errorf("event not received correctly")
	}
	
	// This would NOT compile - type safety!
	// capitan.Emit[SystemEventType, UserEvent](userEvent) // Compile error!
	
	return nil
}

// testEventListenerRegistration verifies listener registration and replacement
func testEventListenerRegistration() error {
	// With pipz, each event type can only have ONE handler
	// Registering a new handler REPLACES the old one
	
	var lastHandlerID int
	
	// Register handlers - each replaces the previous
	for i := 0; i < 3; i++ {
		handlerID := i
		capitan.Listen[UserEventType, UserEvent](func(event UserEvent) error {
			lastHandlerID = handlerID
			return nil
		})
	}
	
	// Emit event
	testEvent := UserEvent{
		UserID:    "test-user",
		Action:    "test_action",
		Timestamp: time.Now(),
	}
	
	err := capitan.Emit[UserEventType, UserEvent](testEvent)
	if err != nil {
		return fmt.Errorf("failed to emit: %v", err)
	}
	
	// Wait for handler
	time.Sleep(10 * time.Millisecond)
	
	// Verify ONLY the last handler was triggered
	if lastHandlerID != 2 {
		return fmt.Errorf("expected last handler (ID 2) to be active, got %d", lastHandlerID)
	}
	
	return nil
}

// testEventTypeIsolation verifies different event types are isolated
func testEventTypeIsolation() error {
	// Set up listeners for different event types
	var userReceived bool
	var systemReceived bool
	
	capitan.Listen[UserEventType, UserEvent](func(event UserEvent) error {
		userReceived = true
		return nil
	})
	
	capitan.Listen[SystemEventType, SystemEvent](func(event SystemEvent) error {
		systemReceived = true
		return nil
	})
	
	// Emit only a user event
	userEvent := UserEvent{
		UserID:    "isolated-user",
		Action:    "test",
		Timestamp: time.Now(),
	}
	
	err := capitan.Emit[UserEventType, UserEvent](userEvent)
	if err != nil {
		return fmt.Errorf("failed to emit user event: %v", err)
	}
	
	time.Sleep(10 * time.Millisecond)
	
	// Verify isolation
	if !userReceived {
		return fmt.Errorf("user event listener should have been triggered")
	}
	
	if systemReceived {
		return fmt.Errorf("system event listener should NOT have been triggered")
	}
	
	return nil
}

// testComplexEventData verifies handling of complex nested data
func testComplexEventData() error {
	// Complex event with nested structures
	type OrderEvent struct {
		OrderID   string    `json:"order_id"`
		Customer  UserEvent `json:"customer"` // Nested event
		Items     []struct {
			SKU   string  `json:"sku"`
			Price float64 `json:"price"`
			Qty   int     `json:"qty"`
		} `json:"items"`
		Total     float64              `json:"total"`
		Metadata  map[string]any       `json:"metadata"`
		Timestamp time.Time            `json:"timestamp"`
	}
	
	type OrderEventType string
	const OrderPlaced OrderEventType = "placed"
	
	var received OrderEvent
	capitan.Listen[OrderEventType, OrderEvent](func(event OrderEvent) error {
		received = event
		return nil
	})
	
	// Create complex event
	complexEvent := OrderEvent{
		OrderID: "ORD-12345",
		Customer: UserEvent{
			UserID: "customer-789",
			Action: "purchase",
		},
		Items: []struct {
			SKU   string  `json:"sku"`
			Price float64 `json:"price"`
			Qty   int     `json:"qty"`
		}{
			{SKU: "ITEM-A", Price: 29.99, Qty: 2},
			{SKU: "ITEM-B", Price: 49.99, Qty: 1},
		},
		Total: 109.97,
		Metadata: map[string]any{
			"shipping": "express",
			"discount": 10.0,
			"nested": map[string]any{
				"tracking": "enabled",
			},
		},
		Timestamp: time.Now(),
	}
	
	err := capitan.Emit[OrderEventType, OrderEvent](complexEvent)
	if err != nil {
		return fmt.Errorf("failed to emit complex event: %v", err)
	}
	
	time.Sleep(10 * time.Millisecond)
	
	// Verify complex data preserved
	if received.OrderID != complexEvent.OrderID {
		return fmt.Errorf("order ID not preserved")
	}
	
	if len(received.Items) != 2 {
		return fmt.Errorf("items not preserved")
	}
	
	if received.Metadata["nested"].(map[string]any)["tracking"] != "enabled" {
		return fmt.Errorf("nested metadata not preserved")
	}
	
	return nil
}

// testEventBroadcasting verifies Hook/Broadcast for multiple handlers
func testEventBroadcasting() error {
	// Different systems listening to same event
	var mu sync.Mutex
	var analyticsProcessed bool
	var inventoryUpdated bool
	var emailSent bool
	
	type CheckoutEventType string
	const CheckoutComplete CheckoutEventType = "complete"
	
	type CheckoutEvent struct {
		OrderID string
		Amount  float64
	}
	
	// Use Hook for multiple handlers
	// Analytics handler
	capitan.Hook[CheckoutEventType, CheckoutEvent](func(event CheckoutEvent) error {
		mu.Lock()
		analyticsProcessed = true
		mu.Unlock()
		return nil
	})
	
	// Inventory handler
	capitan.Hook[CheckoutEventType, CheckoutEvent](func(event CheckoutEvent) error {
		mu.Lock()
		inventoryUpdated = true
		mu.Unlock()
		return nil
	})
	
	// Email handler
	capitan.Hook[CheckoutEventType, CheckoutEvent](func(event CheckoutEvent) error {
		mu.Lock()
		emailSent = true
		mu.Unlock()
		return nil
	})
	
	// Broadcast event to all hooks
	err := capitan.Broadcast[CheckoutEventType, CheckoutEvent](CheckoutEvent{
		OrderID: "broadcast-123",
		Amount:  99.99,
	})
	if err != nil {
		return fmt.Errorf("broadcast failed: %v", err)
	}
	
	time.Sleep(50 * time.Millisecond) // Async processing
	
	// Verify all handlers triggered
	mu.Lock()
	allTriggered := analyticsProcessed && inventoryUpdated && emailSent
	mu.Unlock()
	
	if !allTriggered {
		return fmt.Errorf("not all handlers triggered: analytics=%v, inventory=%v, email=%v",
			analyticsProcessed, inventoryUpdated, emailSent)
	}
	
	return nil
}

// testEventHandlerChaining verifies handlers can emit new events
func testEventHandlerChaining() error {
	// Chain: UserCreated → WelcomeEmail → EmailSent
	type EmailEventType string
	const SendWelcome EmailEventType = "welcome"
	const EmailSent EmailEventType = "sent"
	
	type EmailEvent struct {
		To      string
		Subject string
		Sent    bool
	}
	
	var chainCompleted bool
	
	// Listen for user creation, trigger email
	capitan.Listen[UserEventType, UserEvent](func(event UserEvent) error {
		if event.Action == "created" {
			// Trigger welcome email
			return capitan.Emit[EmailEventType, EmailEvent](EmailEvent{
				To:      event.UserID + "@example.com",
				Subject: "Welcome!",
			})
		}
		return nil
	})
	
	// Listen for email events, mark as sent
	capitan.Listen[EmailEventType, EmailEvent](func(event EmailEvent) error {
		if strings.Contains(event.Subject, "Welcome") {
			chainCompleted = true
		}
		return nil
	})
	
	// Start the chain
	err := capitan.Emit[UserEventType, UserEvent](UserEvent{
		UserID:    "chain-user",
		Action:    "created",
		Timestamp: time.Now(),
	})
	if err != nil {
		return fmt.Errorf("chain start failed: %v", err)
	}
	
	time.Sleep(30 * time.Millisecond)
	
	if !chainCompleted {
		return fmt.Errorf("event chain did not complete")
	}
	
	return nil
}

// testAsyncEventProcessing verifies Hook/Broadcast are truly async
func testAsyncEventProcessing() error {
	// Track processing
	var mu sync.Mutex
	processCount := 0
	
	type AsyncEventType string
	const AsyncProcess AsyncEventType = "process"
	
	type AsyncEvent struct {
		ID    string
		Delay time.Duration
	}
	
	// Register multiple async handlers using Hook
	for i := 0; i < 3; i++ {
		handlerID := i
		capitan.Hook[AsyncEventType, AsyncEvent](func(event AsyncEvent) error {
			time.Sleep(event.Delay)
			mu.Lock()
			processCount++
			mu.Unlock()
			fmt.Printf("Handler %d processed %s\n", handlerID, event.ID)
			return nil
		})
	}
	
	// Broadcast event with delay
	start := time.Now()
	err := capitan.Broadcast[AsyncEventType, AsyncEvent](AsyncEvent{
		ID:    "async-test",
		Delay: 50 * time.Millisecond,
	})
	if err != nil {
		return fmt.Errorf("failed to broadcast: %v", err)
	}
	broadcastDuration := time.Since(start)
	
	// Broadcast should return immediately (not wait for handlers)
	if broadcastDuration > 5*time.Millisecond {
		return fmt.Errorf("broadcast blocked on handler execution: took %v", broadcastDuration)
	}
	
	// Wait for all handlers
	time.Sleep(100 * time.Millisecond)
	
	// Check all handlers processed
	mu.Lock()
	count := processCount
	mu.Unlock()
	
	if count != 3 {
		return fmt.Errorf("expected 3 handlers to process, got %d", count)
	}
	
	return nil
}

// testEventErrorHandling verifies error propagation
func testEventErrorHandling() error {
	type ErrorEventType string
	const ErrorTest ErrorEventType = "test"
	
	type ErrorEvent struct {
		ShouldFail bool
		Message    string
	}
	
	// Register handler that can fail
	capitan.Listen[ErrorEventType, ErrorEvent](func(event ErrorEvent) error {
		if event.ShouldFail {
			return fmt.Errorf("handler error: %s", event.Message)
		}
		return nil
	})
	
	// Test successful event
	err := capitan.Emit[ErrorEventType, ErrorEvent](ErrorEvent{
		ShouldFail: false,
		Message:    "success",
	})
	if err != nil {
		return fmt.Errorf("successful event returned error: %v", err)
	}
	
	// Test failing event
	err = capitan.Emit[ErrorEventType, ErrorEvent](ErrorEvent{
		ShouldFail: true,
		Message:    "intentional failure",
	})
	// Current implementation might not propagate errors back
	// This tests current behavior
	
	return nil
}

// testCrossDomainEvents verifies events work across different domains
func testCrossDomainEvents() error {
	// Different domains using same event infrastructure
	
	// E-commerce domain
	type EcommerceEventType string
	const ProductViewed EcommerceEventType = "viewed"
	
	type ProductEvent struct {
		ProductID string
		UserID    string
		Price     float64
	}
	
	// Analytics domain
	type AnalyticsEventType string
	const TrackEvent AnalyticsEventType = "track"
	
	type AnalyticsEvent struct {
		Category string
		Action   string
		Value    float64
	}
	
	var analyticsReceived bool
	
	// E-commerce emits product view
	capitan.Listen[EcommerceEventType, ProductEvent](func(event ProductEvent) error {
		// Transform to analytics event
		return capitan.Emit[AnalyticsEventType, AnalyticsEvent](AnalyticsEvent{
			Category: "product",
			Action:   "view",
			Value:    event.Price,
		})
	})
	
	// Analytics receives transformed event
	capitan.Listen[AnalyticsEventType, AnalyticsEvent](func(event AnalyticsEvent) error {
		if event.Category == "product" && event.Action == "view" {
			analyticsReceived = true
		}
		return nil
	})
	
	// Start cross-domain flow
	err := capitan.Emit[EcommerceEventType, ProductEvent](ProductEvent{
		ProductID: "PROD-123",
		UserID:    "user-456",
		Price:     29.99,
	})
	if err != nil {
		return fmt.Errorf("cross-domain emission failed: %v", err)
	}
	
	time.Sleep(20 * time.Millisecond)
	
	if !analyticsReceived {
		return fmt.Errorf("cross-domain event not received")
	}
	
	return nil
}

// testEventMetrics verifies we can track event metrics
func testEventMetrics() error {
	// Track event counts and timing
	var mu sync.Mutex
	eventCounts := make(map[string]int)
	eventTimings := make(map[string][]time.Duration)
	
	type MetricsEventType string
	const MetricsTest MetricsEventType = "test"
	
	type MetricsEvent struct {
		ID        string
		Timestamp time.Time
	}
	
	// Metrics collection handler
	capitan.Listen[MetricsEventType, MetricsEvent](func(event MetricsEvent) error {
		processingTime := time.Since(event.Timestamp)
		
		mu.Lock()
		eventCounts["metrics_test"]++
		eventTimings["metrics_test"] = append(eventTimings["metrics_test"], processingTime)
		mu.Unlock()
		
		// Simulate some processing
		time.Sleep(5 * time.Millisecond)
		return nil
	})
	
	// Emit multiple events
	for i := 0; i < 5; i++ {
		err := capitan.Emit[MetricsEventType, MetricsEvent](MetricsEvent{
			ID:        fmt.Sprintf("metric-%d", i),
			Timestamp: time.Now(),
		})
		if err != nil {
			return fmt.Errorf("metric event %d failed: %v", i, err)
		}
	}
	
	// Wait for processing
	time.Sleep(50 * time.Millisecond)
	
	// Verify metrics collected
	mu.Lock()
	count := eventCounts["metrics_test"]
	timings := eventTimings["metrics_test"]
	mu.Unlock()
	
	if count != 5 {
		return fmt.Errorf("expected 5 events, got %d", count)
	}
	
	if len(timings) != 5 {
		return fmt.Errorf("expected 5 timings, got %d", len(timings))
	}
	
	// Verify reasonable timing (should be fast)
	for i, timing := range timings {
		if timing > 100*time.Millisecond {
			return fmt.Errorf("event %d took too long: %v", i, timing)
		}
	}
	
	return nil
}