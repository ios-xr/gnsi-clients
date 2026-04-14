# Certz Client - Project Structure

```
certz-client/
├── README.md                    # Main documentation with quick start
├── BUILD.md                     # Build instructions for all platforms
├── Makefile                     # Build automation
├── go.mod                       # Go module dependencies
├── go.sum                       # Dependency checksums
│
├── cmd/
│   └── certz_client/
│       └── main.go              # CLI interface and flag parsing
│
├── pkg/
│   ├── connection/
│   │   └── connection.go        # gRPC connection management (TLS/mTLS)
│   │
│   ├── operations/
│   │   ├── certz.go            # Certz RPC operations (AddProfile, Rotate, etc.)
│   │   ├── csr.go              # CSR operations and signing
│   │   ├── errors.go           # Error constants and definitions
│   │   └── validation.go       # Input validation helpers
│   │
│   └── logger/
│       └── logger.go           # Structured logging
│
└── docs/
    ├── ARCHITECTURE.md         # This file - project structure
    ├── EXAMPLES.md             # Complete usage examples and workflows
    ├── CONFIGURATION.md        # Configuration best practices
    └── TROUBLESHOOTING.md      # Debugging and problem resolution
```

## Component Overview

### cmd/certz_client/main.go
- CLI interface and command-line flag parsing
- Operation routing to pkg/operations
- Connection setup via pkg/connection

### pkg/connection/connection.go
- gRPC connection management
- TLS/mTLS configuration
- Username/password credential handling

### pkg/operations/
- **certz.go**: Core gNSI Certz RPC operations (AddProfile, DeleteProfile, GetProfileList, Rotate, etc.)
- **csr.go**: CSR generation and signing utilities
- **errors.go**: Centralized error constants
- **validation.go**: Input validation helpers

### pkg/logger/logger.go
- Structured logging with verbosity control

## Documentation Files

- **[README.md](../README.md)**: Command-line flag reference and CSR suite documentation
- **[docs/EXAMPLES.md](EXAMPLES.md)**: Complete usage examples and workflows
- **[docs/CONFIGURATION.md](CONFIGURATION.md)**: Configuration best practices and security
- **[docs/TROUBLESHOOTING.md](TROUBLESHOOTING.md)**: Common issues and debugging
- **[BUILD.md](../BUILD.md)**: Build instructions for all platforms

## gNSI Certz Protocol Overview

This client implements the gNSI Certz v1.0 protocol for certificate management:

### Supported Operations

- **AddProfile**: Create new SSL/TLS profiles
- **DeleteProfile**: Remove SSL/TLS profiles
- **GetProfileList**: Query existing profiles
- **CanGenerateCSR**: Check if device supports CSR generation
- **Rotate**: Certificate lifecycle management with validation

### Certificate Sources

The client supports multiple certificate provisioning methods:

1. **Client-Provided**: Upload certificate and private key from client
2. **Device-Generated CSR**: Device creates CSR, client signs and uploads
3. **Device-Native Certificates**:
   - IDevID (Initial Device Identifier) - Factory-installed
   - oIDevID (Operational IDevID) - Operator-provisioned
   - Self-signed - Device-generated

### Rotation Workflow

1. **Upload**: Client sends certificate/CA/CRL to device
2. **Validation** (optional): Test new credentials before commit
3. **Finalize**: Commit changes or rollback on failure

For detailed examples, see [docs/EXAMPLES.md](EXAMPLES.md).
