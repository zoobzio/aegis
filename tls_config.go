package aegis

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CertificateSource defines how certificates are loaded
type CertificateSource string

const (
	// CertSourceFile loads certificates from files
	CertSourceFile CertificateSource = "file"
	// CertSourceEnv loads certificates from environment variables
	CertSourceEnv CertificateSource = "env"
	// CertSourceVault loads certificates from HashiCorp Vault (future)
	CertSourceVault CertificateSource = "vault"
)

// TLSOptions configures how TLS certificates are loaded
type TLSOptions struct {
	// Source determines where certificates come from
	Source CertificateSource
	
	// For file-based certificates
	CertFile   string
	KeyFile    string
	CAFile     string
	
	// For environment-based certificates
	CertEnvVar string
	KeyEnvVar  string
	CAEnvVar   string
	
	// For Vault-based certificates (future)
	VaultPath   string
	VaultRole   string
	
	// Validation options
	VerifyChain      bool
	AllowExpired     bool
	RequiredSANs     []string
}

// DefaultTLSOptions returns secure default options
func DefaultTLSOptions(nodeID string, certDir string) *TLSOptions {
	return &TLSOptions{
		Source:       CertSourceFile,
		CertFile:     filepath.Join(certDir, fmt.Sprintf("%s-cert.pem", nodeID)),
		KeyFile:      filepath.Join(certDir, fmt.Sprintf("%s-key.pem", nodeID)),
		CAFile:       filepath.Join(certDir, "ca-cert.pem"),
		VerifyChain:  true,
		AllowExpired: false,
	}
}

// LoadTLSConfig loads TLS configuration based on options
func LoadTLSConfig(opts *TLSOptions) (*TLSConfig, error) {
	switch opts.Source {
	case CertSourceFile:
		return loadFromFiles(opts)
	case CertSourceEnv:
		return loadFromEnv(opts)
	default:
		return nil, fmt.Errorf("unsupported certificate source: %s", opts.Source)
	}
}

// loadFromFiles loads certificates from filesystem
func loadFromFiles(opts *TLSOptions) (*TLSConfig, error) {
	// Load certificate and key
	cert, err := tls.LoadX509KeyPair(opts.CertFile, opts.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate: %w", err)
	}

	// Load CA certificate
	caCert, err := os.ReadFile(opts.CAFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to parse CA certificate")
	}

	// Parse certificate for validation
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Validate certificate if requested
	if opts.VerifyChain {
		if err := validateCertificate(x509Cert, certPool, opts); err != nil {
			return nil, fmt.Errorf("certificate validation failed: %w", err)
		}
	}

	return &TLSConfig{
		Certificate: cert,
		CertPool:    certPool,
		ServerName:  x509Cert.Subject.CommonName,
	}, nil
}

// loadFromEnv loads certificates from environment variables
func loadFromEnv(opts *TLSOptions) (*TLSConfig, error) {
	// Get certificate content from environment
	certPEM := os.Getenv(opts.CertEnvVar)
	if certPEM == "" {
		return nil, fmt.Errorf("certificate environment variable %s is empty", opts.CertEnvVar)
	}

	keyPEM := os.Getenv(opts.KeyEnvVar)
	if keyPEM == "" {
		return nil, fmt.Errorf("key environment variable %s is empty", opts.KeyEnvVar)
	}

	caPEM := os.Getenv(opts.CAEnvVar)
	if caPEM == "" {
		return nil, fmt.Errorf("CA environment variable %s is empty", opts.CAEnvVar)
	}

	// Parse certificate and key
	cert, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate/key from env: %w", err)
	}

	// Parse CA certificate
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM([]byte(caPEM)) {
		return nil, fmt.Errorf("failed to parse CA certificate from env")
	}

	// Parse certificate for validation
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Validate certificate if requested
	if opts.VerifyChain {
		if err := validateCertificate(x509Cert, certPool, opts); err != nil {
			return nil, fmt.Errorf("certificate validation failed: %w", err)
		}
	}

	return &TLSConfig{
		Certificate: cert,
		CertPool:    certPool,
		ServerName:  x509Cert.Subject.CommonName,
	}, nil
}

// validateCertificate performs certificate validation
func validateCertificate(cert *x509.Certificate, roots *x509.CertPool, opts *TLSOptions) error {
	// Check expiration unless explicitly allowed
	if !opts.AllowExpired {
		verifyOpts := x509.VerifyOptions{
			Roots: roots,
		}
		
		if _, err := cert.Verify(verifyOpts); err != nil {
			return fmt.Errorf("certificate verification failed: %w", err)
		}
	}

	// Check required SANs
	if len(opts.RequiredSANs) > 0 {
		allSANs := make(map[string]bool)
		for _, dns := range cert.DNSNames {
			allSANs[dns] = true
		}
		for _, ip := range cert.IPAddresses {
			allSANs[ip.String()] = true
		}
		
		for _, required := range opts.RequiredSANs {
			if !allSANs[required] {
				return fmt.Errorf("certificate missing required SAN: %s", required)
			}
		}
	}

	return nil
}

// LoadCABundle loads a CA bundle from file or environment
func LoadCABundle(source CertificateSource, fileOrEnvVar string) (*x509.CertPool, error) {
	var caPEM []byte
	var err error

	switch source {
	case CertSourceFile:
		caPEM, err = os.ReadFile(fileOrEnvVar)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA bundle: %w", err)
		}
	case CertSourceEnv:
		caPEM = []byte(os.Getenv(fileOrEnvVar))
		if len(caPEM) == 0 {
			return nil, fmt.Errorf("CA bundle environment variable %s is empty", fileOrEnvVar)
		}
	default:
		return nil, fmt.Errorf("unsupported source: %s", source)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caPEM) {
		return nil, fmt.Errorf("failed to parse CA bundle")
	}

	return certPool, nil
}

// ParseCertificateChain parses a PEM certificate chain
func ParseCertificateChain(pemData []byte) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate
	
	for len(pemData) > 0 {
		var block *pem.Block
		block, pemData = pem.Decode(pemData)
		if block == nil {
			break
		}
		
		if block.Type != "CERTIFICATE" {
			continue
		}
		
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse certificate: %w", err)
		}
		
		certs = append(certs, cert)
	}
	
	if len(certs) == 0 {
		return nil, fmt.Errorf("no certificates found in PEM data")
	}
	
	return certs, nil
}

// GenerateCertificateRequest generates a CSR for external CA signing
func GenerateCertificateRequest(nodeID string, keyFile string) ([]byte, error) {
	// This would generate a CSR that can be sent to an external CA
	// Implementation depends on your CA infrastructure
	return nil, fmt.Errorf("not implemented: use external tools to generate CSR")
}