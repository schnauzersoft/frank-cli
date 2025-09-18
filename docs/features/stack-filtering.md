# Stack Filtering

Frank CLI provides powerful stack-based filtering that allows you to deploy specific environments, applications, or configurations with precision.

## What is Stack Filtering?

Stack filtering lets you target specific deployments using flexible patterns:

```bash
frank apply                    # Deploy everything
frank apply dev                # Deploy all dev environment stacks
frank apply dev/app            # Deploy all dev/app* configurations
frank apply dev/app.yaml       # Deploy specific configuration file
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

## Filtering Patterns

### 1. Exact Match

Deploy a specific application:

```bash
frank apply app
```

Matches:
- `config/app.yaml` → `myapp-dev-app`
- `config/prod/app.yaml` → `myapp-prod-app`

### 2. Environment Filter

Deploy all applications in an environment:

```bash
frank apply dev
```

Matches:
- `config/dev/app.yaml` → `myapp-dev-app`
- `config/dev/api.yaml` → `myapp-dev-api`
- `config/dev/web.yaml` → `myapp-dev-web`

### 3. Path Pattern

Deploy applications matching a path pattern:

```bash
frank apply dev/app
```

Matches:
- `config/dev/app.yaml` → `myapp-dev-app`
- `config/dev/app-service.yaml` → `myapp-dev-app-service`

### 4. File Pattern

Deploy a specific configuration file:

```bash
frank apply dev/app.yaml
```

Matches:
- `config/dev/app.yaml` → `myapp-dev-app`

## Directory Structure Examples

### Basic Structure

```
config/
├── config.yaml              # Base configuration
├── app.yaml                 # Root app config
├── dev/
│   ├── config.yaml          # Dev environment config
│   ├── app.yaml             # Dev app config
│   └── api.yaml             # Dev API config
├── staging/
│   ├── config.yaml          # Staging environment config
│   ├── app.yaml             # Staging app config
│   └── api.yaml             # Staging API config
└── prod/
    ├── config.yaml          # Production environment config
    ├── app.yaml             # Production app config
    └── api.yaml             # Production API config
```

### Filtering Results

| Command | Matches |
|---------|---------|
| `frank apply` | All configurations |
| `frank apply app` | `config/app.yaml` |
| `frank apply dev` | `config/dev/app.yaml`, `config/dev/api.yaml` |
| `frank apply dev/app` | `config/dev/app.yaml` |
| `frank apply staging` | `config/staging/app.yaml`, `config/staging/api.yaml` |
| `frank apply prod/api` | `config/prod/api.yaml` |

## Advanced Filtering

### Nested Directories

```
config/
├── config.yaml
├── dev/
│   ├── config.yaml
│   ├── web/
│   │   ├── config.yaml
│   │   ├── frontend.yaml
│   │   └── backend.yaml
│   └── api/
│       ├── config.yaml
│       ├── user-service.yaml
│       └── payment-service.yaml
└── prod/
    ├── config.yaml
    ├── web/
    │   ├── config.yaml
    │   ├── frontend.yaml
    │   └── backend.yaml
    └── api/
        ├── config.yaml
        ├── user-service.yaml
        └── payment-service.yaml
```

### Filtering Results

| Command | Matches |
|---------|---------|
| `frank apply dev` | All dev configurations |
| `frank apply dev/web` | `config/dev/web/frontend.yaml`, `config/dev/web/backend.yaml` |
| `frank apply dev/web/frontend` | `config/dev/web/frontend.yaml` |
| `frank apply dev/api` | `config/dev/api/user-service.yaml`, `config/dev/api/payment-service.yaml` |
| `frank apply prod/web/frontend` | `config/prod/web/frontend.yaml` |

## Use Cases

### 1. Environment-Specific Deployments

```bash
# Deploy to development
frank apply dev

# Deploy to staging
frank apply staging

# Deploy to production
frank apply prod
```

### 2. Application-Specific Deployments

```bash
# Deploy web application to all environments
frank apply dev/web
frank apply staging/web
frank apply prod/web

# Deploy API services to all environments
frank apply dev/api
frank apply staging/api
frank apply prod/api
```

### 3. Selective Deployments

```bash
# Deploy only frontend to dev
frank apply dev/web/frontend

