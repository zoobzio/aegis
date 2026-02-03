package aegis

import (
	"context"
	"testing"
)

func TestNodeSendMessage(t *testing.T) {
	node := NewNode("test-node", "Test Node", NodeTypeGeneric, "localhost:8080")
	
	ctx := context.Background()
	
	err := node.SendMessage(ctx, "nonexistent-peer", "Hello")
	if err == nil {
		t.Error("expected error sending message to nonexistent peer")
	}
}

func TestNodeBroadcastMessage(t *testing.T) {
	node := NewNode("test-node", "Test Node", NodeTypeGeneric, "localhost:8080")
	
	ctx := context.Background()
	
	err := node.BroadcastMessage(ctx, "Hello everyone")
	if err != nil {
		t.Errorf("unexpected error broadcasting to zero peers: %v", err)
	}
}

func TestPeerManagerSendMessage(t *testing.T) {
	pm := NewPeerManager("test-node")
	
	ctx := context.Background()
	
	_, err := pm.SendMessage(ctx, "nonexistent", "Hello")
	if err == nil {
		t.Error("expected error sending message to nonexistent peer")
	}
}

func TestMeshServerSendMessage(t *testing.T) {
	node := NewNode("test-node", "Test Node", NodeTypeGeneric, "localhost:8080")
	ms := NewMeshServer(node)
	
	ctx := context.Background()
	req := &MessageRequest{
		SenderId:  "sender-node",
		Message:   "Test message",
		Timestamp: 123456789,
	}
	
	resp, err := ms.SendMessage(ctx, req)
	if err != nil {
		t.Errorf("unexpected error handling message: %v", err)
	}
	
	if resp.ReceiverId != "test-node" {
		t.Errorf("expected receiver_id 'test-node', got '%s'", resp.ReceiverId)
	}
	
	if !resp.Received {
		t.Error("expected message to be marked as received")
	}
	
	if resp.Timestamp <= 0 {
		t.Error("expected timestamp to be set")
	}
}