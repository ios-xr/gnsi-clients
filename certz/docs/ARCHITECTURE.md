# gNSI Certz Client - Architecture & Usage Guide

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Authentication Modes Explained](#authentication-modes-explained)
- [IOS XR Integration](#ios-xr-integration)
- [Advanced Usage](#advanced-usage)
- [Error Handling](#error-handling)
- [Security Best Practices](#security-best-practices)

## Architecture Overview

### Component Structure

```
certz-client/
│
├── cmd/certz_client/main.go
│   ├── CLI interface
│   ├── Flag parsing
│   └── Operation routing
│
├── pkg/connection/
│   ├── TLS configuration
│   ├── mTLS configuration
│   ├── Username/password credentials
│   └── gRPC connection management
│
└── pkg/operations/
    ├── AddProfile
    ├── DeleteProfile
    ├── GetProfileList
    ├── CanGenerateCSR
    ├── Rotate (with sub-operations)
    └── CSR signing utilities
```

### Connection Flow

```
Client → connection.New(config) → gRPC Connection
                                        ↓
                                  TLS/mTLS/Plaintext
                                        ↓
                                  + Username/Password (optional)
                                        ↓
                                  certzpb.CertzClient
                                        ↓
                                  RPC Operations
```

### Rotate Operation Flow

```
Client → Rotate Request
           ↓
    [Optional: Generate CSR on device]
           ↓
    [Upload Certificate/CA/CRL]
           ↓
    [Optional: Copy from existing profile]
           ↓
    Finalize Rotation
           ↓
    Success (or automatic rollback on error)
```

## Authentication Modes

Authentication mode is automatically detected based on certificate flags.

### Mode 1: TLS with Username/Password

**Use Case**: IOS XR devices with username/password authentication

**Auto-detected when**: No client certificate provided

**Requirements**:
- `-target_name` (server hostname)
- `-username` and `-password` (for AAA authentication)
- `-insecure_skip_verify` (optional, for self-signed server certs)
- `-ca_cert` (optional, for server verification)

**How it works**:
1. TLS connection established (server-side TLS only)
2. Client sends username/password for AAA authentication with each RPC call
3. Server authenticates username/password
4. RPC is processed

**Example (IOS XR)**:
```bash
-target_name ems.cisco.com \
-username cisco \
-password cisco123 \
-insecure_skip_verify
```

### Mode 2: Mutual TLS (mTLS)

**Use Case**: High-security environments, certificate-based authentication

**Auto-detected when**: Both `-client_cert` and `-client_key` are provided

**Requirements**:
- `-target_name`
- `-client_cert` and `-client_key`
- `-ca_cert` (optional, for server verification)

**How it works**:
1. Client presents certificate during TLS handshake
2. Server verifies client certificate
3. Client verifies server certificate (if CA provided)
4. Both parties authenticated at TLS layer
5. Encrypted channel established
6. If username/password provided → used for AAA auth; otherwise username extracted from certificate CN

**Example**:
```bash
-target_name router.example.com \
-client_cert client.pem \
-client_key client_key.pem \
-ca_cert ca.pem
```

### Mode 3: mTLS + Username/Password

**Use Case**: Maximum security with dual authentication

**Auto-detected when**: Client cert/key AND username/password are both provided

**Requirements**:
- Both mTLS credentials (client cert/key) and username/password

**How it works**:
1. mTLS authentication at TLS layer (from client certificate)
2. Username/password authentication at application layer
3. Both must succeed for RPC to be processed

**Example**:
```bash
-target_name router.example.com \
-client_cert client.pem \
-client_key client_key.pem \
-ca_cert ca.pem \
-username admin \
-password admin123
```

### Testing Mode: Insecure Skip Verify

**Use Case**: Testing with self-signed certificates

**Requirements**:
- `-insecure_skip_verify` flag

**Warning**: Only use in testing! Vulnerable to man-in-the-middle attacks.

**Example**:
```bash
-insecure_skip_verify \
-username admin \
-password admin123
```

## IOS XR Integration

### Cisco IOS XR gRPC Server

IOS XR devices run a gRPC server (default port: 57400) that supports:
- TLS with username/password authentication
- Optional mTLS
- gNSI Certz service

### Typical IOS XR Configuration

```
grpc
 port 57400
 tls
 service-layer
 address-family ipv4
 !
!
```

### Authentication on IOS XR

IOS XR typically uses **username/password authentication** over TLS:

```bash
./certz_client \
  -target_addr <ios-xr-ip>:57400 \
  -target_name <hostname> \
  -tls \
  -username <user> \
  -password <password> \
  -insecure_skip_verify \
  -op <operation>
```

### IOS XR Certificate Sources

IOS XR devices have built-in certificates:

1. **IDevID** (Initial Device Identifier)
   - Manufacturer-installed certificate
   - Cannot be changed
   - Use with `-cert_source IDevID`

2. **oIDevID** (Operational IDevID)
   - Operator-provisioned certificate
   - Can be updated
   - Use with `-cert_source oIDevID`

3. **Self-signed**
   - Device-generated certificate
   - Use with `-cert_source self-signed`

### Example: IOS XR Workflow

```bash
# 1. List existing profiles
./certz_client \
  -target_addr 1.2.27.3:57400 \
  -target_name ems.cisco.com \
  -username cisco -password cisco123 \
  -insecure_skip_verify \
  -op get-profile-list

# 2. Create new profile
./certz_client \
  -target_addr 1.2.27.3:57400 \
  -target_name ems.cisco.com \
  -username cisco -password cisco123 \
  -insecure_skip_verify \
  -op add-profile \
  -profile_id my_app

# 3. Rotate using IDevID
./certz_client \
  -target_addr 1.2.27.3:57400 \
  -target_name ems.cisco.com \
  -username cisco -password cisco123 \
  -insecure_skip_verify \
  -op rotate \
  -profile_id my_app \
  -cert_source IDevID \
  -ca_bundle_file ca.pem \
  -version "1.0"
```

## Advanced Usage

### Combining Multiple Entity Updates

Update certificate, CA bundle, and CRL in one atomic operation:

```bash
./certz_client \
  -op rotate \
  -profile_id my_service \
  -cert_file device.pem \
  -key_file device_key.pem \
  -ca_bundle_file ca_chain.pem \
  -crl_bundle_file revoked.pem \
  -version "2.0"
```

### Device-Generated CSR with SANs

Generate CSR on device with Subject Alternative Names:

```bash
./certz_client \
  -op rotate \
  -profile_id web_service \
  -generate_csr \
  -common_name web.example.com \
  -san_dns "web.example.com,backup.example.com,admin.example.com" \
  -san_ips "192.168.1.1,10.0.0.1" \
  -san_emails "admin@example.com" \
  -san_uris "https://api.example.com" \
  -country US \
  -state CA \
  -organization "My Company" \
  -csr_suite ecdsa-p256-sha256 \
  -ca_sign_cert ca.pem \
  -ca_sign_key ca_key.pem
```

### Copying Entities Between Profiles

Reuse CA bundle from another profile:

```bash
./certz_client \
  -op rotate \
  -profile_id new_service \
  -copy_entity_from trusted_service \
  -entity_type ca_bundle \
  -version "1.0"
```

### Updating Only CA Bundle

```bash
./certz_client \
  -op rotate \
  -profile_id my_service \
  -ca_bundle_file new_ca_chain.pem \
  -version "1.1"
```

### Updating Only CRL

```bash
./certz_client \
  -op rotate \
  -profile_id my_service \
  -crl_bundle_file updated_crl.pem \
  -version "1.1"
```

## Error Handling

### Common Errors and Solutions

#### Connection Errors

**Error**: `connection refused`
```
Solution:
- Verify target_addr is correct
- Check if gRPC server is running
- Verify firewall rules
```

**Error**: `certificate verification failed`
```
Solution:
- Use -insecure_skip_verify for testing
- Verify CA certificate is correct
- Check that target_name matches certificate
```

#### Authentication Errors

**Error**: `authentication failed`
```
Solution:
- Verify username and password
- Check client certificate (for mTLS)
- Ensure certificates are not expired
```

#### Operation Errors

**Error**: `profile already exists`
```
Solution:
- Use different profile ID
- Or delete existing profile first
```

**Error**: `profile not found`
```
Solution:
- Create profile with add-profile first
- Verify profile_id spelling
```

**Error**: `version already exists`
```
Solution:
- Use different version string
- Or use -force_overwrite=true
```

### Verbose Mode

Enable verbose output for debugging:

```bash
-v
```

This shows:
- Connection details
- Request/response information
- Operation progress

### Timeout Adjustment

For slow connections or large operations:

```bash
-timeout 60s
```

## Security Best Practices

### 1. Production Environment

✅ **Do**:
- Use mTLS with valid certificates
- Use strong key sizes (RSA 3072+, ECDSA P-256+)
- Use strong signature algorithms (SHA-256+)
- Rotate certificates before expiry
- Store private keys securely
- Use meaningful version strings
- Test in staging first

❌ **Don't**:
- Use insecure_skip_verify in production
- Use plaintext mode
- Share or commit private keys
- Use weak crypto (RSA 1024, MD5, SHA-1)

### 2. Certificate Management

- Monitor certificate expiration dates
- Maintain certificate inventory
- Document which profiles use which certificates
- Keep backup of current configuration
- Have rollback procedure ready

### 3. Access Control

- Use separate credentials for different environments
- Implement least-privilege access
- Rotate credentials regularly
- Audit certificate operations

### 4. Network Security

- Use VPN or secure network for management
- Limit gRPC server access by IP
- Monitor for unauthorized access attempts
- Enable logging on devices

## Troubleshooting Tips

### Debug Connection Issues

1. Test basic connectivity:
   ```bash
   telnet <target> 57400
   ```

2. Verify TLS:
   ```bash
   openssl s_client -connect <target>:57400
   ```

3. Try insecure mode first:
   ```bash
   -insecure_skip_verify
   ```

### Verify Certificates

Check certificate validity:
```bash
openssl x509 -in cert.pem -text -noout
```

Check certificate and key match:
```bash
openssl x509 -noout -modulus -in cert.pem | openssl md5
openssl rsa -noout -modulus -in key.pem | openssl md5
```

### Test CSR Generation

Check if specific CSR suite is supported:
```bash
./certz_client \
  -op can-generate-csr \
  -common_name test.example.com \
  -csr_suite ecdsa-p256-sha256
```

If false, try different suite (see README for all available options).

## Performance Considerations

### Timeout Values

- **Fast networks**: 10-30s
- **Slow networks**: 60s+
- **Large CRLs**: 60s+

### Batch Operations

For multiple profiles:
- Use shell scripts
- Add error handling
- Include delays between operations

### Large Files

For large CA bundles or CRLs:
- Increase timeout
- Monitor device memory
- Consider splitting if possible

## References

- [gNSI Certz Specification](https://github.com/openconfig/gnsi/tree/main/certz)
- [gRPC Authentication Guide](https://grpc.io/docs/guides/auth/)
- [Go crypto/tls Package](https://pkg.go.dev/crypto/tls)
- [X.509 Certificate Standard](https://www.ietf.org/rfc/rfc5280.txt)

## Support

For issues or questions:
1. Check this documentation
2. Review example scripts in `examples/`
3. Enable verbose mode (`-v`)
4. Contact your organization's support
