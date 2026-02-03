package aegis

import (
	"encoding/json"
	"testing"
)

func TestNewNode(t *testing.T) {
	node := NewNode("node-1", "test-node", NodeTypeGateway, "localhost:8080")
	
	if node.ID != "node-1" {
		t.Errorf("expected ID 'node-1', got '%s'", node.ID)
	}
	if node.Name != "test-node" {
		t.Errorf("expected name 'test-node', got '%s'", node.Name)
	}
	if node.Type != NodeTypeGateway {
		t.Errorf("expected type %s, got %s", NodeTypeGateway, node.Type)
	}
	if node.Address != "localhost:8080" {
		t.Errorf("expected address 'localhost:8080', got '%s'", node.Address)
	}
}

func TestNodeString(t *testing.T) {
	node := NewNode("node-1", "test-node", NodeTypeProcessor, "127.0.0.1:9090")
	expected := "Node[processor:node-1] test-node (processor) @ 127.0.0.1:9090"
	
	if node.String() != expected {
		t.Errorf("expected string '%s', got '%s'", expected, node.String())
	}
}

func TestNodeJSONMarshal(t *testing.T) {
	node := NewNode("node-1", "test-node", NodeTypeStorage, "10.0.0.1:5432")
	
	data, err := json.Marshal(node)
	if err != nil {
		t.Fatalf("failed to marshal node: %v", err)
	}
	
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}
	
	if result["id"] != "node-1" {
		t.Errorf("expected id 'node-1', got '%v'", result["id"])
	}
	if result["name"] != "test-node" {
		t.Errorf("expected name 'test-node', got '%v'", result["name"])
	}
	if result["type"] != "storage" {
		t.Errorf("expected type 'storage', got '%v'", result["type"])
	}
	if result["address"] != "10.0.0.1:5432" {
		t.Errorf("expected address '10.0.0.1:5432', got '%v'", result["address"])
	}
	if result["health"] == nil {
		t.Error("expected health field to be present")
	}
}

func TestNodeJSONUnmarshal(t *testing.T) {
	jsonData := `{"id":"node-2","name":"another-node","type":"generic","address":"192.168.1.100:3000"}`
	
	var node Node
	err := json.Unmarshal([]byte(jsonData), &node)
	if err != nil {
		t.Fatalf("failed to unmarshal node: %v", err)
	}
	
	if node.ID != "node-2" {
		t.Errorf("expected ID 'node-2', got '%s'", node.ID)
	}
	if node.Name != "another-node" {
		t.Errorf("expected name 'another-node', got '%s'", node.Name)
	}
	if node.Type != NodeTypeGeneric {
		t.Errorf("expected type %s, got %s", NodeTypeGeneric, node.Type)
	}
	if node.Address != "192.168.1.100:3000" {
		t.Errorf("expected address '192.168.1.100:3000', got '%s'", node.Address)
	}
}

func TestNodeValidate(t *testing.T) {
	tests := []struct {
		name    string
		node    *Node
		wantErr bool
	}{
		{
			name: "valid node",
			node: NewNode("node-1", "test", NodeTypeGateway, "localhost:8080"),
			wantErr: false,
		},
		{
			name: "empty ID",
			node: NewNode("", "test", NodeTypeGateway, "localhost:8080"),
			wantErr: true,
		},
		{
			name: "empty name",
			node: NewNode("node-1", "", NodeTypeGateway, "localhost:8080"),
			wantErr: true,
		},
		{
			name: "empty address",
			node: NewNode("node-1", "test", NodeTypeGateway, ""),
			wantErr: true,
		},
		{
			name: "invalid address format",
			node: NewNode("node-1", "test", NodeTypeGateway, "not-a-valid-address"),
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.node.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Node.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}