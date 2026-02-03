package main

import (
	"context"
	"log"
	"os"
	
	"github.com/zoobzio/aegis"
)

func main() {
	// Example 1: Load configuration from file
	fileBasedNode()
	
	// Example 2: Load certificates from environment
	envBasedNode()
	
	// Example 3: Custom TLS options
	customTLSNode()
}

// fileBasedNode demonstrates file-based certificate loading
func fileBasedNode() {
	log.Println("=== File-based Certificate Example ===")
	
	// Load configuration from file
	config, err := aegis.LoadConfig("config-file.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	// Create node with configuration
	node, err := aegis.NewNodeBuilder().
		WithConfig(config).
		Build()
	if err != nil {
		log.Fatalf("Failed to create node: %v", err)
	}
	
	// Start the node
	if err := node.MeshServer.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer node.MeshServer.Stop()
	
	log.Printf("Node %s started with file-based certificates\n", node.ID)
}

// envBasedNode demonstrates environment-based certificate loading
func envBasedNode() {
	log.Println("=== Environment-based Certificate Example ===")
	
	// Set environment variables (in production, these would be set externally)
	os.Setenv("AEGIS_TLS_CERT", `-----BEGIN CERTIFICATE-----
MIIDXTCCAkWgAwIBAgIJAKl1GZcpQR5rMA0GCSqGSIb3DQEBCwUAMEUxCzAJBgNV
... (certificate content) ...
-----END CERTIFICATE-----`)
	
	os.Setenv("AEGIS_TLS_KEY", `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAvLYcyu8f3skuRyUgeeNpeDvYBCDcgq+LsWap6zbX5f8oLqp4
... (key content) ...
-----END RSA PRIVATE KEY-----`)
	
	os.Setenv("AEGIS_TLS_CA", `-----BEGIN CERTIFICATE-----
MIIDXTCCAkWgAwIBAgIJAKl1GZcpQR5rMA0GCSqGSIb3DQEBCwUAMEUxCzAJBgNV
... (CA certificate content) ...
-----END CERTIFICATE-----`)
	
	// Create TLS options for environment-based loading
	tlsOpts := &aegis.TLSOptions{
		Source:       aegis.CertSourceEnv,
		CertEnvVar:   "AEGIS_TLS_CERT",
		KeyEnvVar:    "AEGIS_TLS_KEY",
		CAEnvVar:     "AEGIS_TLS_CA",
		VerifyChain:  true,
		AllowExpired: false,
	}
	
	// Create node with environment-based certificates
	node, err := aegis.NewNodeBuilder().
		WithID("env-node").
		WithName("Environment Node").
		WithType(aegis.NodeTypeGeneric).
		WithAddress("localhost:8003").
		WithTLSOptions(tlsOpts).
		Build()
	if err != nil {
		log.Fatalf("Failed to create node: %v", err)
	}
	
	log.Printf("Node %s created with environment-based certificates\n", node.ID)
}

// customTLSNode demonstrates custom TLS configuration
func customTLSNode() {
	log.Println("=== Custom TLS Configuration Example ===")
	
	// Create custom TLS options with specific requirements
	tlsOpts := &aegis.TLSOptions{
		Source:       aegis.CertSourceFile,
		CertFile:     "/secure/certs/node-cert.pem",
		KeyFile:      "/secure/certs/node-key.pem",
		CAFile:       "/secure/certs/trusted-ca-bundle.pem",
		VerifyChain:  true,
		AllowExpired: false,
		RequiredSANs: []string{
			"node.aegis.local",
			"*.aegis.local",
			"10.0.1.100",
		},
	}
	
	// Create node with custom TLS options
	node, err := aegis.NewNodeBuilder().
		WithID("secure-node").
		WithName("High Security Node").
		WithType(aegis.NodeTypeGateway).
		WithAddress("10.0.1.100:8004").
		WithTLSOptions(tlsOpts).
		Build()
	if err != nil {
		log.Fatalf("Failed to create node: %v", err)
	}
	
	log.Printf("Node %s created with custom TLS validation\n", node.ID)
}

// Example of connecting nodes with proper certificate validation
func connectNodes() {
	ctx := context.Background()
	
	// Load configuration from environment
	config, err := aegis.LoadConfigFromEnv("AEGIS")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	// Create node
	node, err := aegis.NewNodeBuilder().
		WithConfig(config).
		Build()
	if err != nil {
		log.Fatalf("Failed to create node: %v", err)
	}
	
	// Add a peer - the TLS configuration ensures mutual authentication
	peerInfo := aegis.PeerInfo{
		ID:      "peer-node-1",
		Address: "peer1.aegis.local:8001",
		Type:    aegis.NodeTypeProcessor,
	}
	
	if err := node.PeerManager.AddPeer(peerInfo); err != nil {
		log.Fatalf("Failed to add peer: %v", err)
	}
	
	// Ping the peer - this will use mTLS
	resp, err := node.PeerManager.PingPeer(ctx, peerInfo.ID)
	if err != nil {
		log.Fatalf("Failed to ping peer: %v", err)
	}
	
	log.Printf("Successfully connected to peer %s via mTLS\n", resp.ReceiverId)
}