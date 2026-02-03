package aegis

import (
	"fmt"
	"sync"
	"time"
)

type Room struct {
	ID      string
	Name    string
	HostID  string
	Members map[string]bool
	Created time.Time
	mu      sync.RWMutex
}

func NewRoom(id, name, hostID string) *Room {
	return &Room{
		ID:      id,
		Name:    name,
		HostID:  hostID,
		Members: map[string]bool{hostID: true},
		Created: time.Now(),
	}
}

func (r *Room) AddMember(nodeID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if r.Members[nodeID] {
		return fmt.Errorf("node %s is already a member", nodeID)
	}
	
	r.Members[nodeID] = true
	return nil
}

func (r *Room) RemoveMember(nodeID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if !r.Members[nodeID] {
		return fmt.Errorf("node %s is not a member", nodeID)
	}
	
	if nodeID == r.HostID {
		return fmt.Errorf("host cannot leave the room")
	}
	
	delete(r.Members, nodeID)
	return nil
}

func (r *Room) IsMember(nodeID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	return r.Members[nodeID]
}

func (r *Room) GetMembers() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	members := make([]string, 0, len(r.Members))
	for member := range r.Members {
		members = append(members, member)
	}
	return members
}

func (r *Room) MemberCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	return len(r.Members)
}

type RoomManager struct {
	hostedRooms map[string]*Room
	joinedRooms map[string]string // roomID -> hostID
	mu          sync.RWMutex
}

func NewRoomManager() *RoomManager {
	return &RoomManager{
		hostedRooms: make(map[string]*Room),
		joinedRooms: make(map[string]string),
	}
}

func (rm *RoomManager) CreateRoom(id, name, hostID string) (*Room, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	if _, exists := rm.hostedRooms[id]; exists {
		return nil, fmt.Errorf("room %s already exists", id)
	}
	
	room := NewRoom(id, name, hostID)
	rm.hostedRooms[id] = room
	
	return room, nil
}

func (rm *RoomManager) GetHostedRoom(roomID string) (*Room, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	room, exists := rm.hostedRooms[roomID]
	return room, exists
}

func (rm *RoomManager) DeleteRoom(roomID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	if _, exists := rm.hostedRooms[roomID]; !exists {
		return fmt.Errorf("room %s not found", roomID)
	}
	
	delete(rm.hostedRooms, roomID)
	return nil
}

func (rm *RoomManager) JoinRoom(roomID, hostID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	if currentHost, exists := rm.joinedRooms[roomID]; exists {
		if currentHost == hostID {
			return fmt.Errorf("already joined room %s", roomID)
		}
		return fmt.Errorf("room %s conflict: different hosts", roomID)
	}
	
	rm.joinedRooms[roomID] = hostID
	return nil
}

func (rm *RoomManager) LeaveRoom(roomID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	if _, exists := rm.joinedRooms[roomID]; !exists {
		return fmt.Errorf("not a member of room %s", roomID)
	}
	
	delete(rm.joinedRooms, roomID)
	return nil
}

func (rm *RoomManager) GetJoinedRoomHost(roomID string) (string, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	hostID, exists := rm.joinedRooms[roomID]
	return hostID, exists
}

func (rm *RoomManager) ListHostedRooms() []string {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	rooms := make([]string, 0, len(rm.hostedRooms))
	for roomID := range rm.hostedRooms {
		rooms = append(rooms, roomID)
	}
	return rooms
}

func (rm *RoomManager) ListJoinedRooms() []string {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	rooms := make([]string, 0, len(rm.joinedRooms))
	for roomID := range rm.joinedRooms {
		rooms = append(rooms, roomID)
	}
	return rooms
}