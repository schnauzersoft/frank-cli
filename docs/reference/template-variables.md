# Template Variables

This document describes all template variables available in Jinja templates when using Frank CLI.

## Core Variables

These variables are always available in all templates.

### Stack Information

| Variable | Description | Example |
|----------|-------------|---------|
| `stack_name` | Generated stack name | `myapp-dev-web` |
| `app_name` | App name from config or filename | `web` |
| `version` | Version from config | `1.2.3` |

### Project Information

| Variable | Description | Example |
|----------|-------------|---------|
| `project_code` | Project identifier | `myapp` |
| `context` | Kubernetes context | `dev-cluster` |

### Namespace Information

| Variable | Description | Example |
|----------|-------------|---------|
| `namespace` | Target namespace | `myapp-dev` |
| `k8s_namespace` | Alias for namespace | `myapp-dev` |

## Configuration Variables

Any key in your configuration files becomes available as a template variable.

### Base Configuration Variables

From `config/config.yaml`:

```yaml
# config/config.yaml
project_code: myapp
namespace: myapp
timeout: 10m
replicas: 3
image_name: nginx
port: 8080
```

Available in templates as:
- `{{ project_code }}` → `myapp`
- `{{ namespace }}` → `myapp`
- `{{ timeout }}` → `10m`
- `{{ replicas }}` → `3`
- `{{ image_name }}` → `nginx`
- `{{ port }}` → `8080`

### Environment Configuration Variables

From `config/dev/config.yaml`:

```yaml
# config/dev/config.yaml
context: dev-cluster
namespace: myapp-dev
replicas: 2
image_tag: latest
environment: development
```

Available in templates as:
- `{{ context }}` → `dev-cluster`
- `{{ namespace }}` → `myapp-dev`
- `{{ replicas }}` → `2`
- `{{ image_tag }}` → `latest`
- `{{ environment }}` → `development`

### Application Configuration Variables

From `config/dev/app.yaml`:

```yaml
# config/dev/app.yaml
manifest: app-deployment.jinja
app: web
version: 1.2.3
replicas: 5
image_name: myapp/web
port: 8080
environment: production
```

Available in templates as:
- `{{ app_name }}` → `web`
- `{{ version }}` → `1.2.3`
- `{{ replicas }}` → `5`
- `{{ image_name }}` → `myapp/web`
- `{{ port }}` → `8080`
- `{{ environment }}` → `production`

## Variable Inheritance

Variables are inherited through the configuration hierarchy:

### Inheritance Order

1. **Base configuration** - `config/config.yaml`
2. **Environment configuration** - `config/dev/config.yaml`
3. **Application configuration** - `config/dev/app.yaml`

### Override Behavior

Child configurations override parent configurations:

```yaml
# Base config
project_code: myapp
namespace: myapp
replicas: 3

# Dev config (overrides namespace and replicas)
context: dev-cluster
namespace: myapp-dev
replicas: 2

# App config (inherits from dev)
manifest: app-deployment.jinja
app: web
version: 1.2.3
```

### Final Variables

For `config/dev/app.yaml`, the final variables would be:

```yaml
project_code: myapp          # From base config
context: dev-cluster         # From dev config
namespace: myapp-dev         # From dev config (overrides base)
replicas: 2                  # From dev config (overrides base)
manifest: app-deployment.jinja # From app config
app_name: web                # From app config
version: 1.2.3               # From app config
```

## Template Examples

### Basic Deployment

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
```

### Service with Dynamic Ports

```yaml
# manifests/app-service.jinja
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
    protocol: TCP
  type: {{ service_type | default('ClusterIP') }}
```

### ConfigMap with Environment Variables

```yaml
# manifests/app-configmap.jinja
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ stack_name }}-config
  labels:
    app.kubernetes.io/name: {{ app_name }}
    app.kubernetes.io/version: {{ version }}
    app.kubernetes.io/managed-by: frank
data:
  app.properties: |
    server.port={{ port | default(8080) }}
    app.name={{ app_name }}
    app.version={{ version }}
    environment={{ environment | default('development') }}
    project.code={{ project_code }}
