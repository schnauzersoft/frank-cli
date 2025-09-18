# Delete Command

The `frank delete` command removes Kubernetes resources that are managed by Frank CLI with surgical precision.

## Usage

```bash
frank delete [stack] [flags]
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

### Delete All Frank-Managed Resources

```bash
# Interactive deletion (with confirmation)
frank delete

# Skip confirmation
frank delete --yes
```

### Delete Specific Stack

```bash
# Delete all dev environment resources
frank delete dev

# Delete all dev/app* resources
frank delete dev/app

# Delete specific configuration
frank delete dev/app.yaml

# Delete specific stack
frank delete myapp-dev-web
```

## What Delete Does

The delete command performs the following operations:

1. **Resource Discovery** - Finds all frank-managed resources
2. **Stack Filtering** - Filters resources based on the provided stack argument
3. **Resource Identification** - Identifies resources using stack annotations
4. **Resource Deletion** - Deletes resources from Kubernetes
5. **Status Reporting** - Reports deletion results

## Interactive Confirmation

By default, Frank shows an interactive confirmation before deleting:

```
Do you want to delete 'dev'? [y/N]
```

### Confirmation Details

The confirmation shows:
- **Scope** - What will be deleted (stack name or "all frank-managed resources")
- **Resources** - Number of resources that will be deleted
- **Context** - Kubernetes context that will be used

### Skipping Confirmation

Use the `--yes` flag to skip the confirmation prompt:

```bash
frank delete --yes
frank delete dev --yes
```

## Stack Filtering

Frank supports flexible stack filtering for deletion:

| Filter | Matches | Example |
|--------|---------|---------|
| `app` | Exact app name | `app.yaml` |
| `dev` | All dev environment stacks | `dev/app.yaml`, `dev/api.yaml` |
| `dev/app` | Specific dev app | `dev/app.yaml` |
| `dev/app.yaml` | Specific file | `dev/app.yaml` |

## Resource Identification

Frank identifies resources to delete using:

### 1. Stack Annotations

All frank-managed resources have stack annotations:

```yaml
metadata:
  annotations:
    frankthetank.cloud/stack-name: myapp-dev-web
```

### 2. Stack Filtering

Frank matches resources based on stack name patterns:

- **Exact match** - `myapp-dev-web` matches `myapp-dev-web`
- **Prefix match** - `dev` matches `myapp-dev-web`, `myapp-dev-api`
- **Path match** - `dev/app` matches `myapp-dev-app`

### 3. Resource Types

Frank only deletes supported resource types:

- Deployments
- StatefulSets
- DaemonSets
- Services
- ConfigMaps
- Secrets
- Pods
- Jobs
- CronJobs
- Ingresses

## Output Format

Frank uses a structured output format:

```
<timestamp> - <stack_name> - <operation_status>
```

### Example Output

```
2024-01-15T10:30:00Z - myapp-dev-web - Deleting frank-managed resource
2024-01-15T10:30:01Z - myapp-dev-web - Successfully deleted resource
2024-01-15T10:30:00Z - myapp-dev-api - Deleting frank-managed resource
2024-01-15T10:30:01Z - myapp-dev-api - Successfully deleted resource
```

### Status Messages

| Status | Description | Color |
|--------|-------------|-------|
| `Deleting frank-managed resource` | Deleting resource | Yellow |
| `Successfully deleted resource` | Resource deleted | Green |
| `Failed to delete resource` | Error occurred | Red |

## Error Handling

### Common Errors

**Permission Errors**
```
Error: permission denied: cannot delete resource
```

**Resource Not Found**
```
Error: resource not found
```

**Context Not Found**
```
Error: context 'my-cluster' not found in kubectl config
```

### Debugging

Enable debug logging to see detailed information:

```bash
FRANK_LOG_LEVEL=debug frank delete dev
```

Debug output includes:
- Resource discovery process
- Stack filtering results
- Deletion operations
- Error details

## Safety Features

### 1. Stack-Based Filtering

Frank only deletes resources it manages:

- Resources must have `frankthetank.cloud/stack-name` annotation
- Resources must match the provided stack filter
- Only supported resource types are deleted

### 2. Confirmation Prompts

Interactive confirmation prevents accidental deletions:

```bash
frank delete
# Do you want to delete 'all frank-managed resources'? [y/N]
```

### 3. Resource Identification

Frank clearly identifies what will be deleted:

```bash
frank delete dev
# Do you want to delete 'dev'? [y/N]
# This will delete resources with stack names matching 'dev'
```

## Use Cases

### 1. Environment Cleanup

```bash
# Clean up development environment
frank delete dev

# Clean up staging environment
frank delete staging

# Clean up production environment
frank delete prod
```

### 2. Application Cleanup

```bash
# Clean up specific application
frank delete dev/app

# Clean up specific service
frank delete dev/api

# Clean up specific configuration
frank delete dev/app.yaml
```

### 3. CI/CD Cleanup

```bash
# Clean up after failed deployment
frank delete dev --yes

# Clean up before new deployment
frank delete prod --yes
```

### 4. Maintenance Cleanup

```bash
# Clean up old resources
frank delete old-stack

# Clean up test resources
frank delete test
```

## Best Practices

### 1. Use Stack Filtering

```bash
# Delete specific environment
frank delete dev

# Delete specific application
frank delete dev/app

# Delete everything (use with caution)
frank delete
```

### 2. Always Confirm

```bash
# Always confirm before deleting
frank delete

# Only skip confirmation in CI/CD
frank delete --yes
```

### 3. Test First

```bash
# Test in dev first
frank delete dev

# Then staging
frank delete staging

# Finally production
frank delete prod
```

### 4. Monitor Resources

```bash
# Check what will be deleted
kubectl get all -l app.kubernetes.io/managed-by=frank

# Check specific stack
kubectl get all -l frankthetank.cloud/stack-name=myapp-dev-web
```

## Troubleshooting

### Common Issues

**"No resources found"**
- Check if resources have frank annotations
- Verify stack filter pattern
- Check if resources exist in the cluster

**"Permission denied"**
- Check RBAC settings
- Verify service account permissions
- Use appropriate context

**"Context not found"**
- Check available contexts: `kubectl config get-contexts`
- Update context in configuration
- Switch to correct context

### Debug Commands

```bash
# Enable debug logging
FRANK_LOG_LEVEL=debug frank delete dev

# Check frank-managed resources
kubectl get all -l app.kubernetes.io/managed-by=frank

# Check specific stack resources
kubectl get all -l frankthetank.cloud/stack-name=myapp-dev-web
```

## Integration with Other Commands

### Delete and Redeploy

```bash
# Clean up and redeploy
frank delete dev
frank apply dev
```

### Delete and Apply

```bash
# Delete specific application and redeploy
frank delete dev/app
frank apply dev/app
```

### Delete and Clean

```bash
# Clean up everything
frank delete
# Then redeploy
frank apply
```

## Next Steps

- [Apply Command](apply.md) - Learn about deploying resources
- [Configuration](configuration.md) - Understand configuration options
- [Resource Management](../features/resource-management.md) - Learn about resource lifecycle
