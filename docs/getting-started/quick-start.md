# Quick Start

Get up and running with Frank in minutes! This guide will walk you through building Frank, setting up a basic project, and deploying your first Kubernetes resources.

## Prerequisites

Before you begin, make sure you have:

- **Go 1.25 or later** - [Download Go](https://golang.org/dl/)
- **Kubernetes cluster access** - Local (minikube/kind) or remote cluster
- **kubectl configured** - `kubectl get nodes` should work

## Building Frank

```bash
# Clone the repository
git clone https://github.com/schnauzersoft/frank-cli
cd frank-cli

# Build the binary
go build -o frank .

# Verify it works
./frank --help
```

## Basic Project Setup

Create a simple project structure:

```
my-project/
├── config/
│   ├── config.yaml          # Base configuration
│   └── app.yaml             # App-specific config
└── manifests/
    └── app-deployment.yaml  # Kubernetes manifest
```

### 1. Base Configuration

Create `config/config.yaml`:

```yaml
context: my-cluster
project_code: myapp
namespace: myapp-namespace
```

### 2. App Configuration

Create `config/app.yaml`:

```yaml
manifest: app-deployment.yaml
timeout: 5m
```

### 3. Kubernetes Manifest

Create `manifests/app-deployment.yaml`:

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

## Your First Deployment

Navigate to your project directory and deploy:

```bash
cd my-project

# Deploy with confirmation prompt
frank apply

# Or skip confirmation
frank apply --yes
```

You should see output like:

```
2024-01-15T10:30:00Z - myapp-my-cluster-app - Creating Deployment
2024-01-15T10:30:01Z - myapp-my-cluster-app - Waiting for resource to be ready
2024-01-15T10:30:05Z - myapp-my-cluster-app - Resource is ready
```

## Verify Deployment

Check that your deployment was created:

```bash
kubectl get deployments
kubectl get pods
```

You should see your `myapp` deployment running with 3 replicas.

## Using Jinja Templates

Let's create a more dynamic deployment using Jinja templating:

### 1. Update App Config

Update `config/app.yaml`:

```yaml
manifest: app-deployment.jinja
app: myapp
version: 1.2.3
```

### 2. Create Jinja Template

Create `manifests/app-deployment.jinja`:

```yaml
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
        image: nginx:{{ version }}
        ports:
        - containerPort: {{ port | default(80) }}
---
apiVersion: v1
kind: Service
metadata:
  name: {{ stack_name }}-service
  labels:
    app.kubernetes.io/name: {{ app_name }}
    app.kubernetes.io/version: {{ version }}
    app.kubernetes.io/managed-by: frank
spec:
  selector:
    app.kubernetes.io/name: {{ app_name }}
  ports:
  - port: 80
    targetPort: {{ port | default(80) }}
```

### 3. Deploy Template

```bash
frank apply
```

This will create both a Deployment and Service with dynamic names and labels!

## Using HCL Templates

For simpler templating needs, you can use HCL (HashiCorp Configuration Language) syntax:

### 1. Update App Config

Update `config/app.yaml`:

```yaml
manifest: app-deployment.hcl
app: myapp
version: 1.2.3
vars:
  replicas: 3
  port: 8080
```

### 2. Create HCL Template

Create `manifests/app-deployment.hcl`:

```yaml
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
        image: nginx:${version}
        ports:
        - containerPort: ${port}
```

### 3. Deploy HCL Template

```bash
frank apply
```

HCL templates use `${variable}` syntax and are perfect for simple variable substitution without complex logic.

## Stack Filtering

Frank supports powerful stack filtering:

```bash
# Deploy everything
frank apply

# Deploy specific app
frank apply app

# Deploy all dev environment stacks (if you have dev/ directory)
frank apply dev

# Deploy specific dev app
frank apply dev/app
```

## Clean Up

When you're done experimenting:

```bash
# Remove all frank-managed resources
frank delete

# Or remove specific stack
frank delete app
```

## Next Steps

Now that you have Frank working, explore these topics:

- [Configuration](configuration.md) - Learn about all configuration options
- [Jinja Templating](../features/jinja-templating.md) - Master dynamic templates
- [HCL Templating](../features/hcl-templating.md) - Simple variable substitution
- [Stack Filtering](../features/stack-filtering.md) - Organize your deployments
- [Multi-Environment Setup](../advanced/multi-environment.md) - Scale to multiple environments

## Troubleshooting

### Common Issues

**"config directory with config.yaml not found"**
- Make sure you're in a directory with a `config/` subdirectory
- Ensure `config/config.yaml` exists

**"context not found"**
- Check your Kubernetes context: `kubectl config current-context`
- Update the `context` field in your config

**"permission denied"**
- Ensure your kubectl context has the necessary permissions
- Check RBAC settings for your user/service account

### Getting Help

- Check the [Troubleshooting Guide](../reference/troubleshooting.md)
- Open an [issue on GitHub](https://github.com/schnauzersoft/frank-cli/issues)
- Join the [discussions](https://github.com/schnauzersoft/frank-cli/discussions)
