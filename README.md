# gNSI Certz Client

A complete Go client implementation for the [gNSI (gRPC Network Security Interface) Certificate Management Service (Certz)](https://github.com/openconfig/gnsi/tree/main/certz). Designed for Cisco IOS XR and other gNSI-compliant network devices.

> **Specification**: This client implements the [gNSI Certz Protocol](https://github.com/openconfig/gnsi/tree/main/certz)

## 📚 Documentation

- **[EXAMPLES.md](EXAMPLES.md)** - Comprehensive examples, complete workflows, and quick start guide
- **[docs/CONFIGURATION.md](docs/CONFIGURATION.md)** - Configuration best practices and security guidelines
- **[docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md)** - Common issues and debugging techniques
- **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)** - Technical architecture and design details
- **[BUILD.md](BUILD.md)** - Build instructions


## Features

### ✅ Flexible Authentication

- **Client Certificate (mTLS)**: Optional - provide `-client_cert` and `-client_key` for mutual TLS authentication
- **Username/Password (AAA)**: Required for AAA authentication, unless metadata authentication is disabled on the server
- **Username Extraction**: When metadata authentication is disabled, username is automatically extracted from the client certificate's Common Name (CN)

### ✅ Complete Certz RPC Support

- `AddProfile` - Create new SSL profiles
- `DeleteProfile` - Remove SSL profiles
- `GetProfileList` - List all SSL profiles
- `CanGenerateCSR` - Check CSR generation capability
- `Rotate` - Certificate/CA/CRL rotation with multiple modes:
  - Client-provided certificates
  - Device-generated CSR
  - Device-native certificates (IDevID, oIDevID, self-signed)
  - Copy entities between profiles

### ✅ IOS XR Compatibility

- Tested with Cisco IOS XR gRPC server
- Supports username+password authentication
- Supports certificate-based authentication
- Configurable timeouts and connection parameters

## Installation

### Prerequisites

- Go 1.19 or later
- Access to a gNSI-enabled device (e.g., Cisco IOS XR)

### Build

```bash
cd cmd/certz_client
go build -o certz_client
```

## Command-Line Options Reference

### Connection Options

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-target_addr` | string | localhost:9339 | Target device address in format host:port (e.g., 192.168.1.1:57400) |
| `-target_name` | string | - | Target server hostname for TLS certificate verification |
| `-timeout` | duration | 30s | Operation timeout (e.g., 30s, 1m, 90s) |
| `-insecure_skip_verify` | bool | false | Skip TLS certificate verification (⚠️ testing only) |

### Certificate Files

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-ca_cert` | string | - | CA certificate file path for verifying server certificate (optional) |
| `-client_cert` | string | - | Client certificate file path (optional, enables mTLS). Supports certificate chains (leaf + intermediates) |
| `-client_key` | string | - | Client private key file path (optional, required with `-client_cert`). Must match the leaf certificate |

**Certificate Chain Format:**
Both `-client_cert` and `-cert_file` (in rotation) support certificate chains. The PEM file should contain certificates in order:
1. Leaf certificate (first)
2. Intermediate CA certificates (if any)
3. Root CA (optional)

Example PEM format:
```pem
-----BEGIN CERTIFICATE-----
[Leaf Certificate]
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
[Intermediate CA]
-----END CERTIFICATE-----
```

### Username/Password (AAA Authentication)

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-username` | string | - | Username for AAA authentication. **Required** unless metadata authentication is disabled on the server. When metadata auth is disabled, username is extracted from client certificate CN |
| `-password` | string | - | Password for AAA authentication. **Required** unless metadata authentication is disabled on the server |

### Operation Options

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-op` | string | get-profile-list | Operation to perform: `add-profile`, `delete-profile`, `get-profile-list`, `can-generate-csr`, `rotate` |

### Profile Management Options

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-profile_id` | string | - | SSL profile ID to operate on (required for most operations) |
| `-from_profile_id` | string | - | Source profile ID when copying entities between profiles |

### Certificate Rotation Options

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-force_overwrite` | bool | true | Force overwrite if version already exists |
| `-version` | string | 1.0 | Version identifier for the rotated entity |
| `-created_on` | uint64 | current timestamp | Entity creation timestamp (Unix epoch seconds) |
| `-cert_file` | string | - | Certificate file path (PEM format) for rotation. Supports certificate chains |
| `-key_file` | string | - | Private key file path (PEM format) for rotation. Must match the leaf certificate |
| `-cert_source` | string | - | Use device-native certificate: `IDevID`, `oIDevID`, or `self-signed` |

### CA Bundle Options

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-ca_bundle_file` | string | - | CA certificate bundle file path (PEM format, can contain multiple certificates) |

### CRL Bundle Options

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-crl_bundle_file` | string | - | Certificate Revocation List bundle file path (PEM format) |

### CSR Generation Options

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-generate_csr` | bool | false | Request device to generate Certificate Signing Request |
| `-csr_suite` | string | rsa-2048-sha256 | CSR suite (key type and signature algorithm). See [CSR Suite Reference](#csr-suite-reference) for available options |
| `-common_name` | string | - | Common Name (CN) for certificate (required for CSR generation) |
| `-country` | string | - | Country (C) for certificate subject |
| `-state` | string | - | State or Province (ST) for certificate subject |
| `-city` | string | - | City or Locality (L) for certificate subject |
| `-organization` | string | - | Organization (O) for certificate subject |
| `-organizational_unit` | string | - | Organizational Unit (OU) for certificate subject |
| `-ip_address` | string | - | IP address for certificate |
| `-email_id` | string | - | Email address for certificate |

### Subject Alternative Name (SAN) Options

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-san_dns` | string | - | DNS names for SAN extension (comma-separated, e.g., "host1.com,host2.com") |
| `-san_emails` | string | - | Email addresses for SAN extension (comma-separated) |
| `-san_ips` | string | - | IP addresses for SAN extension (comma-separated) |
| `-san_uris` | string | - | URIs for SAN extension (comma-separated) |

### CSR Signing Options (for device-generated CSR)

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-ca_sign_cert` | string | - | CA certificate file for signing device-generated CSR |
| `-ca_sign_key` | string | - | CA private key file for signing device-generated CSR |

**Note**: These are required when using `-generate_csr` option. The client will sign the device-generated CSR and upload the signed certificate.

### Entity Copy Options

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-copy_entity_from` | string | - | Source profile ID to copy entity from |
| `-entity_type` | string | - | Type of entity to copy: `cert`, `ca_bundle`, or `crl_bundle` |

### Misc Options

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-v` | bool | false | Enable verbose output for debugging |

## Quick Start

For complete examples and workflows, see **[EXAMPLES.md](EXAMPLES.md)** which includes:
- Complete onboarding workflow (add profile → rotate cert/CA/CRL → verify)
- Validation before finalize pattern
- IOS XR specific examples
- Advanced use cases

### Basic Examples

**List SSL Profiles:**
```bash
./certz_client -target_addr 192.168.1.1:57400 \
  -target_name router.example.com \
  -username admin -password admin123 \
  -insecure_skip_verify \
  -op get-profile-list
```

**Add New Profile:**
```bash
./certz_client -target_addr 192.168.1.1:57400 \
  -target_name router.example.com \
  -username admin -password admin123 \
  -insecure_skip_verify \
  -op add-profile -profile_id my_service
```

**Rotate Certificate:**
```bash
./certz_client -target_addr 192.168.1.1:57400 \
  -target_name router.example.com \
  -username admin -password admin123 \
  -insecure_skip_verify \
  -op rotate -profile_id my_service \
  -cert_file cert.pem -key_file key.pem \
  -ca_bundle_file ca.pem -version "1.0"
```

See [EXAMPLES.md](EXAMPLES.md) for more detailed examples including:
- Device-generated CSR with SANs
- Validation before finalize
- Copying entities between profiles  
- Using device-native certificates (IDevID)
- Batch operations and scripts

## Usage Examples

**For comprehensive examples and complete workflows, see [EXAMPLES.md](EXAMPLES.md)**

The EXAMPLES.md file includes:
- Complete onboarding workflow (get profiles → add profile → rotate cert/CA/CRL → verify)
- Validation before finalize pattern with rollback capability
- IOS XR specific examples
- Device-generated CSR with SANs
- Copying entities between profiles
- Using device-native certificates (IDevID, oIDevID, self-signed)
- Batch operations and automation scripts
- Troubleshooting guide

### IOS XR Examples

#### Connect to IOS XR with Username/Password

```bash
./certz_client \
  -target_addr 1.2.27.3:57400 \
  -target_name ems.cisco.com \
  -username cisco \
  -password cisco123 \
  -insecure_skip_verify \
  -op get-profile-list \
  -v
```

#### Rotate Using IDevID Certificate

```bash
./certz_client \
  -target_addr 1.2.27.3:57400 \
  -target_name ems.cisco.com \
  -username cisco \
  -password cisco123 \
  -insecure_skip_verify \
  -op rotate \
  -profile_id test \
  -cert_source IDevID \
  -ca_bundle_file ems.pem \
  -version "2.0" \
  -v
```

#### Add Profile and Rotate Certificate

```bash
# First, add the profile
./certz_client \
  -target_addr 1.2.27.3:57400 \
  -target_name ems.cisco.com \
  -username cisco \
  -password cisco123 \
  -insecure_skip_verify \
  -op add-profile \
  -profile_id production_service

# Then, rotate certificate
./certz_client \
  -target_addr 1.2.27.3:57400 \
  -target_name ems.cisco.com \
  -username cisco \
  -password cisco123 \
  -insecure_skip_verify \
  -op rotate \
  -profile_id production_service \
  -cert_file device.pem \
  -key_file device_key.pem \
  -ca_bundle_file ca_chain.pem \
  -version "1.0"
```

**For more examples** including device CSR with SANs, copying entities, and advanced use cases, see [EXAMPLES.md](EXAMPLES.md).

## CSR Suite Reference

CSR suites specify the key type and signature algorithm combination. Use the string names (not integer values) with the `-csr_suite` flag.

| Suite Name | Key Type | Signature Algorithm |
|------------|----------|---------------------|
| `rsa-2048-sha256` | RSA 2048 | SHA-256 |
| `rsa-2048-sha384` | RSA 2048 | SHA-384 |
| `rsa-2048-sha512` | RSA 2048 | SHA-512 |
| `rsa-3072-sha256` | RSA 3072 | SHA-256 |
| `rsa-3072-sha384` | RSA 3072 | SHA-384 |
| `rsa-3072-sha512` | RSA 3072 | SHA-512 |
| `rsa-4096-sha256` | RSA 4096 | SHA-256 |
| `rsa-4096-sha384` | RSA 4096 | SHA-384 |
| `rsa-4096-sha512` | RSA 4096 | SHA-512 |
| `ecdsa-p256-sha256` | ECDSA P-256 | SHA-256 |
| `ecdsa-p256-sha384` | ECDSA P-256 | SHA-384 |
| `ecdsa-p256-sha512` | ECDSA P-256 | SHA-512 |
| `ecdsa-p384-sha256` | ECDSA P-384 | SHA-256 |
| `ecdsa-p384-sha384` | ECDSA P-384 | SHA-384 |
| `ecdsa-p384-sha512` | ECDSA P-384 | SHA-512 |
| `ecdsa-p521-sha256` | ECDSA P-521 | SHA-256 |
| `ecdsa-p521-sha384` | ECDSA P-521 | SHA-384 |
| `ecdsa-p521-sha512` | ECDSA P-521 | SHA-512 |
| `eddsa-ed25519` | EdDSA | Ed25519 |

**Example Usage:**
```bash
-csr_suite ecdsa-p256-sha256
```

## Project Structure

```
certz-client/
├── cmd/
│   └── certz_client/     # Main client application
│       └── main.go
├── pkg/
│   ├── connection/       # Connection management (TLS/mTLS/insecure)
│   │   └── connection.go
│   └── operations/       # Certz RPC implementations
│       ├── certz.go      # Main operations
│       └── csr.go        # CSR signing functions
├── docs/                 # Documentation
├── examples/             # Example scripts
├── go.mod                # Go module definition
└── README.md             # This file
```

## Troubleshooting

### Connection Issues

**Error: "connection refused"**
- Check if gRPC server is running on the device
- Verify the port number (commonly 57400 for IOS XR)
- Check firewall rules

**Error: "certificate verification failed"**
- Use `-insecure_skip_verify` for testing
- Ensure CA certificate is correct
- Verify `-target_name` matches certificate CN or SAN

### Authentication Issues

**Error: "authentication failed"**
- Verify username and password
- Check if client certificate is valid
- Ensure clock is synchronized (certificate validity)

### Operation Errors

**Error: "profile already exists"**
- Use a different profile ID
- Delete existing profile first

**Error: "profile not found"**
- Create the profile with `-op add-profile` first
- Verify the profile ID spelling

## Best Practices

1. **Use Certificate-Based Authentication** - Use client certificates in production environments
2. **Rotate Before Expiry** - Monitor certificate expiration dates
3. **Test in Staging** - Validate rotations in non-production first
4. **Use Strong Crypto** - Prefer ECDSA or RSA 3072+ with SHA-256+
5. **Version Appropriately** - Use meaningful version strings
6. **Backup Configs** - Save current configuration before changes

## References

- [gNSI Certz Specification](https://github.com/openconfig/gnsi/tree/main/certz)
- [OpenConfig gNSI](https://github.com/openconfig/gnsi)
- [Cisco IOS XR gRPC Documentation](https://www.cisco.com/c/en/us/td/docs/routers/asr9000/software/asr9k-r7-0/programmability/configuration/guide/b-programmability-cg-asr9000-70x.html)

## License

Copyright © 2024 Cisco Systems, Inc.

Licensed under the Apache License, Version 2.0. See LICENSE file for details.

## Support

For issues or questions, contact our organization's support channels or open an issue on GitHub.
