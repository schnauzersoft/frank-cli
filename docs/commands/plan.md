# Plan Command

The `frank plan` command shows the diff between what you have configured locally and what's running in Kubernetes.

## Usage

```bash
$ frank plan [stack] [flags]
```

## Arguments

| Argument | Description | Example |
|----------|-------------|---------|
| `stack` | Optional stack filter | `dev`, `dev/app`, `prod/api.yaml` |

## Examples

### Plan All Stacks

```bash
# Interactive deployment (with confirmation)
$ frank plan
```

### Plan Specific Stack

```bash
# Plan specific app
$ frank plan app

# Plan all dev environment stacks
$ frank plan dev

# Plan specific dev app
$ frank plan dev/app

# Plan specific configuration file
$ frank plan dev/app.yaml
```
