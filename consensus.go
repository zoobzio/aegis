package aegis

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type ConsensusManager struct {
	nodeID       string
	topology     *Topology
	activeVotes  map[string]*VoteTracker
	voteTimeout  time.Duration
	onApproved   func(JoinRequest)
	onRejected   func(JoinRequest, string, string)
	mu           sync.RWMutex
}

func NewConsensusManager(nodeID string, topology *Topology) *ConsensusManager {
	return &ConsensusManager{
		nodeID:      nodeID,
		topology:    topology,
		activeVotes: make(map[string]*VoteTracker),
		voteTimeout: 30 * time.Second,
	}
}

func (cm *ConsensusManager) SetCallbacks(onApproved func(JoinRequest), onRejected func(JoinRequest, string, string)) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.onApproved = onApproved
	cm.onRejected = onRejected
}

func (cm *ConsensusManager) InitiateJoinRequest(nodeInfo NodeInfo) (*JoinRequest, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	request := JoinRequest{
		NodeInfo:  nodeInfo,
		RequestID: uuid.New().String(),
		Timestamp: time.Now(),
	}
	
	totalNodes := cm.topology.NodeCount()
	if totalNodes == 0 {
		// First node in mesh - auto-approve
		if cm.onApproved != nil {
			cm.onApproved(request)
		}
		return &request, nil
	}
	
	tracker := NewVoteTracker(request, totalNodes)
	cm.activeVotes[request.RequestID] = tracker
	
	// Start timeout timer
	go cm.timeoutVote(request.RequestID)
	
	return &request, nil
}

func (cm *ConsensusManager) SubmitVote(vote Vote) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	tracker, exists := cm.activeVotes[vote.RequestID]
	if !exists {
		return fmt.Errorf("no active vote for request %s", vote.RequestID)
	}
	
	if err := tracker.AddVote(vote); err != nil {
		return err
	}
	
	if tracker.IsComplete() {
		cm.finalizeVote(vote.RequestID)
	}
	
	return nil
}

func (cm *ConsensusManager) GetActiveVote(requestID string) (*VoteTracker, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	tracker, exists := cm.activeVotes[requestID]
	return tracker, exists
}

func (cm *ConsensusManager) finalizeVote(requestID string) {
	tracker, exists := cm.activeVotes[requestID]
	if !exists {
		return
	}
	
	approved, vetoNode, reason := tracker.GetResult()
	
	if approved {
		if cm.onApproved != nil {
			cm.onApproved(tracker.Request)
		}
	} else {
		if cm.onRejected != nil {
			cm.onRejected(tracker.Request, vetoNode, reason)
		}
	}
	
	delete(cm.activeVotes, requestID)
}

func (cm *ConsensusManager) timeoutVote(requestID string) {
	time.Sleep(cm.voteTimeout)
	
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	tracker, exists := cm.activeVotes[requestID]
	if !exists {
		return
	}
	
	// Timeout is treated as rejection
	if cm.onRejected != nil {
		approvals, _ := tracker.GetVoteCount()
		cm.onRejected(tracker.Request, "system", 
			fmt.Sprintf("timeout: only %d/%d votes received", approvals, tracker.TotalNodes))
	}
	
	delete(cm.activeVotes, requestID)
}

func (cm *ConsensusManager) CastLocalVote(requestID string, approve bool, reason string) error {
	vote := Vote{
		NodeID:    cm.nodeID,
		RequestID: requestID,
		Approve:   approve,
		Reason:    reason,
		Timestamp: time.Now(),
	}
	
	return cm.SubmitVote(vote)
}

type VetoReason string

const (
	VetoReasonBlacklisted    VetoReason = "blacklisted"
	VetoReasonResourceLimit  VetoReason = "resource_limit"
	VetoReasonSuspicious     VetoReason = "suspicious"
	VetoReasonVersionMismatch VetoReason = "version_mismatch"
	VetoReasonCustom         VetoReason = "custom"
)

func (cm *ConsensusManager) EvaluateJoinRequest(ctx context.Context, nodeInfo NodeInfo) (bool, VetoReason, string) {
	// Basic validation
	if nodeInfo.ID == "" || nodeInfo.Address == "" {
		return false, VetoReasonSuspicious, "missing required node information"
	}
	
	// Check if already in topology
	if _, exists := cm.topology.GetNode(nodeInfo.ID); exists {
		return false, VetoReasonSuspicious, "node already in topology"
	}
	
	// Check resource limits (example: max 100 nodes)
	if cm.topology.NodeCount() >= 100 {
		return false, VetoReasonResourceLimit, "mesh at capacity"
	}
	
	// Future: Add blacklist check, version compatibility, etc.
	
	return true, "", ""
}