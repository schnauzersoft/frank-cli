# Frank CLI

Simple multi-environment management of Kubernetes resources.

Frank is a CLI tool for applying templated Kubernetes manifest files to clusters with intelligent configuration management, stack-based filtering, and interactive safety prompts.

## Quick Start

### 1. Installation

```bash
git clone <repository-url>
cd frank-cli
go build -o frank .
```

### 2. Basic Setup

Create a simple project structure:

```
my-project/
├── config/
│   ├── config.yaml          # Base configuration
│   └── app.yaml             # App-specific config
└── manifests/
    └── app-deployment.yaml  # Kubernetes manifest
```

### 3. Configuration

**Base config** (`config/config.yaml`):
```yaml
context: my-cluster
project_code: myapp
namespace: myapp-namespace
```

**App config** (`config/app.yaml`):
```yaml
manifest: app-deployment.yaml
timeout: 5m
```

**Kubernetes manifest** (`manifests/app-deployment.yaml`):
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  replicas: 3
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
      - name: myapp
        image: nginx:alpine
        ports:
        - containerPort: 80
```

### 4. Deploy

```bash
# Interactive deployment (with confirmation)
frank apply

# Skip confirmation
frank apply --yes

# Deploy specific stack
frank apply app

# Deploy all dev environment stacks
frank apply dev
```

## Features

### **Smart Deployments**
- Creates new resources or updates existing ones intelligently
- Adds stack tracking annotations (`frankthetank.cloud/stack-name`)
- Waits patiently for deployments to be ready
- Runs multiple deployments in parallel for speed

### **Stack-Based Filtering**
Target specific environments or applications:

```bash
frank apply                    # Deploy everything
frank apply dev                # Deploy all dev environment stacks
frank apply dev/app            # Deploy all dev/app* configurations
frank apply dev/app.yaml       # Deploy specific configuration file
```

### **Hierarchical Configuration**
Organize your environments with inheritance:

```
config/
├── config.yaml              # Base: context=prod, project_code=myapp
├── dev/
│   ├── config.yaml          # Override: context=dev
│   └── app.yaml             # App config
├── staging/
│   ├── config.yaml          # Override: context=staging
│   └── web/
│       ├── config.yaml      # Override: context=web-staging
│       └── api.yaml         # API config
└── prod/
    ├── config.yaml          # Override: context=prod-cluster
    └── api.yaml             # API config
```

### **Namespace Management**
Smart namespace handling with conflict detection:

```yaml
# config/config.yaml
context: my-cluster
namespace: myapp-namespace    # Global namespace

# config/dev/config.yaml
context: dev-cluster
# Inherits namespace from parent

# manifests/app.yaml (no namespace) -> uses config namespace
# manifests/app.yaml (with namespace) -> ERROR if config also has one
```

### **Clean Resource Management**
Delete resources with surgical precision:

```bash
frank delete                    # Remove all frank-managed resources
frank delete dev                # Remove all dev environment resources
frank delete dev/app            # Remove all dev/app* stack resources
frank delete myapp-dev-web      # Remove specific stack
```

## Configuration Reference

### Base Configuration (`config.yaml`)

```yaml
context: my-cluster           # Required: Kubernetes context
project_code: myapp          # Optional: Project identifier
namespace: myapp-namespace   # Optional: Default namespace
```

### App Configuration (`*.yaml` files)

```yaml
manifest: app-deployment.yaml  # Required: Manifest file name
timeout: 10m                   # Optional: Deployment timeout (default: 10m)
```

### Environment Variables

```bash
export FRANK_LOG_LEVEL=debug  # Set log level (debug, info, warn, error)
```

### Configuration Precedence

1. Environment variables (`FRANK_LOG_LEVEL`)
2. `.frank.yaml` (current directory)
3. `$HOME/.frank/config.yaml`
4. `/etc/frank/config.yaml`

## Commands

### `frank apply [stack]`

Deploy Kubernetes manifests to clusters.

**Options:**
- `-y, --yes` - Skip confirmation prompt

**Examples:**
```bash
frank apply                    # Deploy all stacks
frank apply dev                # Deploy dev environment
frank apply dev/app --yes      # Deploy dev/app without confirmation
```

### `frank delete [stack]`

Remove frank-managed Kubernetes resources.

**Options:**
- `-y, --yes` - Skip confirmation prompt

**Examples:**
```bash
frank delete                   # Remove all frank-managed resources
frank delete dev               # Remove dev environment resources
frank delete prod --yes        # Remove prod resources without confirmation
```

## Advanced Usage

### Multi-Environment Setup

```bash
# Deploy to different environments
frank apply dev                # Deploy to dev cluster
frank apply staging            # Deploy to staging cluster
frank apply prod               # Deploy to production cluster

# Clean up environments
frank delete dev               # Remove dev resources
frank delete staging           # Remove staging resources
```

### CI/CD Integration

```bash
# In your CI pipeline
frank apply prod --yes         # Deploy without prompts
frank delete staging --yes     # Clean up staging
```

### Debugging

```bash
# Enable debug logging
FRANK_LOG_LEVEL=debug frank apply dev

# Check what would be deployed
frank apply dev                # Shows confirmation with scope
```

## Project Structure

```
frank-cli/
├── cmd/                       # Command implementations
│   ├── root.go               # Root command
│   ├── apply.go              # Apply command
│   ├── delete.go             # Delete command
│   └── utils.go              # Shared utilities
├── pkg/
│   ├── config/               # Configuration management
│   ├── deploy/               # Deployment orchestration
│   ├── kubernetes/           # Kubernetes operations
│   └── stack/                # Stack management
├── config/                   # Example configurations
├── manifests/                # Example manifests
└── main.go                   # Application entry point
```

## Development

### Prerequisites

- Go 1.21 or later
- Kubernetes cluster access
- kubectl configured

### Building

```bash
go mod tidy
go build -o frank .
```

### Testing

```bash
# Test basic functionality
./frank apply --yes

# Test with debug logging
FRANK_LOG_LEVEL=debug ./frank apply dev

# Test stack filtering
./frank apply dev
./frank delete dev
```
