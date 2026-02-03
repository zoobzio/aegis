package aegis

import (
	"context"
	"testing"
)

func TestNewRoom(t *testing.T) {
	room := NewRoom("room-1", "Test Room", "host-node")
	
	if room.ID != "room-1" {
		t.Errorf("expected room ID 'room-1', got '%s'", room.ID)
	}
	
	if room.Name != "Test Room" {
		t.Errorf("expected room name 'Test Room', got '%s'", room.Name)
	}
	
	if room.HostID != "host-node" {
		t.Errorf("expected host ID 'host-node', got '%s'", room.HostID)
	}
	
	if !room.IsMember("host-node") {
		t.Error("host should be a member of the room")
	}
	
	if room.MemberCount() != 1 {
		t.Errorf("expected 1 member, got %d", room.MemberCount())
	}
}

func TestRoomAddRemoveMember(t *testing.T) {
	room := NewRoom("room-1", "Test Room", "host-node")
	
	err := room.AddMember("member-1")
	if err != nil {
		t.Errorf("failed to add member: %v", err)
	}
	
	if !room.IsMember("member-1") {
		t.Error("member-1 should be in the room")
	}
	
	if room.MemberCount() != 2 {
		t.Errorf("expected 2 members, got %d", room.MemberCount())
	}
	
	err = room.AddMember("member-1")
	if err == nil {
		t.Error("expected error adding duplicate member")
	}
	
	err = room.RemoveMember("member-1")
	if err != nil {
		t.Errorf("failed to remove member: %v", err)
	}
	
	if room.IsMember("member-1") {
		t.Error("member-1 should not be in the room")
	}
	
	err = room.RemoveMember("host-node")
	if err == nil {
		t.Error("expected error removing host from room")
	}
}

func TestRoomGetMembers(t *testing.T) {
	room := NewRoom("room-1", "Test Room", "host-node")
	room.AddMember("member-1")
	room.AddMember("member-2")
	
	members := room.GetMembers()
	if len(members) != 3 {
		t.Errorf("expected 3 members, got %d", len(members))
	}
	
	memberMap := make(map[string]bool)
	for _, m := range members {
		memberMap[m] = true
	}
	
	if !memberMap["host-node"] || !memberMap["member-1"] || !memberMap["member-2"] {
		t.Error("missing expected members")
	}
}

func TestNewRoomManager(t *testing.T) {
	rm := NewRoomManager()
	
	if len(rm.ListHostedRooms()) != 0 {
		t.Error("expected no hosted rooms initially")
	}
	
	if len(rm.ListJoinedRooms()) != 0 {
		t.Error("expected no joined rooms initially")
	}
}

func TestRoomManagerCreateRoom(t *testing.T) {
	rm := NewRoomManager()
	
	room, err := rm.CreateRoom("room-1", "Test Room", "host-node")
	if err != nil {
		t.Errorf("failed to create room: %v", err)
	}
	
	if room.ID != "room-1" {
		t.Errorf("expected room ID 'room-1', got '%s'", room.ID)
	}
	
	_, err = rm.CreateRoom("room-1", "Duplicate Room", "host-node")
	if err == nil {
		t.Error("expected error creating duplicate room")
	}
	
	hostedRooms := rm.ListHostedRooms()
	if len(hostedRooms) != 1 || hostedRooms[0] != "room-1" {
		t.Error("expected room-1 in hosted rooms")
	}
}

func TestRoomManagerJoinLeaveRoom(t *testing.T) {
	rm := NewRoomManager()
	
	err := rm.JoinRoom("room-1", "host-node")
	if err != nil {
		t.Errorf("failed to join room: %v", err)
	}
	
	host, exists := rm.GetJoinedRoomHost("room-1")
	if !exists || host != "host-node" {
		t.Error("expected to find joined room with correct host")
	}
	
	err = rm.JoinRoom("room-1", "different-host")
	if err == nil {
		t.Error("expected error joining room with different host")
	}
	
	err = rm.LeaveRoom("room-1")
	if err != nil {
		t.Errorf("failed to leave room: %v", err)
	}
	
	_, exists = rm.GetJoinedRoomHost("room-1")
	if exists {
		t.Error("room should not exist after leaving")
	}
}

func TestNodeRoomMethods(t *testing.T) {
	node := NewNode("test-node", "Test Node", NodeTypeGeneric, "localhost:8080")
	
	if node.Rooms == nil {
		t.Error("expected room manager to be initialized")
	}
	
	err := node.CreateRoom("room-1", "Test Room")
	if err != nil {
		t.Errorf("failed to create room: %v", err)
	}
	
	hosted, joined := node.ListRooms()
	if len(hosted) != 1 || hosted[0] != "room-1" {
		t.Error("expected room-1 in hosted rooms")
	}
	if len(joined) != 0 {
		t.Error("expected no joined rooms")
	}
}

func TestNodeInviteToRoom(t *testing.T) {
	node := NewNode("test-node", "Test Node", NodeTypeGeneric, "localhost:8080")
	node.CreateRoom("room-1", "Test Room")
	
	ctx := context.Background()
	
	err := node.InviteToRoom(ctx, "room-1", "nonexistent-peer")
	if err == nil {
		t.Error("expected error inviting nonexistent peer")
	}
	
	err = node.InviteToRoom(ctx, "nonexistent-room", "peer-1")
	if err == nil {
		t.Error("expected error inviting to nonexistent room")
	}
}

func TestNodeSendToRoom(t *testing.T) {
	node := NewNode("test-node", "Test Node", NodeTypeGeneric, "localhost:8080")
	node.CreateRoom("room-1", "Test Room")
	
	ctx := context.Background()
	
	err := node.SendToRoom(ctx, "room-1", "Test message")
	if err != nil {
		t.Errorf("unexpected error sending to hosted room with no other members: %v", err)
	}
	
	err = node.SendToRoom(ctx, "nonexistent-room", "Test message")
	if err == nil {
		t.Error("expected error sending to nonexistent room")
	}
}