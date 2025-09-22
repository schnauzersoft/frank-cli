# Delete Command

The `frank delete` command removes Kubernetes resources that are managed by **frank**.

## Usage

```bash
$ frank delete [stack] [flags]
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

### Delete All frank-Managed Resources

```bash
# Interactive deletion (with confirmation)
$ frank delete

# Skip confirmation
$ frank delete --yes
```

### Delete Specific Stack

```bash
# Delete all dev environment resources
$ frank delete dev

# Delete all dev/app* resources
$ frank delete dev/app

# Delete specific configuration
$ frank delete dev/app.yaml

# Delete specific stack
$ frank delete myapp-dev-web
```

## What Delete Does

The delete command performs the following operations:

1. **Resource Discovery** - Finds all frank-managed resources
2. **Stack Filtering** - Filters resources based on the provided stack argument
3. **Resource Identification** - Identifies resources using stack annotations
4. **Resource Deletion** - Deletes resources from Kubernetes
5. **Status Reporting** - Reports deletion results

## Interactive Confirmation

By default, **frank** shows an interactive confirmation before deleting:

```
Do you want to delete 'dev'? [y/N]
```

### Skipping Confirmation

Use the `--yes` flag to skip the confirmation prompt:

```bash
$ frank delete --yes
$ frank delete dev --yes
```
