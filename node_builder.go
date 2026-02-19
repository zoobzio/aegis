package aegis

import (
	"fmt"
)

// NodeBuilder provides a fluent interface for creating nodes with required TLS.
type NodeBuilder struct {
	id           string
	name         string
	nodeType     NodeType
	address      string
	services     []ServiceInfo
	registrars   []ServiceRegistrar
	certDir      string
	tlsOptions   *TLSOptions
}

// NewNodeBuilder creates a new node builder.
func NewNodeBuilder() *NodeBuilder {
	return &NodeBuilder{
		nodeType: NodeTypeGeneric,
		certDir:  "./certs",
	}
}

// WithID sets the node ID.
func (nb *NodeBuilder) WithID(id string) *NodeBuilder {
	nb.id = id
	return nb
}

// WithName sets the node name.
func (nb *NodeBuilder) WithName(name string) *NodeBuilder {
	nb.name = name
	return nb
}

// WithType sets the node type.
func (nb *NodeBuilder) WithType(nodeType NodeType) *NodeBuilder {
	nb.nodeType = nodeType
	return nb
}

// WithAddress sets the node address.
func (nb *NodeBuilder) WithAddress(address string) *NodeBuilder {
	nb.address = address
	return nb
}

// WithServices sets the services this node provides.
func (nb *NodeBuilder) WithServices(services ...ServiceInfo) *NodeBuilder {
	nb.services = services
	return nb
}

// WithServiceRegistration adds a callback to register gRPC services on the server.
func (nb *NodeBuilder) WithServiceRegistration(r ServiceRegistrar) *NodeBuilder {
	nb.registrars = append(nb.registrars, r)
	return nb
}

// WithCertDir sets the certificate directory.
func (nb *NodeBuilder) WithCertDir(certDir string) *NodeBuilder {
	nb.certDir = certDir
	return nb
}

// WithTLSOptions sets custom TLS options.
func (nb *NodeBuilder) WithTLSOptions(opts *TLSOptions) *NodeBuilder {
	nb.tlsOptions = opts
	return nb
}

// Build creates the node with TLS enabled.
func (nb *NodeBuilder) Build() (*Node, error) {
	if nb.id == "" {
		return nil, fmt.Errorf("node ID is required")
	}
	if nb.name == "" {
		return nil, fmt.Errorf("node name is required")
	}
	if nb.address == "" {
		return nil, fmt.Errorf("node address is required")
	}

	node := NewNode(nb.id, nb.name, nb.nodeType, nb.address)

	// Set services and update topology entry
	if len(nb.services) > 0 {
		node.Services = nb.services
		nodeInfo := NodeInfo{
			ID:       nb.id,
			Name:     nb.name,
			Type:     nb.nodeType,
			Address:  nb.address,
			Services: nb.services,
		}
		_ = node.Topology.UpdateNode(nodeInfo)
	}

	var tlsConfig *TLSConfig
	var err error

	switch {
	case nb.tlsOptions != nil:
		tlsConfig, err = LoadTLSConfig(nb.tlsOptions)
	case nb.certDir != "":
		tlsConfig, err = LoadOrGenerateTLS(nb.id, nb.certDir)
	default:
		opts := DefaultTLSOptions(nb.id, "./certs")
		tlsConfig, err = LoadTLSConfig(opts)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load TLS configuration: %w", err)
	}

	node.TLSConfig = tlsConfig
	node.MeshServer.SetTLSConfig(tlsConfig)
	node.PeerManager.SetTLSConfig(tlsConfig)

	// Register service registrars
	for _, r := range nb.registrars {
		node.MeshServer.RegisterService(r)
	}

	return node, nil
}

// NewSecureNode is a convenience function that creates a node with TLS enabled.
func NewSecureNode(id, name string, nodeType NodeType, address, certDir string) (*Node, error) {
	return NewNodeBuilder().
		WithID(id).
		WithName(name).
		WithType(nodeType).
		WithAddress(address).
		WithCertDir(certDir).
		Build()
}
