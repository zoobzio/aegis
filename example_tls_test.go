package aegis

import (
	"context"
	"fmt"
	"log"
	"time"
)

// ExampleTLSMeshSetup demonstrates how to set up a mesh network with mTLS
func ExampleTLSMeshSetup() {
	// Create two nodes with mandatory TLS
	node1, err := NewSecureNode("node-1", "Gateway Node", NodeTypeGateway, "localhost:8001", "./certs")
	if err != nil {
		log.Fatalf("Failed to create node1: %v", err)
	}
	
	node2, err := NewSecureNode("node-2", "Processor Node", NodeTypeProcessor, "localhost:8002", "./certs")
	if err != nil {
		log.Fatalf("Failed to create node2: %v", err)
	}
	
	// Start the mesh servers with TLS enabled
	if err := node1.MeshServer.Start(); err != nil {
		log.Fatalf("Failed to start node1 server: %v", err)
	}
	defer node1.MeshServer.Stop()
	
	if err := node2.MeshServer.Start(); err != nil {
		log.Fatalf("Failed to start node2 server: %v", err)
	}
	defer node2.MeshServer.Stop()
	
	// Give servers time to start
	time.Sleep(100 * time.Millisecond)
	
	// Now node2 can connect to node1 using mTLS
	peerInfo := PeerInfo{
		ID:      node1.ID,
		Address: node1.Address,
		Type:    node1.Type,
	}
	
	if err := node2.PeerManager.AddPeer(peerInfo); err != nil {
		log.Fatalf("Failed to add peer: %v", err)
	}
	
	// Test the secure connection
	ctx := context.Background()
	resp, err := node2.PeerManager.PingPeer(ctx, node1.ID)
	if err != nil {
		log.Fatalf("Failed to ping peer: %v", err)
	}
	
	fmt.Printf("Secure ping successful! Response from %s\n", resp.ReceiverId)
}

// ExampleGenerateNodeCertificates demonstrates generating certificates for a new node
func ExampleGenerateNodeCertificates() {
	// This will generate a new CA (if needed) and node certificates
	tlsConfig, err := LoadOrGenerateTLS("my-node-id", "./certs")
	if err != nil {
		log.Fatalf("Failed to generate certificates: %v", err)
	}
	
	fmt.Printf("Generated certificates for node: %s\n", tlsConfig.ServerName)
	fmt.Printf("Certificate will be valid for DNS names and IPs configured in generateNodeCertificate()\n")
}

// ExampleNodeBuilderPattern demonstrates using the builder pattern
func ExampleNodeBuilderPattern() {
	// Use the builder pattern for more control
	node, err := NewNodeBuilder().
		WithID("custom-node").
		WithName("My Custom Node").
		WithType(NodeTypeStorage).
		WithAddress("localhost:8003").
		WithCertDir("/custom/cert/path").
		Build()
		
	if err != nil {
		log.Fatalf("Failed to build node: %v", err)
	}
	
	// Node is created with TLS already enabled
	fmt.Printf("Node %s created with mandatory TLS\n", node.ID)
}