# HCL Templating

Frank supports HCL (HashiCorp Configuration Language) templating for Kubernetes manifest generation. This provides legitimate HCL syntax that is parsed and converted to Kubernetes YAML in memory.

## Supported File Extensions

Frank automatically detects and processes HCL templates with these extensions:

- `.hcl` - Standard HCL template files
- `.tf` - Terraform-style HCL files

## Basic Template Example

Here's a simple example of an HCL template:

```hcl
# manifests/app-deployment.hcl
resource "kubernetes_deployment" "app" {
  metadata = {
    name = "${stack_name}"
    labels = {
      "app.kubernetes.io/name" = "${app}"
      "app.kubernetes.io/version" = "${version}"
      "app.kubernetes.io/managed-by" = "frank"
    }
  }

  spec = {
    replicas = ${replicas}

    selector = {
      matchLabels = {
        "app.kubernetes.io/name" = "${app}"
      }
    }

    template = {
      metadata = {
        labels = {
          "app.kubernetes.io/name" = "${app}"
          "app.kubernetes.io/version" = "${version}"
        }
      }

      spec = {
        containers = [
          {
            name  = "${app}"
            image = "${image_name}:${version}"
            ports = [
              {
                containerPort = ${port}
              }
            ]
            env = [
              {
                name  = "ENVIRONMENT"
                value = "${environment}"
              },
              {
                name  = "PROJECT_CODE"
                value = "${project_code}"
              },
              {
                name  = "NAMESPACE"
                value = "${k8s_namespace}"
              }
            ]
          }
        ]
      }
    }
  }
}
```

## Template Context Variables

HCL templates use the same context variables as Jinja templates:

### Core Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `stack_name` | Generated stack name | `myapp-dev-web` |
| `app` | App name from config or filename | `web` |
| `version` | Version from config | `1.2.3` |
| `project_code` | Project identifier | `myapp` |
| `context` | Kubernetes context | `dev-cluster` |
| `namespace` | Target namespace | `myapp-dev` |
| `k8s_namespace` | Alias for namespace | `myapp-dev` |

### Configuration Variables

Any key in your configuration files becomes available as a template variable:

```yaml
# config/app.yaml
manifest: app-deployment.hcl
app: web
version: 1.2.3
replicas: 5
image_name: nginx
port: 8080
environment: production
```

These become available in your template as:
- `${replicas}` → `5`
- `${image_name}` → `nginx`
- `${port}` → `8080`
- `${environment}` → `production`

## HCL Syntax Requirements

HCL templates use proper HCL syntax with the following requirements:

- Use `=` for attribute assignment (not blocks)
- Use `[]` for arrays/lists
- Use `{}` for objects/maps
- Field names should use camelCase (e.g., `matchLabels`, `containerPort`)
- Use `${variable_name}` syntax for variable substitution

```hcl
# Basic substitution
name = "${app_name}"
image = "${image_name}:${version}"

# Arrays and objects
containers = [
  {
    name  = "${app}"
    image = "${image_name}:${version}"
    ports = [
      {
        containerPort = ${port}
      }
    ]
  }
]
```

## Configuration Example

Here's a complete example of using HCL templating:

```yaml
# config/web-app.yaml
manifest: web-deployment.hcl
version: alpine
vars:
  replicas: 3
  image_name: nginx
  port: 8080
  environment: development
```

```hcl
# manifests/web-deployment.hcl
resource "kubernetes_deployment" "app" {
  metadata = {
    name = "${stack_name}"
    labels = {
      "app.kubernetes.io/name" = "${app}"
      "app.kubernetes.io/version" = "${version}"
      "app.kubernetes.io/managed-by" = "frank"
    }
  }

  spec = {
    replicas = ${replicas}

    selector = {
      matchLabels = {
        "app.kubernetes.io/name" = "${app}"
      }
    }

    template = {
      metadata = {
        labels = {
          "app.kubernetes.io/name" = "${app}"
          "app.kubernetes.io/version" = "${version}"
        }
      }

      spec = {
        containers = [
          {
            name  = "${app}"
            image = "${image_name}:${version}"
            ports = [
              {
                containerPort = ${port}
              }
            ]
            env = [
              {
                name  = "ENVIRONMENT"
                value = "${environment}"
              },
              {
                name  = "PROJECT_CODE"
                value = "${project_code}"
              },
              {
                name  = "NAMESPACE"
                value = "${k8s_namespace}"
              }
            ]
          }
        ]
      }
    }
  }
}
```

## Multi-Document YAML

HCL templates can contain multiple Kubernetes resources separated by `---`:

```yaml
# manifests/full-stack.hcl
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
        image: ${image_name}:${version}
        ports:
        - containerPort: ${port}
---
apiVersion: v1
kind: Service
metadata:
  name: ${stack_name}-service
  labels:
    app.kubernetes.io/name: ${app}
    app.kubernetes.io/version: ${version}
    app.kubernetes.io/managed-by: frank
spec:
  selector:
    app.kubernetes.io/name: ${app}
  ports:
  - port: 80
    targetPort: ${port}
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ${stack_name}-ingress
  labels:
    app.kubernetes.io/name: ${app}
    app.kubernetes.io/version: ${version}
    app.kubernetes.io/managed-by: frank
spec:
  rules:
  - host: ${app}.${environment}.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: ${stack_name}-service
            port:
              number: 80
```

