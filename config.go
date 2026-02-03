package aegis

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// MeshConfig holds the configuration for a mesh node
type MeshConfig struct {
	// Node configuration
	Node NodeConfig `json:"node"`
	
	// TLS configuration
	TLS TLSConfigOptions `json:"tls"`
	
	// Network configuration
	Network NetworkConfig `json:"network"`
	
	// Health check configuration
	Health HealthConfig `json:"health"`
}

// NodeConfig contains node-specific settings
type NodeConfig struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Type    NodeType `json:"type"`
	Address string   `json:"address"`
}

// TLSConfigOptions contains TLS/mTLS settings
type TLSConfigOptions struct {
	// Source of certificates (file, env, vault)
	Source CertificateSource `json:"source"`
	
	// File-based certificate paths
	CertFile string `json:"cert_file,omitempty"`
	KeyFile  string `json:"key_file,omitempty"`
	CAFile   string `json:"ca_file,omitempty"`
	
	// Environment variable names
	CertEnv string `json:"cert_env,omitempty"`
	KeyEnv  string `json:"key_env,omitempty"`
	CAEnv   string `json:"ca_env,omitempty"`
	
	// Certificate validation
	VerifyChain      bool     `json:"verify_chain"`
	AllowExpired     bool     `json:"allow_expired"`
	RequiredSANs     []string `json:"required_sans,omitempty"`
	
	// mTLS settings
	RequireClientCert bool `json:"require_client_cert"`
	VerifyClientCert  bool `json:"verify_client_cert"`
}

// NetworkConfig contains network-related settings
type NetworkConfig struct {
	// Connection timeouts
	DialTimeout      time.Duration `json:"dial_timeout"`
	RequestTimeout   time.Duration `json:"request_timeout"`
	
	// Retry configuration
	MaxRetries       int           `json:"max_retries"`
	RetryBackoff     time.Duration `json:"retry_backoff"`
	
	// Connection pool settings
	MaxConnections   int           `json:"max_connections"`
	IdleTimeout      time.Duration `json:"idle_timeout"`
}

// HealthConfig contains health check settings
type HealthConfig struct {
	// Health check intervals
	CheckInterval    time.Duration `json:"check_interval"`
	Timeout         time.Duration `json:"timeout"`
	
	// Failure thresholds
	UnhealthyThreshold int `json:"unhealthy_threshold"`
	HealthyThreshold   int `json:"healthy_threshold"`
}

// DefaultMeshConfig returns a configuration with secure defaults
func DefaultMeshConfig() *MeshConfig {
	return &MeshConfig{
		Node: NodeConfig{
			Type: NodeTypeGeneric,
		},
		TLS: TLSConfigOptions{
			Source:            CertSourceFile,
			VerifyChain:       true,
			AllowExpired:      false,
			RequireClientCert: true,
			VerifyClientCert:  true,
		},
		Network: NetworkConfig{
			DialTimeout:    10 * time.Second,
			RequestTimeout: 30 * time.Second,
			MaxRetries:     3,
			RetryBackoff:   time.Second,
			MaxConnections: 100,
			IdleTimeout:    5 * time.Minute,
		},
		Health: HealthConfig{
			CheckInterval:      30 * time.Second,
			Timeout:           5 * time.Second,
			UnhealthyThreshold: 3,
			HealthyThreshold:   2,
		},
	}
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(filename string) (*MeshConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	config := DefaultMeshConfig()
	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	
	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	return config, nil
}

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv(prefix string) (*MeshConfig, error) {
	config := DefaultMeshConfig()
	
	// Load node configuration
	if id := os.Getenv(prefix + "_NODE_ID"); id != "" {
		config.Node.ID = id
	}
	if name := os.Getenv(prefix + "_NODE_NAME"); name != "" {
		config.Node.Name = name
	}
	if addr := os.Getenv(prefix + "_NODE_ADDRESS"); addr != "" {
		config.Node.Address = addr
	}
	if nodeType := os.Getenv(prefix + "_NODE_TYPE"); nodeType != "" {
		config.Node.Type = NodeType(nodeType)
	}
	
	// Load TLS configuration
	if source := os.Getenv(prefix + "_TLS_SOURCE"); source != "" {
		config.TLS.Source = CertificateSource(source)
	}
	
	// Environment-based certificates
	if config.TLS.Source == CertSourceEnv {
		config.TLS.CertEnv = prefix + "_TLS_CERT"
		config.TLS.KeyEnv = prefix + "_TLS_KEY"
		config.TLS.CAEnv = prefix + "_TLS_CA"
	}
	
	return config, nil
}

// Validate checks if the configuration is valid
func (c *MeshConfig) Validate() error {
	// Validate node configuration
	if c.Node.ID == "" {
		return fmt.Errorf("node ID is required")
	}
	if c.Node.Name == "" {
		return fmt.Errorf("node name is required")
	}
	if c.Node.Address == "" {
		return fmt.Errorf("node address is required")
	}
	
	// Validate TLS configuration
	switch c.TLS.Source {
	case CertSourceFile:
		if c.TLS.CertFile == "" || c.TLS.KeyFile == "" || c.TLS.CAFile == "" {
			return fmt.Errorf("certificate files are required for file source")
		}
	case CertSourceEnv:
		if c.TLS.CertEnv == "" || c.TLS.KeyEnv == "" || c.TLS.CAEnv == "" {
			return fmt.Errorf("environment variables are required for env source")
		}
	default:
		return fmt.Errorf("unsupported TLS source: %s", c.TLS.Source)
	}
	
	// Validate network configuration
	if c.Network.DialTimeout <= 0 {
		return fmt.Errorf("dial timeout must be positive")
	}
	if c.Network.MaxConnections <= 0 {
		return fmt.Errorf("max connections must be positive")
	}
	
	return nil
}

// ToTLSOptions converts config to TLSOptions
func (c *MeshConfig) ToTLSOptions() *TLSOptions {
	return &TLSOptions{
		Source:       c.TLS.Source,
		CertFile:     c.TLS.CertFile,
		KeyFile:      c.TLS.KeyFile,
		CAFile:       c.TLS.CAFile,
		CertEnvVar:   c.TLS.CertEnv,
		KeyEnvVar:    c.TLS.KeyEnv,
		CAEnvVar:     c.TLS.CAEnv,
		VerifyChain:  c.TLS.VerifyChain,
		AllowExpired: c.TLS.AllowExpired,
		RequiredSANs: c.TLS.RequiredSANs,
	}
}

// SaveConfig writes configuration to a JSON file
func (c *MeshConfig) SaveConfig(filename string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	if err := os.WriteFile(filename, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}