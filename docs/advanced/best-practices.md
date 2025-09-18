# Best Practices

This guide covers best practices for using Frank CLI effectively in production environments.

## Configuration Management

### 1. Use Hierarchical Structure

**Good** - Clear hierarchy with inheritance:
```
config/
├── config.yaml              # Base: project_code, timeout
├── dev/
│   ├── config.yaml          # Override: context=dev
│   └── app.yaml             # App config
├── staging/
│   ├── config.yaml          # Override: context=staging
│   └── app.yaml             # App config
└── prod/
    ├── config.yaml          # Override: context=prod
    └── app.yaml             # App config
```

**Avoid** - Flat structure with duplication:
```
config/
├── dev-config.yaml
├── staging-config.yaml
├── prod-config.yaml
├── dev-app.yaml
├── staging-app.yaml
└── prod-app.yaml
```

### 2. Keep App Configs Simple

**Good** - Minimal app config:
```yaml
# config/app.yaml
manifest: app-deployment.jinja
app: web
version: 1.2.3
```

**Avoid** - Duplicating base config:
```yaml
# config/app.yaml
manifest: app-deployment.jinja
context: dev-cluster
project_code: myapp
namespace: myapp-dev
app: web
version: 1.2.3
```

### 3. Use Environment Variables for Secrets

**Good** - Use environment variables:
```bash
# Don't put secrets in config files
export DOCKER_REGISTRY_PASSWORD=secret123
export API_KEY=your-api-key
```

**Avoid** - Secrets in config files:
```yaml
# config/prod/config.yaml
docker_registry_password: secret123  # DON'T DO THIS!
api_key: your-api-key                # DON'T DO THIS!
```

## Template Design

### 1. Use Meaningful Variable Names

**Good** - Clear variable names:
```yaml
image: {{ app_image }}:{{ app_version }}
replicas: {{ app_replicas | default(3) }}
```

**Avoid** - Unclear variable names:
```yaml
image: {{ img }}:{{ v }}
replicas: {{ r | default(3) }}
```

### 2. Provide Sensible Defaults

**Good** - Always provide defaults:
```yaml
replicas: {{ replicas | default(3) }}
port: {{ port | default(80) }}
memory: {{ memory | default('256Mi') }}
```

**Avoid** - No defaults:
```yaml
replicas: {{ replicas }}  # Will error if not defined
port: {{ port }}          # Will error if not defined
```

### 3. Use Kubernetes Best Practice Labels

**Good** - Standard labels:
```yaml
metadata:
  labels:
    app.kubernetes.io/name: {{ app_name }}
    app.kubernetes.io/version: {{ version }}
    app.kubernetes.io/managed-by: frank
```

**Avoid** - Custom labels:
```yaml
metadata:
  labels:
    myapp-name: {{ app_name }}
    myapp-version: {{ version }}
    managed-by: frank
```

### 4. Keep Templates Readable

**Good** - Clear structure:
```yaml
spec:
  replicas: {{ replicas | default(3) }}
  selector:
    matchLabels:
      app.kubernetes.io/name: {{ app_name }}
  template:
    metadata:
      labels:
        app.kubernetes.io/name: {{ app_name }}
```

**Avoid** - Too much logic:
```yaml
spec:
  replicas: {{ replicas if replicas else 3 if environment == 'prod' else 2 if environment == 'staging' else 1 }}
```

## Deployment Strategies

### 1. Test in Dev First

**Good** - Progressive deployment:
```bash
# Test in dev first
frank apply dev

# Then staging
frank apply staging

# Finally production
frank apply prod
```

**Avoid** - Direct to production:
```bash
# DON'T DO THIS!
frank apply prod
```

### 2. Use Stack Filtering

**Good** - Deploy specific applications:
```bash
frank apply dev/web
frank apply dev/api
```

**Avoid** - Deploy everything:
```bash
frank apply dev  # Unless you really need everything
```

### 3. Use Confirmation Prompts

**Good** - Always confirm before deploying:
```bash
frank apply  # Shows confirmation with scope
```

**Avoid** - Skip confirmation in development:
```bash
frank apply --yes  # Only in CI/CD
```

## Resource Management

### 1. Use Appropriate Resource Limits

**Good** - Set resource limits:
```yaml
resources:
  requests:
    memory: "256Mi"
    cpu: "100m"
  limits:
    memory: "512Mi"
    cpu: "500m"
```

**Avoid** - No resource limits:
```yaml
# No resources section - can cause issues
```

### 2. Use Health Checks

**Good** - Include health checks:
```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10
readinessProbe:
  httpGet:
    path: /ready
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
```

### 3. Use ConfigMaps for Configuration

**Good** - Use ConfigMaps:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ stack_name }}-config
data:
  app.properties: |
    server.port={{ port | default(8080) }}
    app.name={{ app_name }}
---
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: {{ app_name }}
        envFrom:
        - configMapRef:
            name: {{ stack_name }}-config
```

## Security Best Practices

### 1. Use Namespace Isolation

**Good** - Separate namespaces:
```yaml
# config/dev/config.yaml
namespace: myapp-dev

