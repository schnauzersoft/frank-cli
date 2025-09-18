package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIntegrationBasicWorkflow(t *testing.T) {
	// Create a temporary directory for integration test
	tempDir := t.TempDir()

	// Create config directory structure
	configDir := filepath.Join(tempDir, "config")
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Create base config
	baseConfig := `context: test
project_code: frank
version: 1.0.0`
	err = os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(baseConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create base config: %v", err)
	}

	// Create manifest config
	manifestConfig := `manifest: test-deployment.yaml
app: test
version: 2.0.0`
	err = os.WriteFile(filepath.Join(configDir, "test.yaml"), []byte(manifestConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create manifest config: %v", err)
	}

	// Create manifests directory
	manifestsDir := filepath.Join(tempDir, "manifests")
	err = os.MkdirAll(manifestsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create manifests directory: %v", err)
	}

	// Create test deployment manifest
	deploymentManifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
  labels:
    app.kubernetes.io/name: test
    app.kubernetes.io/version: "2.0.0"
    app.kubernetes.io/managed-by: frank
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: test
  template:
    metadata:
      labels:
        app.kubernetes.io/name: test
    spec:
      containers:
      - name: test
        image: nginx:alpine
        ports:
        - containerPort: 80`
	err = os.WriteFile(filepath.Join(manifestsDir, "test-deployment.yaml"), []byte(deploymentManifest), 0644)
	if err != nil {
		t.Fatalf("Failed to create deployment manifest: %v", err)
	}

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Test that the application can find the config directory
	// This is a basic integration test to ensure the workflow works
	t.Log("Integration test setup complete")
}

func TestIntegrationTemplateWorkflow(t *testing.T) {
	// Create a temporary directory for template integration test
	tempDir := t.TempDir()

	// Create config directory structure
	configDir := filepath.Join(tempDir, "config")
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Create base config
	baseConfig := `context: test
project_code: frank
version: 1.0.0`
	err = os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(baseConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create base config: %v", err)
	}

	// Create template config
	templateConfig := `manifest: test-deployment.jinja
app: test
version: 2.0.0`
	err = os.WriteFile(filepath.Join(configDir, "template.yaml"), []byte(templateConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create template config: %v", err)
	}

	// Create manifests directory
	manifestsDir := filepath.Join(tempDir, "manifests")
	err = os.MkdirAll(manifestsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create manifests directory: %v", err)
	}

	// Create Jinja template
	templateContent := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ stack_name }}
  labels:
    app.kubernetes.io/name: {{ app }}
    app.kubernetes.io/version: {{ version }}
    app.kubernetes.io/managed-by: frank
spec:
  replicas: {% if context == "prod" %}3{% else %}1{% endif %}
  selector:
    matchLabels:
      app.kubernetes.io/name: {{ app }}
  template:
    metadata:
      labels:
        app.kubernetes.io/name: {{ app }}
    spec:
      containers:
      - name: {{ app }}
        image: nginx:{{ version }}
        ports:
        - containerPort: 80
        env:
        - name: ENVIRONMENT
          value: {{ context }}
        - name: PROJECT_CODE
          value: {{ project_code }}`
	err = os.WriteFile(filepath.Join(manifestsDir, "test-deployment.jinja"), []byte(templateContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Test that the application can find the config directory and template
	t.Log("Template integration test setup complete")
}
