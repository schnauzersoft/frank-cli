# Building

This guide covers how to build Frank CLI from source for development and production use.

## Prerequisites

Before building Frank, ensure you have:

- **Go 1.25 or later** - [Download Go](https://golang.org/dl/)
- **Git** - [Download Git](https://git-scm.com/downloads)
- **Make** (optional) - For using Makefile commands

## Building from Source

### 1. Clone the Repository

```bash
git clone https://github.com/schnauzersoft/frank-cli
cd frank-cli
```

### 2. Install Dependencies

```bash
go mod tidy
```

### 3. Build the Binary

```bash
# Build for current platform
go build -o frank .

# Build with version information
go build -ldflags "-X main.version=$(git describe --tags --always --dirty)" -o frank .

# Build with additional flags
go build -ldflags "-X main.version=$(git describe --tags --always --dirty) -X main.commit=$(git rev-parse HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o frank .
```

### 4. Verify the Build

```bash
# Check if binary was created
ls -la frank

# Test the binary
./frank --help

# Check version
./frank --version
```

## Cross-Platform Building

### Build for Different Platforms

```bash
# Build for Linux AMD64
GOOS=linux GOARCH=amd64 go build -o frank-linux-amd64 .

# Build for Windows AMD64
GOOS=windows GOARCH=amd64 go build -o frank-windows-amd64.exe .

# Build for macOS AMD64
GOOS=darwin GOARCH=amd64 go build -o frank-darwin-amd64 .

# Build for macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o frank-darwin-arm64 .
```

### Build Script

Create a build script for multiple platforms:

```bash
#!/bin/bash
# build.sh

set -e

VERSION=$(git describe --tags --always --dirty)
COMMIT=$(git rev-parse HEAD)
DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS="-X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE"

echo "Building Frank CLI $VERSION"

# Build for different platforms
echo "Building for Linux AMD64..."
GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o frank-linux-amd64 .

echo "Building for Windows AMD64..."
GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o frank-windows-amd64.exe .

echo "Building for macOS AMD64..."
GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFLAGS" -o frank-darwin-amd64 .

echo "Building for macOS ARM64..."
GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o frank-darwin-arm64 .

echo "Build complete!"
ls -la frank-*
```

## Development Building

### Hot Reload

For development, you can use tools like `air` for hot reloading:

```bash
# Install air
go install github.com/cosmtrek/air@latest

# Run with hot reload
air
```

### Debug Build

Build with debug information:

```bash
# Build with debug symbols
go build -gcflags="all=-N -l" -o frank .

# Build with race detection
go build -race -o frank .

# Build with coverage
go build -cover -o frank .
```

## Production Building

### Optimized Build

Build with optimizations for production:

```bash
# Build with optimizations
go build -ldflags "-s -w" -o frank .

# Build with additional optimizations
go build -ldflags "-s -w -X main.version=$(git describe --tags --always --dirty)" -o frank .
```

### Build with CGO Disabled

For better compatibility:

```bash
# Build with CGO disabled
CGO_ENABLED=0 go build -o frank .
```

## Makefile

Create a Makefile for common build tasks:

```makefile
# Makefile

.PHONY: build clean test install

# Variables
VERSION := $(shell git describe --tags --always --dirty)
COMMIT := $(shell git rev-parse HEAD)
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Default target
all: build

# Build for current platform
build:
	go build $(LDFLAGS) -o frank .

# Build for all platforms
build-all:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o frank-linux-amd64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o frank-windows-amd64.exe .
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o frank-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o frank-darwin-arm64 .

# Clean build artifacts
clean:
	rm -f frank frank-*

# Run tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -cover ./...

# Install to $GOPATH/bin
install:
	go install $(LDFLAGS) .

# Run linting
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Run all checks
check: fmt lint test

# Build with optimizations
build-optimized:
	go build -ldflags "-s -w $(LDFLAGS)" -o frank .

# Build with race detection
build-race:
	go build -race $(LDFLAGS) -o frank .

# Build with debug symbols
build-debug:
	go build -gcflags="all=-N -l" $(LDFLAGS) -o frank .
```

### Using the Makefile

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test

# Run all checks
make check

# Clean build artifacts
make clean
```

## Docker Building

### Dockerfile

Create a Dockerfile for containerized builds:

```dockerfile
# Dockerfile

# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o frank .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary
COPY --from=builder /app/frank .

# Copy example configs
COPY --from=builder /app/config ./config
COPY --from=builder /app/manifests ./manifests

# Set the binary as executable
RUN chmod +x frank

# Default command
CMD ["./frank"]
```

### Build Docker Image

```bash
# Build Docker image
docker build -t frank-cli .

# Build with specific tag
docker build -t frank-cli:latest .

# Build with version tag
docker build -t frank-cli:$(git describe --tags --always --dirty) .
```

## CI/CD Building

### GitHub Actions

```yaml
# .github/workflows/build.yml
name: Build

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  build:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.25'
    
    - name: Build
      run: |
        go build -o frank .
        chmod +x frank
    
    - name: Test
      run: go test ./...
    
    - name: Upload binary
      uses: actions/upload-artifact@v3
      with:
        name: frank-binary
        path: frank
```

### GitLab CI

```yaml
# .gitlab-ci.yml
stages:
  - build
  - test

build:
  stage: build
  image: golang:1.25-alpine
  script:
    - go build -o frank .
    - chmod +x frank
  artifacts:
    paths:
      - frank
    expire_in: 1 hour

test:
  stage: test
  image: golang:1.25-alpine
  script:
    - go test ./...
```

## Building for Distribution

### Create Release Archives

```bash
#!/bin/bash
# release.sh

set -e

VERSION=$(git describe --tags --always --dirty)
COMMIT=$(git rev-parse HEAD)
DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS="-X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE"

echo "Creating release archives for $VERSION"

# Create release directory
mkdir -p release

# Build for different platforms
echo "Building for Linux AMD64..."
GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o frank-linux-amd64 .
tar -czf release/frank-linux-amd64.tar.gz frank-linux-amd64

echo "Building for Windows AMD64..."
GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o frank-windows-amd64.exe .
zip -q release/frank-windows-amd64.zip frank-windows-amd64.exe

echo "Building for macOS AMD64..."
GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFLAGS" -o frank-darwin-amd64 .
tar -czf release/frank-darwin-amd64.tar.gz frank-darwin-amd64

echo "Building for macOS ARM64..."
GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o frank-darwin-arm64 .
tar -czf release/frank-darwin-arm64.tar.gz frank-darwin-arm64

echo "Release archives created in release/ directory"
ls -la release/
```

## Troubleshooting

### Common Build Issues

**"go: module not found"**
- Run `go mod tidy` to download dependencies
- Check if you're in the correct directory

**"go: build constraints exclude all Go files"**
- Check if you have the correct Go version
- Verify the source code is present

**"CGO_ENABLED=0" issues**
- Some dependencies might require CGO
- Try building without CGO disabled

**"ldflags" issues**
- Check if the variables are properly set
- Verify the syntax of ldflags

### Debug Build Issues

```bash
# Enable verbose output
go build -v -o frank .

# Check Go version
go version

# Check module status
go mod verify

# Check for build constraints
go list -f '{{.BuildConstraints}}' ./...
```

## Next Steps

- [Testing](testing.md) - Learn about testing Frank CLI
- [Contributing](contributing.md) - Contribute to Frank CLI development
- [Architecture](architecture.md) - Understand Frank CLI architecture
