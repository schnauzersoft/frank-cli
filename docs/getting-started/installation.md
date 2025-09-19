# Installation

This guide covers different ways to install Frank CLI on your system.

## Prerequisites

Before installing Frank, ensure you have:

- **Go 1.25 or later** - [Download Go](https://golang.org/dl/)
- **Kubernetes cluster access** - Local (minikube/kind) or remote cluster
- **kubectl configured** - `kubectl get nodes` should work

## Installation Methods

### 1. Build from Source (Recommended)

This is the recommended method as it gives you the latest version and full control over the build process.

#### Clone and Build

```bash
# Clone the repository
git clone https://github.com/schnauzersoft/frank-cli
cd frank-cli

# Build the binary
go build -o frank .

# Verify installation
./frank --help
```

#### Install to System Path

```bash
# Install to /usr/local/bin (requires sudo)
sudo cp frank /usr/local/bin/

# Or install to ~/bin (add to PATH)
mkdir -p ~/bin
cp frank ~/bin/
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# Verify installation
frank --help
```

### 2. Go Install

Install directly from the repository using Go's install command:

```bash
# Install latest version
go install github.com/schnauzersoft/frank-cli@latest

# Install specific version
go install github.com/schnauzersoft/frank-cli@v1.0.0

# Verify installation
frank --help
```

### 3. Download Pre-built Binaries

Download pre-built binaries from the [GitHub Releases](https://github.com/schnauzersoft/frank-cli/releases) page.

#### Linux

```bash
# Download latest release
wget https://github.com/schnauzersoft/frank-cli/releases/latest/download/frank-linux-amd64.tar.gz

# Extract and install
tar -xzf frank-linux-amd64.tar.gz
sudo mv frank /usr/local/bin/

# Verify installation
frank --help
```

#### macOS

```bash
# Download latest release
wget https://github.com/schnauzersoft/frank-cli/releases/latest/download/frank-darwin-amd64.tar.gz

# Extract and install
tar -xzf frank-darwin-amd64.tar.gz
sudo mv frank /usr/local/bin/

# Verify installation
frank --help
```

#### Windows

```powershell
# Download latest release
Invoke-WebRequest -Uri "https://github.com/schnauzersoft/frank-cli/releases/latest/download/frank-windows-amd64.zip" -OutFile "frank-windows-amd64.zip"

# Extract and install
Expand-Archive -Path "frank-windows-amd64.zip" -DestinationPath "C:\Program Files\frank"
# Add C:\Program Files\frank to your PATH

# Verify installation
frank --help
```

### 4. Package Managers

#### Homebrew (macOS)

```bash
# Add tap (if not already added)
brew tap schnauzersoft/frank-cli

# Install
brew install frank-cli

# Verify installation
frank --help
```

#### Scoop (Windows)

```powershell
# Add bucket
scoop bucket add frank-cli https://github.com/schnauzersoft/frank-cli

# Install
scoop install frank-cli

# Verify installation
frank --help
```

#### AUR (Arch Linux)

```bash
# Install from AUR
yay -S frank-cli

# Or with paru
paru -S frank-cli

# Verify installation
frank --help
```

## Verification

After installation, verify that Frank is working correctly:

```bash
# Check version
frank --version

# Check help
frank --help

# Check kubectl connectivity
kubectl get nodes
```

## Configuration

### Initial Setup

1. **Create a project directory**:
   ```bash
   mkdir my-frank-project
   cd my-frank-project
   ```

2. **Set up basic configuration**:
   ```bash
   mkdir config manifests
   
   # Create base config
   cat > config/config.yaml << EOF
   context: my-cluster
   project_code: myapp
   namespace: myapp-namespace
   EOF
   
   # Create app config
   cat > config/app.yaml << EOF
   manifest: app-deployment.yaml
   timeout: 10m
   EOF
   ```

3. **Create a simple manifest**:
   ```bash
   cat > manifests/app-deployment.yaml << EOF
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: myapp
   spec:
     replicas: 3
     selector:
       matchLabels:
         app: myapp
     template:
       metadata:
         labels:
           app: myapp
       spec:
         containers:
         - name: myapp
           image: nginx:alpine
           ports:
           - containerPort: 80
   EOF
   ```

4. **Test deployment**:
   ```bash
   frank apply
   ```

### Environment Variables

Set up environment variables for Frank configuration:

```bash
# Set log level
export FRANK_LOG_LEVEL=info

# Add to shell profile
echo 'export FRANK_LOG_LEVEL=info' >> ~/.bashrc
source ~/.bashrc
```

## Troubleshooting

### Common Installation Issues

#### "command not found: frank"

**Problem**: Frank is not in your PATH.

**Solutions**:
1. Check if Frank is installed: `which frank`
2. Add Frank to your PATH
3. Restart your terminal

```bash
# Check if installed
which frank

# Add to PATH
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

#### "permission denied"

**Problem**: Insufficient permissions to install or run Frank.

**Solutions**:
1. Use `sudo` for system-wide installation
2. Install to user directory
3. Check file permissions

```bash
# Install to user directory
mkdir -p ~/bin
cp frank ~/bin/
chmod +x ~/bin/frank
```

#### "kubectl not found"

**Problem**: kubectl is not installed or not in PATH.

**Solutions**:
1. Install kubectl: [kubectl installation guide](https://kubernetes.io/docs/tasks/tools/)
2. Add kubectl to your PATH
3. Verify kubectl works: `kubectl version`

#### "context not found"

**Problem**: Kubernetes context is not configured.

**Solutions**:
1. Configure kubectl context
2. Check available contexts: `kubectl config get-contexts`
3. Switch context: `kubectl config use-context <context-name>`

### Build Issues

#### "go: module github.com/schnauzersoft/frank-cli: cannot find module"

**Problem**: Go module path issue.

**Solutions**:
1. Ensure you're using Go 1.25 or later
2. Check your Go module configuration
3. Try clearing module cache: `go clean -modcache`

#### "build failed"

**Problem**: Build process failed.

**Solutions**:
1. Check Go version: `go version`
2. Update dependencies: `go mod tidy`
3. Check for compilation errors: `go build -v .`

### Runtime Issues

#### "config directory with config.yaml not found"

**Problem**: Frank can't find configuration.

**Solutions**:
1. Ensure you're in a directory with `config/` subdirectory
2. Check that `config/config.yaml` exists
3. Verify file permissions

#### "context not found"

**Problem**: Kubernetes context doesn't exist.

**Solutions**:
1. Check contexts: `kubectl config get-contexts`
2. Update context in config
3. Switch context: `kubectl config use-context <context-name>`

## Uninstallation

### Remove from System Path

```bash
# Remove from /usr/local/bin
sudo rm /usr/local/bin/frank

# Remove from ~/bin
rm ~/bin/frank
```

### Remove from Package Managers

#### Homebrew

```bash
brew uninstall frank-cli
```

#### Scoop

```powershell
scoop uninstall frank-cli
```

#### AUR

```bash
yay -R frank-cli
# or
paru -R frank-cli
```

### Clean Up Configuration

```bash
# Remove configuration files
rm -rf ~/.frank
rm -f .frank.yaml
```

## Updating

### Update from Source

```bash
# Pull latest changes
git pull origin main

# Rebuild
go build -o frank .

# Install updated version
sudo cp frank /usr/local/bin/
```

### Update with Go Install

```bash
# Update to latest version
go install github.com/schnauzersoft/frank-cli@latest
```

### Update with Package Managers

#### Homebrew

```bash
brew update
brew upgrade frank-cli
```

#### Scoop

```powershell
scoop update frank-cli
```

## Next Steps

After successful installation:

1. **Follow the [Quick Start Guide](quick-start.md)** to get up and running
2. **Read the [Configuration Guide](configuration.md)** to understand configuration options
3. **Explore [Jinja Templating](../features/jinja-templating.md)** for dynamic manifests
4. **Check out [Advanced Usage](../advanced/multi-environment.md)** for complex setups

## Support

If you encounter issues during installation:

- Check the [Troubleshooting Guide](../reference/troubleshooting.md)
- Open an [issue on GitHub](https://github.com/schnauzersoft/frank-cli/issues)
- Join the [discussions](https://github.com/schnauzersoft/frank-cli/discussions)
