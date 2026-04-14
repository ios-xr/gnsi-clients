# Configuration Guide

## Overview

This guide provides best practices and recommendations for configuring the gNSI Certz client.

## Connection Configuration

### Production Deployments

**Always use proper TLS verification:**
```bash
# DO: Specify target_name and CA certificate
./certz_client \
  -target_addr router.example.com:57400 \
  -target_name router.example.com \
  -ca_cert /path/to/ca-bundle.pem \
  -client_cert client.pem \
  -client_key client_key.pem

# DON'T: Skip verification in production
./certz_client -insecure_skip_verify  # TESTING ONLY!
```

### Test/Lab Environments

For testing with self-signed certificates:
```bash
./certz_client \
  -target_addr 192.168.1.1:57400 \
  -target_name router.lab.local \
  -insecure_skip_verify \
  -username admin \
  -password admin123
```

## Authentication Methods

The gNSI Certz service supports multiple authentication methods that can be combined based on your server configuration.

### Understanding Authentication Requirements

**Username/Password (AAA) Authentication:**
- **Required by default** for AAA authentication to the device
- Can be **omitted only if** metadata authentication is disabled on the server
- When metadata auth is disabled, username is extracted from the client certificate's Common Name (CN)

**Client Certificate (mTLS) Authentication:**
- **Required if** the server is configured to require mutual TLS (mTLS)
- ⚠️ **Connection will fail** if `-client_cert` and `-client_key` are not provided when the server requires mTLS
- The server determines whether mTLS is required - check your server's TLS configuration
- If the server requires client certificates and they are not provided, you'll see TLS handshake errors

### Method 1: mTLS + Username/Password (Most Common)

**Best for:** Production environments with AAA enabled

```bash
./certz_client \
  -target_addr router.example.com:57400 \
  -target_name router.example.com \
  -client_cert /secure/certs/client.pem \
  -client_key /secure/keys/client_key.pem \
  -ca_cert /secure/ca/ca-bundle.pem \
  -username admin \
  -password "${ROUTER_PASSWORD}"
```

**Security:**
- ✅ Strong cryptographic authentication (mTLS)
- ✅ AAA authentication and authorization
- ✅ Supports certificate chains
- ⚠️ Use environment variables for passwords

### Method 2: mTLS Only (Metadata Auth Disabled)

**Best for:** Environments where AAA metadata authentication is disabled

```bash
./certz_client \
  -target_addr router.example.com:57400 \
  -target_name router.example.com \
  -client_cert /secure/certs/client.pem \
  -client_key /secure/keys/client_key.pem \
  -ca_cert /secure/ca/ca-bundle.pem
```

**Requirements:**
- ✅ Server must have metadata authentication disabled
- ✅ Username extracted from certificate CN
- ✅ No passwords in command line
- ✅ Certificate must have valid CN with username

### Method 3: TLS with Server Verification Only (Username/Password + CA)

**Best for:** Servers that don't require mTLS but use proper TLS with server certificate validation

```bash
./certz_client \
  -target_addr router.example.com:57400 \
  -target_name router.example.com \
  -ca_cert /secure/ca/ca-bundle.pem \
  -username admin \
  -password "${ROUTER_PASSWORD}"
```

**Security:**
- ✅ Server certificate is properly validated using CA cert
- ✅ Encrypted TLS connection
- ✅ AAA authentication and authorization
- ⚠️ No client certificate authentication (one-way TLS)
- ℹ️ Server must NOT require mTLS (client certificates)

### Method 4: Username/Password Only (Insecure Transport)

**Best for:** Lab/test environments only - NOT recommended for production

```bash
./certz_client \
  -target_addr 192.168.1.1:57400 \
  -target_name router.lab.local \
  -insecure_skip_verify \
  -username admin \
  -password "${ROUTER_PASSWORD}"
```

**Security:**
- ⚠️ No client certificate authentication
- ⚠️ Requires insecure_skip_verify in most cases
- ⚠️ Testing/lab use only

## File Organization

### Recommended Directory Structure

```
/etc/certz/
├── certs/
│   ├── ca-bundle.pem        # CA certificates
│   ├── client-cert.pem      # Client certificate (or chain: leaf + intermediates)
│   └── client-key.pem       # Client private key (chmod 600) - for leaf cert
├── entities/
│   ├── device-certs/        # Device certificates to rotate
│   ├── ca-bundles/          # CA bundles to upload
│   └── crl-bundles/         # CRL bundles
└── configs/
    └── devices.conf         # Device connection info
```

### Certificate Chain Support

Both `-client_cert` and `-cert_file` (for rotation) support certificate chains in PEM format:

```pem
-----BEGIN CERTIFICATE-----
[Leaf Certificate - device/client cert]
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
[Intermediate CA Certificate]
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
[Root CA Certificate - optional]
-----END CERTIFICATE-----
```

**Important:** The private key file (`-client_key` or `-key_file`) must contain ONLY the private key for the leaf certificate.

