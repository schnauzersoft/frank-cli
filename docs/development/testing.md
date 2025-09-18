# Testing

This guide covers testing Frank CLI, including unit tests, integration tests, and testing best practices.

## Test Structure

Frank CLI uses Go's built-in testing framework with a comprehensive test suite:

```
frank-cli/
├── pkg/
│   ├── config/
│   │   └── config_test.go
│   ├── deploy/
│   │   └── deployer_test.go
│   ├── kubernetes/
│   │   └── comparison_test.go
│   ├── stack/
│   │   ├── stack_test.go
│   │   └── stack_error_test.go
│   └── template/
│       └── renderer_test.go
├── main_test.go
├── integration_test.go
└── benchmark_test.go
```

## Running Tests

### Run All Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Run tests with coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Run Specific Test Packages

```bash
# Test specific package
go test ./pkg/stack

# Test specific package with verbose output
go test -v ./pkg/stack

# Test specific package with coverage
go test -cover ./pkg/stack
```

### Run Specific Tests

```bash
# Run specific test function
go test -run TestGenerateStackName ./pkg/stack

# Run tests matching pattern
go test -run TestConfig ./pkg/config

# Run tests with timeout
go test -timeout 30s ./...
```

## Unit Tests

### Configuration Tests

Test configuration loading and validation:

```go
// pkg/config/config_test.go
func TestLoadConfig(t *testing.T) {
    tests := []struct {
        name     string
        envVars  map[string]string
        expected *Config
    }{
        {
            name: "default config",
            envVars: map[string]string{},
            expected: &Config{
                LogLevel: "info",
            },
        },
        {
            name: "custom log level",
            envVars: map[string]string{
                "FRANK_LOG_LEVEL": "debug",
            },
            expected: &Config{
                LogLevel: "debug",
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Set environment variables
            for k, v := range tt.envVars {
                os.Setenv(k, v)
                defer os.Unsetenv(k)
            }

            config, err := LoadConfig()
            if err != nil {
                t.Fatalf("LoadConfig() error = %v", err)
            }

            if config.LogLevel != tt.expected.LogLevel {
                t.Errorf("LoadConfig() LogLevel = %v, want %v", config.LogLevel, tt.expected.LogLevel)
            }
        })
    }
}
```

### Stack Tests

Test stack name generation and configuration inheritance:

```go
// pkg/stack/stack_test.go
func TestGenerateStackName(t *testing.T) {
    tests := []struct {
        name        string
        projectCode string
        context     string
        appName     string
        expected    string
    }{
        {
            name:        "basic stack name",
            projectCode: "myapp",
            context:     "dev",
            appName:     "web",
            expected:    "myapp-dev-web",
        },
        {
            name:        "empty project code",
            projectCode: "",
            context:     "dev",
            appName:     "web",
            expected:    "dev-web",
        },
        {
            name:        "empty context",
            projectCode: "myapp",
            context:     "",
            appName:     "web",
            expected:    "myapp-web",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := GenerateStackName(tt.projectCode, tt.context, tt.appName)
            if result != tt.expected {
                t.Errorf("GenerateStackName() = %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### Template Tests

Test Jinja template rendering:

```go
// pkg/template/renderer_test.go
func TestRenderManifest(t *testing.T) {
    renderer := NewRenderer()

    tests := []struct {
        name     string
        template string
        context  map[string]interface{}
        expected string
    }{
        {
            name:     "basic template",
            template: "Hello {{ name }}!",
            context:  map[string]interface{}{"name": "World"},
            expected: "Hello World!",
        },
        {
            name:     "template with default",
            template: "Port: {{ port | default(80) }}",
            context:  map[string]interface{}{},
            expected: "Port: 80",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := renderer.RenderManifest(tt.template, tt.context)
            if err != nil {
                t.Fatalf("RenderManifest() error = %v", err)
            }

            if result != tt.expected {
                t.Errorf("RenderManifest() = %v, want %v", result, tt.expected)
            }
        })
    }
}
```

## Integration Tests

### End-to-End Tests

Test complete workflows:

```go
// integration_test.go
func TestIntegrationBasicWorkflow(t *testing.T) {
    // Create temporary directory
    tempDir := t.TempDir()
    
    // Set up test configuration
    configDir := filepath.Join(tempDir, "config")
    manifestsDir := filepath.Join(tempDir, "manifests")
    
    os.MkdirAll(configDir, 0755)
    os.MkdirAll(manifestsDir, 0755)
    
    // Create test files
    createTestConfig(t, configDir)
    createTestManifest(t, manifestsDir)
    
    // Test frank apply
    cmd := exec.Command("frank", "apply", "--yes")
    cmd.Dir = tempDir
    output, err := cmd.CombinedOutput()
    
    if err != nil {
        t.Fatalf("frank apply failed: %v\nOutput: %s", err, output)
    }
    
    // Verify resources were created
    // (This would require kubectl access)
}
```

### Template Integration Tests

Test Jinja template workflows:

```go
func TestIntegrationTemplateWorkflow(t *testing.T) {
    tempDir := t.TempDir()
    
    // Set up test configuration with Jinja template
    configDir := filepath.Join(tempDir, "config")
    manifestsDir := filepath.Join(tempDir, "manifests")
    
    os.MkdirAll(configDir, 0755)
    os.MkdirAll(manifestsDir, 0755)
    
    // Create test files
    createTestConfigWithTemplate(t, configDir)
    createTestJinjaTemplate(t, manifestsDir)
    
    // Test frank apply with template
    cmd := exec.Command("frank", "apply", "--yes")
    cmd.Dir = tempDir
    output, err := cmd.CombinedOutput()
    
    if err != nil {
        t.Fatalf("frank apply with template failed: %v\nOutput: %s", err, output)
    }
}
```

## Benchmark Tests

### Performance Tests

Test performance of critical functions:

```go
// benchmark_test.go
func BenchmarkGenerateStackName(b *testing.B) {
    for i := 0; i < b.N; i++ {
        GenerateStackName("myapp", "dev", "web")
    }
}

