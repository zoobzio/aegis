package aegis

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type PeerHealthMonitor struct {
	node         *Node
	interval     time.Duration
	timeout      time.Duration
	
	stopCh       chan struct{}
	running      bool
	mu           sync.RWMutex
	
	peerStatus   map[string]HealthStatus
	statusMu     sync.RWMutex
}

func NewPeerHealthMonitor(node *Node, interval time.Duration) *PeerHealthMonitor {
	return &PeerHealthMonitor{
		node:       node,
		interval:   interval,
		timeout:    10 * time.Second,
		stopCh:     make(chan struct{}),
		running:    false,
		peerStatus: make(map[string]HealthStatus),
	}
}

func (phm *PeerHealthMonitor) SetTimeout(timeout time.Duration) {
	phm.mu.Lock()
	defer phm.mu.Unlock()
	phm.timeout = timeout
}

func (phm *PeerHealthMonitor) IsRunning() bool {
	phm.mu.RLock()
	defer phm.mu.RUnlock()
	return phm.running
}

func (phm *PeerHealthMonitor) Start() error {
	phm.mu.Lock()
	defer phm.mu.Unlock()
	
	if phm.running {
		return fmt.Errorf("peer health monitor is already running")
	}
	
	if phm.node == nil {
		return fmt.Errorf("node cannot be nil")
	}
	
	if phm.node.PeerManager == nil {
		return fmt.Errorf("peer manager not initialized")
	}
	
	phm.running = true
	phm.stopCh = make(chan struct{})
	
	go phm.monitor()
	
	return nil
}

func (phm *PeerHealthMonitor) Stop() {
	phm.mu.Lock()
	defer phm.mu.Unlock()
	
	if !phm.running {
		return
	}
	
	phm.running = false
	close(phm.stopCh)
}

func (phm *PeerHealthMonitor) monitor() {
	ticker := time.NewTicker(phm.interval)
	defer ticker.Stop()
	
	phm.checkAllPeers()
	
	for {
		select {
		case <-ticker.C:
			phm.checkAllPeers()
		case <-phm.stopCh:
			return
		}
	}
}

func (phm *PeerHealthMonitor) checkAllPeers() {
	peers := phm.node.GetAllPeers()
	if len(peers) == 0 {
		return
	}
	
	phm.mu.RLock()
	timeout := phm.timeout
	phm.mu.RUnlock()
	
	var wg sync.WaitGroup
	for _, peer := range peers {
		wg.Add(1)
		go func(p *Peer) {
			defer wg.Done()
			phm.checkPeerHealth(p, timeout)
		}(peer)
	}
	
	wg.Wait()
}

func (phm *PeerHealthMonitor) checkPeerHealth(peer *Peer, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	healthResp, err := phm.node.GetPeerHealth(ctx, peer.Info.ID)
	
	
	if err != nil {
		currentStatus := HealthStatusUnhealthy
		
		phm.statusMu.Lock()
		phm.peerStatus[peer.Info.ID] = currentStatus
		phm.statusMu.Unlock()
		
		return
	}
	
	currentStatus := HealthStatus(healthResp.Status)
	
	phm.statusMu.Lock()
	phm.peerStatus[peer.Info.ID] = currentStatus
	phm.statusMu.Unlock()
}

func (phm *PeerHealthMonitor) GetPeerStatus(peerID string) (HealthStatus, bool) {
	phm.statusMu.RLock()
	defer phm.statusMu.RUnlock()
	
	status, exists := phm.peerStatus[peerID]
	return status, exists
}

func (phm *PeerHealthMonitor) GetAllPeerStatuses() map[string]HealthStatus {
	phm.statusMu.RLock()
	defer phm.statusMu.RUnlock()
	
	result := make(map[string]HealthStatus)
	for k, v := range phm.peerStatus {
		result[k] = v
	}
	return result
}

func (phm *PeerHealthMonitor) String() string {
	running := phm.IsRunning()
	statuses := phm.GetAllPeerStatuses()
	
	runningStr := "stopped"
	if running {
		runningStr = "running"
	}
	
	healthyCount := 0
	for _, status := range statuses {
		if status == HealthStatusHealthy {
			healthyCount++
		}
	}
	
	return fmt.Sprintf("PeerHealthMonitor[%s] monitoring %d peers, %d healthy", 
		runningStr, len(statuses), healthyCount)
}