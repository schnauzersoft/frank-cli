# Resource Management

Frank CLI provides comprehensive resource management capabilities for creating, updating, and cleaning up Kubernetes resources.

## Resource Lifecycle

### 1. Resource Creation

Frank creates resources when they don't exist:

```bash
frank apply dev
# Creates new resources in the dev environment
```

### 2. Resource Updates

Frank intelligently updates existing resources:

```bash
frank apply dev
# Updates existing resources if changes are detected
```

### 3. Resource Cleanup

Frank can clean up resources it manages:

```bash
frank delete dev
# Removes all frank-managed resources in dev
```

## Resource Tracking

### Stack Annotations

All frank-managed resources are tagged with stack information:

```yaml
metadata:
  annotations:
    frankthetank.cloud/stack-name: myapp-dev-web
```

### Kubernetes Labels

Frank adds standard Kubernetes labels:

```yaml
metadata:
  labels:
    app.kubernetes.io/name: web
    app.kubernetes.io/version: 1.2.3
    app.kubernetes.io/managed-by: frank
```

## Resource Types

Frank supports these Kubernetes resource types:

### Workload Resources

- **Deployments** - Stateless applications
- **StatefulSets** - Stateful applications
- **DaemonSets** - Node-level services
- **Pods** - Individual containers

### Service Resources

- **Services** - Network access to applications
- **Ingress** - HTTP/HTTPS routing

### Configuration Resources

- **ConfigMaps** - Configuration data
- **Secrets** - Sensitive data

### Batch Resources

- **Jobs** - One-time tasks
- **CronJobs** - Scheduled tasks

## Resource Comparison

Frank compares existing and desired resources to determine if updates are needed:

### Deployment Comparison

Frank compares these fields:

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

## Resource Status

Frank monitors resource status and provides real-time updates:

### Status Messages

| Status | Description | Color |
|--------|-------------|-------|
| `Creating <Resource>` | Creating new resource | Yellow |
| `Updating <Resource>` | Updating existing resource | Yellow |
| `Resource is ready` | Resource is ready | Green |
| `Resource is already up to date` | No changes needed | Green |
| `Apply failed` | Error occurred | Red |

### Status Monitoring

Frank waits for resources to be ready:

```bash
frank apply dev
# 2024-01-15T10:30:00Z - myapp-dev-web - Creating Deployment
# 2024-01-15T10:30:01Z - myapp-dev-web - Waiting for resource to be ready
# 2024-01-15T10:30:05Z - myapp-dev-web - Resource is ready
```

## Resource Cleanup

### Delete Command

Remove frank-managed resources:

```bash
# Delete all frank-managed resources
frank delete

# Delete specific environment
frank delete dev

# Delete specific application
frank delete dev/app

# Delete specific stack
frank delete myapp-dev-web
```

### Stack Filtering

Frank supports stack-based filtering for deletion:

```bash
# Delete all dev environment resources
frank delete dev

# Delete all dev/app* resources
frank delete dev/app

# Delete specific configuration
frank delete dev/app.yaml
```

### Resource Identification

Frank identifies resources to delete using:

1. **Stack annotations** - `frankthetank.cloud/stack-name`
2. **Stack filtering** - Match stack name patterns
3. **Resource types** - Only delete supported resource types

## Resource Management Best Practices

### 1. Use Stack Annotations

All resources should have stack annotations:

```yaml
metadata:
  annotations:
    frankthetank.cloud/stack-name: myapp-dev-web
```

### 2. Use Standard Labels

Use Kubernetes best practice labels:

```yaml
metadata:
  labels:
    app.kubernetes.io/name: web
    app.kubernetes.io/version: 1.2.3
    app.kubernetes.io/managed-by: frank
```

### 3. Organize by Environment

Structure resources by environment:

```yaml
# config/dev/config.yaml
namespace: myapp-dev

# config/prod/config.yaml
namespace: myapp-prod
```

### 4. Use Resource Limits

Set appropriate resource limits:

```yaml
resources:
  requests:
    memory: "256Mi"
    cpu: "100m"
  limits:
    memory: "512Mi"
    cpu: "500m"
```

