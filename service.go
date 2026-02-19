package aegis

// ServiceInfo represents a service that a node provides.
// This is the internal representation; Service from mesh.pb.go is used for wire format.
type ServiceInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// GetServiceProviders returns all nodes that provide the specified service.
func (t *Topology) GetServiceProviders(name, version string) []NodeInfo {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var providers []NodeInfo
	for _, node := range t.Nodes {
		for _, svc := range node.Services {
			if svc.Name == name && svc.Version == version {
				providers = append(providers, node)
				break
			}
		}
	}
	return providers
}

// GetNodesByService returns all nodes that provide any version of the specified service.
func (t *Topology) GetNodesByService(name string) []NodeInfo {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var providers []NodeInfo
	for _, node := range t.Nodes {
		for _, svc := range node.Services {
			if svc.Name == name {
				providers = append(providers, node)
				break
			}
		}
	}
	return providers
}
