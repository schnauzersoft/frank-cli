# Contributing

Thank you for your interest in contributing to Frank CLI! This guide will help you get started with contributing to the project.

## Getting Started

### Prerequisites

Before contributing, ensure you have:

- **Go 1.25 or later** - [Download Go](https://golang.org/dl/)
- **Git** - [Download Git](https://git-scm.com/downloads)
- **Kubernetes cluster access** - Local (orbstack/minikube/kind) or remote cluster
- **kubectl configured** - `kubectl get nodes` should work

### Development Setup

1. **Fork the repository** on GitHub
2. **Clone your fork**:
   ```bash
   git clone https://github.com/your-username/frank-cli
   cd frank-cli
   ```

3. **Add upstream remote**:
   ```bash
   git remote add upstream https://github.com/schnauzersoft/frank-cli
   ```

4. **Install dependencies**:
   ```bash
   go mod tidy
   ```

5. **Build the project**:
   ```bash
   go build -o frank .
   ```

6. **Run tests**:
   ```bash
   go test ./...
   ```

## Development Workflow

### 1. Create a Feature Branch

```bash
# Create and switch to feature branch
git checkout -b feature/your-feature-name

# Or for bug fixes
git checkout -b fix/your-bug-description
```

### 2. Make Changes

- Write your code following the project's coding standards
- Add tests for new functionality
- Update documentation as needed
- Ensure all tests pass

### 3. Test Your Changes

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./pkg/stack

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...
```

### 4. Format and Lint

```bash
# Format code
go fmt ./...

# Run static analysis
go run honnef.co/go/tools/cmd/staticcheck@latest ./...

# Check for complexity issues
gocyclo -over 10 .
gocognit -over 10 .
```

### 5. Commit Changes

```bash
# Add changes
git add .

# Commit with descriptive message
git commit -m "feat: add new feature description"

- Detailed description of changes
- Any breaking changes
- Related issues"
```

### 6. Push and Create Pull Request

```bash
# Push to your fork
git push origin feature/your-feature-name

# Create pull request on GitHub
```

## Coding Standards

### Go Code Style

- Follow standard Go formatting with `gofmt`
- Use meaningful variable and function names
- Write clear, concise comments
- Keep functions small and focused
- Use interfaces where appropriate

### Code Organization

- Keep related functionality in the same package
- Use private functions for internal logic
- Export only what's necessary
- Follow the existing package structure

### Error Handling

- Use descriptive error messages
- Wrap errors with context when appropriate
- Handle errors at the appropriate level
- Log errors with appropriate levels

### Testing

- Write unit tests for all new functionality
- Use table-driven tests where appropriate
- Mock external dependencies
- Test error conditions and edge cases

## Project Structure

Understanding the project structure will help you contribute effectively:

```
frank-cli/
├── cmd/                    # Command-line interface
│   ├── root.go            # Root command and global flags
│   ├── apply.go           # Apply command implementation
│   ├── delete.go          # Delete command implementation
│   └── utils.go           # Shared utilities
├── pkg/
│   ├── config/            # Configuration management
│   ├── deploy/            # Deployment orchestration
│   ├── kubernetes/        # Kubernetes operations
│   ├── stack/             # Stack management
│   └── template/          # Jinja templating
├── docs/                  # Documentation
├── examples/              # Example configurations
├── .github/
│   └── workflows/         # GitHub Actions
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
├── mkdocs.yml            # MkDocs configuration
└── README.md             # Project README
```

## Types of Contributions

### Bug Reports

When reporting bugs, please include:

- **Description** - Clear description of the bug
- **Steps to reproduce** - Detailed steps to reproduce the issue
- **Expected behavior** - What should happen
- **Actual behavior** - What actually happens
- **Environment** - OS, Go version, Kubernetes version
- **Logs** - Relevant log output

### Feature Requests

When requesting features, please include:

- **Use case** - Why is this feature needed?
- **Proposed solution** - How should it work?
- **Alternatives** - What other solutions have you considered?
- **Additional context** - Any other relevant information

### Code Contributions

#### Small Changes

- Bug fixes
- Documentation updates
- Small feature additions
- Test improvements

#### Large Changes

- New major features
- Architectural changes
- Breaking changes
- New dependencies

For large changes, please:

1. **Open an issue first** to discuss the change
2. **Get feedback** from maintainers
3. **Create a draft PR** for early feedback
4. **Break down** large changes into smaller PRs

## Pull Request Guidelines

### Before Submitting

- [ ] Code follows project style guidelines
- [ ] All tests pass
- [ ] Code is properly formatted
- [ ] Documentation is updated
- [ ] No breaking changes (or clearly documented)
- [ ] Commit messages are descriptive

### Pull Request Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update
- [ ] Test improvement

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing performed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] No breaking changes (or documented)
```

### Review Process

1. **Automated checks** - CI/CD pipeline runs tests and checks
2. **Code review** - Maintainers review the code
3. **Feedback** - Address any feedback or requested changes
4. **Approval** - Once approved, changes are merged

## Development Tools

### Recommended IDE

- **VS Code** with Go extension
- **GoLand** by JetBrains
- **Vim/Neovim** with Go plugins

### Useful Commands

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test -run TestFunctionName ./pkg/stack

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o frank-linux .
GOOS=windows GOARCH=amd64 go build -o frank-windows.exe .

# Format code
go fmt ./...

# Check for issues
go vet ./...

# Run static analysis
go run honnef.co/go/tools/cmd/staticcheck@latest ./...

# Check complexity
gocyclo -over 10 .
gocognit -over 10 .
```

### Debugging

```bash
# Run with debug logging
FRANK_LOG_LEVEL=debug ./frank apply

# Run with race detection
go run -race . apply

# Profile CPU usage
go run . apply -cpuprofile=cpu.prof
go tool pprof cpu.prof
```

## Documentation

### Updating Documentation

When making changes that affect functionality:

1. **Update relevant documentation** in the `docs/` directory
2. **Update README.md** if needed
3. **Add examples** for new features
4. **Update configuration reference** for new config options

### Documentation Structure

- **Getting Started** - Installation and quick start
- **Features** - Detailed feature documentation
- **Commands** - Command reference
- **Advanced** - Advanced usage patterns
- **Reference** - Configuration and troubleshooting
- **Development** - Contributing and architecture

## Release Process

### Versioning

Frank uses [Semantic Versioning](https://semver.org/):

- **MAJOR** - Breaking changes
- **MINOR** - New features (backward compatible)
- **PATCH** - Bug fixes (backward compatible)

### Release Steps

1. **Update version** in relevant files
2. **Update CHANGELOG.md** with changes
3. **Create release tag** on GitHub
4. **Build and upload** binaries
5. **Update documentation** if needed

## Community Guidelines

### Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and grow
- Follow the [Contributor Covenant](https://www.contributor-covenant.org/)

### Communication

- **GitHub Issues** - Bug reports and feature requests
- **GitHub Discussions** - Questions and general discussion
- **Pull Requests** - Code changes and reviews

### Getting Help

- Check existing issues and discussions
- Ask questions in GitHub Discussions
- Join the community conversations
- Be patient and respectful

## Recognition

Contributors will be recognized in:

- **CONTRIBUTORS.md** - List of all contributors
- **Release notes** - Major contributors for each release
- **GitHub contributors** - Automatic recognition on GitHub

## License

By contributing to Frank CLI, you agree that your contributions will be licensed under the same license as the project (MIT License).

## Questions?

If you have questions about contributing:

Reach out to [ya@bsapp.ru](mailto:ya@bsapp.ru).

Thank you for contributing to Frank CLI! 🎉
