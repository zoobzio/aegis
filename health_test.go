package aegis

import (
	"context"
	"fmt"
	"testing"
)

func TestHealthInfo(t *testing.T) {
	health := NewHealthInfo()
	
	if health.Status != HealthStatusUnknown {
		t.Errorf("expected initial status %s, got %s", HealthStatusUnknown, health.Status)
	}
	
	if !health.IsHealthy() == false {
		t.Error("new health info should not be healthy")
	}
}

func TestHealthInfoUpdate(t *testing.T) {
	health := NewHealthInfo()
	
	health.Update(HealthStatusHealthy, "All good", nil)
	
	status, _, message, errMsg := health.Get()
	if status != HealthStatusHealthy {
		t.Errorf("expected status %s, got %s", HealthStatusHealthy, status)
	}
	if message != "All good" {
		t.Errorf("expected message 'All good', got '%s'", message)
	}
	if errMsg != "" {
		t.Errorf("expected no error message, got '%s'", errMsg)
	}
	if !health.IsHealthy() {
		t.Error("health should be healthy")
	}
}

func TestHealthInfoUpdateWithError(t *testing.T) {
	health := NewHealthInfo()
	
	testErr := fmt.Errorf("test error")
	health.Update(HealthStatusUnhealthy, "Something went wrong", testErr)
	
	status, _, message, errMsg := health.Get()
	if status != HealthStatusUnhealthy {
		t.Errorf("expected status %s, got %s", HealthStatusUnhealthy, status)
	}
	if message != "Something went wrong" {
		t.Errorf("expected message 'Something went wrong', got '%s'", message)
	}
	if errMsg != "test error" {
		t.Errorf("expected error message 'test error', got '%s'", errMsg)
	}
	if health.IsHealthy() {
		t.Error("health should not be healthy")
	}
}

func TestPingHealthChecker(t *testing.T) {
	checker := NewPingHealthChecker("test-ping")
	
	if checker.Name() != "test-ping" {
		t.Errorf("expected name 'test-ping', got '%s'", checker.Name())
	}
	
	ctx := context.Background()
	err := checker.Check(ctx)
	if err != nil {
		t.Errorf("ping health checker should not return error, got: %v", err)
	}
}

func TestFunctionHealthChecker(t *testing.T) {
	checkFn := func(ctx context.Context) error {
		return nil
	}
	
	checker := NewFunctionHealthChecker("test-function", checkFn)
	
	if checker.Name() != "test-function" {
		t.Errorf("expected name 'test-function', got '%s'", checker.Name())
	}
	
	ctx := context.Background()
	err := checker.Check(ctx)
	if err != nil {
		t.Errorf("function health checker should not return error, got: %v", err)
	}
}

func TestFunctionHealthCheckerWithError(t *testing.T) {
	checkFn := func(ctx context.Context) error {
		return fmt.Errorf("health check failed")
	}
	
	checker := NewFunctionHealthChecker("test-function", checkFn)
	
	ctx := context.Background()
	err := checker.Check(ctx)
	if err == nil {
		t.Error("function health checker should return error")
	}
	if err.Error() != "health check failed" {
		t.Errorf("expected error 'health check failed', got '%s'", err.Error())
	}
}

func TestFunctionHealthCheckerNilFunction(t *testing.T) {
	checker := NewFunctionHealthChecker("test-nil", nil)
	
	ctx := context.Background()
	err := checker.Check(ctx)
	if err == nil {
		t.Error("function health checker with nil function should return error")
	}
}

func TestNodeHealthMethods(t *testing.T) {
	node := NewNode("test-node", "Test Node", NodeTypeGeneric, "localhost:8080")
	
	if node.Health == nil {
		t.Error("new node should have health initialized")
	}
	
	if node.IsHealthy() {
		t.Error("new node should not be healthy initially")
	}
	
	status, message := node.GetHealth()
	if status != HealthStatusUnknown {
		t.Errorf("expected initial status %s, got %s", HealthStatusUnknown, status)
	}
	
	node.SetHealth(HealthStatusHealthy, "Node is healthy", nil)
	
	if !node.IsHealthy() {
		t.Error("node should be healthy after setting health")
	}
	
	status, message = node.GetHealth()
	if status != HealthStatusHealthy {
		t.Errorf("expected status %s, got %s", HealthStatusHealthy, status)
	}
	if message != "Node is healthy" {
		t.Errorf("expected message 'Node is healthy', got '%s'", message)
	}
}

func TestNodeCheckHealth(t *testing.T) {
	node := NewNode("test-node", "Test Node", NodeTypeGeneric, "localhost:8080")
	checker := NewPingHealthChecker("test-ping")
	
	ctx := context.Background()
	err := node.CheckHealth(ctx, checker)
	if err != nil {
		t.Errorf("health check should succeed, got error: %v", err)
	}
	
	if !node.IsHealthy() {
		t.Error("node should be healthy after successful check")
	}
}

func TestNodeCheckHealthWithError(t *testing.T) {
	node := NewNode("test-node", "Test Node", NodeTypeGeneric, "localhost:8080")
	
	checkFn := func(ctx context.Context) error {
		return fmt.Errorf("health check failed")
	}
	checker := NewFunctionHealthChecker("failing-check", checkFn)
	
	ctx := context.Background()
	err := node.CheckHealth(ctx, checker)
	if err == nil {
		t.Error("health check should fail")
	}
	
	if node.IsHealthy() {
		t.Error("node should not be healthy after failed check")
	}
}

func TestNodeCheckHealthNilChecker(t *testing.T) {
	node := NewNode("test-node", "Test Node", NodeTypeGeneric, "localhost:8080")
	
	ctx := context.Background()
	err := node.CheckHealth(ctx, nil)
	if err == nil {
		t.Error("health check with nil checker should fail")
	}
	
	if node.IsHealthy() {
		t.Error("node should not be healthy after nil checker")
	}
}