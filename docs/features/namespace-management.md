# Namespace Management

Frank CLI provides intelligent namespace management with conflict detection and flexible configuration options.

## How Namespace Management Works

Frank handles namespaces through a hierarchical system with conflict detection:

1. **Configuration namespaces** - Set in `config.yaml` files
2. **Manifest namespaces** - Set in Kubernetes manifest files
3. **Template namespaces** - Set in Jinja templates
4. **Conflict detection** - Prevents namespace conflicts

## Namespace Sources

### 1. Configuration Files

Set namespaces in your configuration files:

```yaml
# config/config.yaml
namespace: myapp

# config/dev/config.yaml
namespace: myapp-dev

# config/prod/config.yaml
namespace: myapp-prod
```

### 2. Kubernetes Manifests

Set namespaces directly in manifest files:

```yaml
# manifests/app-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  namespace: myapp-dev  # Namespace in manifest
spec:
  # ... deployment spec
```

### 3. Jinja Templates

Set namespaces in Jinja templates:

```yaml
# manifests/app-deployment.jinja
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ stack_name }}
  namespace: {{ namespace }}  # Namespace from template context
spec:
  # ... deployment spec
```

## Namespace Inheritance

Frank uses hierarchical namespace inheritance:

### Base Configuration

```yaml
# config/config.yaml
namespace: myapp  # Base namespace
```

### Environment Overrides

```yaml
# config/dev/config.yaml
namespace: myapp-dev  # Overrides base namespace

# config/prod/config.yaml
namespace: myapp-prod  # Overrides base namespace
```

### Final Namespace Resolution

| Config File | Base Namespace | Environment Namespace | Final Namespace |
|-------------|----------------|----------------------|-----------------|
| `config/dev/app.yaml` | `myapp` | `myapp-dev` | `myapp-dev` |
| `config/prod/app.yaml` | `myapp` | `myapp-prod` | `myapp-prod` |
| `config/app.yaml` | `myapp` | - | `myapp` |

## Conflict Detection

Frank prevents namespace conflicts by detecting when both configuration and manifest specify namespaces.

### Valid Configurations

**Option 1: Configuration namespace only**
```yaml
# config/config.yaml
namespace: myapp-dev

# manifests/app-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  # No namespace - uses config namespace
```

**Option 2: Manifest namespace only**
```yaml
# config/config.yaml
# No namespace field

# manifests/app-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  namespace: myapp-dev  # Namespace in manifest
```

### Invalid Configurations

**Namespace conflict**:
```yaml
# config/config.yaml
namespace: myapp-dev

# manifests/app-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  namespace: myapp-prod  # ERROR: Conflict with config namespace
```

Frank will error with:
```
Error: namespace conflict: both config and manifest specify namespace
```

## Template Context Variables

Frank provides namespace variables in Jinja templates:

### Available Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `namespace` | Resolved namespace | `myapp-dev` |
| `k8s_namespace` | Alias for namespace | `myapp-dev` |

### Template Examples

```yaml
# manifests/app-deployment.jinja
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ stack_name }}
  namespace: {{ namespace }}  # Uses resolved namespace
spec:
  # ... deployment spec
---
apiVersion: v1
kind: Service
metadata:
  name: {{ stack_name }}-service
  namespace: {{ k8s_namespace }}  # Alias for namespace
spec:
  # ... service spec
```

## Best Practices

### 1. Use Configuration Namespaces

**Recommended approach** - Set namespaces in configuration files:

```yaml
# config/config.yaml
namespace: myapp

# config/dev/config.yaml
namespace: myapp-dev

# config/prod/config.yaml
namespace: myapp-prod
```

**Benefits**:
- Centralized namespace management
- Easy to change across all applications
- Consistent with environment structure

### 2. Avoid Manifest Namespaces

**Avoid** - Setting namespaces in manifest files:

```yaml
# DON'T DO THIS!
# manifests/app-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  namespace: myapp-dev  # Hardcoded namespace
```

**Problems**:
- Hardcoded values
- Difficult to change
- Can cause conflicts

