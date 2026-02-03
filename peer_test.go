package aegis

import (
	"context"
	"testing"
	"time"
)

func TestNewPeerManager(t *testing.T) {
	pm := NewPeerManager("test-node")
	
	if pm.nodeID != "test-node" {
		t.Errorf("expected nodeID 'test-node', got '%s'", pm.nodeID)
	}
	
	if pm.Count() != 0 {
		t.Errorf("expected 0 peers, got %d", pm.Count())
	}
}

func TestPeerManagerAddRemovePeer(t *testing.T) {
	pm := NewPeerManager("test-node")
	
	peerInfo := PeerInfo{
		ID:      "peer-1",
		Address: "localhost:8081",
		Type:    NodeTypeGeneric,
	}
	
	err := pm.AddPeer(peerInfo)
	if err == nil {
		t.Error("expected error when adding peer without running server")
	}
	
	if pm.Count() != 0 {
		t.Errorf("expected 0 peers after failed add, got %d", pm.Count())
	}
}

func TestPeerManagerGetPeer(t *testing.T) {
	pm := NewPeerManager("test-node")
	
	_, exists := pm.GetPeer("nonexistent")
	if exists {
		t.Error("expected peer to not exist")
	}
	
	peers := pm.GetAllPeers()
	if len(peers) != 0 {
		t.Errorf("expected 0 peers, got %d", len(peers))
	}
}

func TestPeerManagerGetPeersByType(t *testing.T) {
	pm := NewPeerManager("test-node")
	
	peers := pm.GetPeersByType(NodeTypeGeneric)
	if len(peers) != 0 {
		t.Errorf("expected 0 peers of type generic, got %d", len(peers))
	}
}

func TestPeerManagerPingNonexistentPeer(t *testing.T) {
	pm := NewPeerManager("test-node")
	
	ctx := context.Background()
	_, err := pm.PingPeer(ctx, "nonexistent")
	if err == nil {
		t.Error("expected error when pinging nonexistent peer")
	}
}

func TestPeerManagerGetHealthNonexistentPeer(t *testing.T) {
	pm := NewPeerManager("test-node")
	
	ctx := context.Background()
	_, err := pm.GetPeerHealth(ctx, "nonexistent")
	if err == nil {
		t.Error("expected error when getting health of nonexistent peer")
	}
}

func TestPeerManagerClose(t *testing.T) {
	pm := NewPeerManager("test-node")
	
	err := pm.Close()
	if err != nil {
		t.Errorf("unexpected error closing empty peer manager: %v", err)
	}
	
	if pm.Count() != 0 {
		t.Errorf("expected 0 peers after close, got %d", pm.Count())
	}
}

func TestNodePeerMethods(t *testing.T) {
	node := NewNode("test-node", "Test Node", NodeTypeGeneric, "localhost:8080")
	
	if node.PeerManager == nil {
		t.Error("expected peer manager to be initialized")
	}
	
	if node.MeshServer == nil {
		t.Error("expected mesh server to be initialized")
	}
	
	peers := node.GetAllPeers()
	if len(peers) != 0 {
		t.Errorf("expected 0 peers, got %d", len(peers))
	}
	
	_, exists := node.GetPeer("nonexistent")
	if exists {
		t.Error("expected peer to not exist")
	}
}

func TestNodePeerCommunication(t *testing.T) {
	node := NewNode("test-node", "Test Node", NodeTypeGeneric, "localhost:8080")
	
	ctx := context.Background()
	
	_, err := node.PingPeer(ctx, "nonexistent")
	if err == nil {
		t.Error("expected error pinging nonexistent peer")
	}
	
	_, err = node.GetPeerHealth(ctx, "nonexistent")
	if err == nil {
		t.Error("expected error getting health of nonexistent peer")
	}
	
	err = node.NotifyPeerFailure(ctx, "nonexistent")
	if err != nil {
		t.Errorf("unexpected error notifying about nonexistent peer: %v", err)
	}
}

func TestNodeShutdown(t *testing.T) {
	node := NewNode("test-node", "Test Node", NodeTypeGeneric, "localhost:8080")
	
	err := node.Shutdown()
	if err != nil {
		t.Errorf("unexpected error during shutdown: %v", err)
	}
}

func TestNewPeerHealthMonitor(t *testing.T) {
	node := NewNode("test-node", "Test Node", NodeTypeGeneric, "localhost:8080")
	interval := 5 * time.Second
	
	phm := NewPeerHealthMonitor(node, interval)
	
	if phm.node != node {
		t.Error("expected correct node reference")
	}
	
	if phm.interval != interval {
		t.Errorf("expected interval %v, got %v", interval, phm.interval)
	}
	
	if phm.IsRunning() {
		t.Error("new peer health monitor should not be running")
	}
	
	statuses := phm.GetAllPeerStatuses()
	if len(statuses) != 0 {
		t.Errorf("expected 0 peer statuses, got %d", len(statuses))
	}
}

func TestPeerHealthMonitorStart(t *testing.T) {
	node := NewNode("test-node", "Test Node", NodeTypeGeneric, "localhost:8080")
	phm := NewPeerHealthMonitor(node, 100*time.Millisecond)
	
	err := phm.Start()
	if err != nil {
		t.Errorf("failed to start peer health monitor: %v", err)
	}
	
	if !phm.IsRunning() {
		t.Error("peer health monitor should be running after start")
	}
	
	time.Sleep(50 * time.Millisecond)
	
	phm.Stop()
	
	if phm.IsRunning() {
		t.Error("peer health monitor should not be running after stop")
	}
}

func TestPeerHealthMonitorStartAlreadyRunning(t *testing.T) {
	node := NewNode("test-node", "Test Node", NodeTypeGeneric, "localhost:8080")
	phm := NewPeerHealthMonitor(node, 1*time.Second)
	
	err := phm.Start()
	if err != nil {
		t.Errorf("failed to start peer health monitor: %v", err)
	}
	defer phm.Stop()
	
	err = phm.Start()
	if err == nil {
		t.Error("starting already running peer health monitor should return error")
	}
}