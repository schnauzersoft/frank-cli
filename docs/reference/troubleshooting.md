# Troubleshooting

This guide helps you diagnose and resolve common issues with Frank CLI.

## Common Issues

### Configuration Issues

#### "config directory with config.yaml not found"

**Problem**: Frank can't find the configuration directory.

**Solutions**:
1. Ensure you're in a directory with a `config/` subdirectory
2. Check that `config/config.yaml` exists
3. Verify the file permissions

```bash
# Check directory structure
ls -la config/
# Should show config.yaml

# Check file exists
ls -la config/config.yaml
```

#### "context not found"

**Problem**: The Kubernetes context specified in your config doesn't exist.

**Solutions**:
1. Check available contexts: `kubectl config get-contexts`
2. Update the `context` field in your config
3. Switch to the correct context: `kubectl config use-context <context-name>`

```bash
# List available contexts
kubectl config get-contexts

# Switch context
kubectl config use-context my-cluster

# Update config
# config/config.yaml
context: my-cluster
```

#### "namespace conflict"

**Problem**: Both configuration and manifest specify namespaces.

**Solutions**:
1. Choose either config namespace or manifest namespace
2. Don't specify both

```yaml
# Option 1: Use config namespace (recommended)
# config/config.yaml
namespace: myapp-dev

# manifests/app-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  # No namespace - uses config namespace

# Option 2: Use manifest namespace
# config/config.yaml
# No namespace field

# manifests/app-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  namespace: myapp-prod
```

### Deployment Issues

#### "failed to create Deployment: resource already exists"

**Problem**: A resource with the same name already exists.

**Solutions**:
1. Use `kubectl delete` to remove the existing resource
2. Update the resource name in your manifest
3. Use a different namespace

```bash
# Delete existing resource
kubectl delete deployment myapp

# Or delete by label
kubectl delete deployment -l app.kubernetes.io/managed-by=frank
```

#### "failed to update Service: field is immutable"

**Problem**: Trying to update an immutable field in a Service.

**Solutions**:
1. Delete and recreate the Service
2. Use a different Service name
3. Check which fields are immutable

```bash
# Delete and recreate
kubectl delete service myapp-service
frank apply
```

#### "timeout waiting for resource to be ready"

**Problem**: Resource didn't become ready within the timeout period.

**Solutions**:
1. Check if your cluster has sufficient resources
2. Increase the timeout value in your config
3. Check for resource conflicts
4. Verify the resource configuration

```yaml
# Increase timeout
# config/app.yaml
timeout: 30m  # Increase from default 10m
```

```bash
# Check cluster resources
kubectl top nodes
kubectl describe nodes

# Check resource status
kubectl describe deployment myapp
kubectl logs deployment/myapp
```

### Template Issues

#### "template rendering failed: variable 'replicas' not found"

**Problem**: Template variable is not defined in configuration.

**Solutions**:
1. Add the variable to your config file
2. Use a default value in the template
3. Check variable name spelling

```yaml
# Add variable to config
# config/app.yaml
replicas: 5

# Or use default in template
# manifests/app-deployment.jinja
replicas: {{ replicas | default(3) }}
```

#### "invalid YAML after template rendering"

**Problem**: Template rendered invalid YAML.

**Solutions**:
1. Check template syntax
2. Use `| tojson` filter for complex values
3. Be careful with conditional blocks and indentation

```yaml
# Use tojson filter for complex values
# manifests/app-deployment.jinja
env:
{% for key, value in environment_vars.items() %}
- name: {{ key }}
  value: {{ value | tojson }}
{% endfor %}
```

#### "template syntax error: unexpected end of template"

**Problem**: Jinja template has syntax errors.

**Solutions**:
1. Check for unmatched `{% %}` or `{{ }}` blocks
2. Validate Jinja syntax
3. Use a Jinja template validator

```yaml
# Check for unmatched blocks
# Good
{% if condition %}
  content
{% endif %}

# Bad
{% if condition %}
  content
{% end %}  # Should be {% endif %}
```

### Permission Issues

#### "permission denied"

**Problem**: Insufficient permissions to access Kubernetes resources.

**Solutions**:
1. Check RBAC settings for your user/service account
2. Ensure you have the necessary permissions
3. Use a different context with appropriate permissions