func BenchmarkRenderTemplate(b *testing.B) {
    renderer := NewRenderer()
    template := "Hello {{ name }}! Your version is {{ version }}."
    context := map[string]interface{}{
        "name":    "World",
        "version": "1.2.3",
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        renderer.RenderManifest(template, context)
    }
}

func BenchmarkConfigLoading(b *testing.B) {
    for i := 0; i < b.N; i++ {
        LoadConfig()
    }
}
```

### Run Benchmarks

```bash
# Run all benchmarks
go test -bench=. ./...

# Run specific benchmark
go test -bench=BenchmarkGenerateStackName ./pkg/stack

# Run benchmarks with memory profiling
go test -bench=. -benchmem ./...

# Run benchmarks with CPU profiling
go test -bench=. -cpuprofile=cpu.prof ./...
go tool pprof cpu.prof
```

## Test Utilities

### Test Helpers

Create helper functions for common test operations:

```go
// test_helpers.go
func createTestConfig(t *testing.T, configDir string) {
    configContent := `
context: test-cluster
project_code: testapp
namespace: test-namespace
`
    
    configFile := filepath.Join(configDir, "config.yaml")
    err := os.WriteFile(configFile, []byte(configContent), 0644)
    if err != nil {
        t.Fatalf("Failed to create test config: %v", err)
    }
}

func createTestManifest(t *testing.T, manifestsDir string) {
    manifestContent := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test-app
  template:
    metadata:
      labels:
        app: test-app
    spec:
      containers:
      - name: test-app
        image: nginx:alpine
        ports:
        - containerPort: 80
`
    
    manifestFile := filepath.Join(manifestsDir, "test-deployment.yaml")
    err := os.WriteFile(manifestFile, []byte(manifestContent), 0644)
    if err != nil {
        t.Fatalf("Failed to create test manifest: %v", err)
    }
}
```

### Mock Objects

Create mocks for external dependencies:

```go
// mocks.go
type MockKubernetesClient struct {
    deployments map[string]interface{}
}

func (m *MockKubernetesClient) CreateDeployment(deployment interface{}) error {
    // Mock implementation
    return nil
}

func (m *MockKubernetesClient) UpdateDeployment(deployment interface{}) error {
    // Mock implementation
    return nil
}
```

## Test Coverage

### Coverage Analysis

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage in terminal
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Open coverage report
open coverage.html
```

### Coverage Goals

Aim for high test coverage:

- **Unit tests**: > 80% coverage
- **Integration tests**: > 60% coverage
- **Critical functions**: > 90% coverage

## Test Best Practices

### 1. Use Table-Driven Tests

```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"case1", "input1", "expected1"},
        {"case2", "input2", "expected2"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Function(tt.input)
            if result != tt.expected {
                t.Errorf("Function() = %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### 2. Use Test Helpers

```go
func TestWithHelper(t *testing.T) {
    tempDir := createTempDir(t)
    defer os.RemoveAll(tempDir)
    
    // Test implementation
}
```

### 3. Test Error Cases

```go
func TestErrorCases(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid input", "valid", false},
        {"invalid input", "invalid", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := Function(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Function() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### 4. Use Subtests

```go
func TestComplexFunction(t *testing.T) {
    t.Run("valid cases", func(t *testing.T) {
        // Test valid cases
    })
    
    t.Run("error cases", func(t *testing.T) {
        // Test error cases
    })
    
    t.Run("edge cases", func(t *testing.T) {
        // Test edge cases
    })
}
```

## Continuous Integration

### GitHub Actions

```yaml
# .github/workflows/test.yml
name: Test

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.25'
    
    - name: Run tests
      run: go test -v ./...
    
    - name: Run tests with coverage
      run: go test -coverprofile=coverage.out ./...
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
```

### GitLab CI

```yaml
# .gitlab-ci.yml
test:
  stage: test
  image: golang:1.25-alpine
  script:
    - go test -v ./...
    - go test -coverprofile=coverage.out ./...
  coverage: '/coverage: \d+\.\d+%/'
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage.xml
```

## Troubleshooting

### Common Test Issues

**"no test files"**
- Ensure test files end with `_test.go`
- Check if test files are in the correct package

**"test timeout"**
- Increase timeout: `go test -timeout 30s ./...`
- Check for infinite loops in tests

**"race condition"**
- Run with race detection: `go test -race ./...`
- Fix race conditions in code

**"coverage issues"**
- Check if all code paths are tested
- Add tests for missing coverage

### Debug Tests

```bash
# Run tests with verbose output
go test -v ./...

# Run tests with race detection
go test -race ./...

# Run tests with timeout
go test -timeout 30s ./...

# Run specific test with debug
go test -v -run TestSpecificFunction ./pkg/stack
```

## Next Steps

- [Building](building.md) - Learn about building Frank CLI
- [Contributing](contributing.md) - Contribute to Frank CLI development
- [Architecture](architecture.md) - Understand Frank CLI architecture
