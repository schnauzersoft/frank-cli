# Configuration

Frank uses a hierarchical configuration system that allows you to organize your environments and applications with inheritance and overrides.

## Configuration Structure

Frank looks for configuration in the following order:

1. **Environment variables** (highest priority)
2. **`.frank.yaml`** (current directory)
3. **`$HOME/.frank/config.yaml`** (user home)
4. **`/etc/frank/config.yaml`** (system-wide)

## Directory Structure

Frank expects a `config/` directory in your project root:

```
my-project/
├── config/
│   ├── config.yaml              # Base configuration
│   ├── dev/
│   │   ├── config.yaml          # Dev environment overrides
│   │   ├── app.yaml             # App-specific config
│   │   └── api.yaml             # Another app
│   ├── staging/
│   │   ├── config.yaml          # Staging overrides
│   │   └── web/
│   │       ├── config.yaml      # Web-specific staging config
│   │       └── frontend.yaml    # Frontend app config
│   └── prod/
│       ├── config.yaml          # Production overrides
│       └── api.yaml             # Production API config
└── manifests/
    ├── app-deployment.yaml      # Static manifests
    ├── api-deployment.jinja     # Jinja templates
    └── web-deployment.jinja
```

## Base Configuration (`config.yaml`)

The base configuration file defines global settings:

```yaml
# config/config.yaml
context: my-cluster              # Required: Kubernetes context
project_code: myapp             # Optional: Project identifier
namespace: myapp-namespace      # Optional: Default namespace
```

### Required Fields

| Field | Description | Example |
|-------|-------------|---------|
| `context` | Kubernetes context name | `my-cluster`, `dev-cluster` |

### Optional Fields

| Field | Description | Default | Example |
|-------|-------------|---------|---------|
| `project_code` | Project identifier for stack names | `""` | `myapp`, `frank` |
| `namespace` | Default namespace for resources | `default` | `myapp-dev`, `production` |

## App Configuration (`*.yaml` files)

App-specific configuration files define individual applications:

```yaml
# config/app.yaml
manifest: app-deployment.yaml    # Required: Manifest file name
timeout: 10m                     # Optional: Deployment timeout
app: myapp                       # Optional: App name (defaults to filename)
version: 1.2.3                   # Optional: Version for templates
```

### Required Fields

| Field | Description | Example |
|-------|-------------|---------|
| `manifest` | Manifest file name (with extension) | `app-deployment.yaml`, `api-deployment.jinja` |

### Optional Fields

| Field | Description | Default | Example |
|-------|-------------|---------|---------|
| `timeout` | Deployment timeout | `10m` | `5m`, `30m`, `1h` |
| `app` | App name for templates | filename | `myapp`, `api`, `web` |
| `version` | Version for templates | `""` | `1.2.3`, `latest` |

## Configuration Inheritance

Frank uses a hierarchical inheritance system where child configurations override parent configurations:

### Example Inheritance Chain

```
config/config.yaml (base)
├── context: prod-cluster
├── project_code: myapp
└── namespace: myapp-prod

config/dev/config.yaml (dev overrides)
├── context: dev-cluster
└── namespace: myapp-dev
# project_code inherited from parent: myapp

config/dev/app.yaml (app config)
├── manifest: app-deployment.jinja
├── app: web
└── version: 1.0.0
# context inherited: dev-cluster
# project_code inherited: myapp
# namespace inherited: myapp-dev
```

### Final Configuration

For `config/dev/app.yaml`, the final configuration would be:

```yaml
context: dev-cluster              # From dev/config.yaml
project_code: myapp              # From base config.yaml
namespace: myapp-dev             # From dev/config.yaml
manifest: app-deployment.jinja   # From app.yaml
app: web                         # From app.yaml
version: 1.0.0                   # From app.yaml
```

## Environment Variables

Frank supports configuration via environment variables:

```bash
# Set log level
export FRANK_LOG_LEVEL=debug

# Set configuration file
export FRANK_CONFIG_FILE=/path/to/config.yaml
```

### Supported Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `FRANK_LOG_LEVEL` | Log level (debug, info, warn, error) | `info` |
| `FRANK_CONFIG_FILE` | Path to configuration file | Auto-detected |

## Configuration Files

