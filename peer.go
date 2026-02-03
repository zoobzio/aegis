package aegis

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type PeerInfo struct {
	ID      string   `json:"id"`
	Address string   `json:"address"`
	Type    NodeType `json:"type"`
}

type Peer struct {
	Info   PeerInfo
	Client MeshServiceClient
	Conn   *grpc.ClientConn
}

type PeerManager struct {
	nodeID    string
	peers     map[string]*Peer
	tlsConfig *TLSConfig
	mu        sync.RWMutex
}

func NewPeerManager(nodeID string) *PeerManager {
	return &PeerManager{
		nodeID: nodeID,
		peers:  make(map[string]*Peer),
	}
}

func (pm *PeerManager) SetTLSConfig(tlsConfig *TLSConfig) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.tlsConfig = tlsConfig
}

func (pm *PeerManager) AddPeer(info PeerInfo) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.peers[info.ID]; exists {
		return fmt.Errorf("peer %s already exists", info.ID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// TLS is required
	if pm.tlsConfig == nil {
		return fmt.Errorf("TLS configuration is required but not set")
	}
	
	// Configure dial options with TLS
	creds := credentials.NewTLS(pm.tlsConfig.GetClientTLSConfig(info.ID))
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithBlock(),
	}

	conn, err := grpc.DialContext(ctx, info.Address, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to peer %s at %s: %w", info.ID, info.Address, err)
	}

	client := NewMeshServiceClient(conn)

	peer := &Peer{
		Info:   info,
		Client: client,
		Conn:   conn,
	}

	pm.peers[info.ID] = peer
	return nil
}

func (pm *PeerManager) RemovePeer(peerID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	peer, exists := pm.peers[peerID]
	if !exists {
		return fmt.Errorf("peer %s not found", peerID)
	}

	if err := peer.Conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection to peer %s: %w", peerID, err)
	}

	delete(pm.peers, peerID)
	return nil
}

func (pm *PeerManager) GetPeer(peerID string) (*Peer, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	peer, exists := pm.peers[peerID]
	return peer, exists
}

func (pm *PeerManager) GetAllPeers() []*Peer {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	peers := make([]*Peer, 0, len(pm.peers))
	for _, peer := range pm.peers {
		peers = append(peers, peer)
	}
	return peers
}

func (pm *PeerManager) GetPeersByType(nodeType NodeType) []*Peer {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var peers []*Peer
	for _, peer := range pm.peers {
		if peer.Info.Type == nodeType {
			peers = append(peers, peer)
		}
	}
	return peers
}

func (pm *PeerManager) PingPeer(ctx context.Context, peerID string) (*PingResponse, error) {
	peer, exists := pm.GetPeer(peerID)
	if !exists {
		return nil, fmt.Errorf("peer %s not found", peerID)
	}

	req := &PingRequest{
		SenderId:  pm.nodeID,
		Timestamp: time.Now().Unix(),
	}

	return peer.Client.Ping(ctx, req)
}

func (pm *PeerManager) GetPeerHealth(ctx context.Context, peerID string) (*HealthResponse, error) {
	peer, exists := pm.GetPeer(peerID)
	if !exists {
		return nil, fmt.Errorf("peer %s not found", peerID)
	}

	req := &HealthRequest{
		SenderId: pm.nodeID,
	}

	return peer.Client.GetHealth(ctx, req)
}

func (pm *PeerManager) GetPeerNodeInfo(ctx context.Context, peerID string) (*NodeInfoResponse, error) {
	peer, exists := pm.GetPeer(peerID)
	if !exists {
		return nil, fmt.Errorf("peer %s not found", peerID)
	}

	req := &NodeInfoRequest{
		SenderId: pm.nodeID,
	}

	return peer.Client.GetNodeInfo(ctx, req)
}