```bash
# Check current user
kubectl auth whoami

# Check permissions
kubectl auth can-i create deployments
kubectl auth can-i update deployments
kubectl auth can-i delete deployments

# Check RBAC
kubectl get clusterroles
kubectl get clusterrolebindings
```

#### "context not found"

**Problem**: Kubernetes context doesn't exist or is not accessible.

**Solutions**:
1. Check available contexts: `kubectl config get-contexts`
2. Verify context configuration
3. Reconfigure kubectl if necessary

```bash
# List contexts
kubectl config get-contexts

# Check current context
kubectl config current-context

# Switch context
kubectl config use-context my-cluster
```

## Debugging

### Enable Debug Logging

```bash
# Enable debug logging
FRANK_LOG_LEVEL=debug frank apply

# Or set environment variable
export FRANK_LOG_LEVEL=debug
frank apply
```

### Debug Output

Debug logging shows:
- Configuration loading details
- Template rendering process
- Kubernetes API calls
- Resource status updates
- Stack name generation
- Namespace resolution

### Common Debug Messages

```
DEBUG: Loading config from config/config.yaml
DEBUG: Generated stack info: {stack_name: myapp-dev-web, app_name: web, version: 1.2.3}
DEBUG: Rendering template: manifests/app-deployment.jinja
DEBUG: Template rendered successfully
DEBUG: Applying resource: Deployment myapp-dev-web
DEBUG: Resource status: Creating
```

### Verbose Output

For even more detailed output:

```bash
# Enable verbose logging
FRANK_LOG_LEVEL=debug frank apply 2>&1 | tee frank-debug.log
```

## Performance Issues

### Slow Deployments

**Problem**: Deployments are taking too long.

**Solutions**:
1. Check cluster resources
2. Use appropriate timeouts
3. Deploy fewer resources at once
4. Check network connectivity

```bash
# Check cluster resources
kubectl top nodes
kubectl describe nodes

# Check resource usage
kubectl get pods --all-namespaces
kubectl top pods --all-namespaces
```

### Memory Issues

**Problem**: Frank is using too much memory.

**Solutions**:
1. Deploy fewer resources at once
2. Use stack filtering
3. Check for memory leaks in templates

```bash
# Deploy specific stack
frank apply dev/web

# Instead of deploying everything
frank apply
```

## Getting Help

### Check Logs

```bash
# Enable debug logging
FRANK_LOG_LEVEL=debug frank apply

# Save logs to file
FRANK_LOG_LEVEL=debug frank apply 2>&1 | tee frank-debug.log
```

### Verify Configuration

```bash
# Check kubectl configuration
kubectl config view
kubectl config current-context

# Test kubectl connectivity
kubectl get nodes
kubectl get namespaces
```

### Test Templates

```bash
# Test template rendering
FRANK_LOG_LEVEL=debug frank apply dev/web

# Check rendered output
# Look for "Template rendered successfully" in debug output
```

### Community Support

- **GitHub Issues**: [Report bugs or request features](https://github.com/schnauzersoft/frank-cli/issues)
- **GitHub Discussions**: [Ask questions and get help](https://github.com/schnauzersoft/frank-cli/discussions)
- **Documentation**: Check the [full documentation](../index.md)

### Contributing

If you find a bug or have a feature request:

1. Check if the issue already exists
2. Create a new issue with:
   - Description of the problem
   - Steps to reproduce
   - Expected behavior
   - Actual behavior
   - Environment details (OS, Go version, Kubernetes version)
   - Debug logs (if applicable)

### Reporting Issues

When reporting issues, include:

```bash
# System information
go version
kubectl version
kubectl config current-context

# Frank version
./frank --version

# Debug logs
FRANK_LOG_LEVEL=debug frank apply 2>&1 | tee frank-debug.log
```

## Prevention

### Best Practices

1. **Test in dev first** - Always test changes in development before production
2. **Use stack filtering** - Deploy only what you need
3. **Monitor resources** - Keep an eye on cluster resources
4. **Use appropriate timeouts** - Don't set timeouts too high or too low
5. **Validate templates** - Test templates before deploying
6. **Keep configurations simple** - Avoid complex inheritance chains
7. **Use version control** - Track all configuration changes

### Regular Maintenance

1. **Clean up resources** - Regularly clean up unused resources
2. **Update dependencies** - Keep Frank and dependencies updated
3. **Monitor logs** - Check logs for warnings and errors
4. **Test deployments** - Regularly test deployment processes
5. **Backup configurations** - Keep backups of important configurations
