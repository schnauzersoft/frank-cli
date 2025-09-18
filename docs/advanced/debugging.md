# Debugging

This guide helps you debug issues with Frank CLI deployments and troubleshoot common problems.

## Debug Logging

### Enable Debug Logging

```bash
# Set environment variable
export FRANK_LOG_LEVEL=debug

# Or run with debug logging
FRANK_LOG_LEVEL=debug frank apply

# Or for specific commands
FRANK_LOG_LEVEL=debug frank apply dev
FRANK_LOG_LEVEL=debug frank delete prod
```

### Debug Output

Debug logging shows detailed information about:

- Configuration loading and inheritance
- Template rendering process
- Kubernetes API calls
- Resource status updates
- Stack name generation
- Namespace resolution

### Example Debug Output

```
DEBUG: Loading config from config/config.yaml
DEBUG: Generated stack info: {stack_name: myapp-dev-web, app_name: web, version: 1.2.3}
DEBUG: Rendering template: manifests/app-deployment.jinja
DEBUG: Template rendered successfully
DEBUG: Applying resource: Deployment myapp-dev-web
DEBUG: Resource status: Creating
DEBUG: Waiting for resource to be ready
DEBUG: Resource is ready
```

## Common Issues and Solutions

### Configuration Issues

#### "config directory with config.yaml not found"

**Problem**: Frank can't find the configuration directory.

**Debug Steps**:
```bash
# Check current directory
pwd

# Check if config directory exists
ls -la config/

# Check if config.yaml exists
ls -la config/config.yaml

# Check file permissions
ls -la config/
```

**Solutions**:
1. Ensure you're in a directory with a `config/` subdirectory
2. Check that `config/config.yaml` exists
3. Verify file permissions

#### "context not found"

**Problem**: The Kubernetes context specified in your config doesn't exist.

**Debug Steps**:
```bash
# Check available contexts
kubectl config get-contexts

# Check current context
kubectl config current-context

# Check context configuration
kubectl config view
```

**Solutions**:
1. Verify the context name in your config
2. Switch to the correct context: `kubectl config use-context <context-name>`
3. Update the context in your config file

#### "namespace conflict"

**Problem**: Both configuration and manifest specify namespaces.

**Debug Steps**:
```bash
# Check config namespace
grep -r "namespace:" config/

# Check manifest namespace
grep -r "namespace:" manifests/

# Enable debug logging to see namespace resolution
FRANK_LOG_LEVEL=debug frank apply
```

**Solutions**:
1. Choose either config namespace or manifest namespace
2. Don't specify both

### Template Issues

#### "template rendering failed: variable 'replicas' not found"

**Problem**: Template variable is not defined in configuration.

**Debug Steps**:
```bash
# Check template variables
FRANK_LOG_LEVEL=debug frank apply

# Look for "Template context:" in debug output
# Check if all required variables are defined
```

**Solutions**:
1. Add the variable to your config file
2. Use a default value in the template: `{{ replicas | default(3) }}`
3. Check variable name spelling

#### "invalid YAML after template rendering"

**Problem**: Template rendered invalid YAML.

**Debug Steps**:
```bash
# Enable debug logging to see rendered template
FRANK_LOG_LEVEL=debug frank apply

# Look for "Template rendered:" in debug output
# Check the rendered YAML for syntax errors
```

**Solutions**:
1. Check template syntax
2. Use `| tojson` filter for complex values
3. Be careful with conditional blocks and indentation

#### "template syntax error: unexpected end of template"

**Problem**: Jinja template has syntax errors.

**Debug Steps**:
```bash
# Check template syntax
# Look for unmatched {% %} or {{ }} blocks
```

**Solutions**:
1. Check for unmatched blocks
2. Validate Jinja syntax
3. Use a Jinja template validator

### Kubernetes Issues

#### "failed to create Deployment: resource already exists"

**Problem**: A resource with the same name already exists.

**Debug Steps**:
```bash
# Check existing resources
kubectl get deployments
kubectl get services
kubectl get configmaps

# Check for frank-managed resources
kubectl get all -l app.kubernetes.io/managed-by=frank
```

**Solutions**:
1. Delete existing resource: `kubectl delete deployment <name>`
2. Update the resource name in your manifest
3. Use a different namespace

#### "failed to update Service: field is immutable"

**Problem**: Trying to update an immutable field in a Service.

**Debug Steps**:
```bash
# Check current service configuration
kubectl get service <name> -o yaml

# Check what changes are being made
FRANK_LOG_LEVEL=debug frank apply
```

**Solutions**:
1. Delete and recreate the Service
2. Use a different Service name
3. Check which fields are immutable

#### "timeout waiting for resource to be ready"

