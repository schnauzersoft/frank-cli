# Kubernetes Labels

Frank CLI automatically applies Kubernetes best practice labels to all managed resources for better organization and management.

## Standard Labels

Frank applies these standard Kubernetes labels to all resources:

### Core Labels

| Label | Value | Description |
|-------|-------|-------------|
| `app.kubernetes.io/name` | App name | Name of the application |
| `app.kubernetes.io/version` | Version | Version of the application |
| `app.kubernetes.io/managed-by` | `frank` | Management tool identifier |

### Example

```yaml
metadata:
  labels:
    app.kubernetes.io/name: web
    app.kubernetes.io/version: 1.2.3
    app.kubernetes.io/managed-by: frank
```

## Label Sources

### App Name (`app.kubernetes.io/name`)

The app name comes from:

1. **Config file** - `app` field in configuration
2. **Filename** - Extracted from configuration filename

#### Examples

```yaml
# config/app.yaml
app: web
# Result: app.kubernetes.io/name=web

# config/api.yaml (no app field)
# Result: app.kubernetes.io/name=api (from filename)
```

### Version (`app.kubernetes.io/version`)

The version comes from:

1. **Config file** - `version` field in configuration
2. **Default** - Empty string if not specified

#### Examples

```yaml
# config/app.yaml
version: 1.2.3
# Result: app.kubernetes.io/version=1.2.3

# config/app.yaml (no version field)
# Result: app.kubernetes.io/version="" (empty)
```

### Managed By (`app.kubernetes.io/managed-by`)

Always set to `frank` for all frank-managed resources.

## Stack Annotations

Frank also adds stack tracking annotations:

### Stack Name Annotation

```yaml
metadata:
  annotations:
    frankthetank.cloud/stack-name: myapp-dev-web
```

This annotation is used for:
- Resource identification
- Stack-based filtering
- Resource cleanup

## Label Inheritance

Labels are applied to all resource types:

### Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp-dev-web
  labels:
    app.kubernetes.io/name: web
    app.kubernetes.io/version: 1.2.3
    app.kubernetes.io/managed-by: frank
  annotations:
    frankthetank.cloud/stack-name: myapp-dev-web
spec:
  template:
    metadata:
      labels:
        app.kubernetes.io/name: web
        app.kubernetes.io/version: 1.2.3
        app.kubernetes.io/managed-by: frank
```

### Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: myapp-dev-web-service
  labels:
    app.kubernetes.io/name: web
    app.kubernetes.io/version: 1.2.3
    app.kubernetes.io/managed-by: frank
  annotations:
    frankthetank.cloud/stack-name: myapp-dev-web
```

### ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: myapp-dev-web-config
  labels:
    app.kubernetes.io/name: web
    app.kubernetes.io/version: 1.2.3
    app.kubernetes.io/managed-by: frank
  annotations:
    frankthetank.cloud/stack-name: myapp-dev-web
```

## Resource Selection

Labels enable powerful resource selection and management:

### Select by App

```bash
# Get all resources for web app
kubectl get all -l app.kubernetes.io/name=web

# Get all resources for API app
kubectl get all -l app.kubernetes.io/name=api
```

### Select by Version

```bash
# Get all resources with specific version
kubectl get all -l app.kubernetes.io/version=1.2.3

# Get all resources with latest version
kubectl get all -l app.kubernetes.io/version=latest
```

### Select by Management Tool

```bash
# Get all frank-managed resources
kubectl get all -l app.kubernetes.io/managed-by=frank

# Get all non-frank resources
kubectl get all -l '!app.kubernetes.io/managed-by'
```

### Select by Stack

```bash
# Get all resources for specific stack
kubectl get all -l frankthetank.cloud/stack-name=myapp-dev-web

# Get all resources for dev environment
kubectl get all -l frankthetank.cloud/stack-name=myapp-dev-*
```

## Label Management

### Adding Custom Labels

You can add custom labels in your templates:

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
    # Custom labels
    environment: {{ environment | default('development') }}
    team: {{ team | default('platform') }}
    tier: {{ tier | default('frontend') }}
```