## HCL vs Jinja Templating

| Feature | HCL | Jinja |
|---------|-----|-------|
| **Syntax** | `${variable}` | `{{ variable }}` |
| **Conditionals** | ❌ Not supported | ✅ Full support |
| **Loops** | ❌ Not supported | ✅ Full support |
| **Filters** | ❌ Limited | ✅ Full support |
| **Complex Logic** | ❌ Not supported | ✅ Full support |
| **Simplicity** | ✅ Very simple | ⚠️ More complex |
| **Use Case** | Basic substitution | Advanced templating |

## When to Use HCL

Choose HCL templating when you need:

- **Simple variable substitution** without complex logic
- **Basic configuration** that doesn't require conditionals
- **Familiar syntax** if you're used to Terraform or other HCL tools
- **Minimal complexity** for straightforward deployments

## When to Use Jinja

Choose Jinja templating when you need:

- **Conditional logic** (if/else statements)
- **Loops** for generating multiple resources
- **Advanced filters** and string manipulation
- **Complex templating** with multiple environments
- **Dynamic resource generation**

## Template Debugging

### Enable Debug Logging

```bash
FRANK_LOG_LEVEL=debug frank apply
```

This will show you the rendered template content before applying it to Kubernetes.

### Common HCL Template Issues

**Variable not found**
- Check that the variable is defined in your config file
- Ensure the variable name matches exactly (case-sensitive)
- Use the `vars` block for custom variables

**Invalid YAML after rendering**
- HCL templates must output valid YAML
- Check that all `${variable}` placeholders are properly substituted
- Validate the final YAML syntax

**Template syntax errors**
- Use `${variable}` syntax, not `{{ variable }}`
- Ensure all variables are properly closed with `}`
- Check for typos in variable names

## Best Practices

### 1. Use the `vars` Block for Custom Variables

```yaml
# config/app.yaml
manifest: app-deployment.hcl
version: 1.2.3
vars:
  replicas: 5
  image_name: nginx
  port: 8080
  environment: production
```

### 2. Provide Sensible Defaults in Config

```yaml
# config/app.yaml
manifest: app-deployment.hcl
version: alpine  # Good default
vars:
  replicas: 3    # Good default
  port: 80       # Good default
```

### 3. Use Kubernetes Best Practice Labels

```yaml
metadata:
  labels:
    app.kubernetes.io/name: ${app}
    app.kubernetes.io/version: ${version}
    app.kubernetes.io/managed-by: frank
```

### 4. Keep Templates Simple

```yaml
# Good - simple and clear
spec:
  replicas: ${replicas}
  selector:
    matchLabels:
      app.kubernetes.io/name: ${app}

# Avoid - too complex for HCL
spec:
  replicas: ${replicas if environment == 'prod' else 2}
```

### 5. Test Your Templates

```bash
# Test with different configurations
frank apply dev
frank apply prod
frank apply staging
```

## Migration from Jinja to HCL

If you have existing Jinja templates and want to convert them to HCL:

1. **Replace Jinja syntax** with HCL syntax:
   - `{{ variable }}` → `${variable}`
   - Remove `{% if %}` blocks (not supported in HCL)
   - Remove `{% for %}` loops (not supported in HCL)

2. **Move complex logic** to configuration files:
   - Use different config files for different environments
   - Use the `vars` block for environment-specific values

3. **Simplify templates** to basic variable substitution only

## Example Migration

**Before (Jinja):**
```yaml
# manifests/app-deployment.jinja
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ stack_name }}
spec:
  replicas: {% if environment == 'production' %}5{% else %}2{% endif %}
  template:
    spec:
      containers:
      - name: {{ app }}
        image: {{ image_name }}:{{ version }}
        {% if environment == 'production' %}
        resources:
          requests:
            memory: "512Mi"
        {% endif %}
```

**After (HCL):**
```yaml
# manifests/app-deployment.hcl
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${stack_name}
spec:
  replicas: ${replicas}
  template:
    spec:
      containers:
      - name: ${app}
        image: ${image_name}:${version}
        resources:
          requests:
            memory: "${memory_request}"
```

```yaml
# config/prod/app.yaml
manifest: app-deployment.hcl
version: 1.2.3
vars:
  replicas: 5
  memory_request: "512Mi"
  image_name: nginx
```

```yaml
# config/dev/app.yaml
manifest: app-deployment.hcl
version: 1.2.3
vars:
  replicas: 2
  memory_request: "256Mi"
  image_name: nginx
```

## Template Examples

Check out the [examples directory](https://github.com/schnauzersoft/frank-cli/tree/main/examples) for more HCL template patterns and real-world use cases.
