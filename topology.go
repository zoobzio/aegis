package aegis

import (
	"fmt"
	"sync"
	"time"
)

// NodeInfo contains information about a node in the mesh topology.
type NodeInfo struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      NodeType  `json:"type"`
	Address   string    `json:"address"`
	Services  []ServiceInfo `json:"services,omitempty"`
	JoinedAt  time.Time `json:"joined_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Topology maintains the mesh network topology.
type Topology struct {
	Nodes     map[string]NodeInfo `json:"nodes"`
	Version   int64               `json:"version"`
	UpdatedAt time.Time           `json:"updated_at"`
	mu        sync.RWMutex
}

// NewTopology creates a new empty topology.
func NewTopology() *Topology {
	return &Topology{
		Nodes:     make(map[string]NodeInfo),
		Version:   0,
		UpdatedAt: time.Now(),
	}
}

// AddNode adds a node to the topology.
func (t *Topology) AddNode(info NodeInfo) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.Nodes[info.ID]; exists {
		return fmt.Errorf("node %s already exists in topology", info.ID)
	}

	info.JoinedAt = time.Now()
	info.UpdatedAt = info.JoinedAt
	t.Nodes[info.ID] = info
	t.Version++
	t.UpdatedAt = time.Now()

	return nil
}

// RemoveNode removes a node from the topology.
func (t *Topology) RemoveNode(nodeID string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.Nodes[nodeID]; !exists {
		return fmt.Errorf("node %s not found in topology", nodeID)
	}

	delete(t.Nodes, nodeID)
	t.Version++
	t.UpdatedAt = time.Now()

	return nil
}

// UpdateNode updates a node in the topology.
func (t *Topology) UpdateNode(info NodeInfo) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	existing, exists := t.Nodes[info.ID]
	if !exists {
		return fmt.Errorf("node %s not found in topology", info.ID)
	}

	info.JoinedAt = existing.JoinedAt
	info.UpdatedAt = time.Now()
	t.Nodes[info.ID] = info
	t.Version++
	t.UpdatedAt = time.Now()

	return nil
}

// GetNode returns a node by ID.
func (t *Topology) GetNode(nodeID string) (NodeInfo, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	info, exists := t.Nodes[nodeID]
	return info, exists
}

// GetAllNodes returns all nodes in the topology.
func (t *Topology) GetAllNodes() []NodeInfo {
	t.mu.RLock()
	defer t.mu.RUnlock()

	nodes := make([]NodeInfo, 0, len(t.Nodes))
	for _, node := range t.Nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// GetVersion returns the topology version.
func (t *Topology) GetVersion() int64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Version
}

// Clone creates a copy of the topology.
func (t *Topology) Clone() *Topology {
	t.mu.RLock()
	defer t.mu.RUnlock()

	clone := &Topology{
		Nodes:     make(map[string]NodeInfo),
		Version:   t.Version,
		UpdatedAt: t.UpdatedAt,
	}

	for k, v := range t.Nodes {
		clone.Nodes[k] = v
	}

	return clone
}

// Merge merges another topology if it has a higher version.
func (t *Topology) Merge(other *Topology) bool {
	if other == nil || other.Version <= t.GetVersion() {
		return false
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.Nodes = make(map[string]NodeInfo)
	for k, v := range other.Nodes {
		t.Nodes[k] = v
	}
	t.Version = other.Version
	t.UpdatedAt = other.UpdatedAt

	return true
}

// NodeCount returns the number of nodes.
func (t *Topology) NodeCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.Nodes)
}
