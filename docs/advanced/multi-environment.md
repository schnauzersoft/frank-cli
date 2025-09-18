# Multi-Environment Setup

This guide shows you how to set up Frank for managing multiple environments (dev, staging, production) with different configurations and deployments.

## Directory Structure

A typical multi-environment setup looks like this:

```
my-project/
├── config/
│   ├── config.yaml              # Base configuration
│   ├── dev/
│   │   ├── config.yaml          # Dev environment overrides
│   │   ├── app.yaml             # App config for dev
│   │   └── api.yaml             # API config for dev
│   ├── staging/
│   │   ├── config.yaml          # Staging environment overrides
│   │   ├── app.yaml             # App config for staging
│   │   └── api.yaml             # API config for staging
│   └── prod/
│       ├── config.yaml          # Production environment overrides
│       ├── app.yaml             # App config for production
│       └── api.yaml             # API config for production
└── manifests/
    ├── app-deployment.jinja     # App deployment template
    └── api-deployment.jinja     # API deployment template
```

## Base Configuration

Start with a base configuration that defines common settings:

```yaml
# config/config.yaml
project_code: myapp
namespace: myapp
timeout: 10m
```

## Environment-Specific Configurations

### Development Environment

```yaml
# config/dev/config.yaml
context: dev-cluster
namespace: myapp-dev
replicas: 2
image_tag: latest
```

### Staging Environment

```yaml
# config/staging/config.yaml
context: staging-cluster
namespace: myapp-staging
replicas: 3
image_tag: staging
```

### Production Environment

```yaml
# config/prod/config.yaml
context: prod-cluster
namespace: myapp-prod
replicas: 5
image_tag: stable
```

## Application Configurations

### Development App Config

```yaml
# config/dev/app.yaml
manifest: app-deployment.jinja
app: web
version: 1.0.0-dev
replicas: 2
image_name: myapp/web
port: 8080
```

### Staging App Config

```yaml
# config/staging/app.yaml
manifest: app-deployment.jinja
app: web
version: 1.0.0-staging
replicas: 3
image_name: myapp/web
port: 8080
```

### Production App Config

```yaml
# config/prod/app.yaml
manifest: app-deployment.jinja
app: web
version: 1.0.0
replicas: 5
image_name: myapp/web
port: 8080
```

## Jinja Templates

Use Jinja templates to create environment-specific deployments:

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
        image: {{ image_name }}:{{ image_tag | default('latest') }}
        ports:
        - containerPort: {{ port | default(80) }}
        env:
        - name: ENVIRONMENT
          value: {{ context }}
        - name: VERSION
          value: {{ version }}
        resources:
          requests:
            memory: "{{ memory_request | default('256Mi') }}"
            cpu: "{{ cpu_request | default('100m') }}"
          limits:
            memory: "{{ memory_limit | default('512Mi') }}"
            cpu: "{{ cpu_limit | default('500m') }}"
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

## Deployment Commands

### Deploy to Specific Environment

```bash
# Deploy to development
frank apply dev

# Deploy to staging
frank apply staging

# Deploy to production
frank apply prod
```

### Deploy Specific Application

```bash
# Deploy web app to dev
frank apply dev/app

# Deploy API to staging
frank apply staging/api

# Deploy web app to production
frank apply prod/web
```

### Deploy Everything

```bash
# Deploy all environments
frank apply

# Deploy all dev applications
frank apply dev
```

## Environment-Specific Features

### Different Resource Limits

```yaml
# config/dev/config.yaml
memory_request: 128Mi
memory_limit: 256Mi
cpu_request: 50m
cpu_limit: 200m

# config/prod/config.yaml
memory_request: 512Mi
memory_limit: 1Gi
cpu_request: 250m
cpu_limit: 1000m
```

### Different Image Tags

```yaml
# config/dev/config.yaml
image_tag: dev-latest

# config/staging/config.yaml
image_tag: staging-v1.0.0

# config/prod/config.yaml
image_tag: v1.0.0
```

### Different Replica Counts

```yaml
# config/dev/config.yaml
replicas: 1

# config/staging/config.yaml
replicas: 2

# config/prod/config.yaml
replicas: 5
```

## CI/CD Integration

### GitHub Actions Example

```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    branches: [ main, develop ]

jobs:
  deploy-dev:
    if: github.ref == 'refs/heads/develop'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Deploy to dev
        run: |
          frank apply dev --yes

  deploy-staging:
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Deploy to staging
        run: |
          frank apply staging --yes

  deploy-prod:
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    needs: deploy-staging
    steps:
      - uses: actions/checkout@v3
      - name: Deploy to production
        run: |
          frank apply prod --yes
```

## Best Practices

### 1. Use Consistent Naming

```yaml
# Good - consistent naming
config/dev/app.yaml
config/staging/app.yaml
config/prod/app.yaml

# Avoid - inconsistent naming
config/dev/web.yaml
config/staging/app.yaml
config/prod/frontend.yaml
```

### 2. Inherit Common Settings

```yaml
# Base config with common settings
# config/config.yaml
project_code: myapp
timeout: 10m
image_name: myapp/web
port: 8080

# Environment-specific overrides
# config/dev/config.yaml
context: dev-cluster
namespace: myapp-dev
replicas: 2
```

### 3. Use Environment Variables for Secrets

```yaml
# config/prod/config.yaml
context: prod-cluster
namespace: myapp-prod
# Don't put secrets in config files
# Use environment variables instead
```

### 4. Test in Dev First

```bash
# Always test in dev first
frank apply dev

# Then staging
frank apply staging

# Finally production
frank apply prod
```

### 5. Use Stack Filtering

```bash
# Deploy specific applications
frank apply dev/web
frank apply dev/api

# Instead of deploying everything
frank apply dev
```

## Troubleshooting

### Common Issues

**"context not found"**
- Check available contexts: `kubectl config get-contexts`
- Verify context names in config files

**"namespace conflict"**
- Ensure only one source specifies namespace
- Check both config and manifest files

**"template rendering failed"**
- Verify all template variables are defined
- Check Jinja syntax in templates

### Debug Commands

```bash
# Enable debug logging
FRANK_LOG_LEVEL=debug frank apply dev

# Check configuration inheritance
FRANK_LOG_LEVEL=debug frank apply dev/app

# Verify template rendering
FRANK_LOG_LEVEL=debug frank apply staging
```

## Next Steps

- [CI/CD Integration](ci-cd.md) - Set up automated deployments
- [Debugging](debugging.md) - Troubleshoot deployment issues
- [Best Practices](best-practices.md) - Learn advanced patterns
