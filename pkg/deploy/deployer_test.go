package deploy

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/schnauzersoft/frank-cli/pkg/kubernetes"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestFilterConfigFilesByStack(t *testing.T) {
	// Create a temporary directory structure
	tempDir := t.TempDir()

	// Create config files
	createTestConfigFiles(t, tempDir)

	deployer := &Deployer{
		configDir: tempDir,
		logger:    slog.Default(),
	}

	tests := createFilterTestCases(tempDir)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deployer.filterConfigFilesByStack(tt.configFiles, tt.stackFilter)
			validateFilterResult(t, result, tt.expected)
		})
	}
}

// createTestConfigFiles creates the test directory structure and config files.
func createTestConfigFiles(t *testing.T, tempDir string) []string {
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

		err := os.MkdirAll(dir, 0o755)
		if err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}

		err = os.WriteFile(file, []byte("manifest: test.yaml"), 0o600)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", file, err)
		}
	}

	return configFiles
}

// createFilterTestCases creates test cases for filtering.
func createFilterTestCases(tempDir string) []struct {
	name        string
	configFiles []string
	stackFilter string
	expected    []string
} {
	// Create the config files list for all tests
	configFiles := []string{
		filepath.Join(tempDir, "app.yaml"),
		filepath.Join(tempDir, "api.yaml"),
		filepath.Join(tempDir, "web.yaml"),
		filepath.Join(tempDir, "dev", "app.yaml"),
		filepath.Join(tempDir, "dev", "api.yaml"),
		filepath.Join(tempDir, "prod", "web.yaml"),
		filepath.Join(tempDir, "prod", "api.yaml"),
	}

	return []struct {
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
}

// validateFilterResult validates the result of filtering.
func validateFilterResult(t *testing.T, result, expected []string) {
	if len(result) != len(expected) {
		t.Errorf("filterConfigFilesByStack() returned %d files, want %d", len(result), len(expected))

		return
	}

	// Convert to map for easier comparison
	resultMap := make(map[string]bool)
	for _, file := range result {
		resultMap[file] = true
	}

	for _, expectedFile := range expected {
		if !resultMap[expectedFile] {
			t.Errorf("Expected file %s not found in result", expectedFile)
		}
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

	tests := createNamespaceValidationTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary manifest file
			manifestPath := filepath.Join(tempDir, "test-manifest.yaml")

			err := os.WriteFile(manifestPath, []byte(tt.manifestContent), 0o600)
			if err != nil {
				t.Fatalf("Failed to create manifest file: %v", err)
			}

			// Test validation
			err = deployer.validateNamespaceConfiguration(manifestPath, tt.configNamespace)
			validateNamespaceTestResult(t, err, tt.expectError, tt.errorContains)
		})
	}
}

// createNamespaceValidationTestCases creates test cases for namespace validation.
func createNamespaceValidationTestCases() []struct {
	name            string
	manifestContent string
	configNamespace string
	expectError     bool
	errorContains   string
} {
	return []struct {
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
			errorContains:   "invalid YAML format in manifest",
		},
	}
}

// validateNamespaceTestResult validates the result of namespace validation test.
func validateNamespaceTestResult(t *testing.T, err error, expectError bool, errorContains string) {
	if expectError {
		if err == nil {
			t.Errorf("validateNamespaceConfiguration() expected error but got none")

			return
		}

		if errorContains != "" && !contains(err.Error(), errorContains) {
			t.Errorf("validateNamespaceConfiguration() error = %v, want error containing %s", err, errorContains)
		}
	} else if err != nil {
		t.Errorf("validateNamespaceConfiguration() unexpected error: %v", err)
	}
}

// Helper function to check if a string contains a substring.
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

// TestDeployResultErrorHandling tests that DeployResult errors are properly handled.
func TestDeployResultErrorHandling(t *testing.T) {
	deployer := &Deployer{
		logger: slog.Default(),
	}

	// Test successful result
	successResult := &kubernetes.DeployResult{
		Resource: &unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]any{
					"name":      "test-app",
					"namespace": "test-namespace",
				},
			},
		},
		Operation: "created",
		Status:    "Available",
		Error:     nil,
	}

	response := deployer.formatResponse(successResult)

	expectedSuccess := "Applied apps/v1/Deployment: test-app in namespace test-namespace (operation: created, status: Available)"
	if response != expectedSuccess {
		t.Errorf("Expected success response %q, got %q", expectedSuccess, response)
	}

	// Test failed result
	failedResult := &kubernetes.DeployResult{
		Resource: &unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]any{
					"name":      "test-app",
					"namespace": "test-namespace",
				},
			},
		},
		Operation: "failed",
		Status:    "failed",
		Error:     errors.New("namespaces \"test-namespace\" not found"),
	}

	response = deployer.formatResponse(failedResult)

	expectedFailure := "Apply failed: namespaces \"test-namespace\" not found"
	if response != expectedFailure {
		t.Errorf("Expected failure response %q, got %q", expectedFailure, response)
	}
}

// TestFormatResponseSuccess tests successful formatResponse.
func TestFormatResponseSuccess(t *testing.T) {
	deployer := &Deployer{
		logger: slog.Default(),
	}

	result := &kubernetes.DeployResult{
		Resource: &unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]any{
					"name":      "test-app",
					"namespace": "test-namespace",
				},
			},
		},
		Operation: "created",
		Status:    "Available",
		Error:     nil,
	}

	response := deployer.formatResponse(result)

	expected := "Applied apps/v1/Deployment: test-app in namespace test-namespace (operation: created, status: Available)"
	if response != expected {
		t.Errorf("Expected response %q, got %q", expected, response)
	}
}

// TestFormatResponseError tests error formatResponse.
func TestFormatResponseError(t *testing.T) {
	deployer := &Deployer{
		logger: slog.Default(),
	}

	result := &kubernetes.DeployResult{
		Resource: &unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": "v1",
				"kind":       "Service",
				"metadata": map[string]any{
					"name":      "test-service",
					"namespace": "test-namespace",
				},
			},
		},
		Operation: "failed",
		Status:    "failed",
		Error:     errors.New("service already exists"),
	}

	response := deployer.formatResponse(result)

	expected := "Apply failed: service already exists"
	if response != expected {
		t.Errorf("Expected response %q, got %q", expected, response)
	}
}
