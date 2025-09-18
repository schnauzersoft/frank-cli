# Frank CLI

A CLI tool for deploying Jinja templated Kubernetes manifest files to clusters with hierarchical configuration support.

## Overview

Frank is a command-line application that uses a hierarchical configuration system to deploy Kubernetes manifests. It requires a `config` directory structure and supports context inheritance and template specification through separate configuration files.

## Installation

1. Clone the repository
2. Build the application:
   ```bash
   go build -o frank .
   ```

## Usage

### Configuration Structure

Frank uses a hierarchical configuration system with the following rules:

1. **Context Configuration**: Only specified in `config.yaml` files
2. **Template Configuration**: Only specified in non-`config.yaml` files (e.g., `deploy.yaml`, `app.yaml`)
3. **Directory Requirement**: The tool only works when there's a `config` directory in the current location or immediate parent

### Basic Setup

Create a `config` directory with the following structure:

```
project/
├── config/
│   ├── config.yaml          # Base context configuration
│   └── deploy.yaml          # Template configuration
└── manifests/               # Kubernetes manifest templates
```

#### Base Configuration (`config/config.yaml`)

```yaml
context: orbstack
```

This file specifies the Kubernetes context that will be used for all deployments.

#### Template Configuration (`config/deploy.yaml`)

```yaml
template: sample-deployment.yaml
```

This file specifies which manifest template to deploy from the `manifests/` directory at the project root.

### Deploy Command

Deploy manifests to the Kubernetes cluster:

```bash
./frank deploy
```

This command will:
1. Find the `config` directory in the current location or immediate parent
2. Read the context from `config.yaml` files (with inheritance from parent directories)
3. Find template configuration from non-`config.yaml` files
4. Connect to the Kubernetes cluster using the specified context
5. Deploy the specified manifest template from the `manifests/` directory at the project root

### Hierarchical Configuration

Frank supports nested configuration directories with inheritance:

```
project/
├── config/
│   ├── config.yaml          # Base context: "orbstack"
│   ├── deploy.yaml          # Template: "sample-deployment.yaml"
│   ├── prod/
│   │   └── config.yaml      # Override context: "prod-cluster"
│   ├── dev/
│   │   └── config.yaml      # Override context: "dev-cluster"
│   └── staging/
│       ├── config.yaml      # Override context: "staging-cluster"
│       └── web/
│           └── config.yaml  # Override context: "web-staging"
└── manifests/
    ├── sample-deployment.yaml
    ├── prod-deployment.yaml
    └── dev-deployment.yaml
```

#### Context Inheritance Rules

- Context is inherited from parent directories within the config hierarchy
- Child directories can override the context from their parent
- Inheritance stops at the `config` directory root

#### Template Discovery Rules

- Templates are specified in non-`config.yaml` files only
- Discovery follows precedence order: `deploy.yaml` > `deploy.yml` > `app.yaml` > `app.yml` > `service.yaml` > `service.yml`
- Search starts in the current directory and walks up the config hierarchy

### Usage Examples

#### From Project Root
```bash
frank deploy
# Uses: config/config.yaml (context: orbstack)
# Uses: config/deploy.yaml (template: sample-deployment.yaml)
```

#### From Config Directory
```bash
cd config
frank deploy
# Uses: config/config.yaml (context: orbstack)
# Uses: config/deploy.yaml (template: sample-deployment.yaml)
```

#### From Subdirectory (Fails)
```bash
cd config/prod
frank deploy
# Fails: No config directory in current or immediate parent
```

### Directory Structure

```
frank-cli/
├── cmd/                    # Command implementations
│   ├── root.go            # Root command
│   └── deploy.go          # Deploy command
├── main.go               # Application entry point
├── go.mod                # Go module file
└── go.sum                # Go module checksums
```

## Development

### Dependencies

- Go 1.25.0 or later
- Cobra CLI framework
- Kubernetes client-go v0.34.1
- YAML parsing library (gopkg.in/yaml.v3)

### Building

```bash
go mod tidy
go build -o frank .
```

### Testing

```bash
# Test from project root (should work)
frank deploy

# Test from config directory (should work)
cd config && frank deploy

# Test from subdirectory (should fail)
cd config/prod && frank deploy
```

## License

Copyright © 2025 Ben Sapp ya.bsapp.ru