```

## Jinja Filters

Frank supports all standard Jinja filters and functions.

### Default Values

```yaml
replicas: {{ replicas | default(3) }}
port: {{ port | default(80) }}
image: {{ image_name | default('nginx') }}:{{ version | default('latest') }}
```

### String Operations

```yaml
name: {{ app_name | upper }}
namespace: {{ namespace | lower }}
image: {{ image_name | replace('_', '-') }}
```

### Conditional Logic

```yaml
{% if environment == 'production' %}
replicas: 5
{% else %}
replicas: 2
{% endif %}
```

### Loops

```yaml
ports:
{% for port in ports %}
- containerPort: {{ port }}
{% endfor %}
```

## Advanced Template Patterns

### Environment-Specific Configuration

```yaml
# config/prod/app.yaml
manifest: app-deployment.jinja
app: web
version: 2.1.0
replicas: 5
environment: production
image_name: myregistry/web-app
port: 8080
```

```yaml
# manifests/app-deployment.jinja
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ stack_name }}
spec:
  replicas: {{ replicas }}
  template:
    spec:
      containers:
      - name: {{ app_name }}
        image: {{ image_name }}:{{ version }}
        {% if environment == 'production' %}
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "500m"
        {% endif %}
        ports:
        - containerPort: {{ port }}
        env:
        - name: ENVIRONMENT
          value: {{ environment }}
        - name: VERSION
          value: {{ version }}
```

### Conditional Resource Creation

```yaml
# manifests/conditional-resources.jinja
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ stack_name }}
spec:
  replicas: {{ replicas | default(3) }}
  template:
    spec:
      containers:
      - name: {{ app_name }}
        image: {{ image_name }}:{{ version }}
        ports:
        - containerPort: {{ port | default(80) }}
{% if enable_service | default(true) %}
---
apiVersion: v1
kind: Service
metadata:
  name: {{ stack_name }}-service
spec:
  selector:
    app.kubernetes.io/name: {{ app_name }}
  ports:
  - port: 80
    targetPort: {{ port | default(80) }}
{% endif %}
{% if enable_ingress | default(false) %}
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: {{ stack_name }}-ingress
spec:
  rules:
  - host: {{ app_name }}.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: {{ stack_name }}-service
            port:
              number: 80
{% endif %}
```

## Variable Debugging

### Enable Debug Logging

```bash
FRANK_LOG_LEVEL=debug frank apply
```

This will show you the template context:

```
DEBUG: Template context: {
  stack_name: myapp-dev-web,
  app_name: web,
  version: 1.2.3,
  project_code: myapp,
  context: dev-cluster,
  namespace: myapp-dev,
  replicas: 5,
  image_name: nginx,
  port: 8080
}
```

### Common Issues

**Variable not found**:
```
Error: template rendering failed: variable 'replicas' not found
```

**Solution**: Add the variable to your config file or use a default value.

**Invalid YAML after rendering**:
```
Error: invalid YAML after template rendering
```

**Solution**: Use `| tojson` filter for complex values or check template syntax.

## Best Practices

### 1. Use Meaningful Variable Names

```yaml
# Good
image: {{ app_image }}:{{ app_version }}

# Avoid
image: {{ img }}:{{ v }}
```

### 2. Provide Sensible Defaults

```yaml
# Good
replicas: {{ replicas | default(3) }}
port: {{ port | default(80) }}

# Avoid
replicas: {{ replicas }}  # Will error if not defined
```

### 3. Use Kubernetes Best Practice Labels

```yaml
metadata:
  labels:
    app.kubernetes.io/name: {{ app_name }}
    app.kubernetes.io/version: {{ version }}
    app.kubernetes.io/managed-by: frank
```

### 4. Keep Templates Readable

```yaml
# Good - clear structure
spec:
  replicas: {{ replicas | default(3) }}
  selector:
    matchLabels:
      app.kubernetes.io/name: {{ app_name }}

# Avoid - too much logic
spec:
  replicas: {{ replicas if replicas else 3 if environment == 'prod' else 2 }}
```

## Next Steps

- [Configuration Schema](configuration-schema.md) - Learn about configuration structure
- [Kubernetes Labels](kubernetes-labels.md) - Understand label management
- [Troubleshooting](troubleshooting.md) - Debug template issues
