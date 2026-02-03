package aegis

import (
	"fmt"
	"sync"
	"time"
)

type NodeInfo struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      NodeType  `json:"type"`
	Address   string    `json:"address"`
	JoinedAt  time.Time `json:"joined_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Topology struct {
	Nodes     map[string]NodeInfo `json:"nodes"`
	Version   int64               `json:"version"`
	UpdatedAt time.Time           `json:"updated_at"`
	mu        sync.RWMutex
}

func NewTopology() *Topology {
	return &Topology{
		Nodes:     make(map[string]NodeInfo),
		Version:   0,
		UpdatedAt: time.Now(),
	}
}

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

func (t *Topology) GetNode(nodeID string) (NodeInfo, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	info, exists := t.Nodes[nodeID]
	return info, exists
}

func (t *Topology) GetAllNodes() []NodeInfo {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	nodes := make([]NodeInfo, 0, len(t.Nodes))
	for _, node := range t.Nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

func (t *Topology) GetVersion() int64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Version
}

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

func (t *Topology) NodeCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.Nodes)
}

type JoinRequest struct {
	NodeInfo  NodeInfo  `json:"node_info"`
	RequestID string    `json:"request_id"`
	Timestamp time.Time `json:"timestamp"`
}

type Vote struct {
	NodeID    string    `json:"node_id"`
	RequestID string    `json:"request_id"`
	Approve   bool      `json:"approve"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

type VoteTracker struct {
	Request     JoinRequest       `json:"request"`
	Votes       map[string]Vote   `json:"votes"`
	TotalNodes  int               `json:"total_nodes"`
	CreatedAt   time.Time         `json:"created_at"`
	mu          sync.RWMutex
}

func NewVoteTracker(request JoinRequest, totalNodes int) *VoteTracker {
	return &VoteTracker{
		Request:    request,
		Votes:      make(map[string]Vote),
		TotalNodes: totalNodes,
		CreatedAt:  time.Now(),
	}
}

func (vt *VoteTracker) AddVote(vote Vote) error {
	vt.mu.Lock()
	defer vt.mu.Unlock()
	
	if vote.RequestID != vt.Request.RequestID {
		return fmt.Errorf("vote request ID mismatch")
	}
	
	if _, exists := vt.Votes[vote.NodeID]; exists {
		return fmt.Errorf("node %s already voted", vote.NodeID)
	}
	
	vt.Votes[vote.NodeID] = vote
	return nil
}

func (vt *VoteTracker) IsComplete() bool {
	vt.mu.RLock()
	defer vt.mu.RUnlock()
	return len(vt.Votes) >= vt.TotalNodes
}

func (vt *VoteTracker) GetResult() (approved bool, vetoNode string, reason string) {
	vt.mu.RLock()
	defer vt.mu.RUnlock()
	
	for _, vote := range vt.Votes {
		if !vote.Approve {
			return false, vote.NodeID, vote.Reason
		}
	}
	
	return len(vt.Votes) == vt.TotalNodes, "", ""
}

func (vt *VoteTracker) GetVoteCount() (approve, reject int) {
	vt.mu.RLock()
	defer vt.mu.RUnlock()
	
	for _, vote := range vt.Votes {
		if vote.Approve {
			approve++
		} else {
			reject++
		}
	}
	
	return approve, reject
}