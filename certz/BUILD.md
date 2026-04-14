# Build Instructions for gNSI Certz Client

## Quick Build

```bash
cd /nobackup/manbhasi/certz_open_src_client/gnsi/certz/certz-client
cd cmd/certz_client
go build -o certz_client
```

## Prerequisites

1. Go 1.19 or later installed
2. Network access to download dependencies

## Step-by-Step Build

### 1. Navigate to Project

```bash
cd /nobackup/manbhasi/certz_open_src_client/gnsi/certz/certz-client
```

### 2. Download Dependencies

```bash
go mod download
go mod tidy
```

### 3. Build the Client

```bash
cd cmd/certz_client
go build -o certz_client
```

### 4. Verify Build

```bash
./certz_client -help
```

You should see the help output with all available flags.

## Testing the Build

### Test Connection to IOS XR

```bash
./certz_client \
  -target_addr 1.2.27.3:57400 \
  -target_name ems.cisco.com \
  -tls \
  -username cisco \
  -password cisco123 \
  -insecure_skip_verify \
  -op get-profile-list \
  -v
```

### Test with Verbose Output

```bash
./certz_client \
  -target_addr 1.2.27.3:57400 \
  -target_name ems.cisco.com \
  -tls \
  -username cisco \
  -password cisco123 \
  -insecure_skip_verify \
  -op get-profile-list \
  -v
```

## Building for Different Platforms

### Linux (AMD64)

```bash
GOOS=linux GOARCH=amd64 go build -o certz_client_linux
```

### Linux (ARM64)

```bash
GOOS=linux GOARCH=arm64 go build -o certz_client_arm64
```

### macOS (AMD64)

```bash
GOOS=darwin GOARCH=amd64 go build -o certz_client_macos
```

### macOS (ARM64/M1/M2)

```bash
GOOS=darwin GOARCH=arm64 go build -o certz_client_macos_arm64
```

### Windows

```bash
GOOS=windows GOARCH=amd64 go build -o certz_client.exe
```

## Installing Globally

### Option 1: Copy to /usr/local/bin

```bash
sudo cp certz_client /usr/local/bin/
```

### Option 2: Use go install

From the project root:

```bash
go install ./cmd/certz_client
```

This installs to `$GOPATH/bin/certz_client`

## Troubleshooting Build Issues

### Issue: "package not found"

```bash
# Solution: Download dependencies
go mod download
go mod tidy
```

### Issue: "cannot find module"

```bash
# Solution: Initialize module
go mod init github.com/cisco/gnsi-certz-client
go mod tidy
```

### Issue: Build fails on imports

```bash
# Solution: Clean and rebuild
go clean
go mod tidy
go build
```

## Development Build (with debug info)

```bash
go build -gcflags="all=-N -l" -o certz_client_debug
```

## Production Build (optimized)

```bash
go build -ldflags="-s -w" -o certz_client
```

Flags:
- `-s`: Omit symbol table
- `-w`: Omit DWARF debug info
- Results in smaller binary

## Build with Version Information

```bash
VERSION="1.0.0"
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
go build -ldflags="-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}" -o certz_client
```

## Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests verbosely
go test -v ./...
```

## Static Analysis

```bash
# Format code
go fmt ./...

# Vet code
go vet ./...

# Run golint (if installed)
golint ./...
```

## Next Steps

After building:

1. Test with your IOS XR device
2. Review the examples in`examples/`
3. Read the documentation in `docs/`
4. Configure for your environment

## Quick Test Command

Your exact command from the example:

```bash
./certz_client \
  -target_name ems.cisco.com \
  -target_addr 1.2.27.3:57400 \
  -username cisco \
  -password cisco123 \
  -tls \
  -insecure_skip_verify \
  -op add-profile \
  -profile_id test \
  -v
```

Then to rotate:

```bash
./certz_client \
  -target_name ems.cisco.com \
  -target_addr 1.2.27.3:57400 \
  -username cisco \
  -password cisco123 \
  -tls \
  -insecure_skip_verify \
  -op rotate \
  -profile_id test \
  -cert_source IDevID \
  -ca_bundle_file ems.pem \
  -version "2.0" \
  -v
```

## Support

If you encounter build issues:
1. Ensure Go 1.19+ is installed: `go version`
2. Check network connectivity
3. Verify GOPATH is set correctly
4. Try `go clean -modcache` and rebuild
