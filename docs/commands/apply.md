# Apply Command

The `frank apply` command deploys Kubernetes manifests to clusters with intelligent configuration management and stack-based filtering.

## Usage

```bash
frank apply [stack] [flags]
```

## Arguments

| Argument | Description | Example |
|----------|-------------|---------|
| `stack` | Optional stack filter | `dev`, `dev/app`, `prod/api.yaml` |

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--yes` | `-y` | Skip confirmation prompt | `false` |

## Examples

### Deploy All Stacks

```bash
# Interactive deployment (with confirmation)
frank apply

# Skip confirmation
frank apply --yes
```

### Deploy Specific Stack

```bash
# Deploy specific app
frank apply app

# Deploy all dev environment stacks
frank apply dev

# Deploy specific dev app
frank apply dev/app

# Deploy specific configuration file
frank apply dev/app.yaml
```

### Stack Filtering

Frank supports flexible stack filtering:

| Filter | Matches | Example |
|--------|---------|---------|
| `app` | Exact app name | `app.yaml` |
| `dev` | All dev environment stacks | `dev/app.yaml`, `dev/api.yaml` |
| `dev/app` | Specific dev app | `dev/app.yaml` |
| `dev/app.yaml` | Specific file | `dev/app.yaml` |

## What Apply Does

The apply command performs the following operations:

1. **Configuration Discovery** - Finds and loads configuration files
2. **Stack Filtering** - Filters configurations based on the provided stack argument
3. **Template Rendering** - Renders Jinja templates with context variables
4. **Namespace Validation** - Checks for namespace conflicts
5. **Resource Application** - Creates or updates Kubernetes resources
6. **Status Monitoring** - Waits for resources to be ready
7. **Parallel Processing** - Runs multiple deployments concurrently

## Interactive Confirmation

By default, Frank shows an interactive confirmation before deploying:

```
Do you want to apply 'dev'? [y/N]
```

### Confirmation Details

The confirmation shows:
- **Scope** - What will be deployed (stack name or "all stacks")
- **Resources** - Number of configurations that will be processed
- **Context** - Kubernetes context that will be used

### Skipping Confirmation

Use the `--yes` flag to skip the confirmation prompt:

```bash
frank apply --yes
frank apply dev --yes
```

## Output Format

Frank uses a structured output format:

```
<timestamp> - <stack_name> - <operation_status>
```

### Example Output

```
2024-01-15T10:30:00Z - myapp-dev-web - Creating Deployment
2024-01-15T10:30:01Z - myapp-dev-web - Waiting for resource to be ready
2024-01-15T10:30:05Z - myapp-dev-web - Resource is ready
2024-01-15T10:30:00Z - myapp-dev-api - Creating Service
2024-01-15T10:30:01Z - myapp-dev-api - Resource is ready
```

### Status Messages

| Status | Description | Color |
|--------|-------------|-------|
| `Creating <Resource>` | Creating new resource | Yellow |
| `Updating <Resource>` | Updating existing resource | Yellow |
| `Resource is ready` | Resource is ready | Green |
| `Resource is already up to date` | No changes needed | Green |
| `Apply failed` | Error occurred | Red |

## Error Handling

### Common Errors

**Configuration Errors**
```
Error: config directory with config.yaml not found
Error: context 'my-cluster' not found in kubectl config
Error: namespace conflict: both config and manifest specify namespace
```

**Deployment Errors**
```
Error: failed to create Deployment: resource already exists
Error: failed to update Service: field is immutable
Error: timeout waiting for resource to be ready
```

**Template Errors**
```
Error: template rendering failed: variable 'replicas' not found
Error: invalid YAML after template rendering
Error: template syntax error: unexpected end of template
```

### Debugging

Enable debug logging to see detailed information:

```bash
FRANK_LOG_LEVEL=debug frank apply
```

Debug output includes:
- Configuration loading details
- Template rendering process
- Kubernetes API calls
- Resource status updates

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

## Resource Management

### Stack Annotations

Frank adds the following annotation to all managed resources:

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

### Resource Updates

Frank intelligently determines if resources need updates:

- **No changes** - Resource is already up to date
- **Changes detected** - Resource will be updated
- **New resource** - Resource will be created

## Timeout Configuration

Set deployment timeouts in your configuration:

```yaml
# config/app.yaml
manifest: app-deployment.yaml
timeout: 5m  # 5 minutes timeout
```

### Timeout Values

| Value | Description |
|-------|-------------|
| `5m` | 5 minutes |
| `30m` | 30 minutes |
| `1h` | 1 hour |
| `0` | No timeout (wait indefinitely) |

## Parallel Processing

Frank runs multiple deployments in parallel for speed:

- Each configuration is processed concurrently
- Resources within a configuration are applied sequentially
- Status monitoring happens in parallel

### Performance Tips

1. **Use appropriate timeouts** - Don't set timeouts too high
2. **Organize configurations** - Group related resources together
3. **Use stack filtering** - Deploy only what you need
4. **Monitor resource limits** - Don't overwhelm your cluster

## Best Practices

### 1. Use Stack Filtering

```bash
# Deploy specific environment
frank apply dev

# Deploy specific app
frank apply dev/web

# Deploy everything
frank apply
```

### 2. Test Before Production

```bash
# Test in dev first
frank apply dev

# Then deploy to staging
frank apply staging

# Finally deploy to production
frank apply prod
```

### 3. Use Confirmation Prompts

```bash
# Always confirm before deploying
frank apply

# Only skip confirmation in CI/CD
frank apply --yes
```

### 4. Monitor Deployments

```bash
# Watch deployment progress
frank apply dev

# Check resource status
kubectl get pods -l app.kubernetes.io/managed-by=frank
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

**"timeout waiting for resource"**
- Check if your cluster has sufficient resources
- Increase the timeout value in your config
- Check for resource conflicts

### Getting Help

- Check the [Troubleshooting Guide](../reference/troubleshooting.md)
- Open an [issue on GitHub](https://github.com/schnauzersoft/frank-cli/issues)
- Join the [discussions](https://github.com/schnauzersoft/frank-cli/discussions)
