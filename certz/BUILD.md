# Build Instructions for gNSI Certz Client

## Quick Build

```bash
cd cmd/certz_client
go build -o certz_client
```

## Prerequisites

- Go 1.19 or later installed
- Network access to download dependencies

## Step-by-Step Build

### 1. Download Dependencies

```bash
go mod download
go mod tidy
```

### 2. Build the Client

```bash
cd cmd/certz_client
go build -o certz_client
```

### 3. Verify Build

```bash
./certz_client -help
```

You should see the help output with all available flags.

## Testing the Build

After building, test the client with your device. See [docs/EXAMPLES.md](docs/EXAMPLES.md) for usage examples.

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

### Windows (AMD64)

```bash
GOOS=windows GOARCH=amd64 go build -o certz_client.exe
```

## Installation

After building, you can install the binary to your system:

```bash
# Install to $GOPATH/bin
go install

# Or copy to a directory in your PATH
sudo cp certz_client /usr/local/bin/
```

## References

- [README.md](README.md) - Main documentation and command-line reference
- [docs/EXAMPLES.md](docs/EXAMPLES.md) - Usage examples

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

## Support

If you encounter build issues:
1. Ensure Go 1.19+ is installed: `go version`
2. Check network connectivity
3. Verify GOPATH is set correctly
4. Try `go clean -modcache` and rebuild

For usage examples, see [docs/EXAMPLES.md](docs/EXAMPLES.md).