**Problem**: Resource didn't become ready within the timeout period.

**Debug Steps**:
```bash
# Check resource status
kubectl describe deployment <name>
kubectl get pods -l app=<name>

# Check pod logs
kubectl logs deployment/<name>

# Check cluster resources
kubectl top nodes
kubectl describe nodes
```

**Solutions**:
1. Check if your cluster has sufficient resources
2. Increase the timeout value in your config
3. Check for resource conflicts
4. Verify the resource configuration

### Permission Issues

#### "permission denied"

**Problem**: Insufficient permissions to access Kubernetes resources.

**Debug Steps**:
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

**Solutions**:
1. Check RBAC settings for your user/service account
2. Ensure you have the necessary permissions
3. Use a different context with appropriate permissions

## Debugging Tools

### kubectl Commands

```bash
# Check cluster status
kubectl get nodes
kubectl get namespaces

# Check frank-managed resources
kubectl get all -l app.kubernetes.io/managed-by=frank

# Check specific resource
kubectl describe deployment <name>
kubectl get deployment <name> -o yaml

# Check pod logs
kubectl logs deployment/<name>
kubectl logs -f deployment/<name>

# Check events
kubectl get events --sort-by=.metadata.creationTimestamp
```

### Frank Debug Commands

```bash
# Enable debug logging
FRANK_LOG_LEVEL=debug frank apply

# Check configuration loading
FRANK_LOG_LEVEL=debug frank apply dev

# Check template rendering
FRANK_LOG_LEVEL=debug frank apply dev/app

# Check stack filtering
FRANK_LOG_LEVEL=debug frank apply dev
```

### Template Debugging

```bash
# Test template rendering
FRANK_LOG_LEVEL=debug frank apply

# Look for these debug messages:
# - "Template context:"
# - "Template rendered:"
# - "Template rendering failed:"
```

## Performance Debugging

### Slow Deployments

**Debug Steps**:
```bash
# Check cluster resources
kubectl top nodes
kubectl top pods --all-namespaces

# Check resource usage
kubectl describe nodes

# Check for resource conflicts
kubectl get pods --all-namespaces
```

**Solutions**:
1. Check cluster resources
2. Use appropriate timeouts
3. Deploy fewer resources at once
4. Check network connectivity

### Memory Issues

**Debug Steps**:
```bash
# Check memory usage
kubectl top pods --all-namespaces

# Check node resources
kubectl describe nodes
```

**Solutions**:
1. Deploy fewer resources at once
2. Use stack filtering
3. Check for memory leaks in templates

## Log Analysis

### Understanding Log Levels

- **ERROR**: Critical errors that prevent operation
- **WARN**: Warnings about potential issues
- **INFO**: General information about operations
- **DEBUG**: Detailed debugging information

### Log Patterns

**Successful Deployment**:
```
INFO: Creating Deployment myapp-dev-web
WARN: Waiting for resource to be ready
INFO: Resource is ready
```

**Failed Deployment**:
```
ERROR: Failed to create Deployment: resource already exists
ERROR: Apply failed
```

**Template Issues**:
```
DEBUG: Template context: {stack_name: myapp-dev-web, ...}
ERROR: Template rendering failed: variable 'replicas' not found
```

## Troubleshooting Checklist

### Before Deploying

- [ ] Check kubectl connectivity: `kubectl get nodes`
- [ ] Verify context: `kubectl config current-context`
- [ ] Check config files exist: `ls -la config/`
- [ ] Verify template syntax
- [ ] Check required variables are defined

### During Deployment

- [ ] Enable debug logging: `FRANK_LOG_LEVEL=debug`
- [ ] Monitor resource creation: `kubectl get pods -w`
- [ ] Check pod logs: `kubectl logs deployment/<name>`
- [ ] Monitor events: `kubectl get events`

### After Deployment

- [ ] Verify resources are running: `kubectl get all -l app.kubernetes.io/managed-by=frank`
- [ ] Check resource status: `kubectl describe deployment <name>`
- [ ] Test application functionality
- [ ] Monitor logs for errors

## Getting Help

### Community Support

- **GitHub Issues**: [Report bugs or request features](https://github.com/schnauzersoft/frank-cli/issues)
- **GitHub Discussions**: [Ask questions and get help](https://github.com/schnauzersoft/frank-cli/discussions)
- **Email**: [ya@bsapp.ru](mailto:ya@bsapp.ru)

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

### Useful Resources

- [Kubernetes Troubleshooting](https://kubernetes.io/docs/tasks/debug-application-cluster/)
- [kubectl Cheat Sheet](https://kubernetes.io/docs/reference/kubectl/cheatsheet/)
- [Jinja Template Documentation](https://jinja.palletsprojects.com/)
