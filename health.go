package aegis

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

type HealthInfo struct {
	Status      HealthStatus `json:"status"`
	LastChecked time.Time    `json:"last_checked"`
	Message     string       `json:"message,omitempty"`
	Error       string       `json:"error,omitempty"`
	mu          sync.RWMutex `json:"-"`
}

func NewHealthInfo() *HealthInfo {
	return &HealthInfo{
		Status:      HealthStatusUnknown,
		LastChecked: time.Now(),
		Message:     "Not yet checked",
	}
}

func (h *HealthInfo) Update(status HealthStatus, message string, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	h.Status = status
	h.LastChecked = time.Now()
	h.Message = message
	
	if err != nil {
		h.Error = err.Error()
	} else {
		h.Error = ""
	}
}

func (h *HealthInfo) Get() (HealthStatus, time.Time, string, string) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	return h.Status, h.LastChecked, h.Message, h.Error
}

func (h *HealthInfo) IsHealthy() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	return h.Status == HealthStatusHealthy
}

func (h *HealthInfo) String() string {
	status, lastChecked, message, errMsg := h.Get()
	
	result := fmt.Sprintf("Health[%s] %s (checked: %s)", 
		status, message, lastChecked.Format(time.RFC3339))
	
	if errMsg != "" {
		result += fmt.Sprintf(" - Error: %s", errMsg)
	}
	
	return result
}

type HealthChecker interface {
	Check(ctx context.Context) error
	Name() string
}

type PingHealthChecker struct {
	name string
}

func NewPingHealthChecker(name string) *PingHealthChecker {
	return &PingHealthChecker{name: name}
}

func (p *PingHealthChecker) Check(ctx context.Context) error {
	return nil
}

func (p *PingHealthChecker) Name() string {
	return p.name
}

type FunctionHealthChecker struct {
	name     string
	checkFn  func(ctx context.Context) error
}

func NewFunctionHealthChecker(name string, checkFn func(ctx context.Context) error) *FunctionHealthChecker {
	return &FunctionHealthChecker{
		name:    name,
		checkFn: checkFn,
	}
}

func (f *FunctionHealthChecker) Check(ctx context.Context) error {
	if f.checkFn == nil {
		return fmt.Errorf("no check function provided")
	}
	return f.checkFn(ctx)
}

func (f *FunctionHealthChecker) Name() string {
	return f.name
}