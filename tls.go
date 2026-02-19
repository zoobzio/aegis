package aegis

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

// TLSConfig holds the TLS configuration for a node
type TLSConfig struct {
	Certificate tls.Certificate
	CertPool    *x509.CertPool
	ServerName  string
}

// LoadOrGenerateTLS loads existing certificates or generates new ones
func LoadOrGenerateTLS(nodeID string, certDir string) (*TLSConfig, error) {
	// Ensure cert directory exists
	if err := os.MkdirAll(certDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create cert directory: %w", err)
	}

	certFile := filepath.Join(certDir, fmt.Sprintf("%s-cert.pem", nodeID))
	keyFile := filepath.Join(certDir, fmt.Sprintf("%s-key.pem", nodeID))
	caFile := filepath.Join(certDir, "ca-cert.pem")

	// Check if certificates exist
	if _, err := os.Stat(certFile); err == nil {
		return loadTLSConfig(certFile, keyFile, caFile)
	}

	// Generate new certificates
	return generateTLSConfig(nodeID, certDir)
}

// loadTLSConfig loads existing certificates from files
func loadTLSConfig(certFile, keyFile, caFile string) (*TLSConfig, error) {
	// Load certificate and key
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate: %w", err)
	}

	// Load CA certificate
	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to parse CA certificate")
	}

	// Extract server name from certificate
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return &TLSConfig{
		Certificate: cert,
		CertPool:    certPool,
		ServerName:  x509Cert.Subject.CommonName,
	}, nil
}

// generateTLSConfig generates new certificates for the node
func generateTLSConfig(nodeID string, certDir string) (*TLSConfig, error) {
	// Generate CA if it doesn't exist
	caFile := filepath.Join(certDir, "ca-cert.pem")
	caKeyFile := filepath.Join(certDir, "ca-key.pem")
	
	var caCert *x509.Certificate
	var caKey *rsa.PrivateKey
	
	if _, err := os.Stat(caFile); err != nil {
		// Generate CA
		caCert, caKey, err = generateCA(certDir)
		if err != nil {
			return nil, fmt.Errorf("failed to generate CA: %w", err)
		}
	} else {
		// Load existing CA
		caCert, caKey, err = loadCA(caFile, caKeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load CA: %w", err)
		}
	}

	// Generate node certificate
	cert, key, err := generateNodeCertificate(nodeID, caCert, caKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate node certificate: %w", err)
	}

	// Save certificates
	certFile := filepath.Join(certDir, fmt.Sprintf("%s-cert.pem", nodeID))
	keyFile := filepath.Join(certDir, fmt.Sprintf("%s-key.pem", nodeID))

	if err := saveCertificate(certFile, cert); err != nil {
		return nil, fmt.Errorf("failed to save certificate: %w", err)
	}

	if err := savePrivateKey(keyFile, key); err != nil {
		return nil, fmt.Errorf("failed to save private key: %w", err)
	}

	// Create TLS certificate
	tlsCert := tls.Certificate{
		Certificate: [][]byte{cert.Raw},
		PrivateKey:  key,
	}

	// Create cert pool with CA
	certPool := x509.NewCertPool()
	certPool.AddCert(caCert)

	return &TLSConfig{
		Certificate: tlsCert,
		CertPool:    certPool,
		ServerName:  nodeID,
	}, nil
}

// generateCA generates a new Certificate Authority
func generateCA(certDir string) (*x509.Certificate, *rsa.PrivateKey, error) {
	// Generate RSA key
	caKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate CA key: %w", err)
	}

	// Create CA certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"Aegis Mesh Network"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{""},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// Generate certificate
	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create CA certificate: %w", err)
	}

	// Parse certificate
	caCert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	// Save CA certificate and key
	caFile := filepath.Join(certDir, "ca-cert.pem")
	caKeyFile := filepath.Join(certDir, "ca-key.pem")

	if err := saveCertificate(caFile, caCert); err != nil {
		return nil, nil, fmt.Errorf("failed to save CA certificate: %w", err)
	}

	if err := savePrivateKey(caKeyFile, caKey); err != nil {
		return nil, nil, fmt.Errorf("failed to save CA key: %w", err)
	}

	return caCert, caKey, nil
}

// loadCA loads existing CA certificate and key
func loadCA(certFile, keyFile string) (*x509.Certificate, *rsa.PrivateKey, error) {
	// Load certificate
	certPEM, err := os.ReadFile(certFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, nil, fmt.Errorf("failed to decode CA certificate")
	}

	caCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	// Load key
	keyPEM, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read CA key: %w", err)
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return nil, nil, fmt.Errorf("failed to decode CA key")
	}

	caKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse CA key: %w", err)
	}

	return caCert, caKey, nil
}

// generateNodeCertificate generates a certificate for a node
func generateNodeCertificate(nodeID string, caCert *x509.Certificate, caKey *rsa.PrivateKey) (*x509.Certificate, *rsa.PrivateKey, error) {
	// Generate RSA key
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate key: %w", err)
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			CommonName:   nodeID,
			Organization: []string{"Aegis Node"},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(90 * 24 * time.Hour),
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		DNSNames:     []string{nodeID, "localhost"},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
	}

	// Generate certificate signed by CA
	certBytes, err := x509.CreateCertificate(rand.Reader, &template, caCert, &key.PublicKey, caKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Parse certificate
	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, key, nil
}

// saveCertificate saves a certificate to a PEM file
func saveCertificate(filename string, cert *x509.Certificate) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create certificate file: %w", err)
	}
	defer func() { _ = file.Close() }()

	err = pem.Encode(file, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})
	if err != nil {
		return fmt.Errorf("failed to encode certificate: %w", err)
	}

	return nil
}

// savePrivateKey saves a private key to a PEM file
func savePrivateKey(filename string, key *rsa.PrivateKey) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer func() { _ = file.Close() }()

	err = pem.Encode(file, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	if err != nil {
		return fmt.Errorf("failed to encode private key: %w", err)
	}

	return nil
}

// GetServerTLSConfig returns TLS configuration for the server
func (tc *TLSConfig) GetServerTLSConfig() *tls.Config {
	return &tls.Config{
		Certificates: []tls.Certificate{tc.Certificate},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    tc.CertPool,
		MinVersion:   tls.VersionTLS12,
	}
}

// GetClientTLSConfig returns TLS configuration for the client
func (tc *TLSConfig) GetClientTLSConfig(serverName string) *tls.Config {
	return &tls.Config{
		Certificates: []tls.Certificate{tc.Certificate},
		RootCAs:      tc.CertPool,
		ServerName:   serverName,
		MinVersion:   tls.VersionTLS12,
	}
}