### `.frank.yaml` (Current Directory)

```yaml
# .frank.yaml
log_level: debug
```

### `$HOME/.frank/config.yaml` (User Home)

```yaml
# ~/.frank/config.yaml
log_level: info
```

### `/etc/frank/config.yaml` (System-wide)

```yaml
# /etc/frank/config.yaml
log_level: warn
```

## Stack Name Generation

Frank automatically generates stack names using this pattern:

```
{project_code}-{context}-{app_name}
```

### Examples

| Project Code | Context | App Name | Stack Name |
|--------------|---------|----------|------------|
| `myapp` | `dev` | `web` | `myapp-dev-web` |
| `frank` | `prod` | `api` | `frank-prod-api` |
| `test` | `staging` | `frontend` | `test-staging-frontend` |

### Edge Cases

| Project Code | Context | App Name | Stack Name |
|--------------|---------|----------|------------|
| `""` | `dev` | `web` | `dev-web` |
| `myapp` | `""` | `web` | `myapp-web` |
| `myapp` | `dev` | `""` | `myapp-dev` |

## Namespace Management

Frank provides intelligent namespace management with conflict detection:

### Namespace Sources

1. **Configuration files** - `namespace` field in any `config.yaml`
2. **Manifest files** - `metadata.namespace` in Kubernetes manifests
3. **Templates** - `{{ namespace }}` in Jinja templates

### Conflict Detection

Frank will error if both configuration and manifest specify namespaces:

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

### Resolution

Choose one approach:

**Option 1: Configuration namespace (recommended)**
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

**Option 2: Manifest namespace**
```yaml
# config/config.yaml
# No namespace field

# manifests/app-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  namespace: myapp-prod
```

## Template Variables

Configuration values become available in Jinja templates:

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

Available in template as:
- `{{ app_name }}` → `web`
- `{{ version }}` → `1.2.3`
- `{{ replicas }}` → `5`
- `{{ image_name }}` → `nginx`
- `{{ port }}` → `8080`
- `{{ environment }}` → `production`

## Configuration Validation

Frank validates configuration at startup:

### Required Fields
- `context` must be specified in base config
- `manifest` must be specified in app configs

### Field Validation
- `timeout` must be a valid duration (e.g., `5m`, `1h`)
- `context` must exist in kubectl config
- `namespace` conflicts are detected

### Error Messages

```
Error: context 'my-cluster' not found in kubectl config
Error: namespace conflict: both config and manifest specify namespace
Error: invalid timeout '5x': time: unknown unit "x" in duration "5x"
```

## Best Practices

### 1. Use Hierarchical Structure

```
config/
├── config.yaml              # Base: context, project_code
├── dev/
│   ├── config.yaml          # Override: context=dev
│   └── app.yaml             # App config
└── prod/
    ├── config.yaml          # Override: context=prod
    └── app.yaml             # App config
```

### 2. Keep App Configs Simple

```yaml
# Good - minimal app config
manifest: app-deployment.jinja
app: web
version: 1.2.3

# Avoid - duplicating base config
manifest: app-deployment.jinja
context: dev-cluster
project_code: myapp
namespace: myapp-dev
app: web
version: 1.2.3
```

### 3. Use Environment Variables for Secrets

```bash
# Don't put secrets in config files
export DOCKER_REGISTRY_PASSWORD=secret123
```

### 4. Validate Before Deploying

```bash
# Check configuration
frank apply --dry-run  # If implemented

# Check what would be deployed
frank apply  # Shows confirmation with scope
```

## Troubleshooting

### Common Issues

**"config directory with config.yaml not found"**
- Ensure you're in a directory with a `config/` subdirectory
- Check that `config/config.yaml` exists

**"context not found"**
- Verify your Kubernetes context: `kubectl config current-context`
- Update the `context` field in your config

**"namespace conflict"**
- Choose either config namespace or manifest namespace
- Don't specify both

**"invalid timeout"**
- Use valid duration format: `5m`, `1h`, `30s`
- Check for typos in timeout values

### Debug Configuration

```bash
# Enable debug logging to see configuration loading
FRANK_LOG_LEVEL=debug frank apply
```

This will show you:
- Which configuration files are loaded
- Final configuration values
- Inheritance chain
