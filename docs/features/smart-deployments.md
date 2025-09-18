# Smart Deployments

Frank CLI provides intelligent deployment capabilities that automatically handle resource creation, updates, and status monitoring.

## What Makes Deployments "Smart"

### 1. Intelligent Resource Updates

Frank automatically determines if resources need updates by comparing existing and desired states:

- **No changes detected** → Resource is already up to date
- **Changes detected** → Resource will be updated
- **New resource** → Resource will be created

### 2. Stack Tracking

All resources managed by Frank are automatically tagged with stack information:

```yaml
metadata:
  annotations:
    frankthetank.cloud/stack-name: myapp-dev-web
  labels:
    app.kubernetes.io/name: web
    app.kubernetes.io/version: 1.2.3
    app.kubernetes.io/managed-by: frank
```

### 3. Status Monitoring

Frank waits for resources to be ready and provides real-time status updates:

```
2024-01-15T10:30:00Z - myapp-dev-web - Creating Deployment
2024-01-15T10:30:01Z - myapp-dev-web - Waiting for resource to be ready
2024-01-15T10:30:05Z - myapp-dev-web - Resource is ready
```

### 4. Parallel Processing

Multiple deployments run concurrently for faster execution:

```bash
frank apply dev  # Deploys all dev apps in parallel
```

## Resource Comparison Logic

### Deployment Comparison

Frank compares the following fields to determine if a Deployment needs updating:

- **Spec changes**: replicas, image, ports, environment variables
- **Metadata changes**: labels, annotations
- **Template changes**: container specifications

### Service Comparison

For Services, Frank compares:

- **Port configurations**: port numbers, protocols, target ports
- **Selector changes**: label selectors
- **Type changes**: ClusterIP, NodePort, LoadBalancer

### ConfigMap/Secret Comparison

For ConfigMaps and Secrets:

- **Data changes**: configuration values
- **Metadata changes**: labels, annotations

## Status Monitoring

### Resource States

Frank monitors these resource states:

- **Creating** → Resource is being created
- **Updating** → Resource is being updated
- **Ready** → Resource is ready and available
- **Failed** → Resource creation/update failed
- **Already up to date** → No changes needed

### Status Messages

| Status | Description | Color |
|--------|-------------|-------|
| `Creating <Resource>` | Creating new resource | Yellow |
| `Updating <Resource>` | Updating existing resource | Yellow |
| `Resource is ready` | Resource is ready | Green |
| `Resource is already up to date` | No changes needed | Green |
| `Apply failed` | Error occurred | Red |

### Timeout Handling

Frank uses configurable timeouts for resource readiness:

```yaml
# config/app.yaml
timeout: 10m  # 10 minutes timeout
```

If a resource doesn't become ready within the timeout:
- Frank reports the timeout error
- The deployment is marked as failed
- You can investigate and retry

## Kubernetes Best Practices

### Standard Labels

Frank automatically adds Kubernetes best practice labels:

```yaml
metadata:
  labels:
    app.kubernetes.io/name: web           # App name
    app.kubernetes.io/version: 1.2.3     # App version
    app.kubernetes.io/managed-by: frank  # Management tool
```

### Stack Annotations

All resources get stack tracking annotations:

```yaml
metadata:
  annotations:
    frankthetank.cloud/stack-name: myapp-dev-web
```

## Parallel Deployment

### How It Works

Frank processes multiple configurations concurrently:

1. **Configuration Discovery** - Find all config files
2. **Stack Filtering** - Filter by stack argument
3. **Parallel Processing** - Process each config concurrently
4. **Status Monitoring** - Monitor all deployments simultaneously

### Benefits

- **Faster deployments** - Multiple apps deploy simultaneously
- **Better resource utilization** - Efficient use of cluster resources
- **Improved user experience** - Quicker feedback

### Example

```bash
# Deploy all dev applications in parallel
frank apply dev

# Output shows parallel processing:
2024-01-15T10:30:00Z - myapp-dev-web - Creating Deployment
2024-01-15T10:30:00Z - myapp-dev-api - Creating Service
2024-01-15T10:30:01Z - myapp-dev-web - Waiting for resource to be ready
2024-01-15T10:30:01Z - myapp-dev-api - Resource is ready
2024-01-15T10:30:05Z - myapp-dev-web - Resource is ready
```

## Error Handling

### Common Errors

**Resource Already Exists**:
```
Error: failed to create Deployment: resource already exists
```

**Immutable Field Changes**:
```
Error: failed to update Service: field is immutable
```

**Timeout Waiting for Resource**:
```
Error: timeout waiting for resource to be ready
```

### Error Recovery

Frank provides clear error messages and suggestions:

1. **Check existing resources** - Use `kubectl get` commands
2. **Verify permissions** - Check RBAC settings
3. **Check cluster resources** - Ensure sufficient capacity
4. **Review configuration** - Validate YAML syntax

## Configuration Options

### Timeout Configuration

Set deployment timeouts in your configuration:

```yaml
# config/app.yaml
timeout: 5m   # 5 minutes
timeout: 30m  # 30 minutes
timeout: 1h   # 1 hour
timeout: 0    # No timeout (wait indefinitely)
```

### Log Level Configuration

Control verbosity of deployment output:

```bash
# Environment variable
export FRANK_LOG_LEVEL=debug

# Or per command
FRANK_LOG_LEVEL=debug frank apply dev
```

## Monitoring and Observability

### Resource Tracking

All frank-managed resources can be found using labels:

```bash
# List all frank-managed resources
kubectl get all -l app.kubernetes.io/managed-by=frank

# List resources for specific stack
kubectl get all -l frankthetank.cloud/stack-name=myapp-dev-web
```

### Status Monitoring

Monitor deployment status:

```bash
# Check deployment status
kubectl get deployments -l app.kubernetes.io/managed-by=frank

# Check pod status
kubectl get pods -l app.kubernetes.io/managed-by=frank

# Check service status
kubectl get services -l app.kubernetes.io/managed-by=frank
```

## Best Practices

### 1. Use Appropriate Timeouts

```yaml
# Good - reasonable timeout
timeout: 10m

# Avoid - too high or too low
timeout: 0    # No timeout
timeout: 1h   # Too high for most apps
```

### 2. Monitor Resource Status

```bash
# Check status after deployment
kubectl get pods -l app.kubernetes.io/managed-by=frank
kubectl describe deployment myapp
```

### 3. Use Stack Filtering

```bash
# Deploy specific applications
frank apply dev/web
frank apply dev/api

# Instead of deploying everything
frank apply dev
```

### 4. Test Before Production

```bash
# Always test in dev first
frank apply dev

# Then staging
frank apply staging

# Finally production
frank apply prod
```

## Troubleshooting

### Debug Deployment Issues

```bash
# Enable debug logging
FRANK_LOG_LEVEL=debug frank apply dev

# Check resource status
kubectl get pods -l app.kubernetes.io/managed-by=frank
kubectl describe deployment myapp
```

### Common Issues

**Slow Deployments**:
- Check cluster resources
- Use appropriate timeouts
- Deploy fewer resources at once

**Resource Conflicts**:
- Check for existing resources
- Use different names or namespaces
- Clean up old resources

**Permission Issues**:
- Check RBAC settings
- Verify service account permissions
- Use appropriate contexts

## Next Steps

- [Jinja Templating](jinja-templating.md) - Learn about dynamic templates
- [Stack Filtering](stack-filtering.md) - Organize your deployments
- [Namespace Management](namespace-management.md) - Handle namespaces properly