## Resource Monitoring

### List Resources

List frank-managed resources:

```bash
# List all frank-managed resources
kubectl get all -l app.kubernetes.io/managed-by=frank

# List resources for specific stack
kubectl get all -l frankthetank.cloud/stack-name=myapp-dev-web

# List resources in specific namespace
kubectl get all -l app.kubernetes.io/managed-by=frank -n myapp-dev
```

### Check Resource Status

Monitor resource status:

```bash
# Check deployment status
kubectl get deployments -l app.kubernetes.io/managed-by=frank

# Check pod status
kubectl get pods -l app.kubernetes.io/managed-by=frank

# Check service status
kubectl get services -l app.kubernetes.io/managed-by=frank
```

### Resource Details

Get detailed resource information:

```bash
# Describe deployment
kubectl describe deployment myapp-dev-web

# Get deployment YAML
kubectl get deployment myapp-dev-web -o yaml

# Check pod logs
kubectl logs deployment/myapp-dev-web
```

## Resource Troubleshooting

### Common Issues

**Resource Already Exists**:
```
Error: failed to create Deployment: resource already exists
```

**Immutable Field Changes**:
```
Error: failed to update Service: field is immutable
```

**Resource Not Found**:
```
Error: resource not found
```

**Permission Denied**:
```
Error: permission denied
```

### Debug Commands

```bash
# Enable debug logging
FRANK_LOG_LEVEL=debug frank apply dev

# Check resource status
kubectl get pods -l app.kubernetes.io/managed-by=frank
kubectl describe deployment myapp-dev-web

# Check resource events
kubectl get events --sort-by=.metadata.creationTimestamp
```

### Resource Recovery

If resources are in a bad state:

```bash
# Delete and recreate
kubectl delete deployment myapp-dev-web
frank apply dev

# Or force update
kubectl patch deployment myapp-dev-web -p '{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt":"'$(date +%Y-%m-%dT%H:%M:%S%z)'"}}}}}'
```

## Resource Security

### RBAC Configuration

Set up proper RBAC for resource management:

```yaml
# manifests/rbac.jinja
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: frank-manager
rules:
- apiGroups: [""]
  resources: ["pods", "services", "configmaps", "secrets"]
  verbs: ["get", "list", "create", "update", "delete"]
- apiGroups: ["apps"]
  resources: ["deployments", "statefulsets", "daemonsets"]
  verbs: ["get", "list", "create", "update", "delete"]
- apiGroups: ["batch"]
  resources: ["jobs", "cronjobs"]
  verbs: ["get", "list", "create", "update", "delete"]
- apiGroups: ["networking.k8s.io"]
  resources: ["ingresses"]
  verbs: ["get", "list", "create", "update", "delete"]
```

### Resource Quotas

Set resource quotas to prevent resource exhaustion:

```yaml
# manifests/resource-quota.jinja
apiVersion: v1
kind: ResourceQuota
metadata:
  name: {{ namespace }}-quota
  namespace: {{ namespace }}
spec:
  hard:
    requests.cpu: "2"
    requests.memory: 4Gi
    limits.cpu: "4"
    limits.memory: 8Gi
    pods: "10"
    services: "5"
    configmaps: "10"
    secrets: "10"
```

## Resource Lifecycle Management

### 1. Development Lifecycle

```bash
# Create resources
frank apply dev

# Update resources
frank apply dev

# Clean up resources
frank delete dev
```

### 2. Production Lifecycle

```bash
# Deploy to production
frank apply prod

# Update production
frank apply prod

# Rollback if needed
kubectl rollout undo deployment/myapp-prod-web
```

### 3. Maintenance Lifecycle

```bash
# Scale down for maintenance
kubectl scale deployment myapp-prod-web --replicas=0

# Perform maintenance
# ...

# Scale back up
kubectl scale deployment myapp-prod-web --replicas=3
```

## Next Steps

- [Smart Deployments](smart-deployments.md) - Learn about intelligent deployments
- [Stack Filtering](stack-filtering.md) - Organize your deployments
- [Namespace Management](namespace-management.md) - Handle namespaces properly
