# Configuration Command

The `frank configuration` command provides utilities for managing Frank CLI configuration and settings.

## Usage

```bash
frank configuration [command] [flags]
```

## Commands

| Command | Description |
|---------|-------------|
| `validate` | Validate configuration files |
| `show` | Show current configuration |
| `init` | Initialize configuration structure |

## Configuration Validation

### Validate All Configurations

```bash
# Validate all configuration files
frank configuration validate

# Validate specific environment
frank configuration validate dev

# Validate specific application
frank configuration validate dev/app
```

### Validation Checks

Frank validates:

- **YAML syntax** - Valid YAML format
- **Required fields** - All required fields present
- **Field types** - Correct data types
- **Namespace conflicts** - No namespace conflicts
- **Template syntax** - Valid Jinja templates
- **Kubernetes context** - Context exists in kubectl

### Example Output

```
✓ config/config.yaml - Valid
✓ config/dev/config.yaml - Valid
✓ config/dev/app.yaml - Valid
✓ manifests/app-deployment.jinja - Valid
✓ All configurations are valid
```

## Show Configuration

### Show Current Configuration

```bash
# Show resolved configuration
frank configuration show

# Show specific environment
frank configuration show dev

# Show specific application
frank configuration show dev/app
```

### Configuration Details

Frank shows:

- **Base configuration** - Global settings
- **Environment overrides** - Environment-specific settings
- **Application configuration** - App-specific settings
- **Resolved values** - Final configuration after inheritance
- **Template context** - Variables available in templates

### Example Output

```
Configuration for dev/app:
  Base:
    project_code: myapp
    timeout: 10m
  Environment (dev):
    context: dev-cluster
    namespace: myapp-dev
  Application (app):
    manifest: app-deployment.jinja
    app: web
    version: 1.2.3
  Resolved:
    stack_name: myapp-dev-web
    namespace: myapp-dev
    context: dev-cluster
```

## Initialize Configuration

### Create Configuration Structure

```bash
# Initialize basic configuration
frank configuration init

# Initialize with environment structure
frank configuration init --environments dev,staging,prod

# Initialize with applications
frank configuration init --apps web,api
```

### Generated Structure

Frank creates:

```
config/
├── config.yaml              # Base configuration
├── dev/
│   ├── config.yaml          # Dev environment
│   ├── web.yaml             # Web app config
│   └── api.yaml             # API app config
├── staging/
│   ├── config.yaml          # Staging environment
│   ├── web.yaml             # Web app config
│   └── api.yaml             # API app config
└── prod/
    ├── config.yaml          # Production environment
    ├── web.yaml             # Web app config
    └── api.yaml             # API app config
```

## Configuration Files

### Base Configuration

```yaml
# config/config.yaml
project_code: myapp
namespace: myapp
timeout: 10m
```

### Environment Configuration

```yaml
# config/dev/config.yaml
context: dev-cluster
namespace: myapp-dev
replicas: 2
image_tag: latest
```

### Application Configuration

```yaml
# config/dev/web.yaml
manifest: web-deployment.jinja
app: web
version: 1.2.3
replicas: 2
image_name: myapp/web
port: 8080
```

## Configuration Inheritance

Frank uses hierarchical configuration inheritance:

### Inheritance Order

1. **Base configuration** - `config/config.yaml`
2. **Environment configuration** - `config/dev/config.yaml`
3. **Application configuration** - `config/dev/web.yaml`

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

# Web config (inherits from dev)
manifest: web-deployment.jinja
app: web
version: 1.2.3
```

### Final Configuration

For `config/dev/web.yaml`:
```yaml
project_code: myapp          # From base
context: dev-cluster         # From dev
namespace: myapp-dev         # From dev (overrides base)
timeout: 10m                 # From base
manifest: web-deployment.jinja # From web
app: web                     # From web
version: 1.2.3               # From web
```

## Environment Variables

Frank supports configuration via environment variables:

### Supported Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `FRANK_LOG_LEVEL` | Log level (debug, info, warn, error) | `info` |
| `FRANK_CONFIG_FILE` | Path to configuration file | Auto-detected |

### Environment Variable Precedence

1. **Environment variables** (highest priority)
2. **`.frank.yaml`** (current directory)
3. **`$HOME/.frank/config.yaml`** (user home)
4. **`/etc/frank/config.yaml`** (system-wide)

## Configuration Validation

### YAML Syntax Validation

Frank validates YAML syntax:

```bash
frank configuration validate
# ✓ config/config.yaml - Valid YAML
# ✗ config/invalid.yaml - Invalid YAML: line 3, column 5
```

### Required Fields Validation

Frank checks for required fields:

```bash
frank configuration validate
# ✗ config/app.yaml - Missing required field: manifest
# ✗ config/config.yaml - Missing required field: context
```

### Field Type Validation

Frank validates field types:

```bash
frank configuration validate
# ✗ config/app.yaml - Invalid timeout: must be a duration
# ✗ config/config.yaml - Invalid replicas: must be a number
```

### Namespace Conflict Validation

Frank detects namespace conflicts:

```bash
frank configuration validate
# ✗ config/app.yaml - Namespace conflict: both config and manifest specify namespace
```

## Template Validation

### Jinja Template Validation

Frank validates Jinja templates:

```bash
frank configuration validate
# ✓ manifests/app-deployment.jinja - Valid Jinja template
# ✗ manifests/invalid.jinja - Invalid Jinja syntax: unexpected end of template
```

### Template Context Validation

Frank validates template context:

```bash
frank configuration validate
# ✗ config/app.yaml - Template variable 'replicas' not defined
# ✗ config/app.yaml - Template variable 'image_name' not defined
```

## Configuration Management

### Configuration Files

Frank looks for configuration in this order:

1. **`.frank.yaml`** (current directory)
2. **`$HOME/.frank/config.yaml`** (user home)
3. **`/etc/frank/config.yaml`** (system-wide)

### Configuration Structure

Frank expects this directory structure:

```
my-project/
├── config/
│   ├── config.yaml          # Base configuration
│   ├── dev/
│   │   ├── config.yaml      # Dev environment
│   │   └── app.yaml         # App configuration
│   └── prod/
│       ├── config.yaml      # Prod environment
│       └── app.yaml         # App configuration
└── manifests/
    └── app-deployment.jinja # Kubernetes manifests
```

## Troubleshooting

### Common Issues

**"config directory with config.yaml not found"**
- Ensure you're in a directory with a `config/` subdirectory
- Check that `config/config.yaml` exists
- Verify file permissions

**"context not found"**
- Check available contexts: `kubectl config get-contexts`
- Update the `context` field in your config
- Switch to the correct context

**"namespace conflict"**
- Choose either config namespace or manifest namespace
- Don't specify both

**"template rendering failed"**
- Check Jinja template syntax
- Verify all template variables are defined
- Use debug logging to see template context

### Debug Commands

```bash
# Enable debug logging
FRANK_LOG_LEVEL=debug frank configuration show

# Validate with debug output
FRANK_LOG_LEVEL=debug frank configuration validate

# Show configuration with debug info
FRANK_LOG_LEVEL=debug frank configuration show dev
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

### 4. Validate Before Deploying

```bash
# Always validate configuration
frank configuration validate

# Then deploy
frank apply
```

## Next Steps

- [Apply Command](apply.md) - Learn about deploying resources
- [Delete Command](delete.md) - Learn about cleaning up resources
- [Configuration Guide](../getting-started/configuration.md) - Detailed configuration documentation
