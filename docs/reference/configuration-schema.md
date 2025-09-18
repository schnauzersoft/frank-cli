# Configuration Schema

This document describes the complete configuration schema for Frank CLI, including all available fields, types, and validation rules.

## Base Configuration Schema

### `config/config.yaml`

The base configuration file defines global settings for all environments and applications.

```yaml
# Required fields
context: string              # Kubernetes context name

# Optional fields
project_code: string         # Project identifier for stack names
namespace: string           # Default namespace for resources
timeout: duration           # Default deployment timeout
```

#### Field Descriptions

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | Yes | - | Kubernetes context name |
| `project_code` | string | No | `""` | Project identifier for stack names |
| `namespace` | string | No | `default` | Default namespace for resources |
| `timeout` | duration | No | `10m` | Default deployment timeout |

#### Examples

```yaml
# Minimal configuration
context: my-cluster

# Full configuration
context: my-cluster
project_code: myapp
namespace: myapp-namespace
timeout: 15m
```

## Application Configuration Schema

### `config/*.yaml` files

Application-specific configuration files define individual applications.

```yaml
# Required fields
manifest: string            # Manifest file name

# Optional fields
timeout: duration           # Deployment timeout
app: string                # App name for templates
version: string            # Version for templates
```

#### Field Descriptions

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `manifest` | string | Yes | - | Manifest file name (with extension) |
| `timeout` | duration | No | `10m` | Deployment timeout |
| `app` | string | No | filename | App name for templates |
| `version` | string | No | `""` | Version for templates |

#### Examples

```yaml
# Minimal app configuration
manifest: app-deployment.yaml

# Full app configuration
manifest: app-deployment.jinja
timeout: 5m
app: web
version: 1.2.3
```

## Environment Variables Schema

Frank supports configuration via environment variables.

### Supported Variables

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `FRANK_LOG_LEVEL` | string | `info` | Log level (debug, info, warn, error) |
| `FRANK_CONFIG_FILE` | string | auto-detected | Path to configuration file |

### Environment Variable Precedence

1. **Environment variables** (highest priority)
2. **`.frank.yaml`** (current directory)
3. **`$HOME/.frank/config.yaml`** (user home)
4. **`/etc/frank/config.yaml`** (system-wide)

## Data Types

### String

Text values that can contain any characters.

```yaml
context: my-cluster
project_code: myapp
namespace: myapp-dev
```

### Duration

Time duration values in Go duration format.

```yaml
timeout: 5m        # 5 minutes
timeout: 30s       # 30 seconds
timeout: 1h        # 1 hour
timeout: 0         # No timeout (wait indefinitely)
```

**Valid duration formats:**
- `5m` - 5 minutes
- `30s` - 30 seconds
- `1h` - 1 hour
- `1h30m` - 1 hour 30 minutes
- `0` - No timeout

### Boolean

True/false values.

```yaml
enabled: true
disabled: false
```

## Validation Rules

### Required Fields

#### Base Configuration

- `context` - Must be specified in base config

#### Application Configuration

- `manifest` - Must be specified in app configs

### Field Validation

#### Context Validation

- Must exist in kubectl configuration
- Checked with `kubectl config get-contexts`

#### Timeout Validation

- Must be a valid duration format
- Examples: `5m`, `30s`, `1h`, `0`

#### Namespace Validation

- Must be a valid Kubernetes namespace name
- Cannot be empty string
- Checked for conflicts between config and manifest

### YAML Syntax Validation

- Must be valid YAML syntax
- Proper indentation required
- No duplicate keys allowed

## Configuration Inheritance

Frank uses hierarchical configuration inheritance where child configurations override parent configurations.

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
timeout: 10m

# Dev config (overrides namespace)
context: dev-cluster
namespace: myapp-dev

# App config (inherits from dev)
manifest: app-deployment.jinja
app: web
version: 1.2.3
```

### Final Configuration

For `config/dev/app.yaml`:
```yaml
context: dev-cluster         # From dev config
project_code: myapp          # From base config
namespace: myapp-dev         # From dev config (overrides base)
timeout: 10m                 # From base config
manifest: app-deployment.jinja # From app config
app: web                     # From app config
version: 1.2.3               # From app config
```

## Stack Name Generation

Frank generates stack names using this pattern:

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

## Template Context Variables

Configuration values become available in Jinja templates.

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

## Error Messages

### Configuration Errors

```
Error: context 'my-cluster' not found in kubectl config
Error: namespace conflict: both config and manifest specify namespace
Error: invalid timeout '5x': time: unknown unit "x" in duration "5x"
Error: missing required field: context
Error: missing required field: manifest
```

### Validation Errors

```
Error: invalid YAML syntax: line 3, column 5
Error: duplicate key: context
Error: invalid duration: must be a valid duration format
Error: invalid namespace: namespace cannot be empty
```

## Best Practices

### 1. Use Hierarchical Structure

```
config/
├── config.yaml              # Base: project_code, timeout
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
export API_KEY=your-api-key
```

### 4. Validate Configuration

```bash
# Always validate configuration
frank configuration validate

# Check specific environment
frank configuration validate dev

# Check specific application
frank configuration validate dev/app
```

## Next Steps

- [Template Variables](template-variables.md) - Learn about template context
- [Kubernetes Labels](kubernetes-labels.md) - Understand label management
- [Troubleshooting](troubleshooting.md) - Debug configuration issues
