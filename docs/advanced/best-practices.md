# Best Practices

This guide covers best practices for using **frank** effectively in production environments.

## Configuration

### 1. Hierarchical Structure

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
manifest: app-deployment.jinja
app: web
version: 1.2.3
```

**Avoid** - Duplicating base config:
```yaml
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
image: {{ image }}:{{ version }}
replicas: {{ replicas | default(3) }}
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
      app.kubernetes.io/name: {{ app }}
  template:
    metadata:
      labels:
        app.kubernetes.io/name: {{ app }}
```

**Avoid** - Too much logic:
```yaml
spec:
  replicas: {{ replicas if replicas else 3 if environment == 'prod' else 2 if environment == 'staging' else 1 }}
```

## Deployment Strategies

### 1. Test in Dev First (Preferably locally)

**Good** - Progressive deployment:
```bash
$ frank apply dev

$ frank apply staging

$ frank apply prod
```

**Avoid** - Direct to production:
```bash
# DON'T DO THIS WITHOUT TESTING IN DEV FIRST!
$ frank apply prod
```

### 2. Use Stack Filtering

**Good** - Deploy specific applications:
```bash
$ frank apply dev/web
$ frank apply dev/api
```

**Avoid** - Deploy everything:
```bash
$ frank apply dev  # Unless you really need everything
```

### 3. Use Confirmation Prompts

**Good** - Always confirm before deploying:
```bash
$ frank apply  # Shows confirmation with scope
```

**Avoid** - Skip confirmation in development:
```bash
$ frank apply --yes  # Only in CI/CD
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

### 2. Use Rollback Strategies

**Good** - Plan for rollbacks:
```bash
# Keep previous version available
$ kubectl get deployments -l app.kubernetes.io/name=web

# Rollback if needed
$ kubectl rollout undo deployment/web
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
# frank handles this automatically
$ frank apply dev  # Deploys all dev apps in parallel
```

## Troubleshooting

### 1. Use Debug Logging

**Good** - Enable debug when needed:
```bash
$ FRANK_LOG_LEVEL=debug frank apply dev
```

### 2. Monitor Resource Status

**Good** - Check resource status:
```bash
$ kubectl get pods -l app.kubernetes.io/managed-by=frank
$ kubectl describe deployment web
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
$ frank apply prod --yes  # Without testing first
```
