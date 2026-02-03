package aegis

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type MeshServer struct {
	UnimplementedMeshServiceServer
	node      *Node
	server    *grpc.Server
	listener  net.Listener
	tlsConfig *TLSConfig
}

func NewMeshServer(node *Node) *MeshServer {
	return &MeshServer{
		node: node,
	}
}

func (ms *MeshServer) SetTLSConfig(tlsConfig *TLSConfig) {
	ms.tlsConfig = tlsConfig
}

func (ms *MeshServer) Start() error {
	if ms.node == nil {
		return fmt.Errorf("node cannot be nil")
	}
	
	// TLS is required
	if ms.tlsConfig == nil {
		return fmt.Errorf("TLS configuration is required but not set")
	}

	listener, err := net.Listen("tcp", ms.node.Address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", ms.node.Address, err)
	}

	ms.listener = listener
	
	// Configure gRPC server with TLS
	creds := credentials.NewTLS(ms.tlsConfig.GetServerTLSConfig())
	opts := []grpc.ServerOption{grpc.Creds(creds)}
	
	ms.server = grpc.NewServer(opts...)
	RegisterMeshServiceServer(ms.server, ms)

	go func() {
		ms.server.Serve(listener)
	}()

	return nil
}

func (ms *MeshServer) Stop() {
	if ms.server != nil {
		ms.server.GracefulStop()
	}
	if ms.listener != nil {
		ms.listener.Close()
	}
}

func (ms *MeshServer) Ping(ctx context.Context, req *PingRequest) (*PingResponse, error) {
	return &PingResponse{
		ReceiverId: ms.node.ID,
		Timestamp:  time.Now().Unix(),
		Success:    true,
	}, nil
}

func (ms *MeshServer) GetHealth(ctx context.Context, req *HealthRequest) (*HealthResponse, error) {
	if ms.node.Health == nil {
		return &HealthResponse{
			NodeId:      ms.node.ID,
			Status:      string(HealthStatusUnknown),
			LastChecked: 0,
			Message:     "Health not initialized",
			Error:       "",
		}, nil
	}

	status, lastChecked, message, errMsg := ms.node.Health.Get()

	return &HealthResponse{
		NodeId:      ms.node.ID,
		Status:      string(status),
		LastChecked: lastChecked.Unix(),
		Message:     message,
		Error:       errMsg,
	}, nil
}

func (ms *MeshServer) GetNodeInfo(ctx context.Context, req *NodeInfoRequest) (*NodeInfoResponse, error) {
	healthResp, err := ms.GetHealth(ctx, &HealthRequest{SenderId: req.SenderId})
	if err != nil {
		return nil, fmt.Errorf("failed to get health info: %w", err)
	}

	return &NodeInfoResponse{
		Id:      ms.node.ID,
		Name:    ms.node.Name,
		Type:    string(ms.node.Type),
		Address: ms.node.Address,
		Health:  healthResp,
	}, nil
}

func (ms *MeshServer) NotifyHealthChange(ctx context.Context, req *HealthChangeRequest) (*HealthChangeResponse, error) {
	return &HealthChangeResponse{
		Acknowledged: true,
		ReceiverId:   ms.node.ID,
	}, nil
}

func (ms *MeshServer) ExecuteFunction(ctx context.Context, req *FunctionRequest) (*FunctionResponse, error) {
	if ms.node.Functions == nil {
		return &FunctionResponse{
			ReceiverId: ms.node.ID,
			Success:    false,
			Result:     "",
			Error:      "function registry not initialized",
			Timestamp:  time.Now().Unix(),
		}, nil
	}

	result, err := ms.node.Functions.Execute(ctx, req.FunctionName, req.Parameters)
	
	if err != nil {
		return &FunctionResponse{
			ReceiverId: ms.node.ID,
			Success:    false,
			Result:     "",
			Error:      err.Error(),
			Timestamp:  time.Now().Unix(),
		}, nil
	}

	return &FunctionResponse{
		ReceiverId: ms.node.ID,
		Success:    true,
		Result:     result,
		Error:      "",
		Timestamp:  time.Now().Unix(),
	}, nil
}

