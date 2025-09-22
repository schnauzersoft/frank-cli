package stack

import (
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkGenerateStackName(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateStackName("frank", "dev", "app.yaml")
	}
}

func BenchmarkExtractAppNameFromFilename(b *testing.B) {
	filename := "/path/to/deployment.jinja"
	for i := 0; i < b.N; i++ {
		extractAppNameFromFilename(filename)
	}
}

func BenchmarkReadConfigForFile(b *testing.B) {
	// Create a temporary directory for benchmark
	tempDir := b.TempDir()

	// Create base config
	baseConfig := `context: test
project_code: frank
version: 1.0.0`
	err := os.WriteFile(filepath.Join(tempDir, "config.yaml"), []byte(baseConfig), 0o644)
	if err != nil {
		b.Fatalf("Failed to create base config: %v", err)
	}

	// Create child config
	childDir := filepath.Join(tempDir, "dev")
	err = os.Mkdir(childDir, 0o755)
	if err != nil {
		b.Fatalf("Failed to create child directory: %v", err)
	}

	childConfig := `context: dev
namespace: dev-namespace
app: web
version: 2.0.0`
	err = os.WriteFile(filepath.Join(childDir, "config.yaml"), []byte(childConfig), 0o644)
	if err != nil {
		b.Fatalf("Failed to create child config: %v", err)
	}

	configFile := filepath.Join(childDir, "config.yaml")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ReadConfigForFile(configFile)
		if err != nil {
			b.Fatalf("ReadConfigForFile failed: %v", err)
		}
	}
}

func BenchmarkGetStackInfo(b *testing.B) {
	// Create a temporary directory for benchmark
	tempDir := b.TempDir()

	// Create base config
	baseConfig := `context: test
project_code: frank
version: 1.0.0`
	err := os.WriteFile(filepath.Join(tempDir, "config.yaml"), []byte(baseConfig), 0o644)
	if err != nil {
		b.Fatalf("Failed to create base config: %v", err)
	}

	// Create web config
	webDir := filepath.Join(tempDir, "web")
	err = os.Mkdir(webDir, 0o755)
	if err != nil {
		b.Fatalf("Failed to create web directory: %v", err)
	}

	webConfig := `project_code: frank
app: web
version: 2.0.0`
	err = os.WriteFile(filepath.Join(webDir, "config.yaml"), []byte(webConfig), 0o644)
	if err != nil {
		b.Fatalf("Failed to create web config: %v", err)
	}

	configFile := filepath.Join(webDir, "config.yaml")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetStackInfo(configFile)
		if err != nil {
			b.Fatalf("GetStackInfo failed: %v", err)
		}
	}
}
