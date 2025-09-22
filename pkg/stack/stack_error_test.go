package stack

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadConfigForFileErrors(t *testing.T) {
	tests := createErrorTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			err := tt.setup(tempDir)
			if err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			configFile := filepath.Join(tempDir, "config.yaml")
			_, err = ReadConfigForFile(configFile)

			validateErrorTestResult(t, err, tt.expectError)
		})
	}
}

// createErrorTestCases creates test cases for error scenarios
func createErrorTestCases() []struct {
	name        string
	setup       func(string) error
	expectError bool
} {
	return []struct {
		name        string
		setup       func(string) error
		expectError bool
	}{
		{
			name: "non-existent file",
			setup: func(tempDir string) error {
				// Don't create any files
				return nil
			},
			expectError: true,
		},
		{
			name: "invalid YAML",
			setup: func(tempDir string) error {
				invalidYAML := `context: test
project_code: frank
invalid: [`
				return os.WriteFile(filepath.Join(tempDir, "config.yaml"), []byte(invalidYAML), 0o644)
			},
			expectError: true,
		},
		{
			name: "missing project_code",
			setup: func(tempDir string) error {
				config := `context: test
version: 1.0.0`
				return os.WriteFile(filepath.Join(tempDir, "config.yaml"), []byte(config), 0o644)
			},
			expectError: true,
		},
		{
			name: "valid config",
			setup: func(tempDir string) error {
				config := `context: test
project_code: frank
version: 1.0.0`
				return os.WriteFile(filepath.Join(tempDir, "config.yaml"), []byte(config), 0o644)
			},
			expectError: false,
		},
	}
}

// validateErrorTestResult validates the result of an error test
func validateErrorTestResult(t *testing.T, err error, expectError bool) {
	if expectError {
		if err == nil {
			t.Errorf("Expected error but got none")
		}
	} else {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}
}

func TestGenerateStackNameEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		projectCode string
		context     string
		configFile  string
		expected    string
	}{
		{
			name:        "empty strings",
			projectCode: "",
			context:     "",
			configFile:  "app.yaml",
			expected:    "--app",
		},
		{
			name:        "special characters",
			projectCode: "my_project",
			context:     "dev_env",
			configFile:  "app_name.yaml",
			expected:    "my-project-dev-env-app-name",
		},
		{
			name:        "uppercase",
			projectCode: "FRANK",
			context:     "PROD",
			configFile:  "APP.yaml",
			expected:    "frank-prod-app",
		},
		{
			name:        "nested path",
			projectCode: "test",
			context:     "staging",
			configFile:  "/very/deep/nested/path/app.yaml",
			expected:    "test-staging-app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateStackName(tt.projectCode, tt.context, tt.configFile)
			if result != tt.expected {
				t.Errorf("GenerateStackName() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExtractAppNameFromFilenameEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "empty filename",
			filename: "",
			expected: ".",
		},
		{
			name:     "no extension",
			filename: "app",
			expected: "app",
		},
		{
			name:     "multiple extensions",
			filename: "app.backup.yaml",
			expected: "app.backup",
		},
		{
			name:     "hidden file",
			filename: ".app.yaml",
			expected: ".app",
		},
		{
			name:     "path with no file",
			filename: "/path/to/",
			expected: "to",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractAppNameFromFilename(tt.filename)
			if result != tt.expected {
				t.Errorf("extractAppNameFromFilename() = %v, want %v", result, tt.expected)
			}
		})
	}
}
