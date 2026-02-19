//go:build testing

package aegis

import "testing"

func TestGetServiceProviders(t *testing.T) {
	topology := NewTopology()

	// Add nodes with different services
	_ = topology.AddNode(NodeInfo{
		ID:      "morpheus-1",
		Name:    "Morpheus Primary",
		Type:    NodeTypeGeneric,
		Address: "morpheus-1:8443",
		Services: []ServiceInfo{
			{Name: "identity", Version: "v1"},
		},
	})

	_ = topology.AddNode(NodeInfo{
		ID:      "morpheus-2",
		Name:    "Morpheus Secondary",
		Type:    NodeTypeGeneric,
		Address: "morpheus-2:8443",
		Services: []ServiceInfo{
			{Name: "identity", Version: "v1"},
		},
	})

	_ = topology.AddNode(NodeInfo{
		ID:      "vicky-1",
		Name:    "Vicky Primary",
		Type:    NodeTypeGeneric,
		Address: "vicky-1:8443",
		Services: []ServiceInfo{
			{Name: "storage", Version: "v1"},
		},
	})

	// Test GetServiceProviders
	identityProviders := topology.GetServiceProviders("identity", "v1")
	if len(identityProviders) != 2 {
		t.Errorf("expected 2 identity providers, got %d", len(identityProviders))
	}

	storageProviders := topology.GetServiceProviders("storage", "v1")
	if len(storageProviders) != 1 {
		t.Errorf("expected 1 storage provider, got %d", len(storageProviders))
	}

	noProviders := topology.GetServiceProviders("nonexistent", "v1")
	if len(noProviders) != 0 {
		t.Errorf("expected 0 providers for nonexistent service, got %d", len(noProviders))
	}
}

func TestGetNodesByService(t *testing.T) {
	topology := NewTopology()

	_ = topology.AddNode(NodeInfo{
		ID:      "node-1",
		Name:    "Node 1",
		Type:    NodeTypeGeneric,
		Address: "node-1:8443",
		Services: []ServiceInfo{
			{Name: "identity", Version: "v1"},
		},
	})

	_ = topology.AddNode(NodeInfo{
		ID:      "node-2",
		Name:    "Node 2",
		Type:    NodeTypeGeneric,
		Address: "node-2:8443",
		Services: []ServiceInfo{
			{Name: "identity", Version: "v2"},
		},
	})

	// GetNodesByService should return both versions
	providers := topology.GetNodesByService("identity")
	if len(providers) != 2 {
		t.Errorf("expected 2 identity providers (any version), got %d", len(providers))
	}
}

func TestNodeBuilderWithServices(t *testing.T) {
	node := NewNode("test-node", "Test Node", NodeTypeGeneric, "localhost:8443")
	node.Services = []ServiceInfo{
		{Name: "identity", Version: "v1"},
		{Name: "audit", Version: "v1"},
	}

	if len(node.Services) != 2 {
		t.Errorf("expected 2 services, got %d", len(node.Services))
	}

	if node.Services[0].Name != "identity" {
		t.Errorf("expected first service to be identity, got %s", node.Services[0].Name)
	}
}
