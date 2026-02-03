# Aegis Certificate Management

Aegis requires mTLS (mutual TLS) for all node-to-node communication. This document explains how to manage certificates in production environments.

## Certificate Requirements

- **No CA Generation**: Nodes do NOT generate their own CA
- **External CA Required**: Use an existing PKI infrastructure
- **mTLS Mandatory**: All connections require valid client certificates

## Certificate Sources

### 1. File-Based Certificates (Traditional)

Store certificates as files on disk:

```json
{
  "tls": {
    "source": "file",
    "cert_file": "/etc/aegis/certs/node-cert.pem",
    "key_file": "/etc/aegis/certs/node-key.pem",
    "ca_file": "/etc/aegis/certs/ca-cert.pem"
  }
}
```

**Security**: Ensure proper file permissions (0600) on key files.

### 2. Environment Variables (Container-Friendly)

Load certificates from environment variables:

```json
{
  "tls": {
    "source": "env",
    "cert_env": "AEGIS_TLS_CERT",
    "key_env": "AEGIS_TLS_KEY",
    "ca_env": "AEGIS_TLS_CA"
  }
}
```

```bash
export AEGIS_TLS_CERT="-----BEGIN CERTIFICATE-----
MIIDXTCCAkWgAwIBAgIJAKl...
-----END CERTIFICATE-----"

export AEGIS_TLS_KEY="-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAvLYcyu8f3...
-----END RSA PRIVATE KEY-----"

export AEGIS_TLS_CA="-----BEGIN CERTIFICATE-----
MIIDXTCCAkWgAwIBAgIJAKl1...
-----END CERTIFICATE-----"
```

### 3. External CA Integration (Future)

Support for HashiCorp Vault, AWS Certificate Manager, etc.

## Certificate Generation Workflow

### Option 1: Pre-Generated Certificates

1. Generate certificates using your CA infrastructure
2. Distribute certificates to nodes
3. Configure nodes to load certificates

```bash
# Example using OpenSSL
openssl req -new -key node.key -out node.csr \
  -subj "/CN=node-1/O=Aegis Network"
  
# Submit CSR to your CA and get signed certificate
```

### Option 2: Certificate Signing Request (CSR)

1. Node generates private key locally
2. Create CSR and submit to CA
3. Load signed certificate when received

### Option 3: Automated Certificate Management

Use tools like cert-manager, Vault PKI, or ACME for automated certificate lifecycle.

## Configuration Examples

### Minimal Configuration

```go
// Load from default locations
node, err := aegis.NewNodeBuilder().
    WithID("node-1").
    WithName("My Node").
    WithType(aegis.NodeTypeGateway).
    WithAddress("localhost:8001").
    Build()
```

### Custom Certificate Validation

```go
tlsOpts := &aegis.TLSOptions{
    Source:       aegis.CertSourceFile,
    CertFile:     "/secure/certs/node-cert.pem",
    KeyFile:      "/secure/certs/node-key.pem",
    CAFile:       "/secure/certs/ca-bundle.pem",
    VerifyChain:  true,
    RequiredSANs: []string{"*.aegis.local"},
}

node, err := aegis.NewNodeBuilder().
    WithTLSOptions(tlsOpts).
    Build()
```

### Environment-Based (Kubernetes/Docker)

```yaml
# Kubernetes Secret
apiVersion: v1
kind: Secret
metadata:
  name: aegis-tls
type: Opaque
data:
  cert: LS0tLS1CRUdJTi...
  key: LS0tLS1CRUdJTi...
  ca: LS0tLS1CRUdJTi...
```

```yaml
# Pod Spec
env:
  - name: AEGIS_TLS_CERT
    valueFrom:
      secretKeyRef:
        name: aegis-tls
        key: cert
  - name: AEGIS_TLS_KEY
    valueFrom:
      secretKeyRef:
        name: aegis-tls
        key: key
  - name: AEGIS_TLS_CA
    valueFrom:
      secretKeyRef:
        name: aegis-tls
        key: ca
```

## Security Best Practices

1. **Never Generate CA on Nodes**: Use external CA infrastructure
2. **Rotate Certificates**: Implement regular certificate rotation
3. **Validate SANs**: Use `RequiredSANs` to ensure certificates match expected hostnames
4. **Monitor Expiration**: Set up alerts for certificate expiration
5. **Secure Key Storage**: Use hardware security modules (HSM) or key management services
6. **Audit Certificate Usage**: Log all certificate validation failures

## Certificate Validation

Aegis performs the following validations:

1. **Chain of Trust**: Verifies certificate is signed by trusted CA
2. **Expiration**: Checks certificate validity period
3. **Subject Alternative Names**: Validates SANs match expected values
4. **Client Certificate**: Requires valid client cert for all connections

## Troubleshooting

### Common Issues

1. **"TLS configuration is required"**: Node started without certificates
2. **"Certificate verification failed"**: Certificate not signed by trusted CA
3. **"Certificate missing required SAN"**: Certificate doesn't have expected hostname

### Debug TLS Issues

```bash
# Test certificate chain
openssl verify -CAfile ca.pem node-cert.pem

# Check certificate details
openssl x509 -in node-cert.pem -text -noout

# Test mTLS connection
openssl s_client -connect localhost:8001 \
  -cert client-cert.pem \
  -key client-key.pem \
  -CAfile ca.pem
```

## Migration from Insecure Connections

Since Aegis requires mTLS, there's no migration path from insecure connections. All nodes must have valid certificates before joining the mesh.