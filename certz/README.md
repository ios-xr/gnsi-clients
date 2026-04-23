# gNSI Certz Client

A complete Go client implementation for the [gNSI (gRPC Network Security Interface) Certificate Management Service (Certz)](https://github.com/openconfig/gnsi/tree/main/certz). Designed for Cisco IOS XR and other gNSI-compliant network devices.

> **Specification**: This client implements the [gNSI Certz Protocol](https://github.com/openconfig/gnsi/tree/main/certz)

## 📚 Documentation

- **[docs/EXAMPLES.md](docs/EXAMPLES.md)** - Comprehensive examples, complete workflows, and quick start guide
- **[docs/CONFIGURATION.md](docs/CONFIGURATION.md)** - Configuration best practices and security guidelines
- **[docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md)** - Common issues and debugging techniques
- **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)** - Technical architecture and design details
- **[BUILD.md](BUILD.md)** - Build instructions


## Features

### ✅ Flexible Authentication

- **Client Certificate (mTLS)** and **Username/Password (AAA)** authentication
- Multiple authentication modes supported (see [docs/CONFIGURATION.md](docs/CONFIGURATION.md) for details)

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
| `-ca_cert` | string | - | CA certificate file path for verifying server certificate |
| `-client_cert` | string | - | Client certificate file path (enables mTLS). Supports certificate chains |
| `-client_key` | string | - | Client private key file path (required with `-client_cert`) |

> **Note:** Certificate files support chains with leaf + intermediate certificates. See [docs/CONFIGURATION.md](docs/CONFIGURATION.md) for details.

### Username/Password (AAA Authentication)

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-username` | string | - | Username for AAA authentication (see [docs/CONFIGURATION.md](docs/CONFIGURATION.md) for when required) |
| `-password` | string | - | Password for AAA authentication |

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

## Usage Examples

**See [docs/EXAMPLES.md](docs/EXAMPLES.md) for:**
- Quick start examples
- Complete onboarding workflows
- Certificate rotation scenarios

## Additional Documentation

For detailed information, see:

- **Project Architecture & Design** - See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)
- **Configuration Best Practices** - See [docs/CONFIGURATION.md](docs/CONFIGURATION.md)
- **Troubleshooting Guide** - See [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md)

## Best Practices

See [docs/CONFIGURATION.md](docs/CONFIGURATION.md) for detailed best practices and security recommendations.

## References

- [gNSI Certz Specification](https://github.com/openconfig/gnsi/tree/main/certz)
- [OpenConfig gNSI](https://github.com/openconfig/gnsi)
- [Cisco IOS XR gRPC Documentation](https://www.cisco.com/c/en/us/td/docs/routers/asr9000/software/asr9k-r7-0/programmability/configuration/guide/b-programmability-cg-asr9000-70x.html)

## License

Copyright © 2024 Cisco Systems, Inc.

Licensed under the Apache License, Version 2.0. See LICENSE file for details.

## Support

For issues or questions, contact your organization's support channels or open an issue on GitHub.