# config/prod/config.yaml
namespace: myapp-prod
```

### 2. Use Service Accounts

**Good** - Create service accounts:
```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ stack_name }}-sa
  namespace: {{ namespace }}
---
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      serviceAccountName: {{ stack_name }}-sa
```

### 3. Use Resource Quotas

**Good** - Set resource quotas:
```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: {{ stack_name }}-quota
  namespace: {{ namespace }}
spec:
  hard:
    requests.cpu: "2"
    requests.memory: 4Gi
    limits.cpu: "4"
    limits.memory: 8Gi
```

## Monitoring and Observability

### 1. Use Consistent Labeling

**Good** - Consistent labels:
```yaml
metadata:
  labels:
    app.kubernetes.io/name: {{ app_name }}
    app.kubernetes.io/version: {{ version }}
    app.kubernetes.io/managed-by: frank
    environment: {{ context }}
```

### 2. Include Monitoring Annotations

**Good** - Add monitoring annotations:
```yaml
metadata:
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "8080"
    prometheus.io/path: "/metrics"
```

### 3. Use Structured Logging

**Good** - Structured logs:
```yaml
env:
- name: LOG_LEVEL
  value: "{{ log_level | default('info') }}"
- name: LOG_FORMAT
  value: "json"
```

## CI/CD Best Practices

### 1. Use Environment-Specific Configurations

**Good** - Different configs per environment:
```yaml
# .github/workflows/deploy.yml
- name: Deploy to dev
  if: github.ref == 'refs/heads/develop'
  run: frank apply dev --yes

- name: Deploy to prod
  if: github.ref == 'refs/heads/main'
  run: frank apply prod --yes
```

### 2. Use Secrets Management

**Good** - Use Kubernetes secrets:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: {{ stack_name }}-secrets
type: Opaque
stringData:
  password: "{{ secret_password }}"
```

### 3. Use Rollback Strategies

**Good** - Plan for rollbacks:
```bash
# Keep previous version available
kubectl get deployments -l app.kubernetes.io/name=web

# Rollback if needed
kubectl rollout undo deployment/web
```

## Performance Optimization

### 1. Use Appropriate Timeouts

**Good** - Set reasonable timeouts:
```yaml
# config/app.yaml
timeout: 10m  # 10 minutes for most apps
```

**Avoid** - Too high or too low:
```yaml
timeout: 0    # No timeout - can hang forever
timeout: 1h   # Too high - slow feedback
```

### 2. Use Parallel Deployments

**Good** - Deploy multiple apps in parallel:
```bash
# Frank handles this automatically
frank apply dev  # Deploys all dev apps in parallel
```

### 3. Use Resource Optimization

**Good** - Optimize resource usage:
```yaml
resources:
  requests:
    memory: "128Mi"  # Start small
    cpu: "50m"
  limits:
    memory: "256Mi"  # Reasonable limit
    cpu: "200m"
```

## Troubleshooting

### 1. Use Debug Logging

**Good** - Enable debug when needed:
```bash
FRANK_LOG_LEVEL=debug frank apply dev
```

### 2. Monitor Resource Status

**Good** - Check resource status:
```bash
kubectl get pods -l app.kubernetes.io/managed-by=frank
kubectl describe deployment web
```

### 3. Use Health Checks

**Good** - Implement health checks:
```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10
```

## Documentation

### 1. Document Configuration

**Good** - Document config options:
```yaml
# config/config.yaml
# Base configuration for all environments
project_code: myapp        # Project identifier
namespace: myapp          # Default namespace
timeout: 10m              # Default deployment timeout
```

### 2. Document Template Variables

**Good** - Document template variables:
```yaml
# Available template variables:
# - stack_name: Generated stack name
# - app_name: App name from config
# - version: Version from config
# - context: Kubernetes context
# - namespace: Target namespace
```

### 3. Document Deployment Process

**Good** - Document deployment steps:
```bash
# Deployment process:
# 1. Test in dev: frank apply dev
# 2. Test in staging: frank apply staging
# 3. Deploy to prod: frank apply prod
```

## Common Anti-Patterns

### 1. Don't Put Secrets in Config Files

```yaml
# DON'T DO THIS!
# config/prod/config.yaml
database_password: secret123
api_key: your-api-key
```

### 2. Don't Use Hardcoded Values

```yaml
# DON'T DO THIS!
# manifests/app-deployment.yaml
spec:
  replicas: 5  # Hardcoded value
  image: nginx:1.20  # Hardcoded image
```

### 3. Don't Skip Testing

```bash
# DON'T DO THIS!
frank apply prod --yes  # Without testing first
```

### 4. Don't Ignore Resource Limits

```yaml
# DON'T DO THIS!
# No resources section - can cause cluster issues
```

## Next Steps

- [Multi-Environment Setup](multi-environment.md) - Configure multiple environments
- [CI/CD Integration](ci-cd.md) - Set up automated deployments
- [Debugging](debugging.md) - Troubleshoot deployment issues
