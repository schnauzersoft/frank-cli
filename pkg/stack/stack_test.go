package stack

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateStackName(t *testing.T) {
	tests := []struct {
		name        string
		projectCode string
		context     string
		configFile  string
		expected    string
	}{
		{
			name:        "basic stack name",
			projectCode: "frank",
			context:     "dev",
			configFile:  "app.yaml",
			expected:    "frank-dev-app",
		},
		{
			name:        "stack name with subdirectory",
			projectCode: "frank",
			context:     "prod",
			configFile:  "config/prod/web.yaml",
			expected:    "frank-prod-web",
		},
		{
			name:        "stack name with nested path",
			projectCode: "myapp",
			context:     "staging",
			configFile:  "config/staging/api/service.yaml",
			expected:    "myapp-staging-service",
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

func TestExtractAppNameFromFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "yaml file",
			filename: "app.yaml",
			expected: "app",
		},
		{
			name:     "yml file",
			filename: "service.yml",
			expected: "service",
		},
		{
			name:     "jinja file",
			filename: "deployment.jinja",
			expected: "deployment",
		},
		{
			name:     "j2 file",
			filename: "config.j2",
			expected: "config",
		},
		{
			name:     "file with path",
			filename: "/path/to/app.yaml",
			expected: "app",
		},
		{
			name:     "file with multiple dots",
			filename: "app-v1.2.3.yaml",
			expected: "app-v1.2.3",
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

func TestReadConfigForFile(t *testing.T) {
	// Create test directory structure
	tempDir, childDir, apiDir := setupTestConfigStructure(t)

	tests := []struct {
		name           string
		configFile     string
		expectedConfig *Config
		expectError    bool
	}{
		{
			name:       "base config only",
			configFile: filepath.Join(tempDir, "config.yaml"),
			expectedConfig: &Config{
				Context:     "base",
				ProjectCode: "test",
				Version:     "1.0.0",
				Namespace:   "default",
			},
			expectError: false,
		},
		{
			name:       "child config with inheritance",
			configFile: filepath.Join(childDir, "config.yaml"),
			expectedConfig: &Config{
				Context:     "dev",
				ProjectCode: "test", // inherited from parent
				Version:     "2.0.0",
				Namespace:   "dev-namespace",
				App:         "web",
			},
			expectError: false,
		},
		{
			name:       "api config with inheritance",
			configFile: filepath.Join(apiDir, "config.yaml"),
			expectedConfig: &Config{
				Context:     "dev",           // inherited from child
				ProjectCode: "test",          // inherited from base
				Version:     "3.0.0",         // specified in api config
				Namespace:   "dev-namespace", // inherited from child
				App:         "api",           // specified in api config
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ReadConfigForFile(tt.configFile)
			validateConfigResult(t, config, err, tt.expectedConfig, tt.expectError)
		})
	}
}

// setupTestConfigStructure creates the test directory structure and config files
func setupTestConfigStructure(t *testing.T) (string, string, string) {
	tempDir := t.TempDir()

	// Create base config
	baseConfig := `context: base
project_code: test
version: 1.0.0
namespace: default`
	err := os.WriteFile(filepath.Join(tempDir, "config.yaml"), []byte(baseConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create base config: %v", err)
	}

	// Create child config
	childDir := filepath.Join(tempDir, "dev")
	err = os.Mkdir(childDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create child directory: %v", err)
	}

	childConfig := `context: dev
namespace: dev-namespace
app: web
version: 2.0.0`
	err = os.WriteFile(filepath.Join(childDir, "config.yaml"), []byte(childConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create child config: %v", err)
	}

	// Create another child config in a subdirectory
	apiDir := filepath.Join(childDir, "api")
	err = os.Mkdir(apiDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create api directory: %v", err)
	}

	apiConfig := `project_code: test
app: api
version: 3.0.0`
	err = os.WriteFile(filepath.Join(apiDir, "config.yaml"), []byte(apiConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create api config: %v", err)
	}

	return tempDir, childDir, apiDir
}

// validateConfigResult validates the result of ReadConfigForFile
func validateConfigResult(t *testing.T, config *Config, err error, expectedConfig *Config, expectError bool) {
	if expectError {
		if err == nil {
			t.Errorf("ReadConfigForFile() expected error but got none")
		}
		return
	}

	if err != nil {
		t.Errorf("ReadConfigForFile() unexpected error: %v", err)
		return
	}

	if config.Context != expectedConfig.Context {
		t.Errorf("Context = %v, want %v", config.Context, expectedConfig.Context)
	}
	if config.ProjectCode != expectedConfig.ProjectCode {
		t.Errorf("ProjectCode = %v, want %v", config.ProjectCode, expectedConfig.ProjectCode)
	}
	if config.Version != expectedConfig.Version {
		t.Errorf("Version = %v, want %v", config.Version, expectedConfig.Version)
	}
	if config.Namespace != expectedConfig.Namespace {
		t.Errorf("Namespace = %v, want %v", config.Namespace, expectedConfig.Namespace)
	}
	if config.App != expectedConfig.App {
		t.Errorf("App = %v, want %v", config.App, expectedConfig.App)
	}
}

func TestGetStackInfo(t *testing.T) {
	// Create a temporary directory structure
	tempDir := t.TempDir()

	// Create base config
	baseConfig := `context: test
project_code: frank
version: 1.0.0`
	err := os.WriteFile(filepath.Join(tempDir, "config.yaml"), []byte(baseConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create base config: %v", err)
	}

	// Create web config in subdirectory
	webDir := filepath.Join(tempDir, "web")
	err = os.Mkdir(webDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create web directory: %v", err)
	}

	webConfig := `project_code: frank
app: web
version: 2.0.0`
	err = os.WriteFile(filepath.Join(webDir, "config.yaml"), []byte(webConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create web config: %v", err)
	}

	stackInfo, err := GetStackInfo(filepath.Join(webDir, "config.yaml"))
	if err != nil {
		t.Fatalf("GetStackInfo() unexpected error: %v", err)
	}

	expectedStackName := "frank-test-config"
	if stackInfo.Name != expectedStackName {
		t.Errorf("StackName = %v, want %v", stackInfo.Name, expectedStackName)
	}

	if stackInfo.Context != "test" {
		t.Errorf("Context = %v, want test", stackInfo.Context)
	}

	if stackInfo.ProjectCode != "frank" {
		t.Errorf("ProjectCode = %v, want frank", stackInfo.ProjectCode)
	}

	if stackInfo.App != "web" {
		t.Errorf("App = %v, want web", stackInfo.App)
	}

	if stackInfo.Version != "2.0.0" {
		t.Errorf("Version = %v, want 2.0.0", stackInfo.Version)
	}
}

func TestMergeConfigs(t *testing.T) {
	parent := &Config{
		Context:     "base",
		ProjectCode: "test",
		Version:     "1.0.0",
		Namespace:   "default",
		App:         "parent-app",
	}

	child := &Config{
		Context:   "dev",
		Version:   "2.0.0",
		Namespace: "dev-namespace",
		App:       "child-app",
	}

	merged := mergeConfigs(parent, child)

	// Child should override parent values
	if merged.Context != "dev" {
		t.Errorf("Context = %v, want dev", merged.Context)
	}
	if merged.ProjectCode != "test" {
		t.Errorf("ProjectCode = %v, want test", merged.ProjectCode)
	}
	if merged.Version != "2.0.0" {
		t.Errorf("Version = %v, want 2.0.0", merged.Version)
	}
	if merged.Namespace != "dev-namespace" {
		t.Errorf("Namespace = %v, want dev-namespace", merged.Namespace)
	}
	if merged.App != "child-app" {
		t.Errorf("App = %v, want child-app", merged.App)
	}
}
