# Troubleshooting Guide

This guide helps diagnose and resolve common issues with the gNSI Certz client.

## Table of Contents

- [Connection Issues](#connection-issues)
- [Authentication Failures](#authentication-failures)
- [Certificate Problems](#certificate-problems)
- [Rotation Failures](#rotation-failures)
- [Validation Issues](#validation-issues)
- [Performance Problems](#performance-problems)
- [Debugging Techniques](#debugging-techniques)

## Connection Issues

### Error: "connection refused"

**Symptoms:**
```
Error: failed to connect: failed to dial 192.168.1.1:57400: connection refused
```

**Possible Causes & Solutions:**

1. **gRPC server not running on device**
   ```bash
   # On IOS XR, verify gRPC is configured:
   show run grpc
   
   # Should show:
   grpc
    port 57400
    no-tls
   !
   ```

2. **Wrong port number**
   ```bash
   # Verify correct port
   netstat -an | grep 57400  # On device
   nc -zv 192.168.1.1 57400  # From client
   ```

3. **Firewall blocking connection**
   ```bash
   # Test connectivity
   telnet 192.168.1.1 57400
   # Or
   nmap -p 57400 192.168.1.1
   ```

**Fix:**
```bash
# Use correct address and port
-target_addr 192.168.1.1:57400
```

### Error: "context deadline exceeded"

**Symptoms:**
```
Error: context deadline exceeded
```

**Causes:**
- Network latency
- Slow device response
- Firewall dropping packets

**Solutions:**

1. **Increase timeout:**
   ```bash
   -timeout 2m  # Extended timeout for slow connections
   ```

2. **Check network path:**
   ```bash
   ping 192.168.1.1
   traceroute 192.168.1.1
   ```

3. **Verify device load:**
   ```bash
   # On device, check CPU/memory
   show processes cpu
   show memory summary
   ```

### Error: "no such host"

**Symptoms:**
```
Error: failed to dial router.example.com:57400: no such host
```

**Solutions:**

1. **Use IP address instead:**
   ```bash
   -target_addr 192.168.1.1:57400
   ```

2. **Verify DNS resolution:**
   ```bash
   nslookup router.example.com
   dig router.example.com
   ```

3. **Check /etc/hosts:**
   ```bash
   grep router /etc/hosts
   ```

## Authentication Failures

### Error: "authentication failed"

**Symptoms:**
```
Error: rpc error: code = Unauthenticated desc = authentication failed
```

**For Username/Password Authentication:**

1. **Verify credentials:**
   ```bash
   # Test on device directly
   ssh admin@192.168.1.1
   ```

2. **Check for special characters:**
   ```bash
   # Quote password if it contains special characters
   -password 'P@ssw0rd!'
   ```

3. **Verify AAA configuration (IOS XR):**
   ```
   show run aaa
   ```

**For Certificate Authentication:**

1. **Verify certificate is valid:**
   ```bash
   openssl x509 -in client-cert.pem -text -noout
   # Check:
   # - Not expired (Validity dates)
   # - Correct CN (Subject)
   # - Issuer matches server's trusted CA
   ```

2. **Verify private key matches certificate:**
   ```bash
   # Extract public key from cert
   openssl x509 -in client-cert.pem -pubkey -noout > pub1.pem
   
   # Extract public key from private key
   openssl pkey -in client-key.pem -pubout > pub2.pem
   
   # Compare
   diff pub1.pem pub2.pem  # Should be identical
   ```

### Error: "certificate verification failed"

**Symptoms:**
```
Error: x509: certificate signed by unknown authority
```

**Solutions:**

1. **Provide CA certificate:**
   ```bash
   -ca_cert /path/to/ca-bundle.pem
   ```

2. **Use insecure mode for testing:**
   ```bash
   -insecure_skip_verify  # TESTING ONLY!
   ```

3. **Verify CA bundle contains server's CA:**
   ```bash
   # View certificates in bundle
   openssl crl2pkcs7 -nocrl -certfile ca-bundle.pem | \
     openssl pkcs7 -print_certs -text -noout
   ```

### Error: "certificate has expired" or "certificate is not yet valid"

**Symptoms:**
```
Error: x509: certificate has expired or is not yet valid
```

**Causes:**
- Server certificate expired
- Client certificate expired
- System clock not synchronized

**Solutions:**

1. **Check certificate validity:**
   ```bash
   # Check server certificate
   openssl s_client -connect router.example.com:57400 -showcerts
   
   # Check client certificate
   openssl x509 -in client-cert.pem -noout -dates
   ```

2. **Verify system time:**
   ```bash
   date
   # Sync if needed
   ntpdate -u pool.ntp.org
   ```

3. **Renew expired certificates**

### Error: "certificate name does not match"

**Symptoms:**
```
Error: x509: certificate is valid for ems.cisco.com, not router.example.com
```

**Cause:**
The `-target_name` doesn't match the server certificate's CN or SAN

**Solutions:**

1. **Use correct target name:**
   ```bash
   # Must match certificate CN or SAN
   -target_name ems.cisco.com
   ```

2. **Check server certificate details:**
   ```bash
   openssl s_client -connect 192.168.1.1:57400 -showcerts | \
     openssl x509 -noout -text | grep -A2 "Subject:\|DNS:"
   ```

3. **Use insecure mode for testing (not production):**
   ```bash
   -insecure_skip_verify
   ```

### Error: "bad certificate" (mTLS)

**Symptoms:**
```
Error: remote error: tls: bad certificate
Error: tls: bad certificate
```

**Cause:**
Server rejected the client certificate during mTLS handshake

**Common Reasons:**

1. **Client certificate not trusted by server:**
   - Server doesn't have the CA that signed your client cert
   - Certificate chain incomplete

2. **Client certificate expired or not yet valid**

3. **Client private key doesn't match certificate**

4. **Certificate revoked (in server's CRL)**

**Solutions:**

1. **Verify client certificate is trusted:**
   ```bash
   # Check if server has your CA certificate configured
   # On IOS XR: show crypto ca certificates
   ```

2. **Check certificate validity:**
   ```bash
   openssl x509 -in client-cert.pem -noout -dates -subject -issuer
   ```

3. **Verify key matches certificate:**
   ```bash
   # Compare modulus
   openssl x509 -noout -modulus -in client-cert.pem | openssl md5
   openssl rsa -noout -modulus -in client-key.pem | openssl md5
   # Should output the same hash
   ```

4. **Ensure certificate chain is complete:**
   ```bash
   # If using intermediate CAs, include them in client-cert.pem
   # Order: Leaf cert first, then intermediates
   ```

5. **Try without client cert to isolate issue:**
   ```bash
   # Remove -client_cert and -client_key flags
   # Use only -username and -password
   # If this works, issue is with client certificate
   ```

### Error: "tls: handshake failure"

**Symptoms:**
```
Error: remote error: tls: handshake failure
```

**Causes:**
- TLS version mismatch
- Cipher suite mismatch
- Certificate issues

**Solutions:**

1. **Check TLS configuration on server**

2. **Verify certificate is valid:**
   ```bash
   openssl x509 -in cert.pem -text -noout
   ```

3. **Test TLS connection:**
   ```bash
   openssl s_client -connect 192.168.1.1:57400 -tls1_2
   ```

## Certificate Problems

### Error: "failed to load client certificate"

**Symptoms:**
```
Error: failed to load mTLS config: failed to load client certificate: ...
```

**Solutions:**

1. **Verify file paths:**
   ```bash
   ls -la /path/to/client-cert.pem
   ls -la /path/to/client-key.pem
   ```

2. **Check file permissions:**
   ```bash
   # Private key should be readable
   chmod 600 /path/to/client-key.pem
   ```

3. **Verify PEM format:**
   ```bash
   # Certificate should have:
   grep "BEGIN CERTIFICATE" client-cert.pem
   
   # Private key should have:
   grep "BEGIN.*PRIVATE KEY" client-key.pem
   ```

   > **Note:** Certificate files support chains. See [CONFIGURATION.md](CONFIGURATION.md) for certificate chain format details.

4. **Test certificate validity:**
   ```bash
   openssl verify -CAfile ca.pem client-cert.pem
   ```

### Error: "failed to decode PEM"

**Symptoms:**
```
Error: failed to decode PEM block
```

**Causes:**
- Wrong file format (DER instead of PEM)
- Corrupted file
- Not a certificate file

**Solutions:**

1. **Convert DER to PEM:**
   ```bash
   openssl x509 -inform DER -in cert.der -out cert.pem
   ```

2. **Verify file content:**
   ```bash
   file cert.pem  # Should show "PEM certificate"
   head -1 cert.pem  # Should show "-----BEGIN CERTIFICATE-----"
   ```

3. **Re-export certificate:**
   ```bash
   # From PKCS12
   openssl pkcs12 -in cert.p12 -out cert.pem -clcerts -nokeys
   openssl pkcs12 -in cert.p12 -out key.pem -nocerts -nodes
   ```

## Rotation Failures

### Error: "profile not found"

**Symptoms:**
```
Error: RotateRequest failed: profile 'my-profile' not found
```

**Solutions:**

1. **List existing profiles:**
   ```bash
   ./certz_client -op get-profile-list
   ```

2. **Create profile first:**
   ```bash
   ./certz_client -op add-profile -profile_id my-profile
   ```

### Error: "version already exists"

**Symptoms:**
```
Error: version '1.0' already exists
```

**Solutions:**

1. **Use different version:**
   ```bash
   -version "1.1"
   ```

2. **Force overwrite:**
   ```bash
   -force_overwrite=true  # This is actually the default!
   ```

###Error: "CSR signing failed"

**Symptoms:**
```
Error: failed to sign CSR: ...
```

**Solutions:**

1. **Verify CA certificate can sign:**
   ```bash
   openssl x509 -in ca-cert.pem -text -noout | grep -A5 "X509v3 Basic Constraints"
   # Should show: CA:TRUE
   ```

2. **Check CA private key:**
   ```bash
   openssl pkey -in ca-key.pem -check
   ```

3. **Verify key permissions:**
   ```bash
   chmod 600 ca-key.pem
   ```

## Validation Issues

### Error: "validation failed, rotation cancelled"

**Symptoms:**
```
Error: validation failed, rotation cancelled: failed to create test connection
```

**Meaning:**
The rotated certificates were uploaded but failed validation test. The rotation was automatically cancelled (rolled back).

**Debugging:**

1. **Test certificates manually:**
   ```bash
   # Try connecting with rotated certificates
   openssl s_client -connect 192.168.1.1:57400 \
     -cert new-cert.pem \
     -key new-key.pem \
     -CAfile new-ca.pem
   ```

2. **Check certificate-key match:**
   ```bash
   # See "Certificate Problems" section above
   ```

3. **Verify CA bundle:**
   ```bash
   # New CA bundle must contain CA that signed new cert
   openssl verify -CAfile new-ca.pem new-cert.pem
   ```

4. **Retry without validation:**
   ```bash
   # For debugging only - removes safety check!
   # Remove -validate flag
   ```

## Performance Problems

### Slow Upload Operations

**Symptoms:**
Large certificate bundles take long to upload

**Solutions:**

1. **Optimize CA bundle:**
   ```bash
   # Remove unnecessary intermediate CAs
   # Keep only: Root CA + Intermediate CA(s) needed for your cert
   ```

2. **Increase timeout:**
   ```bash
   -timeout 5m
   ```

3. **Use faster network connection**

### High Memory Usage

**Symptoms:**
Client consumes lots of memory

**Solutions:**

1. **Don't load huge CRL bundles**
   - Split into multiple smaller bundles
   - Remove expired CRLs

2. **Process profiles sequentially:**
   ```bash
   # Instead of one script processing all profiles
   # Process one at a time
   for profile in profile1 profile2 profile3; do
       ./certz_client -op rotate -profile_id "$profile" ...
   done
   ```

## Debugging Techniques

### Enable Verbose Mode

```bash
-v  # Shows detailed progress
```

### Capture gRPC Traffic

```bash
# Set gRPC logging environment variables
export GRPC_GO_LOG_VERBOSITY_LEVEL=99
export GRPC_GO_LOG_SEVERITY_LEVEL=info

./certz_client ...
```

### Test with OpenSSL

```bash
# Test server TLS
openssl s_client -connect 192.168.1.1:57400 -showcerts

# Test mutual TLS
openssl s_client -connect 192.168.1.1:57400 \
  -cert client-cert.pem \
  -key client-key.pem \
  -CAfile ca.pem
```

### Verify Certificate Chain

```bash
# Check certificate expiration
openssl x509 -in cert.pem -noout -dates

# Check certificate subject and issuer
openssl x509 -in cert.pem -noout -subject -issuer

# Verify full chain
openssl verify -CAfile ca-bundle.pem -untrusted intermediate.pem cert.pem
```

### Packet Capture

```bash
# Capture gRPC traffic
tcpdump -i any -w certz.pcap port 57400

# Analyze with Wireshark
wireshark certz.pcap
```

### Check Device Logs

**IOS XR:**
```
# Monitor gRPC logs
show logging | include grpc

# Enable debugging (if needed)
debug grpc all
```

### Test Minimal Connection

```bash
# Simplest possible connection test
./certz_client \
  -target_addr 192.168.1.1:57400 \
  -username admin \
  -password admin123 \
  -insecure_skip_verify \
  -op get-profile-list \
  -v
```

## Error Reference Table

| Error Code | Meaning | Common Cause | Quick Fix |
|------------|---------|--------------|-----------|
| `connection refused` | Server not listening | gRPC not configured | Enable gRPC on device |
| `context deadline exceeded` | Timeout | Network latency | Increase `-timeout` |
| `Unauthenticated` | Auth failed | Wrong credentials | Verify username/password |
| `PermissionDenied` | Authorization failed | Insufficient privileges | Use admin account |
| `NotFound` | Profile doesn't exist | Profile not created | Run `add-profile` first |
| `AlreadyExists` | Duplicate resource | Version conflict | Change `-version` or use `-force_overwrite` |
| `Unknown` | Server error | Device issue | Check device logs |

## Getting Help

If issues persist:

1. **Collect diagnostic information:**
   ```bash
   # Run with verbose mode
   ./certz_client -op get-profile-list -v 2>&1 | tee debug.log
   
   # Check versions
   ./certz_client -help | head -5
   go version
   
   # Network info
   ip route get 192.168.1.1
   ```

2. **Check device status:**
   - CPU/memory usage
   - gRPC configuration
   - Certificate trust store
   - AAA configuration

3. **Review documentation:**
   - [README.md](../README.md) - Getting started
   - [EXAMPLES.md](EXAMPLES.md) - Worked examples
   - [CONFIGURATION.md](CONFIGURATION.md) - Best practices

4. **Open GitHub issue** with:
   - Client version
   - Device type and version
   - Complete command line
   - Full error message
   - Debug logs (with `-v`)
