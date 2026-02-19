package aegis

import (
	"context"
	"crypto/x509"
	"errors"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

var (
	// ErrNoPeerInfo is returned when no peer info is found in context.
	ErrNoPeerInfo = errors.New("no peer info in context")
	// ErrNoTLSInfo is returned when the peer has no TLS info.
	ErrNoTLSInfo = errors.New("no TLS info in peer")
	// ErrNoCertificate is returned when no client certificate is present.
	ErrNoCertificate = errors.New("no client certificate")
)

// Caller represents the identity of a calling node.
type Caller struct {
	NodeID      string
	Certificate *x509.Certificate
}

// CallerFromContext extracts the caller's identity from the gRPC context.
// The caller's node ID is extracted from the client certificate's Common Name.
func CallerFromContext(ctx context.Context) (*Caller, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, ErrNoPeerInfo
	}

	tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return nil, ErrNoTLSInfo
	}

	if len(tlsInfo.State.PeerCertificates) == 0 {
		return nil, ErrNoCertificate
	}

	cert := tlsInfo.State.PeerCertificates[0]

	return &Caller{
		NodeID:      cert.Subject.CommonName,
		Certificate: cert,
	}, nil
}

// MustCallerFromContext extracts the caller's identity, panicking on error.
// Use only when mTLS is guaranteed (e.g., after middleware validation).
func MustCallerFromContext(ctx context.Context) *Caller {
	caller, err := CallerFromContext(ctx)
	if err != nil {
		panic(err)
	}
	return caller
}
