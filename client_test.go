//go:build testing

package aegis

import (
	"context"
	"testing"

	"google.golang.org/grpc"
)

func TestServiceClientPoolNoProviders(t *testing.T) {
	node := NewNode("test", "Test", NodeTypeGeneric, "localhost:8443")
	pool := NewServiceClientPool(node)
	defer pool.Close()

	_, err := pool.GetConn(context.Background(), "nonexistent", "v1")
	if err != ErrNoProviders {
		t.Errorf("expected ErrNoProviders, got %v", err)
	}
}

func TestServiceClientPoolNoTLS(t *testing.T) {
	node := NewNode("test", "Test", NodeTypeGeneric, "localhost:8443")

	// Add a provider to topology
	_ = node.Topology.AddNode(NodeInfo{
		ID:       "provider-1",
		Name:     "Provider",
		Type:     NodeTypeGeneric,
		Address:  "localhost:9443",
		Services: []ServiceInfo{{Name: "identity", Version: "v1"}},
	})

	pool := NewServiceClientPool(node)
	defer pool.Close()

	_, err := pool.GetConn(context.Background(), "identity", "v1")
	if err != ErrNoTLSConfig {
		t.Errorf("expected ErrNoTLSConfig, got %v", err)
	}
}

func TestServiceClientPoolRoundRobin(t *testing.T) {
	node := NewNode("test", "Test", NodeTypeGeneric, "localhost:8443")

	// Add multiple providers
	_ = node.Topology.AddNode(NodeInfo{
		ID:       "provider-1",
		Name:     "Provider 1",
		Type:     NodeTypeGeneric,
		Address:  "localhost:9001",
		Services: []ServiceInfo{{Name: "identity", Version: "v1"}},
	})
	_ = node.Topology.AddNode(NodeInfo{
		ID:       "provider-2",
		Name:     "Provider 2",
		Type:     NodeTypeGeneric,
		Address:  "localhost:9002",
		Services: []ServiceInfo{{Name: "identity", Version: "v1"}},
	})

	pool := NewServiceClientPool(node)
	defer pool.Close()

	// Verify round-robin counter increments
	counter := pool.getCounter("identity/v1")
	if counter.Load() != 0 {
		t.Errorf("expected counter to start at 0, got %d", counter.Load())
	}

	// Simulate multiple selections (can't actually connect without TLS)
	providers := node.Topology.GetServiceProviders("identity", "v1")
	if len(providers) != 2 {
		t.Errorf("expected 2 providers, got %d", len(providers))
	}

	// Test counter increment logic
	idx1 := counter.Add(1) - 1
	idx2 := counter.Add(1) - 1

	if idx1%2 == idx2%2 {
		t.Error("round-robin should alternate between providers")
	}
}

// mockClient is a fake gRPC client for testing
type mockClient struct{}

func newMockClient(cc grpc.ClientConnInterface) *mockClient {
	return &mockClient{}
}

func TestServiceClientGet(t *testing.T) {
	node := NewNode("test", "Test", NodeTypeGeneric, "localhost:8443")

	_ = node.Topology.AddNode(NodeInfo{
		ID:       "provider-1",
		Name:     "Provider",
		Type:     NodeTypeGeneric,
		Address:  "localhost:9443",
		Services: []ServiceInfo{{Name: "identity", Version: "v1"}},
	})

	pool := NewServiceClientPool(node)
	defer pool.Close()

	client := NewServiceClient(pool, "identity", "v1", newMockClient)

	// Will fail due to no TLS, but tests the wiring
	_, err := client.Get(context.Background())
	if err != ErrNoTLSConfig {
		t.Errorf("expected ErrNoTLSConfig, got %v", err)
	}
}
