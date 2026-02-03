package aegis

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type NodeType string

const (
	NodeTypeGeneric   NodeType = "generic"
	NodeTypeGateway   NodeType = "gateway"
	NodeTypeProcessor NodeType = "processor"
	NodeTypeStorage   NodeType = "storage"
)

type Node struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        NodeType          `json:"type"`
	Address     string            `json:"address"`
	Health      *HealthInfo       `json:"health"`
	PeerManager *PeerManager      `json:"-"`
	MeshServer  *MeshServer       `json:"-"`
	Functions   *FunctionRegistry `json:"-"`
	Rooms       *RoomManager      `json:"-"`
	Topology    *Topology         `json:"-"`
	Consensus   *ConsensusManager `json:"-"`
	TLSConfig   *TLSConfig        `json:"-"`
}

func NewNode(id, name string, nodeType NodeType, address string) *Node {
	node := &Node{
		ID:      id,
		Name:    name,
		Type:    nodeType,
		Address: address,
		Health:  NewHealthInfo(),
	}
	
	node.PeerManager = NewPeerManager(id)
	node.MeshServer = NewMeshServer(node)
	node.Functions = NewFunctionRegistry()
	node.Rooms = NewRoomManager()
	node.Topology = NewTopology()
	node.Consensus = NewConsensusManager(id, node.Topology)
	
	// Add self to topology
	nodeInfo := NodeInfo{
		ID:      id,
		Name:    name,
		Type:    nodeType,
		Address: address,
	}
	node.Topology.AddNode(nodeInfo)
	
	return node
}

// EnableTLS enables TLS for the node using the specified certificate directory
func (n *Node) EnableTLS(certDir string) error {
	tlsConfig, err := LoadOrGenerateTLS(n.ID, certDir)
	if err != nil {
		return fmt.Errorf("failed to setup TLS: %w", err)
	}
	
	n.TLSConfig = tlsConfig
	
	// Configure TLS for server
	if n.MeshServer != nil {
		n.MeshServer.SetTLSConfig(tlsConfig)
	}
	
	// Configure TLS for peer manager
	if n.PeerManager != nil {
		n.PeerManager.SetTLSConfig(tlsConfig)
	}
	
	return nil
}

func (n *Node) String() string {
	return fmt.Sprintf("Node[%s:%s] %s (%s) @ %s", 
		n.Type, n.ID, n.Name, n.Type, n.Address)
}

func (n *Node) MarshalJSON() ([]byte, error) {
	type Alias Node
	return json.Marshal((*Alias)(n))
}

func (n *Node) UnmarshalJSON(data []byte) error {
	type Alias Node
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(n),
	}
	return json.Unmarshal(data, &aux)
}

func (n *Node) Validate() error {
	if n.ID == "" {
		return fmt.Errorf("node ID cannot be empty")
	}
	if n.Name == "" {
		return fmt.Errorf("node name cannot be empty")
	}
	if n.Address == "" {
		return fmt.Errorf("node address cannot be empty")
	}
	
	_, _, err := net.SplitHostPort(n.Address)
	if err != nil {
		return fmt.Errorf("invalid address format: %w", err)
	}
	
	return nil
}

func (n *Node) SetHealth(status HealthStatus, message string, err error) {
	if n.Health == nil {
		n.Health = NewHealthInfo()
	}
	n.Health.Update(status, message, err)
}

func (n *Node) GetHealth() (HealthStatus, string) {
	if n.Health == nil {
		return HealthStatusUnknown, "Health not initialized"
	}
	status, _, message, errMsg := n.Health.Get()
	if errMsg != "" {
		return status, fmt.Sprintf("%s (Error: %s)", message, errMsg)
	}
	return status, message
}

func (n *Node) IsHealthy() bool {
	if n.Health == nil {
		return false
	}
	return n.Health.IsHealthy()
}

func (n *Node) CheckHealth(ctx context.Context, checker HealthChecker) error {
	if checker == nil {
		n.SetHealth(HealthStatusUnhealthy, "No health checker provided", fmt.Errorf("health checker is nil"))
		return fmt.Errorf("health checker is nil")
	}
	
	err := checker.Check(ctx)
	if err != nil {
		n.SetHealth(HealthStatusUnhealthy, fmt.Sprintf("Health check failed: %s", checker.Name()), err)
		return err
	}
	
	n.SetHealth(HealthStatusHealthy, fmt.Sprintf("Health check passed: %s", checker.Name()), nil)
	return nil
}

func (n *Node) StartMeshServer() error {
	if n.MeshServer == nil {
		return fmt.Errorf("mesh server not initialized")
	}
	return n.MeshServer.Start()
}

func (n *Node) StopMeshServer() {
	if n.MeshServer != nil {
		n.MeshServer.Stop()
	}
}

func (n *Node) AddPeer(info PeerInfo) error {
	if n.PeerManager == nil {
		return fmt.Errorf("peer manager not initialized")
	}
	return n.PeerManager.AddPeer(info)
}