func (pm *PeerManager) NotifyHealthChange(ctx context.Context, failedNodeID, status, message string) error {
	peers := pm.GetAllPeers()
	
	req := &HealthChangeRequest{
		SenderId:     pm.nodeID,
		FailedNodeId: failedNodeID,
		Status:       status,
		Message:      message,
		Timestamp:    time.Now().Unix(),
	}

	var lastErr error
	for _, peer := range peers {
		if peer.Info.ID == failedNodeID {
			continue
		}

		_, err := peer.Client.NotifyHealthChange(ctx, req)
		if err != nil {
			lastErr = err
		}
	}

	return lastErr
}

func (pm *PeerManager) Close() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	var lastErr error
	for _, peer := range pm.peers {
		if err := peer.Conn.Close(); err != nil {
			lastErr = err
		}
	}

	pm.peers = make(map[string]*Peer)
	return lastErr
}

func (pm *PeerManager) Count() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.peers)
}

func (pm *PeerManager) ExecuteFunction(ctx context.Context, peerID, functionName string, parameters []string) (*FunctionResponse, error) {
	peer, exists := pm.GetPeer(peerID)
	if !exists {
		return nil, fmt.Errorf("peer %s not found", peerID)
	}

	req := &FunctionRequest{
		SenderId:     pm.nodeID,
		FunctionName: functionName,
		Parameters:   parameters,
		Timestamp:    time.Now().Unix(),
	}

	return peer.Client.ExecuteFunction(ctx, req)
}

func (pm *PeerManager) SendMessage(ctx context.Context, peerID, message string) (*MessageResponse, error) {
	peer, exists := pm.GetPeer(peerID)
	if !exists {
		return nil, fmt.Errorf("peer %s not found", peerID)
	}

	req := &MessageRequest{
		SenderId:  pm.nodeID,
		Message:   message,
		Timestamp: time.Now().Unix(),
	}

	return peer.Client.SendMessage(ctx, req)
}

func (pm *PeerManager) InviteToRoom(ctx context.Context, peerID, roomID string) error {
	peer, exists := pm.GetPeer(peerID)
	if !exists {
		return fmt.Errorf("peer %s not found", peerID)
	}

	req := &RoomInviteRequest{
		SenderId:  pm.nodeID,
		RoomId:    roomID,
		InviteeId: peerID,
	}

	resp, err := peer.Client.InviteToRoom(ctx, req)
	if err != nil {
		return err
	}

	if !resp.Acknowledged {
		return fmt.Errorf("invitation not acknowledged by peer %s", peerID)
	}

	return nil
}

func (pm *PeerManager) JoinRoom(ctx context.Context, hostID, roomID string) (*JoinRoomResponse, error) {
	peer, exists := pm.GetPeer(hostID)
	if !exists {
		return nil, fmt.Errorf("host peer %s not found", hostID)
	}

	req := &JoinRoomRequest{
		SenderId: pm.nodeID,
		RoomId:   roomID,
		HostId:   hostID,
	}

	return peer.Client.JoinRoom(ctx, req)
}

func (pm *PeerManager) LeaveRoom(ctx context.Context, hostID, roomID string) error {
	peer, exists := pm.GetPeer(hostID)
	if !exists {
		return fmt.Errorf("host peer %s not found", hostID)
	}

	req := &LeaveRoomRequest{
		SenderId: pm.nodeID,
		RoomId:   roomID,
	}

	resp, err := peer.Client.LeaveRoom(ctx, req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("failed to leave room: %s", resp.Error)
	}

	return nil
}

func (pm *PeerManager) SendRoomMessage(ctx context.Context, hostID, roomID, message string) error {
	peer, exists := pm.GetPeer(hostID)
	if !exists {
		return fmt.Errorf("host peer %s not found", hostID)
	}

	req := &RoomMessageRequest{
		SenderId:  pm.nodeID,
		RoomId:    roomID,
		Message:   message,
		Timestamp: time.Now().Unix(),
	}

	resp, err := peer.Client.SendRoomMessage(ctx, req)
	if err != nil {
		return err
	}

	if !resp.Acknowledged {
		return fmt.Errorf("room message not acknowledged: %s", resp.Error)
	}

	return nil
}

func (pm *PeerManager) IsConnected(peerID string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	peer, exists := pm.peers[peerID]
	if !exists {
		return false
	}
	
	return peer.Conn.GetState().String() == "READY"
}