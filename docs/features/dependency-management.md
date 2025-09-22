# Dependency Management

**frank** supports stack dependencies through the `depends_on` configuration field, allowing you to define execution order for your deployments. This ensures that dependent stacks are deployed only after their dependencies have been successfully deployed.

## Configuration

The `depends_on` field is specified in your manifest configuration files (e.g., `app.yaml`, `api.yaml`) and accepts a list of stack names that must be deployed before the current stack.

### Basic Syntax

```yaml
# config/dev/api.yaml
manifest: api-deployment.yaml
app: api
version: 1.0.0
depends_on:
  - frank-dev-database
  - frank-dev-redis
```

### Stack Name Format

Dependencies are specified using the full stack name, which follows the pattern:
```
{project_code}-{context}-{app_name}
```

For example, if your configuration is:
- `project_code: myapp`
- `context: dev`
- `app: api`

The stack name would be: `myapp-dev-api`

## Examples

### Simple Dependency Chain

```yaml
# config/dev/database.yaml
manifest: database-deployment.yaml
app: database

# config/dev/api.yaml
manifest: api-deployment.yaml
app: api
depends_on:
  - frank-dev-database

# config/dev/web.yaml
manifest: web-deployment.yaml
app: web
depends_on:
  - frank-dev-api
```

**Execution Order:**
1. `frank-dev-database` (no dependencies)
2. `frank-dev-api` (depends on database)
3. `frank-dev-web` (depends on api)

### Multiple Dependencies

```yaml
# config/dev/api.yaml
manifest: api-deployment.yaml
app: api
depends_on:
  - frank-dev-database
  - frank-dev-redis
  - frank-dev-messaging
```

The API stack will wait for all three dependencies to complete before deploying.

### Complex Dependency Graph

```yaml
# config/dev/database.yaml
manifest: database-deployment.yaml
app: database

# config/dev/redis.yaml
manifest: redis-deployment.yaml
app: redis

# config/dev/api.yaml
manifest: api-deployment.yaml
app: api
depends_on:
  - frank-dev-database
  - frank-dev-redis

# config/dev/web.yaml
manifest: web-deployment.yaml
app: web
depends_on:
  - frank-dev-api

# config/dev/worker.yaml
manifest: worker-deployment.yaml
app: worker
depends_on:
  - frank-dev-database
  - frank-dev-redis
```

**Execution Order:**
1. `frank-dev-database` and `frank-dev-redis` (parallel, no dependencies)
2. `frank-dev-api` and `frank-dev-worker` (parallel, both depend on database and redis)
3. `frank-dev-web` (depends on api)

## Error Handling

### Circular Dependencies

**frank** detects and prevents circular dependencies:

```yaml
# This will cause an error
# config/dev/app1.yaml
depends_on:
  - frank-dev-app2

# config/dev/app2.yaml
depends_on:
  - frank-dev-app1
```

**Error:** `circular dependency detected: frank-dev-app1 -> frank-dev-app2`

### Missing Dependencies

If a stack depends on a non-existent stack:

```yaml
# config/dev/api.yaml
depends_on:
  - frank-dev-nonexistent
```

**Error:** `stack 'frank-dev-api' depends on 'frank-dev-nonexistent' which does not exist`

## Usage

### Deploy with Dependencies

```bash
# Deploy all stacks in dependency order
frank apply

# Deploy specific environment
frank apply dev

# Deploy specific stack (and its dependencies)
frank apply dev/api
```

### Plan with Dependencies

```bash
# Plan all stacks in dependency order
frank plan

# Plan specific environment
frank plan dev
```

## Best Practices

1. **Keep Dependencies Simple**: Avoid complex dependency graphs that are hard to understand and maintain.

2. **Use Descriptive Names**: Make stack names clear and consistent to avoid confusion in dependencies.

3. **Document Dependencies**: Consider adding comments in configuration files explaining why dependencies exist.

4. **Test Dependency Changes**: When modifying dependencies, test with `frank plan` first to ensure the execution order is correct.

5. **Avoid Deep Chains**: Long dependency chains can slow down deployments and make debugging harder.

## Implementation Details

- Dependencies are resolved using topological sorting
- Circular dependency detection uses depth-first search
- Stack execution is sequential when dependencies exist
- Failed deployments stop the execution of dependent stacks
- Dependencies are validated before any deployment begins
