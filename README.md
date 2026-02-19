# aegis

[![CI Status](https://github.com/zoobzio/aegis/workflows/CI/badge.svg)](https://github.com/zoobzio/aegis/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/zoobzio/aegis/graph/badge.svg?branch=main)](https://codecov.io/gh/zoobzio/aegis)
[![Go Report Card](https://goreportcard.com/badge/github.com/zoobzio/aegis)](https://goreportcard.com/report/github.com/zoobzio/aegis)
[![CodeQL](https://github.com/zoobzio/aegis/workflows/CodeQL/badge.svg)](https://github.com/zoobzio/aegis/security/code-scanning)
[![Go Reference](https://pkg.go.dev/badge/github.com/zoobzio/aegis.svg)](https://pkg.go.dev/github.com/zoobzio/aegis)
[![License](https://img.shields.io/github/license/zoobzio/aegis)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/zoobzio/aegis)](go.mod)
[![Release](https://img.shields.io/github/v/release/zoobzio/aegis)](https://github.com/zoobzio/aegis/releases)

Service mesh for Go microservices — mTLS everywhere, zero configuration. Nodes discover each other, authenticate via certificates, and call domain services without managing PKI infrastructure.

## Zero-Trust by Default

```go
node, _ := aegis.NewNodeBuilder().
    WithID("api-1").
    WithName("API Server").
    WithAddress("localhost:8443").
    WithServices(aegis.ServiceInfo{Name: "identity", Version: "v1"}).
    WithCertDir("./certs").
    Build()

node.StartServer()
// Certificates generated automatically. All connections use mTLS.
// Other nodes can now discover this service and call it securely.
```

## Install

```bash
go get github.com/zoobzio/aegis
```

Requires Go 1.24+.

## Quick Start

A provider node exposes a service. A consumer node discovers and calls it.

```go
package main

import (
    "context"
    "log"

    "github.com/zoobzio/aegis"
    identity "github.com/zoobzio/aegis/proto/identity"
    "google.golang.org/grpc"
)

func main() {
    // Provider: morpheus serves identity
    morpheus, _ := aegis.NewNodeBuilder().
        WithID("morpheus-1").
        WithName("Morpheus").
        WithAddress("localhost:8443").
        WithServices(aegis.ServiceInfo{Name: "identity", Version: "v1"}).
        WithServiceRegistration(func(s *grpc.Server) {
            identity.RegisterIdentityServiceServer(s, &myIdentityServer{})
        }).
        WithCertDir("./certs").
        Build()

    morpheus.StartServer()
    defer morpheus.Shutdown()

    // Consumer: vicky calls identity service
    vicky, _ := aegis.NewNodeBuilder().
        WithID("vicky-1").
        WithName("Vicky").
        WithAddress("localhost:9443").
        WithCertDir("./certs").
        Build()

    pool := aegis.NewServiceClientPool(vicky)
    defer pool.Close()

    client := aegis.NewServiceClient(pool, "identity", "v1", identity.NewIdentityServiceClient)

    // Get a connection (round-robin across providers)
    ctx := context.Background()
    identityClient, _ := client.Get(ctx)
    resp, _ := identityClient.ValidateSession(ctx, &identity.ValidateSessionRequest{
        Token: "user-session-token",
    })

    log.Printf("Session valid: %v, user: %s", resp.Valid, resp.UserId)
}
```

## Capabilities

| Capability | Description | Docs |
|------------|-------------|------|
| Node identity | Build nodes with ID, name, type, address | [node_builder.go](node_builder.go) |
| Automatic mTLS | Certificates generated on first run, loaded thereafter | [tls.go](tls.go) |
| Service registry | Declare services, discover providers across mesh | [service.go](service.go) |
| Topology sync | Nodes share topology; version-based merge | [topology.go](topology.go) |
| Health checks | Extensible health checker interface | [health.go](health.go) |
| Service client | Connection pooling, round-robin load balancing | [client.go](client.go) |
| Caller identity | Extract calling node from mTLS context | [context.go](context.go) |

## Why aegis?

- **Automatic mTLS** — Nodes generate and exchange certificates on startup. No PKI infrastructure to manage.
- **Service discovery built-in** — Declare services, query providers, topology syncs across the mesh.
- **One import** — Node, peer connections, health checks, and gRPC server in a single package.
- **Caller identity on every request** — `CallerFromContext(ctx)` extracts the calling node from mTLS certificates.
- **Round-robin client pooling** — Service clients load-balance across providers automatically.

## The Ecosystem

aegis is the transport layer. Domain services build on top:

| Package | Role |
|---------|------|
| [capitan](https://github.com/zoobzio/capitan) | Event coordination within a process |
| [herald](https://github.com/zoobzio/herald) | Bridge capitan events to message brokers (future: aegis provider) |
| [morpheus](https://github.com/zoobzio/morpheus) | Identity service — implements `IdentityService` |
| [vicky](https://github.com/zoobzio/vicky) | Storage service — consumes identity via mesh |

## Documentation

**Learn**
- [Overview](docs/1.learn/1.overview.md) — What aegis is and why
- [Quickstart](docs/1.learn/2.quickstart.md) — Build your first mesh
- [Concepts](docs/1.learn/3.concepts.md) — Nodes, peers, topology, services
- [Architecture](docs/1.learn/4.architecture.md) — How it works internally

**Guides**
- [Testing](docs/2.guides/1.testing.md) — Testing code that uses aegis
- [Troubleshooting](docs/2.guides/2.troubleshooting.md) — Common errors and solutions
- [Services](docs/2.guides/3.services.md) — Defining and consuming services
- [Certificates](docs/2.guides/4.certificates.md) — Certificate management

**Reference**
- [API](docs/3.reference/1.api.md) — Function signatures
- [Types](docs/3.reference/2.types.md) — Type definitions
- [pkg.go.dev](https://pkg.go.dev/github.com/zoobzio/aegis) — Generated documentation

## Contributing

Contributions welcome — see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License — see [LICENSE](LICENSE).