# Deploy only user service to staging
frank apply staging/api/user-service

# Deploy specific configuration file
frank apply prod/web/frontend.yaml
```

### 4. CI/CD Integration

```bash
# In your CI pipeline
if [ "$BRANCH" = "develop" ]; then
    frank apply dev --yes
elif [ "$BRANCH" = "main" ]; then
    frank apply prod --yes
fi
```

## Configuration Examples

### Base Configuration

```yaml
# config/config.yaml
project_code: myapp
namespace: myapp
timeout: 10m
```

### Environment Configurations

```yaml
# config/dev/config.yaml
context: dev-cluster
namespace: myapp-dev

# config/staging/config.yaml
context: staging-cluster
namespace: myapp-staging

# config/prod/config.yaml
context: prod-cluster
namespace: myapp-prod
```

### Application Configurations

```yaml
# config/dev/app.yaml
manifest: app-deployment.jinja
app: web
version: 1.0.0-dev

# config/dev/api.yaml
manifest: api-deployment.jinja
app: api
version: 1.0.0-dev

# config/prod/app.yaml
manifest: app-deployment.jinja
app: web
version: 1.0.0

# config/prod/api.yaml
manifest: api-deployment.jinja
app: api
version: 1.0.0
```

## Stack Name Examples

Based on the configuration above:

| Config File | Stack Name |
|-------------|------------|
| `config/dev/app.yaml` | `myapp-dev-web` |
| `config/dev/api.yaml` | `myapp-dev-api` |
| `config/staging/app.yaml` | `myapp-staging-web` |
| `config/staging/api.yaml` | `myapp-staging-api` |
| `config/prod/app.yaml` | `myapp-prod-web` |
| `config/prod/api.yaml` | `myapp-prod-api` |

## Best Practices

### 1. Use Consistent Naming

```yaml
# Good - consistent naming
config/dev/app.yaml
config/staging/app.yaml
config/prod/app.yaml

# Avoid - inconsistent naming
config/dev/web.yaml
config/staging/app.yaml
config/prod/frontend.yaml
```

### 2. Organize by Environment

```
config/
├── dev/           # Development environment
├── staging/       # Staging environment
└── prod/          # Production environment
```

### 3. Use Descriptive Names

```yaml
# Good - descriptive names
config/dev/web-frontend.yaml
config/dev/api-user-service.yaml

# Avoid - unclear names
config/dev/app1.yaml
config/dev/service2.yaml
```

### 4. Group Related Applications

```
config/
├── dev/
│   ├── web/       # Web applications
│   │   ├── frontend.yaml
│   │   └── backend.yaml
│   └── api/       # API services
│       ├── user-service.yaml
│       └── payment-service.yaml
```

## Troubleshooting

### Common Issues

**"No configurations found"**
- Check if config files exist in the expected locations
- Verify the filter pattern matches your directory structure
- Ensure config files have valid YAML syntax

**"Stack name not generated"**
- Check that `project_code` is defined in base config
- Verify `context` is defined in environment config
- Ensure `app` is defined in application config

**"Filter not working"**
- Check the exact directory structure
- Verify the filter pattern syntax
- Use debug logging to see what's being matched

### Debug Commands

```bash
# Enable debug logging
FRANK_LOG_LEVEL=debug frank apply dev

# Check what configurations are found
FRANK_LOG_LEVEL=debug frank apply dev/app

# Verify stack name generation
FRANK_LOG_LEVEL=debug frank apply dev/app.yaml
```

## Advanced Patterns

### 1. Wildcard Filtering

```bash
# Deploy all web applications
frank apply */web

# Deploy all API services
frank apply */api
```

### 2. Multiple Filters

```bash
# Deploy specific applications across environments
frank apply dev/app
frank apply staging/app
frank apply prod/app
```

### 3. Exclude Patterns

```bash
# Deploy everything except specific applications
frank apply dev  # Deploy all dev apps
# Then manually deploy specific ones
frank apply dev/app
```

## Next Steps

- [Smart Deployments](smart-deployments.md) - Learn about intelligent deployments
- [Jinja Templating](jinja-templating.md) - Use dynamic templates
- [Multi-Environment Setup](../advanced/multi-environment.md) - Set up multiple environments
