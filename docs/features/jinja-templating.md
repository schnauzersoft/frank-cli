# Jinja Templating

Frank supports powerful Jinja templating for dynamic Kubernetes manifest generation. This allows you to create flexible, reusable templates that adapt to different environments and configurations.

## Supported File Extensions

Frank automatically detects and processes Jinja templates with these extensions:

- `.jinja` - Standard Jinja template files
- `.j2` - Alternative Jinja extension

## Basic Template Example

Here's a simple example of a Jinja template:

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
        image: {{ image_name }}:{{ version }}
        ports:
        - containerPort: {{ port | default(80) }}
```

## Template Context Variables

Frank provides a rich context for your templates with the following variables:

### Core Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `stack_name` | Generated stack name | `myapp-dev-web` |
| `app_name` | App name from config or filename | `web` |
| `version` | Version from config | `1.2.3` |
| `project_code` | Project identifier | `myapp` |
| `context` | Kubernetes context | `dev-cluster` |
| `namespace` | Target namespace | `myapp-dev` |
| `k8s_namespace` | Alias for namespace | `myapp-dev` |

### Configuration Variables

Any key in your configuration files becomes available as a template variable:

```yaml
# config/app.yaml
manifest: app-deployment.jinja
app: web
version: 1.2.3
replicas: 5
image_name: nginx
port: 8080
environment: production
```

These become available in your template as:
- `{{ replicas }}` → `5`
- `{{ image_name }}` → `nginx`
- `{{ port }}` → `8080`
- `{{ environment }}` → `production`

## Jinja Filters and Functions

Frank supports all standard Jinja filters and functions:

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

## Multi-Document YAML

Templates can contain multiple Kubernetes resources separated by `---`:

```yaml
# manifests/full-stack.jinja
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
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: {{ stack_name }}-ingress
  labels:
    app.kubernetes.io/name: {{ app_name }}
    app.kubernetes.io/version: {{ version }}
    app.kubernetes.io/managed-by: frank
spec:
  rules:
  - host: {{ app_name }}.{{ environment | default('dev') }}.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: {{ stack_name }}-service
            port:
              number: 80
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

## Template Debugging

### Enable Debug Logging

```bash
FRANK_LOG_LEVEL=debug frank apply
```

This will show you the rendered template content before applying it to Kubernetes.

### Common Template Issues

**Variable not found**
- Check that the variable is defined in your config file
- Ensure the variable name matches exactly (case-sensitive)

**Invalid YAML after rendering**
- Use Jinja's `| tojson` filter for complex values
- Be careful with conditional blocks and indentation

**Template syntax errors**
- Validate your Jinja syntax
- Check for unmatched `{% %}` or `{{ }}` blocks

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

# Avoid - too much logic in one place
spec:
  replicas: {{ replicas if replicas else 3 if environment == 'prod' else 2 }}
```

### 5. Test Your Templates

```bash
# Test with different configurations
frank apply dev
frank apply prod
frank apply staging
```

## Template Examples

Check out the [examples directory](https://github.com/schnauzersoft/frank-cli/tree/main/examples) for more template patterns and real-world use cases.