### File Permissions

```bash
# CA and client certificates - readable
chmod 644 /etc/certz/certs/ca-bundle.pem
chmod 644 /etc/certz/certs/client-cert.pem

# Private keys - strictly protected
chmod 600 /etc/certz/certs/client-key.pem
chown certz-user:certz-group /etc/certz/certs/client-key.pem
```

## Timeout Configuration

### Default Timeout (30s)
Suitable for most operations

### Extended Timeout for Slow Links
```bash
-timeout 2m  # For CSR generation over slow connections
```

### Quick Operations
```bash
-timeout 10s  # For simple queries like get-profile-list
```

## CSR Suite Selection

### Security vs Compatibility

**High Security (Recommended):**
```bash
-csr_suite ecdsa-p384-sha384  # ECDSA P-384, strong and fast
```

**Balanced:**
```bash
-csr_suite ecdsa-p256-sha256  # ECDSA P-256, widely compatible
```

**Legacy Compatibility:**
```bash
-csr_suite rsa-2048-sha256    # RSA 2048, legacy systems
```

**Maximum Security:**
```bash
-csr_suite ecdsa-p521-sha512  # ECDSA P-521, maximum strength
```

## Validation Best Practices

### Always Validate Before Finalize (Production)

```bash
# Test the rotated certificates before committing
-validate

# What this does:
# 1. Uploads new certificates
# 2. Creates test connection with new certs
# 3. Verifies connectivity (calls CanGenerateCSR)
# 4. If success → finalizes rotation
# 5. If failure → cancels rotation (automatic rollback)
```

## Version Management

### Semantic Versioning for Certificates

```bash
# Initial deployment
-version "1.0"

# Minor update (renewed cert, same key)
-version "1.1"

# Major update (new key pair)
-version "2.0"
```

### Force Overwrite

```bash
# Overwrite existing version (default)
-force_overwrite=true

# Prevent accidental overwrite
-force_overwrite=false
```

## Environment Variables

### Secure Password Management

```bash
# .env file (never commit to git)
export ROUTER_ADDRESS="192.168.1.1:57400"
export ROUTER_USER="admin"
export ROUTER_PASS="secret123"
export CA_CERT="/etc/certz/certs/ca-bundle.pem"

# Usage
source .env
./certz_client \
  -target_addr "$ROUTER_ADDRESS" \
  -username "$ROUTER_USER" \
  -password "$ROUTER_PASS" \
  -ca_cert "$CA_CERT"
```

### Automation Scripts

```bash
#!/bin/bash
# Load configuration
source /etc/certz/configs/production.env

# Validate environment
required_vars=("ROUTER_ADDRESS" "CLIENT_CERT" "CLIENT_KEY")
for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        echo "Error: $var not set"
        exit 1
    fi
done

# Execute operation
./certz_client \
  -target_addr "$ROUTER_ADDRESS" \
  -client_cert "$CLIENT_CERT" \
  -client_key "$CLIENT_KEY" \
  "$@"
```

## Profile Management

### Naming Conventions

```bash
# Use descriptive, consistent names
-profile_id "grpc-server"        # gRPC server profile
-profile_id "netconf-server"     # NETCONF server profile  
-profile_id "bgp-peers"          # BGP peer authentication
-profile_id "syslog-client"      # Syslog client certificates
```

### Profile Lifecycle

```bash
# 1. Create profile
./certz_client -op add-profile -profile_id new-service

# 2. Rotate initial certificates
./certz_client -op rotate -profile_id new-service \
  -cert_file init-cert.pem -key_file init-key.pem

# 3. Update periodically (before expiration)
./certz_client -op rotate -profile_id new-service \
  -cert_file renewed-cert.pem -key_file renewed-key.pem \
  -validate

# 4. Clean up (when service is decommissioned)
./certz_client -op delete-profile -profile_id new-service
```

## Troubleshooting

### Enable Verbose Output

```bash
-v  # Shows connection details, operation progress
```

### Common Issues

**Certificate Not Found:**
```bash
# Verify file paths are absolute or relative to current directory
ls -la /path/to/cert.pem
file /path/to/cert.pem  # Should show "PEM certificate"
```

**Permission Denied:**
```bash
# Check file permissions
ls -l /path/to/client-key.pem
# Should be readable by current user
```

**Connection Timeout:**
```bash
# Increase timeout
-timeout 2m

# Check network connectivity
nc -zv router.example.com 57400
```

## References

- [EXAMPLES.md](EXAMPLES.md) - Comprehensive usage examples
- [TROUBLESHOOTING.md](TROUBLESHOOTING.md) - Problem resolution
- [ARCHITECTURE.md](ARCHITECTURE.md) - Project structure
- [README.md](../README.md) - Command-line reference
- [gNSI Certz Specification](https://github.com/openconfig/gnsi/tree/main/certz) - Official protocol specification
