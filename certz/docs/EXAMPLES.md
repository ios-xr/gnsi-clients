# gNSI Certz Client - Examples and Quick Start Guide

Complete examples showing how to use the certz client for certificate management on network devices.

## Table of Contents

- [Quick Start Examples](#quick-start-examples)
- [Complete Onboarding Workflow](#complete-onboarding-workflow)
- [Certificate Rotation Examples](#certificate-rotation-examples)
- [Validation and Finalize Pattern](#validation-and-finalize-pattern)
- [IOS XR Specific Examples](#ios-xr-specific-examples)
- [Advanced Use Cases](#advanced-use-cases)

## Quick Start Examples

> **Authentication**: For authentication options and requirements, see [CONFIGURATION.md](CONFIGURATION.md).

### 1. List SSL Profiles (Production - mTLS + Username/Password)

This is the **recommended production configuration** with both certificate authentication and AAA credentials:

```bash
./certz_client \
  -target_addr 192.168.1.1:57400 \
  -target_name router.example.com \
  -client_cert client.pem \
  -client_key client_key.pem \
  -ca_cert ca.pem \
  -username admin \
  -password admin123 \
  -op get-profile-list
```

### 2. List SSL Profiles (mTLS Only)

```bash
./certz_client \
  -target_addr 192.168.1.1:57400 \
  -target_name router.example.com \
  -client_cert client.pem \
  -client_key client_key.pem \
  -ca_cert ca.pem \
  -op get-profile-list
```

### 3. List SSL Profiles (Lab/Test - Username/Password Only)

```bash
./certz_client \
  -target_addr 192.168.1.1:57400 \
  -target_name router.example.com \
  -username admin \
  -password admin123 \
  -insecure_skip_verify \
  -op get-profile-list
```

### 4. Add New Profile

```bash
./certz_client \
  -target_addr 192.168.1.1:57400 \
  -target_name router.example.com \
  -client_cert client.pem \
  -client_key client_key.pem \
  -ca_cert ca.pem \
  -username admin \
  -password admin123 \
  -op add-profile \
  -profile_id my_service
```

### 5. Rotate Certificate (Client-Provided)

```bash
./certz_client \
  -target_addr 192.168.1.1:57400 \
  -target_name router.example.com \
  -client_cert client.pem \
  -client_key client_key.pem \
  -ca_cert ca.pem \
  -username admin \
  -password admin123 \
  -op rotate \
  -profile_id my_service \
  -cert_file new_device_cert.pem \
  -key_file new_device_key.pem \
  -version "2.0"
```

### 6. Rotate with Device-Generated CSR

```bash
./certz_client \
  -target_addr 192.168.1.1:57400 \
  -target_name router.example.com \
  -client_cert client.pem \
  -client_key client_key.pem \
  -ca_cert ca.pem \
  -username admin \
  -password admin123 \
  -op rotate \
  -profile_id my_service \
  -generate_csr \
  -common_name device.example.com \
  -country US \
  -state CA \
  -organization "Cisco Systems" \
  -csr_suite ecdsa-p256-sha256 \
  -ca_sign_cert signing_ca.pem \
  -ca_sign_key signing_ca_key.pem
```

## Complete Onboarding Workflow

This section demonstrates a complete workflow for onboarding a new SSL profile with certificates.

### Step 1: Check Current Profiles

First, list all existing SSL profiles to see what's already configured:

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

**Expected Output:**
```
Profile List:
- system_default_profile
- existing_service_profile
```

### Step 2: Add New SSL Profile

Create a new SSL profile for your service:

```bash
./certz_client \
  -target_addr 1.2.27.3:57400 \
  -target_name ems.cisco.com \
  -username cisco \
  -password cisco123 \
  -insecure_skip_verify \
  -op add-profile \
  -profile_id production_app \
  -v
```

**Expected Output:**
```
✓ Profile 'production_app' added successfully
```

### Step 3: Rotate Certificate and Key

Upload certificate and private key to the new profile:

```bash
./certz_client \
  -target_addr 1.2.27.3:57400 \
  -target_name ems.cisco.com \
  -username cisco \
  -password cisco123 \
  -insecure_skip_verify \
  -op rotate \
  -profile_id production_app \
  -cert_file production_cert.pem \
  -key_file production_key.pem \
  -version "1.0" \
  -v
```

### Step 4: Rotate CA Bundle

Add trusted CA certificates to the profile:

```bash
./certz_client \
  -target_addr 1.2.27.3:57400 \
  -target_name ems.cisco.com \
  -username cisco \
  -password cisco123 \
  -insecure_skip_verify \
  -op rotate \
  -profile_id production_app \
  -ca_bundle_file ca_chain.pem \
  -version "1.0" \
  -v
```

### Step 5: Rotate CRL Bundle

Add certificate revocation list:

```bash
./certz_client \
  -target_addr 1.2.27.3:57400 \
  -target_name ems.cisco.com \
  -username cisco \
  -password cisco123 \
  -insecure_skip_verify \
  -op rotate \
  -profile_id production_app \
  -crl_bundle_file revoked_certs.crl \
  -version "1.0" \
  -v
```

### Step 6: Verify the Profile

List profiles again to confirm everything is configured:

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

**Expected Output:**
```
Profile List:
- system_default_profile
- existing_service_profile
- production_app ✓
```

## Certificate Rotation Examples

### Rotate All Entities Together

Rotate certificate, CA bundle, and CRL in a single operation:

```bash
./certz_client \
  -target_addr 192.168.1.1:57400 \
  -target_name router.example.com \
  -client_cert client.pem \
  -client_key client_key.pem \
  -ca_cert ca.pem \
  -username admin \
  -password admin123 \
  -op rotate \
  -profile_id my_service \
  -cert_file new_device_cert.pem \
  -key_file new_device_key.pem \
  -ca_bundle_file new_ca_chain.pem \
  -crl_bundle_file new_crl.pem \
  -version "2.0" \
  -force_overwrite \
  -v
```

### Rotate Using Device CSR with SANs

Request device to generate CSR with Subject Alternative Names:

```bash
./certz_client \
  -target_addr 192.168.1.1:57400 \
  -target_name router.example.com \
  -client_cert client.pem \
  -client_key client_key.pem \
  -ca_cert ca.pem \
  -username admin \
  -password admin123 \
  -op rotate \
  -profile_id web_service \
  -generate_csr \
  -common_name web.example.com \
  -san_dns "web.example.com,backup.example.com,admin.example.com" \
  -san_ips "192.168.1.1,10.0.0.1" \
  -san_emails "admin@example.com" \
  -country US \
  -state CA \
  -city "San Jose" \
  -organization "Cisco Systems" \
  -organizational_unit "Network Security" \
  -csr_suite ecdsa-p384-sha384 \
  -ca_sign_cert ca.pem \
  -ca_sign_key ca_key.pem \
  -version "3.0" \
  -v
```

### Copy Entities Between Profiles

Reuse CA bundle from an existing profile:

```bash
./certz_client \
  -target_addr 192.168.1.1:57400 \
  -target_name router.example.com \
  -client_cert client.pem \
  -client_key client_key.pem \
  -ca_cert ca.pem \
  -username admin \
  -password admin123 \
  -op rotate \
  -profile_id new_service \
  -copy_entity_from existing_service \
  -entity_type ca_bundle \
  -version "1.0" \
  -v
```

### Use Device-Native Certificates

Rotate using device's built-in IDevID certificate:

```bash
./certz_client \
  -target_addr 1.2.27.3:57400 \
  -target_name ems.cisco.com \
  -username cisco \
  -password cisco123 \
  -insecure_skip_verify \
  -op rotate \
  -profile_id production_app \
  -cert_source IDevID \
  -ca_bundle_file ca.pem \
  -version "1.0" \
  -v
```

**Other cert_source options:**
- `IDevID` - Initial Device Identifier (manufacturer certificate)
- `oIDevID` - Operational IDevID
- `self-signed` - Device-generated self-signed certificate

## Validation and Finalize Pattern

**Important**: The Rotate operation supports validation before committing changes. Here's how it works:

### Rotation Workflow with Validation

1. **Upload Phase**: Client uploads certificate/CA/CRL to the device
2. **Validation Phase** (Optional): Client tests the new certificates by establishing a new connection
3. **Finalize Phase**: Client sends finalize message to commit, or cancels to rollback

### Example: Rotate with Validation

```bash
# The client internally handles the streaming protocol:
# 1. Send Upload Request (cert, CA, CRL)
# 2. Receive Upload Response
# 3. [OPTIONAL] Validate by attempting new connection with rotated credentials
# 4. If validation successful → Send FinalizeRequest
# 5. If validation fails → Cancel stream (automatic rollback)

./certz_client \
  -target_addr 192.168.1.1:57400 \
  -target_name router.example.com \
  -client_cert client.pem \
  -client_key client_key.pem \
  -ca_cert ca.pem \
  -op rotate \
  -profile_id my_service \
  -cert_file new_cert.pem \
  -key_file new_key.pem \
  -ca_bundle_file new_ca.pem \
  -validate \
  -version "2.0" \
  -v
```

**With `-validate` flag:**
- After upload, client attempts CanGenerateCSR to verify connectivity
- If successful → automatically sends FinalizeRequest
- If fails → cancels stream, device rolls back to previous certificates

**Without `-validate` flag:**
- Immediately sends FinalizeRequest after upload
- Faster, but no pre-commit validation

## IOS XR Specific Examples

### Connect to IOS XR with Username/Password

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

### Rotate Using IDevID on IOS XR

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

### Complete IOS XR Onboarding

```bash
# 1. List existing profiles
./certz_client -target_addr 1.2.27.3:57400 -target_name ems.cisco.com \
  -username cisco -password cisco123 -insecure_skip_verify \
  -op get-profile-list

# 2. Add new profile
./certz_client -target_addr 1.2.27.3:57400 -target_name ems.cisco.com \
  -username cisco -password cisco123 -insecure_skip_verify \
  -op add-profile -profile_id grpc_service

# 3. Rotate certificate using IDevID
./certz_client -target_addr 1.2.27.3:57400 -target_name ems.cisco.com \
  -username cisco -password cisco123 -insecure_skip_verify \
  -op rotate -profile_id grpc_service \
  -cert_source IDevID -ca_bundle_file ca.pem -version "1.0"

# 4. Verify
./certz_client -target_addr 1.2.27.3:57400 -target_name ems.cisco.com \
  -username cisco -password cisco123 -insecure_skip_verify \
  -op get-profile-list
```

## Advanced Use Cases

### Check CSR Generation Capability

Before requesting a device-generated CSR, check if the device supports it:

```bash
./certz_client \
  -target_addr 192.168.1.1:57400 \
  -target_name router.example.com \
  -client_cert client.pem \
  -client_key client_key.pem \
  -ca_cert ca.pem \
  -op can-generate-csr \
  -common_name test.example.com \
  -csr_suite ecdsa-p256-sha256 \
  -country US \
  -state CA \
  -organization "Cisco Systems" \
  -v
```

**Response:**
```
CanGenerateCSRResponse: {
  can_generate: true
}
```

### Rotate Only CA Bundle

Update only the CA trust bundle without changing certificates:

```bash
./certz_client \
  -target_addr 192.168.1.1:57400 \
  -target_name router.example.com \
  -client_cert client.pem \
  -client_key client_key.pem \
  -ca_cert ca.pem \
  -op rotate \
  -profile_id my_service \
  -ca_bundle_file updated_ca_chain.pem \
  -version "1.1" \
  -v
```

### Rotate Only CRL Bundle

Update only the certificate revocation list:

```bash
./certz_client \
  -target_addr 192.168.1.1:57400 \
  -target_name router.example.com \
  -client_cert client.pem \
  -client_key client_key.pem \
  -ca_cert ca.pem \
  -op rotate \
  -profile_id my_service \
  -crl_bundle_file updated_crl.pem \
  -version "1.2" \
  -v
```

### Delete Profile

Remove an SSL profile when no longer needed:

```bash
./certz_client \
  -target_addr 192.168.1.1:57400 \
  -target_name router.example.com \
  -client_cert client.pem \
  -client_key client_key.pem \
  -ca_cert ca.pem \
  -op delete-profile \
  -profile_id old_service \
  -v
```

### Batch Operations Script

Example shell script for rotating certificates across multiple profiles:

```bash
#!/bin/bash
# rotate_all_services.sh

TARGET="192.168.1.1:57400"
TARGET_NAME="router.example.com"

PROFILES=("service1" "service2" "service3")

for PROFILE in "${PROFILES[@]}"; do
  echo "Rotating $PROFILE..."
  ./certz_client \
    -target_addr "$TARGET" \
    -target_name "$TARGET_NAME" \
    -client_cert client.pem \
    -client_key client_key.pem \
    -ca_cert ca.pem \
    -op rotate \
    -profile_id "$PROFILE" \
    -cert_file "${PROFILE}_cert.pem" \
    -key_file "${PROFILE}_key.pem" \
    -ca_bundle_file ca_bundle.pem \
    -version "$(date +%Y.%m.%d)" \
    -validate \
    -v
  
  if [ $? -eq 0 ]; then
    echo "✓ $PROFILE rotated successfully"
  else
    echo "✗ $PROFILE rotation failed"
    exit 1
  fi
done

echo "All profiles rotated successfully!"
```

## Best Practices

1. **Always Validate**: Use `-validate` flag for production rotations
2. **Version Management**: Use meaningful versions (see [CONFIGURATION.md](CONFIGURATION.md))
3. **Test First**: Validate in non-production environments
4. **Monitor Expiration**: Rotate before certificates expire
5. **Backup**: Keep copies of current certificates
6. **Strong Crypto**: See [CONFIGURATION.md](CONFIGURATION.md) for recommendations

## Troubleshooting

For troubleshooting common issues, see [TROUBLESHOOTING.md](TROUBLESHOOTING.md).

## References

- [README.md](../README.md) - Command-line flag reference
- [CONFIGURATION.md](CONFIGURATION.md) - Configuration best practices and authentication
- [TROUBLESHOOTING.md](TROUBLESHOOTING.md) - Debugging and problem resolution
- [ARCHITECTURE.md](ARCHITECTURE.md) - Project structure
- [gNSI Certz Specification](https://github.com/openconfig/gnsi/tree/main/certz) - Official protocol specification
