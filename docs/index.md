# Frank CLI

<div class="grid cards" markdown>

-   :material-rocket-launch:{ .lg .middle } **Quick Start**

    ---

    Get up and running with Frank in minutes

    [:octicons-arrow-right-24: Quick Start](getting-started/quick-start.md)

-   :material-cog:{ .lg .middle } **Configuration**

    ---

    Learn how to configure Frank for your environment

    [:octicons-arrow-right-24: Configuration](getting-started/configuration.md)

-   :material-code-tags:{ .lg .middle } **Jinja Templating**

    ---

    Dynamic manifest generation with powerful templating

    [:octicons-arrow-right-24: Jinja Templating](features/jinja-templating.md)

-   :material-kubernetes:{ .lg .middle } **Kubernetes Integration**

    ---

    Smart deployments with stack-based filtering

    [:octicons-arrow-right-24: Smart Deployments](features/smart-deployments.md)

</div>

## What is Frank?

Frank is a CLI tool for applying templated Kubernetes manifest files to clusters with intelligent configuration management and stack-based filtering. It simplifies multi-environment deployments by providing:

- **🎯 Smart Deployments** - Creates or updates resources intelligently
- **📝 Jinja Templating** - Dynamic manifest generation with powerful templating
- **🏷️ Stack Management** - Organize and filter deployments by environment
- **🔧 Configuration Inheritance** - Hierarchical config management
- **🧹 Resource Cleanup** - Surgical precision resource deletion
- **⚡ Parallel Processing** - Fast multi-stack deployments

## Key Features

### Smart Deployments
Frank intelligently creates new resources or updates existing ones, adding stack tracking annotations and waiting patiently for deployments to be ready.

### Jinja Templating
Generate dynamic Kubernetes manifests using Jinja templates with access to stack information, configuration values, and custom variables.

### Stack-Based Filtering
Target specific environments or applications with flexible filtering:
```bash
frank apply                    # Deploy everything
frank apply dev                # Deploy all dev environment stacks
frank apply dev/app            # Deploy all dev/app* configurations
frank apply dev/app.yaml       # Deploy specific configuration file
```

### Hierarchical Configuration
Organize your environments with inheritance and override patterns that make sense for your team.

## Quick Example

```yaml
# config/config.yaml
context: my-cluster
project_code: myapp
namespace: myapp-namespace

# config/app.yaml
manifest: app-deployment.jinja
app: myapp
version: 1.2.3

# manifests/app-deployment.jinja
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ stack_name }}
  labels:
    app.kubernetes.io/name: {{ app_name }}
    app.kubernetes.io/version: {{ version }}
    app.kubernetes.io/managed-by: frank
```

```bash
# Deploy with confirmation
frank apply

# Deploy specific stack
frank apply app

# Deploy without prompts
frank apply --yes
```

## Why Frank?

Frank fills the gap between simple `kubectl apply` and complex tools like Helm. It provides:

- **Simplicity** - No complex charts or templates to learn
- **Flexibility** - Jinja templating for dynamic content
- **Organization** - Stack-based filtering and management
- **Safety** - Interactive confirmations and conflict detection
- **Speed** - Parallel deployments and intelligent updates

## Getting Started

Ready to get started? Check out our [Quick Start Guide](getting-started/quick-start.md) or dive into the [Configuration](getting-started/configuration.md) documentation.

---

<div class="grid cards" markdown>

-   :material-github:{ .lg .middle } **GitHub**

    ---

    View source code and contribute

    [:octicons-arrow-right-24: schnauzersoft/frank-cli](https://github.com/schnauzersoft/frank-cli)

-   :material-bug:{ .lg .middle } **Issues**

    ---

    Report bugs or request features

    [:octicons-arrow-right-24: GitHub Issues](https://github.com/schnauzersoft/frank-cli/issues)

-   :material-help:{ .lg .middle } **Support**

    ---

    Get help and ask questions

    [:octicons-arrow-right-24: GitHub Discussions](https://github.com/schnauzersoft/frank-cli/discussions)

</div>