func (n *Node) RemovePeer(peerID string) error {
	if n.PeerManager == nil {
		return fmt.Errorf("peer manager not initialized")
	}
	return n.PeerManager.RemovePeer(peerID)
}

func (n *Node) GetPeer(peerID string) (*Peer, bool) {
	if n.PeerManager == nil {
		return nil, false
	}
	return n.PeerManager.GetPeer(peerID)
}

func (n *Node) GetAllPeers() []*Peer {
	if n.PeerManager == nil {
		return nil
	}
	return n.PeerManager.GetAllPeers()
}

func (n *Node) PingPeer(ctx context.Context, peerID string) (*PingResponse, error) {
	if n.PeerManager == nil {
		return nil, fmt.Errorf("peer manager not initialized")
	}
	return n.PeerManager.PingPeer(ctx, peerID)
}

func (n *Node) GetPeerHealth(ctx context.Context, peerID string) (*HealthResponse, error) {
	if n.PeerManager == nil {
		return nil, fmt.Errorf("peer manager not initialized")
	}
	return n.PeerManager.GetPeerHealth(ctx, peerID)
}

func (n *Node) NotifyPeerFailure(ctx context.Context, failedNodeID string) error {
	if n.PeerManager == nil {
		return fmt.Errorf("peer manager not initialized")
	}
	return n.PeerManager.NotifyHealthChange(ctx, failedNodeID, string(HealthStatusUnhealthy), "Peer health check failed")
}

func (n *Node) RegisterFunction(name string, fn NodeFunction) error {
	if n.Functions == nil {
		return fmt.Errorf("function registry not initialized")
	}
	return n.Functions.Register(name, fn)
}

func (n *Node) UnregisterFunction(name string) error {
	if n.Functions == nil {
		return fmt.Errorf("function registry not initialized")
	}
	return n.Functions.Unregister(name)
}

func (n *Node) ExecuteFunction(ctx context.Context, name string, parameters []string) (string, error) {
	if n.Functions == nil {
		return "", fmt.Errorf("function registry not initialized")
	}
	return n.Functions.Execute(ctx, name, parameters)
}

func (n *Node) ListFunctions() []string {
	if n.Functions == nil {
		return nil
	}
	return n.Functions.ListFunctions()
}

func (n *Node) ExecutePeerFunction(ctx context.Context, peerID, functionName string, parameters []string) (*FunctionResponse, error) {
	if n.PeerManager == nil {
		return nil, fmt.Errorf("peer manager not initialized")
	}
	return n.PeerManager.ExecuteFunction(ctx, peerID, functionName, parameters)
}

func (n *Node) SendMessage(ctx context.Context, peerID, message string) error {
	if n.PeerManager == nil {
		return fmt.Errorf("peer manager not initialized")
	}
	
	resp, err := n.PeerManager.SendMessage(ctx, peerID, message)
	if err != nil {
		return err
	}
	
	if !resp.Received {
		return fmt.Errorf("message not received by peer %s", peerID)
	}
	
	return nil
}

func (n *Node) BroadcastMessage(ctx context.Context, message string) error {
	if n.PeerManager == nil {
		return fmt.Errorf("peer manager not initialized")
	}
	
	peers := n.GetAllPeers()
	var lastErr error
	
	for _, peer := range peers {
		if err := n.SendMessage(ctx, peer.Info.ID, message); err != nil {
			lastErr = err
		}
	}
	
	return lastErr
}

func (n *Node) CreateRoom(roomID, roomName string) error {
	if n.Rooms == nil {
		return fmt.Errorf("room manager not initialized")
	}
	
	_, err := n.Rooms.CreateRoom(roomID, roomName, n.ID)
	return err
}

func (n *Node) InviteToRoom(ctx context.Context, roomID, peerID string) error {
	if n.Rooms == nil {
		return fmt.Errorf("room manager not initialized")
	}
	
	room, exists := n.Rooms.GetHostedRoom(roomID)
	if !exists {
		return fmt.Errorf("room %s not found or not hosted by this node", roomID)
	}
	
	if n.PeerManager == nil {
		return fmt.Errorf("peer manager not initialized")
	}
	
	return n.PeerManager.InviteToRoom(ctx, peerID, room.ID)
}

func (n *Node) SendToRoom(ctx context.Context, roomID, message string) error {
	if n.Rooms == nil {
		return fmt.Errorf("room manager not initialized")
	}
	
	room, isHosted := n.Rooms.GetHostedRoom(roomID)
	if isHosted {
		members := room.GetMembers()
		var lastErr error
		for _, memberID := range members {
			if memberID != n.ID {
				if err := n.SendMessage(ctx, memberID, fmt.Sprintf("[Room %s] %s: %s", roomID, n.ID, message)); err != nil {
					lastErr = err
				}
			}
		}
		return lastErr
	}
	
	hostID, isJoined := n.Rooms.GetJoinedRoomHost(roomID)
	if !isJoined {
		return fmt.Errorf("not a member of room %s", roomID)
	}
	
	if n.PeerManager == nil {
		return fmt.Errorf("peer manager not initialized")
	}
	
	return n.PeerManager.SendRoomMessage(ctx, hostID, roomID, message)
}

