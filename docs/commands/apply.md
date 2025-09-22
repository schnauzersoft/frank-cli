# Apply Command

The `frank apply` command deploys or updates Kubernetes resources based on your configuration.

## Usage

```bash
$ frank apply [stack] [flags]
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
$ frank apply

# Skip confirmation
$ frank apply --yes
```

### Deploy Specific Stack

```bash
# Deploy specific app
$ frank apply app

# Deploy all dev environment stacks
$ frank apply dev

# Deploy specific dev app
$ frank apply dev/app

# Deploy specific configuration file
$ frank apply dev/app.yaml
```

## What Apply Does

The apply command performs the following operations:

1. **Configuration Discovery** - Finds and loads configuration files
2. **Stack Filtering** - Filters configurations based on the provided stack argument
3. **Template Rendering** - Renders Jinja and HCL templates with context variables
4. **Namespace Validation** - Checks for namespace conflicts
5. **Resource Application** - Creates or updates Kubernetes resources
6. **Status Monitoring** - Waits for resources to be ready
7. **Parallel Processing** - Runs multiple deployments concurrently

## Interactive Confirmation

By default, **frank** shows an interactive confirmation before deploying:

```
Do you want to apply 'dev'? [y/N]
```

### Skipping Confirmation

Use the `--yes` flag to skip the confirmation prompt:

```bash
$ frank apply --yes
$ frank apply dev --yes
```

## Output Format

**frank** uses a structured output format:

```
<timestamp> - <stack_name> - <operation_status>
```

### Status Messages

| Status | Description | Color |
|--------|-------------|-------|
| `Creating <Resource>` | Creating new resource | Yellow |
| `Updating <Resource>` | Updating existing resource | Yellow |
| `Resource is ready` | Resource is ready | Green |
| `Resource is already up to date` | No changes needed | Green |
| `Apply failed` | Error occurred | Red |
