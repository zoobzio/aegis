package aegis

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	// ErrNoProviders is returned when no providers are found for a service.
	ErrNoProviders = errors.New("no providers available for service")
	// ErrNoTLSConfig is returned when the node has no TLS configuration.
	ErrNoTLSConfig = errors.New("node has no TLS configuration")
)

// ServiceClientPool manages gRPC connections to service providers.
type ServiceClientPool struct {
	node  *Node
	conns map[string]*grpc.ClientConn
	mu    sync.RWMutex

	// Round-robin counters per service
	counters map[string]*atomic.Uint64
}

// NewServiceClientPool creates a connection pool for service clients.
func NewServiceClientPool(node *Node) *ServiceClientPool {
	return &ServiceClientPool{
		node:     node,
		conns:    make(map[string]*grpc.ClientConn),
		counters: make(map[string]*atomic.Uint64),
	}
}

// getOrCreateConn returns an existing connection or creates a new one.
func (p *ServiceClientPool) getOrCreateConn(ctx context.Context, address string) (*grpc.ClientConn, error) {
	p.mu.RLock()
	conn, exists := p.conns[address]
	p.mu.RUnlock()

	if exists {
		return conn, nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check after acquiring write lock
	if conn, exists = p.conns[address]; exists {
		return conn, nil
	}

	if p.node.TLSConfig == nil {
		return nil, ErrNoTLSConfig
	}

	creds := credentials.NewTLS(p.node.TLSConfig.GetClientTLSConfig(address))
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, err
	}

	p.conns[address] = conn
	return conn, nil
}

// getCounter returns the round-robin counter for a service key.
func (p *ServiceClientPool) getCounter(serviceKey string) *atomic.Uint64 {
	p.mu.RLock()
	counter, exists := p.counters[serviceKey]
	p.mu.RUnlock()

	if exists {
		return counter
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if counter, exists = p.counters[serviceKey]; exists {
		return counter
	}

	counter = &atomic.Uint64{}
	p.counters[serviceKey] = counter
	return counter
}

// GetConn returns a connection to a provider of the specified service.
// Uses round-robin to distribute calls across providers.
func (p *ServiceClientPool) GetConn(ctx context.Context, name, version string) (*grpc.ClientConn, error) {
	if p.node.Topology == nil {
		return nil, ErrNoProviders
	}

	providers := p.node.Topology.GetServiceProviders(name, version)
	if len(providers) == 0 {
		return nil, ErrNoProviders
	}

	// Round-robin selection
	serviceKey := name + "/" + version
	counter := p.getCounter(serviceKey)
	idx := counter.Add(1) - 1
	provider := providers[idx%uint64(len(providers))]

	return p.getOrCreateConn(ctx, provider.Address)
}

// Close closes all connections in the pool.
func (p *ServiceClientPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var lastErr error
	for addr, conn := range p.conns {
		if err := conn.Close(); err != nil {
			lastErr = err
		}
		delete(p.conns, addr)
	}
	return lastErr
}

// ServiceClient provides a typed client for a specific service.
type ServiceClient[T any] struct {
	pool       *ServiceClientPool
	name       string
	version    string
	newClient  func(grpc.ClientConnInterface) T
}

// NewServiceClient creates a typed service client.
func NewServiceClient[T any](pool *ServiceClientPool, name, version string, newClient func(grpc.ClientConnInterface) T) *ServiceClient[T] {
	return &ServiceClient[T]{
		pool:      pool,
		name:      name,
		version:   version,
		newClient: newClient,
	}
}

// Get returns a client instance connected to a provider.
// Each call may return a client connected to a different provider (round-robin).
func (sc *ServiceClient[T]) Get(ctx context.Context) (T, error) {
	var zero T
	conn, err := sc.pool.GetConn(ctx, sc.name, sc.version)
	if err != nil {
		return zero, err
	}
	return sc.newClient(conn), nil
}