func (ms *MeshServer) SendMessage(ctx context.Context, req *MessageRequest) (*MessageResponse, error) {
	return &MessageResponse{
		ReceiverId: ms.node.ID,
		Received:   true,
		Timestamp:  time.Now().Unix(),
	}, nil
}

func (ms *MeshServer) CreateRoom(ctx context.Context, req *CreateRoomRequest) (*CreateRoomResponse, error) {
	if ms.node.Rooms == nil {
		return &CreateRoomResponse{
			Success: false,
			RoomId:  "",
			Error:   "room manager not initialized",
		}, nil
	}
	
	_, err := ms.node.Rooms.CreateRoom(req.RoomId, req.RoomName, ms.node.ID)
	if err != nil {
		return &CreateRoomResponse{
			Success: false,
			RoomId:  "",
			Error:   err.Error(),
		}, nil
	}
	
	return &CreateRoomResponse{
		Success: true,
		RoomId:  req.RoomId,
		Error:   "",
	}, nil
}

func (ms *MeshServer) InviteToRoom(ctx context.Context, req *RoomInviteRequest) (*RoomInviteResponse, error) {
	return &RoomInviteResponse{
		ReceiverId:   ms.node.ID,
		Acknowledged: true,
	}, nil
}

func (ms *MeshServer) JoinRoom(ctx context.Context, req *JoinRoomRequest) (*JoinRoomResponse, error) {
	if ms.node.Rooms == nil {
		return &JoinRoomResponse{
			Success: false,
			Error:   "room manager not initialized",
			Members: nil,
		}, nil
	}
	
	room, exists := ms.node.Rooms.GetHostedRoom(req.RoomId)
	if !exists {
		return &JoinRoomResponse{
			Success: false,
			Error:   "room not found",
			Members: nil,
		}, nil
	}
	
	err := room.AddMember(req.SenderId)
	if err != nil {
		return &JoinRoomResponse{
			Success: false,
			Error:   err.Error(),
			Members: nil,
		}, nil
	}
	
	return &JoinRoomResponse{
		Success: true,
		Error:   "",
		Members: room.GetMembers(),
	}, nil
}

func (ms *MeshServer) LeaveRoom(ctx context.Context, req *LeaveRoomRequest) (*LeaveRoomResponse, error) {
	if ms.node.Rooms == nil {
		return &LeaveRoomResponse{
			Success: false,
			Error:   "room manager not initialized",
		}, nil
	}
	
	room, exists := ms.node.Rooms.GetHostedRoom(req.RoomId)
	if !exists {
		return &LeaveRoomResponse{
			Success: false,
			Error:   "room not found",
		}, nil
	}
	
	err := room.RemoveMember(req.SenderId)
	if err != nil {
		return &LeaveRoomResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}
	
	return &LeaveRoomResponse{
		Success: true,
		Error:   "",
	}, nil
}

func (ms *MeshServer) SendRoomMessage(ctx context.Context, req *RoomMessageRequest) (*RoomMessageResponse, error) {
	if ms.node.Rooms == nil {
		return &RoomMessageResponse{
			Acknowledged: false,
			Error:        "room manager not initialized",
		}, nil
	}
	
	room, exists := ms.node.Rooms.GetHostedRoom(req.RoomId)
	if !exists {
		return &RoomMessageResponse{
			Acknowledged: false,
			Error:        "room not found",
		}, nil
	}
	
	if !room.IsMember(req.SenderId) {
		return &RoomMessageResponse{
			Acknowledged: false,
			Error:        "sender is not a room member",
		}, nil
	}
	
	// Relay message to all members
	members := room.GetMembers()
	for _, memberID := range members {
		if memberID != req.SenderId && memberID != ms.node.ID {
			ms.node.SendMessage(ctx, memberID, fmt.Sprintf("[Room %s] %s: %s", req.RoomId, req.SenderId, req.Message))
		}
	}
	
	return &RoomMessageResponse{
		Acknowledged: true,
		Error:        "",
	}, nil
}