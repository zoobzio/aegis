package aegis

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// PeerInfo contains information about a peer node.
type PeerInfo struct {
	ID      string   `json:"id"`
	Address string   `json:"address"`
	Type    NodeType `json:"type"`
}

// Peer represents a connected peer node.
type Peer struct {
	Info   PeerInfo
	Client MeshServiceClient
	Conn   *grpc.ClientConn
}

// PeerManager manages connections to peer nodes.
type PeerManager struct {
	nodeID    string
	peers     map[string]*Peer
	tlsConfig *TLSConfig
	mu        sync.RWMutex
}

// NewPeerManager creates a new peer manager.
func NewPeerManager(nodeID string) *PeerManager {
	return &PeerManager{
		nodeID: nodeID,
		peers:  make(map[string]*Peer),
	}
}

// SetTLSConfig sets the TLS configuration for peer connections.
func (pm *PeerManager) SetTLSConfig(tlsConfig *TLSConfig) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.tlsConfig = tlsConfig
}

// AddPeer adds a new peer connection.
func (pm *PeerManager) AddPeer(info PeerInfo) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.peers[info.ID]; exists {
		return fmt.Errorf("peer %s already exists", info.ID)
	}

	if pm.tlsConfig == nil {
		return fmt.Errorf("TLS configuration is required but not set")
	}

	creds := credentials.NewTLS(pm.tlsConfig.GetClientTLSConfig(info.ID))
	conn, err := grpc.NewClient(info.Address, grpc.WithTransportCredentials(creds))
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

// RemovePeer removes a peer connection.
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

// GetPeer returns a peer by ID.
func (pm *PeerManager) GetPeer(peerID string) (*Peer, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	peer, exists := pm.peers[peerID]
	return peer, exists
}

// GetAllPeers returns all connected peers.
func (pm *PeerManager) GetAllPeers() []*Peer {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	peers := make([]*Peer, 0, len(pm.peers))
	for _, peer := range pm.peers {
		peers = append(peers, peer)
	}
	return peers
}

// GetPeersByType returns peers of a specific type.
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

// PingPeer sends a ping request to a peer.
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

// GetPeerHealth retrieves the health status of a peer.
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

// GetPeerNodeInfo retrieves node information from a peer.
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

// SyncTopology requests topology synchronization from a peer.
func (pm *PeerManager) SyncTopology(ctx context.Context, peerID string, version int64) (*TopologySyncResponse, error) {
	peer, exists := pm.GetPeer(peerID)
	if !exists {
		return nil, fmt.Errorf("peer %s not found", peerID)
	}

	req := &TopologySyncRequest{
		SenderId: pm.nodeID,
		Version:  version,
	}

	return peer.Client.SyncTopology(ctx, req)
}

// Close closes all peer connections.
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

// Count returns the number of connected peers.
func (pm *PeerManager) Count() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.peers)
}

// IsConnected checks if a peer connection is in READY state.
func (pm *PeerManager) IsConnected(peerID string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	peer, exists := pm.peers[peerID]
	if !exists {
		return false
	}

	return peer.Conn.GetState().String() == "READY"
}
