<div align="center">

# frank

## Simple multi-environment management of Kubernetes resources.

[![Go](https://img.shields.io/badge/go-1.25-00ADD8.svg?logo=go)](https://tip.golang.org/doc/go1.25)
[![Go Report Card](https://goreportcard.com/badge/github.com/schnauzersoft/frank-cli)](https://goreportcard.com/report/github.com/schnauzersoft/frank-cli)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE.md)
[![Test Status](https://github.com/schnauzersoft/frank-cli/actions/workflows/test.yml/badge.svg)](https://github.com/schnauzersoft/frank-cli/actions/workflows/test.yml)

> [!CAUTION]
> This project is pre-release! It's still actively being tested. An official GitHub release will be added once it's ready.

</div>

---

## Quick Start

frank is a CLI tool for applying templated Kubernetes manifest files to clusters with intelligent configuration management, stack-based filtering.

### 1. Building

```bash
git clone https://github.com/schnauzersoft/frank-cli
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

### **Template Support**
Dynamic manifest generation with powerful templating:

#### **Jinja Templating**
Advanced templating with conditionals, loops, and filters:

```yaml
# manifests/app-deployment.jinja
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ stack_name }}
  labels:
    app.kubernetes.io/name: {{ app_name }}
    app.kubernetes.io/version: {{ version }}
    app.kubernetes.io/managed-by: frank
spec:
  replicas: {{ replicas | default(3) }}
  selector:
    matchLabels:
      app.kubernetes.io/name: {{ app_name }}
  template:
    metadata:
      labels:
        app.kubernetes.io/name: {{ app_name }}
        app.kubernetes.io/version: {{ version }}
    spec:
      containers:
      - name: {{ app_name }}
        image: {{ image_name }}:{{ version }}
        ports:
        - containerPort: {{ port | default(80) }}
```

#### **HCL Templating**
Simple variable substitution with familiar syntax:

```yaml
# manifests/app-deployment.hcl
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${stack_name}
  labels:
    app.kubernetes.io/name: ${app}
    app.kubernetes.io/version: ${version}
    app.kubernetes.io/managed-by: frank
spec:
  replicas: ${replicas}
  selector:
    matchLabels:
      app.kubernetes.io/name: ${app}
  template:
    metadata:
      labels:
        app.kubernetes.io/name: ${app}
        app.kubernetes.io/version: ${version}
    spec:
      containers:
      - name: ${app}
        image: ${image_name}:${version}
        ports:
        - containerPort: ${port}
```

**Template Context Variables:**
- `stack_name` - Generated stack name (e.g., `myapp-dev-web`)
- `app_name` - App name from config or filename
- `version` - Version from config
- `project_code` - Project identifier
- `context` - Kubernetes context
- `namespace` - Target namespace
- `k8s_namespace` - Alias for namespace

**Multi-Document Support:**
Templates can contain multiple Kubernetes resources separated by `---`:

```yaml
# manifests/full-stack.jinja
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ stack_name }}
  labels:
    app.kubernetes.io/name: {{ app_name }}
---
apiVersion: v1
kind: Service
metadata:
  name: {{ stack_name }}-service
  labels:
    app.kubernetes.io/name: {{ app_name }}
spec:
  selector:
    app.kubernetes.io/name: {{ app_name }}
  ports:
  - port: 80
    targetPort: {{ port | default(80) }}
```

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
app: myapp                     # Optional: App name (defaults to filename)
version: 1.2.3                 # Optional: Version for templates
```

### Template Files

frank supports both Jinja and HCL templating for dynamic manifest generation:

**Supported Extensions:**
- `.jinja` - Jinja template files
- `.j2` - Alternative Jinja extension
- `.hcl` - HCL template files
- `.tf` - Terraform-style HCL files

**Jinja Example:**
```yaml
# config/app.yaml
manifest: app-deployment.jinja  # Points to template file
app: myapp
version: 1.2.3

# manifests/app-deployment.jinja
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ stack_name }}
  labels:
    app.kubernetes.io/name: {{ app_name }}
    app.kubernetes.io/version: {{ version }}
```

**HCL Example:**
```yaml
# config/app.yaml
manifest: app-deployment.hcl    # Points to template file
app: myapp
version: 1.2.3

# manifests/app-deployment.hcl
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${stack_name}
  labels:
    app.kubernetes.io/name: ${app}
    app.kubernetes.io/version: ${version}
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
│   ├── stack/                # Stack management
│   └── template/             # Jinja and HCL templating engine
├── config/                   # Example configurations
├── manifests/                # Example manifests
└── main.go                   # Application entry point
```

## Development

### Prerequisites

- Go 1.25 or later
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
