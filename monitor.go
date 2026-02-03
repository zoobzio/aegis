package aegis

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type HealthMonitor struct {
	node        *Node
	checker     HealthChecker
	interval    time.Duration
	timeout     time.Duration
	
	stopCh      chan struct{}
	running     bool
	mu          sync.RWMutex
}

func NewHealthMonitor(node *Node, checker HealthChecker, interval time.Duration) *HealthMonitor {
	return &HealthMonitor{
		node:     node,
		checker:  checker,
		interval: interval,
		timeout:  30 * time.Second,
		stopCh:   make(chan struct{}),
		running:  false,
	}
}

func (hm *HealthMonitor) SetTimeout(timeout time.Duration) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.timeout = timeout
}

func (hm *HealthMonitor) IsRunning() bool {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return hm.running
}

func (hm *HealthMonitor) Start() error {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	
	if hm.running {
		return fmt.Errorf("health monitor is already running")
	}
	
	if hm.node == nil {
		return fmt.Errorf("node cannot be nil")
	}
	
	if hm.checker == nil {
		return fmt.Errorf("health checker cannot be nil")
	}
	
	hm.running = true
	hm.stopCh = make(chan struct{})
	
	go hm.monitor()
	
	return nil
}

func (hm *HealthMonitor) Stop() {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	
	if !hm.running {
		return
	}
	
	hm.running = false
	close(hm.stopCh)
}

func (hm *HealthMonitor) monitor() {
	ticker := time.NewTicker(hm.interval)
	defer ticker.Stop()
	
	hm.performHealthCheck()
	
	for {
		select {
		case <-ticker.C:
			hm.performHealthCheck()
		case <-hm.stopCh:
			return
		}
	}
}

func (hm *HealthMonitor) performHealthCheck() {
	hm.mu.RLock()
	timeout := hm.timeout
	hm.mu.RUnlock()
	
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	hm.node.CheckHealth(ctx, hm.checker)
}

func (hm *HealthMonitor) GetStatus() (bool, HealthStatus, string) {
	if hm.node == nil || hm.node.Health == nil {
		return hm.IsRunning(), HealthStatusUnknown, "Node or health not initialized"
	}
	
	status, message := hm.node.GetHealth()
	return hm.IsRunning(), status, message
}

func (hm *HealthMonitor) String() string {
	running, status, message := hm.GetStatus()
	
	runningStr := "stopped"
	if running {
		runningStr = "running"
	}
	
	return fmt.Sprintf("HealthMonitor[%s] %s - %s (%s)", 
		runningStr, hm.node.ID, status, message)
}