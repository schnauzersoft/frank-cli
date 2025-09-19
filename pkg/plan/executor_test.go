/*
Copyright © 2025 Ben Sapp ya.bsapp.ru
*/

package plan

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestExecutor_NewExecutor(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Create a logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	// Test creating executor
	executor, err := NewExecutor(configDir, logger)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	if executor == nil {
		t.Fatal("Expected executor to be created")
	}

	if executor.configDir != configDir {
		t.Errorf("Expected configDir %s, got %s", configDir, executor.configDir)
	}

	if executor.logger == nil {
		t.Error("Expected logger to be set")
	}

	if executor.k8sDeployer == nil {
		t.Error("Expected k8sDeployer to be set")
	}

	if executor.templateRenderer == nil {
		t.Error("Expected templateRenderer to be set")
	}

	if executor.planner == nil {
		t.Error("Expected planner to be set")
	}
}

func TestExecutor_isConfigFile(t *testing.T) {
	executor := createTestExecutor(t)

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "yaml config file",
			path:     "config.yaml",
			expected: true,
		},
		{
			name:     "yml config file",
			path:     "config.yml",
			expected: true,
		},
		{
			name:     "jinja template file",
			path:     "template.jinja",
			expected: false,
		},
		{
			name:     "hcl template file",
			path:     "template.hcl",
			expected: false,
		},
		{
			name:     "non-yaml file",
			path:     "config.txt",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.isConfigFile(tt.path)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for path %s", tt.expected, result, tt.path)
			}
		})
	}
}

func TestExecutor_isYAMLFile(t *testing.T) {
	executor := createTestExecutor(t)

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "yaml file",
			path:     "test.yaml",
			expected: true,
		},
		{
			name:     "yml file",
			path:     "test.yml",
			expected: true,
		},
		{
			name:     "YAML file (uppercase)",
			path:     "test.YAML",
			expected: true,
		},
		{
			name:     "YML file (uppercase)",
			path:     "test.YML",
			expected: true,
		},
		{
			name:     "non-yaml file",
			path:     "test.txt",
			expected: false,
		},
		{
			name:     "jinja file",
			path:     "test.jinja",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.isYAMLFile(tt.path)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for path %s", tt.expected, result, tt.path)
			}
		})
	}
}

func TestExecutor_isConfigYAML(t *testing.T) {
	executor := createTestExecutor(t)

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "yaml config file",
			path:     "config.yaml",
			expected: true,
		},
		{
			name:     "jinja template file",
			path:     "template.jinja",
			expected: false,
		},
		{
			name:     "hcl template file",
			path:     "template.hcl",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.isConfigYAML(tt.path)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for path %s", tt.expected, result, tt.path)
			}
		})
	}
}

func TestExecutor_matchesStackFilter(t *testing.T) {
	executor := createTestExecutor(t)

	tests := []struct {
		name      string
		stackName string
		filter    string
		expected  bool
	}{
		{
			name:      "exact match",
			stackName: "frank-dev-app",
			filter:    "frank-dev-app",
			expected:  true,
		},
		{
			name:      "partial match",
			stackName: "frank-dev-app",
			filter:    "dev",
			expected:  true,
		},
		{
			name:      "no match",
			stackName: "frank-dev-app",
			filter:    "prod",
			expected:  false,
		},
		{
			name:      "filter with config prefix",
			stackName: "frank-dev-app",
			filter:    "config/dev",
			expected:  true,
		},
		{
			name:      "filter with yaml extension",
			stackName: "frank-dev-app",
			filter:    "dev.yaml",
			expected:  true,
		},
		{
			name:      "filter with yml extension",
			stackName: "frank-dev-app",
			filter:    "dev.yml",
			expected:  true,
		},
		{
			name:      "empty filter",
			stackName: "frank-dev-app",
			filter:    "",
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.matchesStackFilter(tt.stackName, tt.filter)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for stack %s and filter %s", tt.expected, result, tt.stackName, tt.filter)
			}
		})
	}
}

func TestExecutor_readManifestConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Create a test config file
	configFile := filepath.Join(configDir, "test.yaml")
	configContent := `manifest: test-deployment.yaml
timeout: 300
version: "1.0.0"
vars:
  replicas: 3
  environment: test`

	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	executor := createTestExecutor(t)
	executor.configDir = configDir

	// Test reading config
	config, err := executor.readManifestConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to read manifest config: %v", err)
	}

	if config.Manifest != "test-deployment.yaml" {
		t.Errorf("Expected manifest 'test-deployment.yaml', got '%s'", config.Manifest)
	}

	if config.Timeout != 300 {
		t.Errorf("Expected timeout 300, got %d", config.Timeout)
	}

	if config.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", config.Version)
	}

	if config.Vars == nil {
		t.Error("Expected vars to be set")
	}

	if config.Vars["replicas"] != 3 {
		t.Errorf("Expected replicas 3, got %v", config.Vars["replicas"])
	}

	if config.Vars["environment"] != "test" {
		t.Errorf("Expected environment 'test', got %v", config.Vars["environment"])
	}
}

func TestExecutor_findManifestFile(t *testing.T) {
	// Create a temporary directory structure
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	manifestsDir := filepath.Join(tempDir, "manifests")

	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	err = os.MkdirAll(manifestsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create manifests directory: %v", err)
	}

	// Create a test manifest file
	manifestFile := filepath.Join(manifestsDir, "test-deployment.yaml")
	manifestContent := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment`

	err = os.WriteFile(manifestFile, []byte(manifestContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write manifest file: %v", err)
	}

	executor := createTestExecutor(t)
	executor.configDir = configDir

	// Test finding manifest file
	foundPath, err := executor.findManifestFile("test-deployment.yaml")
	if err != nil {
		t.Fatalf("Failed to find manifest file: %v", err)
	}

	if foundPath != manifestFile {
		t.Errorf("Expected path %s, got %s", manifestFile, foundPath)
	}

	// Test finding non-existent file
	_, err = executor.findManifestFile("non-existent.yaml")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestExecutor_filterConfigFilesByStack(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Create test config files
	configFiles := []string{
		filepath.Join(configDir, "dev-app.yaml"),
		filepath.Join(configDir, "dev-db.yaml"),
		filepath.Join(configDir, "prod-app.yaml"),
		filepath.Join(configDir, "prod-db.yaml"),
	}

	for _, file := range configFiles {
		content := `manifest: test.yaml`
		err = os.WriteFile(file, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write config file %s: %v", file, err)
		}
	}

	executor := createTestExecutor(t)

	// Test filtering by "dev"
	filtered := executor.filterConfigFilesByStack(configFiles, "dev")
	if len(filtered) != 2 {
		t.Errorf("Expected 2 files, got %d", len(filtered))
	}

	// Test filtering by "prod"
	filtered = executor.filterConfigFilesByStack(configFiles, "prod")
	if len(filtered) != 2 {
		t.Errorf("Expected 2 files, got %d", len(filtered))
	}

	// Test filtering by "app"
	filtered = executor.filterConfigFilesByStack(configFiles, "app")
	if len(filtered) != 2 {
		t.Errorf("Expected 2 files, got %d", len(filtered))
	}

	// Test filtering by non-existent pattern
	filtered = executor.filterConfigFilesByStack(configFiles, "nonexistent")
	if len(filtered) != 0 {
		t.Errorf("Expected 0 files, got %d", len(filtered))
	}
}

// Helper function to create a test executor
func createTestExecutor(t *testing.T) *Executor {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Create a logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	// Create executor
	executor, err := NewExecutor(configDir, logger)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	return executor
}