### 3. Use Template Variables

**Good** - Use template variables for namespaces:

```yaml
# manifests/app-deployment.jinja
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ stack_name }}
  namespace: {{ namespace }}  # Dynamic namespace
```

**Benefits**:
- Dynamic namespace resolution
- No hardcoded values
- Flexible and maintainable

## Environment-Specific Namespaces

### Development Environment

```yaml
# config/dev/config.yaml
context: dev-cluster
namespace: myapp-dev
```

### Staging Environment

```yaml
# config/staging/config.yaml
context: staging-cluster
namespace: myapp-staging
```

### Production Environment

```yaml
# config/prod/config.yaml
context: prod-cluster
namespace: myapp-prod
```

## Namespace Validation

Frank validates namespaces at deployment time:

### 1. Conflict Detection

Frank checks for namespace conflicts before deploying:

```bash
frank apply dev
# Error: namespace conflict: both config and manifest specify namespace
```

### 2. Namespace Resolution

Frank resolves the final namespace using inheritance:

```bash
FRANK_LOG_LEVEL=debug frank apply dev
# DEBUG: Resolved namespace: myapp-dev
```

### 3. Validation Errors

Common namespace validation errors:

```
Error: namespace conflict: both config and manifest specify namespace
Error: invalid namespace: namespace cannot be empty
Error: namespace not found: namespace 'myapp-dev' does not exist
```

## Troubleshooting

### Common Issues

**"namespace conflict"**
- Check both config and manifest files
- Choose one approach (config or manifest)
- Don't specify both

**"namespace not found"**
- Verify the namespace exists in Kubernetes
- Check namespace name spelling
- Create the namespace if needed

**"invalid namespace"**
- Check namespace name format
- Ensure namespace is not empty
- Verify namespace syntax

### Debug Commands

```bash
# Enable debug logging to see namespace resolution
FRANK_LOG_LEVEL=debug frank apply dev

# Check available namespaces
kubectl get namespaces

# Check namespace configuration
kubectl config view --minify
```

### Namespace Creation

If a namespace doesn't exist, create it:

```bash
# Create namespace
kubectl create namespace myapp-dev

# Or create with labels
kubectl create namespace myapp-dev --labels=environment=dev
```

## Advanced Patterns

### 1. Dynamic Namespaces

Use Jinja templates for dynamic namespaces:

```yaml
# manifests/app-deployment.jinja
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ stack_name }}
  namespace: {{ namespace }}-{{ environment | default('dev') }}
```

### 2. Conditional Namespaces

Use conditional logic for namespaces:

```yaml
# manifests/app-deployment.jinja
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ stack_name }}
  {% if environment == 'prod' %}
  namespace: {{ namespace }}-prod
  {% else %}
  namespace: {{ namespace }}-{{ environment }}
  {% endif %}
```

### 3. Namespace Labels

Add labels to namespaces:

```yaml
# manifests/namespace.jinja
apiVersion: v1
kind: Namespace
metadata:
  name: {{ namespace }}
  labels:
    environment: {{ environment }}
    project: {{ project_code }}
    managed-by: frank
```

## Integration with Kubernetes

### RBAC Configuration

Set up RBAC for namespaces:

```yaml
# manifests/rbac.jinja
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: frank-role
  namespace: {{ namespace }}
rules:
- apiGroups: [""]
  resources: ["pods", "services", "configmaps"]
  verbs: ["get", "list", "create", "update", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: frank-rolebinding
  namespace: {{ namespace }}
subjects:
- kind: ServiceAccount
  name: frank-sa
  namespace: {{ namespace }}
roleRef:
  kind: Role
  name: frank-role
  apiGroup: rbac.authorization.k8s.io
```

### Resource Quotas

Set resource quotas for namespaces:

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
```

## Next Steps

- [Smart Deployments](smart-deployments.md) - Learn about intelligent deployments
- [Stack Filtering](stack-filtering.md) - Organize your deployments
- [Multi-Environment Setup](../advanced/multi-environment.md) - Set up multiple environments
