<div align="center">

<img src="docs/assets/logo.png" alt="frank">

Simple multi-environment management of Kubernetes resources.

[![Go](https://img.shields.io/badge/go-1.25-00ADD8.svg?logo=go)](https://tip.golang.org/doc/go1.25)
[![Go Report Card](https://goreportcard.com/badge/github.com/schnauzersoft/frank-cli)](https://goreportcard.com/report/github.com/schnauzersoft/frank-cli)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE.md)
[![Test Status](https://github.com/schnauzersoft/frank-cli/actions/workflows/test.yml/badge.svg)](https://github.com/schnauzersoft/frank-cli/actions/workflows/ci.yml)

</div>

---

## Overview

**frank** is a CLI tool for managing resources in Kubernetes. This project is inspired by many different projects, such as:

- [kubectl](https://kubernetes.io/docs/reference/kubectl/)
- [terraform](https://developer.hashicorp.com/terraform)
- [sceptre](https://github.com/Sceptre/sceptre/tree/master)

The namesake of this project, however, is the ultimate inspiration for this application as this project follows similar patterns to how that individual managed cloud-native deployments regardless of the tooling at hand. The intention of this project is to create a stateless (so no terraform), packageless (so no helm), stack-based (so no vanilla kubectl), template-driven mechanism for managing Kubernetes resources. The closest off-the-shelf project that's comparable to the intentions of **frank** is probably [kustomize](https://kustomize.io/). However, **frank** is intented to be more generic and accessible (by leveraging Jinja and HCL) than the esoteric templating patterns of kustomize.

Philosphy of the project:

1. This is a templated deployment project first and foremost. It is not a packaging project, it is not a linting project, it is a project designed to make deploying a single simple boilerplate collection of resources (such a "backend API" Kubernetes deployment and service combination) as easy as possible. The belief is that the vast majority of anyone ever using Kubernetes (a choice which is not entirely optional these days) only need a simple "docker compose" (by level of complexity) configuration to deploy their containers. In the experience of the **frank** maintainers, the standard tooling out there for Kubernetes today lacks that simplicity.
2. This tool exists to enable people to "do the DevOps" - meaning: the project is designed for deploying essentially the exact same collection of resources (defined as a "stack" by **frank**) to multiple environments, including simultaneously in parallel. This enables users to manage local development, alpha, staging, production environments as Gene Kim intended.


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
# manifests/app-deployment.j2
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

```hcl
# manifests/app-deployment.hcl
resource "kubernetes_deployment" "app" {
  metadata = {
    name = "${stack_name}"
    labels = {
      "app.kubernetes.io/name" = "${app}"
      "app.kubernetes.io/version" = "${version}"
      "app.kubernetes.io/managed-by" = "frank"
    }
  }

  spec = {
    replicas = ${replicas}

    selector = {
      matchLabels = {
        "app.kubernetes.io/name" = "${app}"
      }
    }

    template = {
      metadata = {
        labels = {
          "app.kubernetes.io/name" = "${app}"
          "app.kubernetes.io/version" = "${version}"
        }
      }

      spec = {
        containers = [
          {
            name  = "${app}"
            image = "${image_name}:${version}"
            ports = [
              {
                containerPort = ${port}
              }
            ]
            env = [
              {
                name  = "ENVIRONMENT"
                value = "${environment}"
              },
              {
                name  = "PROJECT_CODE"
                value = "${project_code}"
              },
              {
                name  = "NAMESPACE"
                value = "${k8s_namespace}"
              }
            ]
          }
        ]
      }
    }
  }
}
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

# Show changes to environments
frank plan dev                 # Show diff of expected changes to dev resources
frank plan prod/app1-backend    # Show diff of changes to a single stack
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
frank plan dev
```

## Development

### Prerequisites

- Go 1.25 or later
- Kubernetes cluster access

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
FRANK_LOG_LEVEL=debug ./frank apply local

# Test stack filtering
./frank apply local
./frank delete local

# Run unit tests
go test ./...
```
