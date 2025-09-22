# frank

<div class="grid cards" markdown>

-   :material-code-tags:{ .lg .middle } **Templating**

    ---

    Dynamic manifest generation with powerful templating

    [:octicons-arrow-right-24: Templating](features/templating.md)

-   :material-link:{ .lg .middle } **Dependency Management**

    ---

    Define execution order with stack dependencies

    [:octicons-arrow-right-24: Dependency Management](features/dependency-management.md)

</div>

## What is frank?

**frank** is a CLI tool for applying templated Kubernetes manifest files to clusters with intelligent configuration management and stack-based filtering. It simplifies multi-environment deployments by providing:

- **🎯 Smart Deployments** - Creates or updates resources intelligently
- **📝 Jinja Templating** - Dynamic manifest generation with powerful templating
- **🏷️ Stack Management** - Organize and filter deployments by environment
- **🔧 Configuration Inheritance** - Hierarchical config management
- **🔗 Dependency Management** - Define execution order with stack dependencies
- **🧹 Resource Cleanup** - Surgical precision resource deletion
- **⚡ Parallel Processing** - Fast multi-stack deployments

## Key Features

### Templating
Generate dynamic Kubernetes manifests using Jinja templates and HCL.

### Stack-Based Filtering
Target specific environments or applications with flexible filtering:
```bash
$ frank apply                    # Deploy everything
$ frank apply dev                # Deploy all dev environment stacks
$ frank apply dev/app            # Deploy all dev/app* configurations
$ frank apply dev/app.yaml       # Deploy specific configuration file
```

### Hierarchical Configuration
Organize your environments with inheritance and override patterns that make sense for your team.

### Dependency Management
Define execution order for your stacks using the `depends_on` field, ensuring dependent stacks are deployed only after their dependencies complete successfully.

## Why frank?

**frank** fills the gap between simple `kubectl apply` and complex tools like Helm. It provides:

- **Simplicity** - No package lifecyle to manage
- **Flexibility** - Jinja and HCL templating for dynamic content
- **Organization** - Stack-based filtering and management
- **Speed** - Parallel deployments and intelligent updates