func (n *Node) ListRooms() (hosted []string, joined []string) {
	if n.Rooms == nil {
		return nil, nil
	}
	
	return n.Rooms.ListHostedRooms(), n.Rooms.ListJoinedRooms()
}

func (n *Node) JoinMesh(ctx context.Context, entryNodeAddress string) error {
	if n.Topology == nil || n.Consensus == nil {
		return fmt.Errorf("topology or consensus manager not initialized")
	}
	
	// Create temporary peer connection to entry node
	tempPeer := PeerInfo{
		ID:      "entry-node",
		Address: entryNodeAddress,
		Type:    NodeTypeGeneric,
	}
	
	if err := n.PeerManager.AddPeer(tempPeer); err != nil {
		return fmt.Errorf("failed to connect to entry node: %w", err)
	}
	defer n.PeerManager.RemovePeer("entry-node")
	
	// Request to join mesh
	nodeInfo := NodeInfo{
		ID:      n.ID,
		Name:    n.Name,
		Type:    n.Type,
		Address: n.Address,
	}
	
	joinReq, err := n.Consensus.InitiateJoinRequest(nodeInfo)
	if err != nil {
		return err
	}
	
	// Send join request to entry node
	peer, _ := n.PeerManager.GetPeer("entry-node")
	req := &JoinMeshRequest{
		NodeId:    n.ID,
		Name:      n.Name,
		Type:      string(n.Type),
		Address:   n.Address,
		RequestId: joinReq.RequestID,
	}
	
	resp, err := peer.Client.RequestJoinMesh(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to request join: %w", err)
	}
	
	if !resp.Pending {
		return fmt.Errorf("join request rejected: %s", resp.Message)
	}
	
	return nil
}

func (n *Node) GetTopologyVersion() int64 {
	if n.Topology == nil {
		return 0
	}
	return n.Topology.GetVersion()
}

func (n *Node) GetMeshNodes() []NodeInfo {
	if n.Topology == nil {
		return nil
	}
	return n.Topology.GetAllNodes()
}

func (n *Node) SyncTopologyWithPeers(ctx context.Context) error {
	if n.Topology == nil || n.PeerManager == nil {
		return fmt.Errorf("topology or peer manager not initialized")
	}
	
	peers := n.GetAllPeers()
	for _, peer := range peers {
		if err := n.syncWithPeer(ctx, peer.Info.ID); err != nil {
			// Log error but continue with other peers
			continue
		}
	}
	
	return nil
}

func (n *Node) syncWithPeer(ctx context.Context, peerID string) error {
	peer, exists := n.GetPeer(peerID)
	if !exists {
		return fmt.Errorf("peer %s not found", peerID)
	}
	
	req := &TopologySyncRequest{
		SenderId: n.ID,
		Version:  n.Topology.GetVersion(),
	}
	
	resp, err := peer.Client.SyncTopology(ctx, req)
	if err != nil {
		return err
	}
	
	if resp.Version > n.Topology.GetVersion() {
		// Build new topology from response
		newTopology := NewTopology()
		for _, nodeProto := range resp.Nodes {
			nodeInfo := NodeInfo{
				ID:        nodeProto.Id,
				Name:      nodeProto.Name,
				Type:      NodeType(nodeProto.Type),
				Address:   nodeProto.Address,
				JoinedAt:  time.Unix(nodeProto.JoinedAt, 0),
				UpdatedAt: time.Unix(nodeProto.UpdatedAt, 0),
			}
			newTopology.AddNode(nodeInfo)
		}
		newTopology.Version = resp.Version
		newTopology.UpdatedAt = time.Unix(resp.UpdatedAt, 0)
		
		// Merge with our topology
		n.Topology.Merge(newTopology)
		
		// Update peer connections based on new topology
		n.updatePeerConnections(ctx)
	}
	
	return nil
}

func (n *Node) updatePeerConnections(ctx context.Context) error {
	if n.Topology == nil || n.PeerManager == nil {
		return fmt.Errorf("topology or peer manager not initialized")
	}
	
	// Get all nodes from topology
	nodes := n.Topology.GetAllNodes()
	
	// Add peers that are in topology but not connected
	for _, nodeInfo := range nodes {
		if nodeInfo.ID == n.ID {
			continue // Skip self
		}
		
		if _, exists := n.GetPeer(nodeInfo.ID); !exists {
			peerInfo := PeerInfo{
				ID:      nodeInfo.ID,
				Address: nodeInfo.Address,
				Type:    nodeInfo.Type,
			}
			n.PeerManager.AddPeer(peerInfo)
		}
	}
	
	return nil
}

func (n *Node) Shutdown() error {
	n.StopMeshServer()
	
	if n.PeerManager != nil {
		return n.PeerManager.Close()
	}
	
	return nil
}