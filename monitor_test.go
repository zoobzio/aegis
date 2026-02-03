package aegis

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNewHealthMonitor(t *testing.T) {
	node := NewNode("test-node", "Test Node", NodeTypeGeneric, "localhost:8080")
	checker := NewPingHealthChecker("test-ping")
	interval := 5 * time.Second
	
	monitor := NewHealthMonitor(node, checker, interval)
	
	if monitor.node != node {
		t.Error("monitor should have correct node reference")
	}
	if monitor.checker != checker {
		t.Error("monitor should have correct checker reference")
	}
	if monitor.interval != interval {
		t.Errorf("expected interval %v, got %v", interval, monitor.interval)
	}
	if monitor.IsRunning() {
		t.Error("new monitor should not be running")
	}
}

func TestHealthMonitorStart(t *testing.T) {
	node := NewNode("test-node", "Test Node", NodeTypeGeneric, "localhost:8080")
	checker := NewPingHealthChecker("test-ping")
	interval := 100 * time.Millisecond
	
	monitor := NewHealthMonitor(node, checker, interval)
	
	err := monitor.Start()
	if err != nil {
		t.Errorf("failed to start monitor: %v", err)
	}
	
	if !monitor.IsRunning() {
		t.Error("monitor should be running after start")
	}
	
	time.Sleep(150 * time.Millisecond)
	
	if !node.IsHealthy() {
		t.Error("node should be healthy after monitor checks")
	}
	
	monitor.Stop()
	
	if monitor.IsRunning() {
		t.Error("monitor should not be running after stop")
	}
}

func TestHealthMonitorStartAlreadyRunning(t *testing.T) {
	node := NewNode("test-node", "Test Node", NodeTypeGeneric, "localhost:8080")
	checker := NewPingHealthChecker("test-ping")
	interval := 1 * time.Second
	
	monitor := NewHealthMonitor(node, checker, interval)
	
	err := monitor.Start()
	if err != nil {
		t.Errorf("failed to start monitor: %v", err)
	}
	defer monitor.Stop()
	
	err = monitor.Start()
	if err == nil {
		t.Error("starting already running monitor should return error")
	}
}

func TestHealthMonitorStartNilNode(t *testing.T) {
	checker := NewPingHealthChecker("test-ping")
	interval := 1 * time.Second
	
	monitor := NewHealthMonitor(nil, checker, interval)
	
	err := monitor.Start()
	if err == nil {
		t.Error("starting monitor with nil node should return error")
	}
}

func TestHealthMonitorStartNilChecker(t *testing.T) {
	node := NewNode("test-node", "Test Node", NodeTypeGeneric, "localhost:8080")
	interval := 1 * time.Second
	
	monitor := NewHealthMonitor(node, nil, interval)
	
	err := monitor.Start()
	if err == nil {
		t.Error("starting monitor with nil checker should return error")
	}
}

func TestHealthMonitorWithFailingChecker(t *testing.T) {
	node := NewNode("test-node", "Test Node", NodeTypeGeneric, "localhost:8080")
	
	checkFn := func(ctx context.Context) error {
		return fmt.Errorf("health check failed")
	}
	checker := NewFunctionHealthChecker("failing-check", checkFn)
	interval := 100 * time.Millisecond
	
	monitor := NewHealthMonitor(node, checker, interval)
	
	err := monitor.Start()
	if err != nil {
		t.Errorf("failed to start monitor: %v", err)
	}
	defer monitor.Stop()
	
	time.Sleep(150 * time.Millisecond)
	
	if node.IsHealthy() {
		t.Error("node should not be healthy with failing checker")
	}
	
	running, status, _ := monitor.GetStatus()
	if !running {
		t.Error("monitor should be running")
	}
	if status != HealthStatusUnhealthy {
		t.Errorf("expected status %s, got %s", HealthStatusUnhealthy, status)
	}
}

func TestHealthMonitorTimeout(t *testing.T) {
	node := NewNode("test-node", "Test Node", NodeTypeGeneric, "localhost:8080")
	checker := NewPingHealthChecker("test-ping")
	interval := 1 * time.Second
	
	monitor := NewHealthMonitor(node, checker, interval)
	monitor.SetTimeout(500 * time.Millisecond)
	
	err := monitor.Start()
	if err != nil {
		t.Errorf("failed to start monitor: %v", err)
	}
	defer monitor.Stop()
	
	if !monitor.IsRunning() {
		t.Error("monitor should be running")
	}
}

func TestHealthMonitorString(t *testing.T) {
	node := NewNode("test-node", "Test Node", NodeTypeGeneric, "localhost:8080")
	checker := NewPingHealthChecker("test-ping")
	interval := 1 * time.Second
	
	monitor := NewHealthMonitor(node, checker, interval)
	
	str := monitor.String()
	if str == "" {
		t.Error("monitor string representation should not be empty")
	}
	
	err := monitor.Start()
	if err != nil {
		t.Errorf("failed to start monitor: %v", err)
	}
	defer monitor.Stop()
	
	str = monitor.String()
	if str == "" {
		t.Error("running monitor string representation should not be empty")
	}
}