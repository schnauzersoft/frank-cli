package deploy

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestFilterConfigFilesByStack(t *testing.T) {
	// Create a temporary directory structure
	tempDir := t.TempDir()

	// Create config files
	configFiles := []string{
		filepath.Join(tempDir, "app.yaml"),
		filepath.Join(tempDir, "api.yaml"),
		filepath.Join(tempDir, "web.yaml"),
		filepath.Join(tempDir, "dev", "app.yaml"),
		filepath.Join(tempDir, "dev", "api.yaml"),
		filepath.Join(tempDir, "prod", "web.yaml"),
		filepath.Join(tempDir, "prod", "api.yaml"),
	}

	// Create the directory structure
	for _, file := range configFiles {
		dir := filepath.Dir(file)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(file, []byte("manifest: test.yaml"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", file, err)
		}
	}

	deployer := &Deployer{
		configDir: tempDir,
		logger:    slog.Default(),
	}

	tests := []struct {
		name        string
		configFiles []string
		stackFilter string
		expected    []string
	}{
		{
			name:        "no filter should return no files",
			configFiles: configFiles,
			stackFilter: "",
			expected:    []string{},
		},
		{
			name:        "exact match should return single file",
			configFiles: configFiles,
			stackFilter: "app",
			expected: []string{
				filepath.Join(tempDir, "app.yaml"),
			},
		},
		{
			name:        "partial match should return multiple files",
			configFiles: configFiles,
			stackFilter: "api",
			expected: []string{
				filepath.Join(tempDir, "api.yaml"),
			},
		},
		{
			name:        "directory filter should return files in directory",
			configFiles: configFiles,
			stackFilter: "dev",
			expected: []string{
				filepath.Join(tempDir, "dev", "app.yaml"),
				filepath.Join(tempDir, "dev", "api.yaml"),
			},
		},
		{
			name:        "path filter should return files matching path",
			configFiles: configFiles,
			stackFilter: "dev/app",
			expected: []string{
				filepath.Join(tempDir, "dev", "app.yaml"),
			},
		},
		{
			name:        "non-matching filter should return empty",
			configFiles: configFiles,
			stackFilter: "nonexistent",
			expected:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deployer.filterConfigFilesByStack(tt.configFiles, tt.stackFilter)

			if len(result) != len(tt.expected) {
				t.Errorf("filterConfigFilesByStack() returned %d files, want %d", len(result), len(tt.expected))
				return
			}

			// Convert to map for easier comparison
			resultMap := make(map[string]bool)
			for _, file := range result {
				resultMap[file] = true
			}

			for _, expectedFile := range tt.expected {
				if !resultMap[expectedFile] {
					t.Errorf("Expected file %s not found in result", expectedFile)
				}
			}
		})
	}
}

func TestMatchesStackFilter(t *testing.T) {
	deployer := &Deployer{
		logger: slog.Default(),
	}

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
			name:      "prefix match with dash",
			stackName: "frank-dev-app",
			filter:    "frank-dev",
			expected:  true,
		},
		{
			name:      "prefix match without dash",
			stackName: "frank-dev-app",
			filter:    "frank",
			expected:  true,
		},
		{
			name:      "path pattern match",
			stackName: "dev-app",
			filter:    "dev/app",
			expected:  true,
		},
		{
			name:      "no match",
			stackName: "frank-dev-app",
			filter:    "prod",
			expected:  false,
		},
		{
			name:      "partial match in middle",
			stackName: "frank-dev-app",
			filter:    "dev-app",
			expected:  false,
		},
		{
			name:      "empty filter should not match",
			stackName: "frank-dev-app",
			filter:    "",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deployer.matchesStackFilter(tt.stackName, tt.filter)
			if result != tt.expected {
				t.Errorf("matchesStackFilter(%s, %s) = %v, want %v", tt.stackName, tt.filter, result, tt.expected)
			}
		})
	}
}

func TestValidateNamespaceConfiguration(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	deployer := &Deployer{
		logger: slog.Default(),
	}

	tests := []struct {
		name            string
		manifestContent string
		configNamespace string
		expectError     bool
		errorContains   string
	}{
		{
			name: "no namespace conflict - config only",
			manifestContent: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  replicas: 3`,
			configNamespace: "dev-namespace",
			expectError:     false,
		},
		{
			name: "no namespace conflict - manifest only",
			manifestContent: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
  namespace: manifest-namespace
spec:
  replicas: 3`,
			configNamespace: "",
			expectError:     false,
		},
		{
			name: "no namespace conflict - both default",
			manifestContent: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  replicas: 3`,
			configNamespace: "",
			expectError:     false,
		},
		{
			name: "namespace conflict - both specified",
			manifestContent: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
  namespace: manifest-namespace
spec:
  replicas: 3`,
			configNamespace: "config-namespace",
			expectError:     true,
			errorContains:   "namespace specified in both config file",
		},
		{
			name: "no conflict - manifest namespace is empty string",
			manifestContent: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
  namespace: ""
spec:
  replicas: 3`,
			configNamespace: "config-namespace",
			expectError:     false,
		},
		{
			name: "invalid YAML should error",
			manifestContent: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
  namespace: manifest-namespace
spec:
  replicas: 3
  invalid: [`,
			configNamespace: "config-namespace",
			expectError:     true,
			errorContains:   "failed to parse manifest YAML",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary manifest file
			manifestPath := filepath.Join(tempDir, "test-manifest.yaml")
			err := os.WriteFile(manifestPath, []byte(tt.manifestContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create manifest file: %v", err)
			}

			// Test validation
			err = deployer.validateNamespaceConfiguration(manifestPath, tt.configNamespace)

			if tt.expectError {
				if err == nil {
					t.Errorf("validateNamespaceConfiguration() expected error but got none")
					return
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("validateNamespaceConfiguration() error = %v, want error containing %s", err, tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("validateNamespaceConfiguration() unexpected error: %v", err)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsInMiddle(s, substr))))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
