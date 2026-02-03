package aegis

import (
	"fmt"
)

// NodeBuilder provides a fluent interface for creating nodes with required TLS
type NodeBuilder struct {
	id         string
	name       string
	nodeType   NodeType
	address    string
	certDir    string
	config     *MeshConfig
	tlsOptions *TLSOptions
}

// NewNodeBuilder creates a new node builder
func NewNodeBuilder() *NodeBuilder {
	return &NodeBuilder{
		nodeType: NodeTypeGeneric,
		certDir:  "./certs", // default certificate directory
	}
}

// WithID sets the node ID
func (nb *NodeBuilder) WithID(id string) *NodeBuilder {
	nb.id = id
	return nb
}

// WithName sets the node name
func (nb *NodeBuilder) WithName(name string) *NodeBuilder {
	nb.name = name
	return nb
}

// WithType sets the node type
func (nb *NodeBuilder) WithType(nodeType NodeType) *NodeBuilder {
	nb.nodeType = nodeType
	return nb
}

// WithAddress sets the node address
func (nb *NodeBuilder) WithAddress(address string) *NodeBuilder {
	nb.address = address
	return nb
}

// WithCertDir sets the certificate directory
func (nb *NodeBuilder) WithCertDir(certDir string) *NodeBuilder {
	nb.certDir = certDir
	return nb
}

// WithConfig sets the mesh configuration
func (nb *NodeBuilder) WithConfig(config *MeshConfig) *NodeBuilder {
	nb.config = config
	return nb
}

// WithTLSOptions sets custom TLS options
func (nb *NodeBuilder) WithTLSOptions(opts *TLSOptions) *NodeBuilder {
	nb.tlsOptions = opts
	return nb
}

// Build creates the node with TLS enabled
func (nb *NodeBuilder) Build() (*Node, error) {
	// Validate required fields
	if nb.id == "" {
		return nil, fmt.Errorf("node ID is required")
	}
	if nb.name == "" {
		return nil, fmt.Errorf("node name is required")
	}
	if nb.address == "" {
		return nil, fmt.Errorf("node address is required")
	}
	
	// Create the node
	node := NewNode(nb.id, nb.name, nb.nodeType, nb.address)
	
	// Determine TLS configuration source
	var tlsConfig *TLSConfig
	var err error
	
	if nb.tlsOptions != nil {
		// Use provided TLS options
		tlsConfig, err = LoadTLSConfig(nb.tlsOptions)
	} else if nb.config != nil {
		// Use mesh configuration
		tlsConfig, err = LoadTLSConfig(nb.config.ToTLSOptions())
	} else if nb.certDir != "" {
		// Use legacy certificate directory
		tlsConfig, err = LoadOrGenerateTLS(nb.id, nb.certDir)
	} else {
		// Use default file-based options
		opts := DefaultTLSOptions(nb.id, "./certs")
		tlsConfig, err = LoadTLSConfig(opts)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS configuration: %w", err)
	}
	
	// Apply TLS configuration
	node.TLSConfig = tlsConfig
	node.MeshServer.SetTLSConfig(tlsConfig)
	node.PeerManager.SetTLSConfig(tlsConfig)
	
	// Apply additional configuration if provided
	if nb.config != nil {
		// Apply network settings, health settings, etc.
		// This could be expanded based on needs
	}
	
	return node, nil
}

// NewSecureNode is a convenience function that creates a node with TLS enabled
func NewSecureNode(id, name string, nodeType NodeType, address, certDir string) (*Node, error) {
	return NewNodeBuilder().
		WithID(id).
		WithName(name).
		WithType(nodeType).
		WithAddress(address).
		WithCertDir(certDir).
		Build()
}