### Label Selectors

Use labels in selectors:

```yaml
# manifests/app-service.jinja
apiVersion: v1
kind: Service
metadata:
  name: {{ stack_name }}-service
spec:
  selector:
    app.kubernetes.io/name: {{ app_name }}
    app.kubernetes.io/version: {{ version }}
    app.kubernetes.io/managed-by: frank
```

## Monitoring and Observability

### Prometheus Integration

Labels enable Prometheus monitoring:

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
    # Prometheus labels
    prometheus.io/scrape: "true"
    prometheus.io/port: "8080"
    prometheus.io/path: "/metrics"
```

### Grafana Dashboards

Use labels in Grafana dashboards:

```yaml
# Query for all frank-managed resources
kubectl get all -l app.kubernetes.io/managed-by=frank

# Query for specific app
kubectl get all -l app.kubernetes.io/name=web

# Query for specific version
kubectl get all -l app.kubernetes.io/version=1.2.3
```

## Resource Organization

### By Application

```bash
# List all applications
kubectl get all -l app.kubernetes.io/managed-by=frank --show-labels | grep app.kubernetes.io/name

# Get resources for specific app
kubectl get all -l app.kubernetes.io/name=web
```

### By Environment

```bash
# Get all dev resources
kubectl get all -l frankthetank.cloud/stack-name=myapp-dev-*

# Get all prod resources
kubectl get all -l frankthetank.cloud/stack-name=myapp-prod-*
```

### By Version

```bash
# Get all resources with specific version
kubectl get all -l app.kubernetes.io/version=1.2.3

# Get all resources with latest version
kubectl get all -l app.kubernetes.io/version=latest
```

## Best Practices

### 1. Use Standard Labels

Always use the standard Kubernetes labels:

```yaml
metadata:
  labels:
    app.kubernetes.io/name: {{ app_name }}
    app.kubernetes.io/version: {{ version }}
    app.kubernetes.io/managed-by: frank
```

### 2. Add Meaningful Custom Labels

Add custom labels that provide value:

```yaml
metadata:
  labels:
    app.kubernetes.io/name: {{ app_name }}
    app.kubernetes.io/version: {{ version }}
    app.kubernetes.io/managed-by: frank
    environment: {{ environment }}
    team: {{ team }}
    tier: {{ tier }}
```

### 3. Use Consistent Label Values

Ensure label values are consistent across resources:

```yaml
# Good - consistent values
app.kubernetes.io/name: web
app.kubernetes.io/version: 1.2.3

# Avoid - inconsistent values
app.kubernetes.io/name: web
app.kubernetes.io/version: v1.2.3  # Different format
```

### 4. Use Labels in Selectors

Use labels in selectors for proper resource association:

```yaml
spec:
  selector:
    app.kubernetes.io/name: {{ app_name }}
    app.kubernetes.io/version: {{ version }}
    app.kubernetes.io/managed-by: frank
```

## Troubleshooting

### Common Issues

**Labels not applied**:
- Check if resources are created by Frank
- Verify template syntax
- Check for label conflicts

**Selectors not working**:
- Verify label names and values
- Check for typos in selectors
- Ensure labels are applied to target resources

**Resource not found**:
- Check label values
- Verify resource exists
- Check namespace

### Debug Commands

```bash
# Check labels on resources
kubectl get all -l app.kubernetes.io/managed-by=frank --show-labels

# Check specific resource labels
kubectl get deployment myapp-dev-web --show-labels

# Check label selectors
kubectl get all --selector app.kubernetes.io/name=web
```

## Next Steps

- [Configuration Schema](configuration-schema.md) - Learn about configuration structure
- [Template Variables](template-variables.md) - Understand template context
- [Troubleshooting](troubleshooting.md) - Debug label issues
