package template

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsTemplateFile(t *testing.T) {
	renderer := NewRenderer(nil)

	tests := []struct {
		name     string
		filePath string
		expected bool
	}{
		{
			name:     "jinja file",
			filePath: "deployment.jinja",
			expected: true,
		},
		{
			name:     "j2 file",
			filePath: "config.j2",
			expected: true,
		},
		{
			name:     "hcl file",
			filePath: "deployment.hcl",
			expected: true,
		},
		{
			name:     "tf file",
			filePath: "config.tf",
			expected: true,
		},
		{
			name:     "yaml file",
			filePath: "deployment.yaml",
			expected: false,
		},
		{
			name:     "yml file",
			filePath: "config.yml",
			expected: false,
		},
		{
			name:     "file with path",
			filePath: "/path/to/deployment.jinja",
			expected: true,
		},
		{
			name:     "uppercase extension",
			filePath: "deployment.JINJA",
			expected: true,
		},
		{
			name:     "mixed case extension",
			filePath: "deployment.J2",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderer.IsTemplateFile(tt.filePath)
			if result != tt.expected {
				t.Errorf("IsTemplateFile() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestBuildTemplateContext(t *testing.T) {
	renderer := NewRenderer(nil)

	tests := []struct {
		name           string
		stackName      string
		context        string
		projectCode    string
		namespace      string
		app            string
		version        string
		vars           map[string]interface{}
		expectedKeys   []string
		expectedValues map[string]string
	}{
		{
			name:        "full context",
			stackName:   "frank-dev-app",
			context:     "dev",
			projectCode: "frank",
			namespace:   "dev-namespace",
			app:         "app",
			version:     "1.2.3",
			vars:        nil,
			expectedKeys: []string{
				"stack_name", "context", "project_code", "namespace", "app", "version", "k8s_namespace", "app_name",
			},
			expectedValues: map[string]string{
				"stack_name":    "frank-dev-app",
				"context":       "dev",
				"project_code":  "frank",
				"namespace":     "dev-namespace",
				"app":           "app",
				"version":       "1.2.3",
				"k8s_namespace": "dev-namespace",
				"app_name":      "app",
			},
		},
		{
			name:        "empty namespace defaults to default",
			stackName:   "frank-prod-app",
			context:     "prod",
			projectCode: "frank",
			namespace:   "",
			app:         "app",
			version:     "2.0.0",
			vars:        nil,
			expectedValues: map[string]string{
				"k8s_namespace": "default",
			},
		},
		{
			name:        "with vars",
			stackName:   "frank-dev-app",
			context:     "dev",
			projectCode: "frank",
			namespace:   "dev-namespace",
			app:         "app",
			version:     "1.2.3",
			vars: map[string]interface{}{
				"replicas":    3,
				"image_name":  "nginx",
				"port":        8080,
				"environment": "development",
			},
			expectedKeys: []string{
				"stack_name", "context", "project_code", "namespace", "app", "version", "k8s_namespace", "app_name",
				"replicas", "image_name", "port", "environment",
			},
			expectedValues: map[string]string{
				"stack_name":    "frank-dev-app",
				"context":       "dev",
				"project_code":  "frank",
				"namespace":     "dev-namespace",
				"app":           "app",
				"version":       "1.2.3",
				"k8s_namespace": "dev-namespace",
				"app_name":      "app",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context := renderer.BuildTemplateContext(
				tt.stackName, tt.context, tt.projectCode, tt.namespace, tt.app, tt.version, tt.vars,
			)

			validateTemplateContext(t, context, tt.expectedKeys, tt.expectedValues)
		})
	}
}

// validateTemplateContext validates the template context against expected keys and values
func validateTemplateContext(t *testing.T, context map[string]interface{}, expectedKeys []string, expectedValues map[string]string) {
	validateExpectedKeys(t, context, expectedKeys)
	validateExpectedValues(t, context, expectedValues)
}

// validateExpectedKeys checks that all expected keys exist in the context
func validateExpectedKeys(t *testing.T, context map[string]interface{}, expectedKeys []string) {
	for _, key := range expectedKeys {
		if _, exists := context[key]; !exists {
			t.Errorf("Expected key %s not found in context", key)
		}
	}
}

// validateExpectedValues checks specific values in the context
func validateExpectedValues(t *testing.T, context map[string]interface{}, expectedValues map[string]string) {
	for key, expectedValue := range expectedValues {
		actualValue, exists := context[key]
		if !exists {
			t.Errorf("Expected key %s not found in context", key)
			continue
		}

		if !valuesMatch(actualValue, expectedValue) {
			t.Errorf("Context[%s] = %v, want %v", key, actualValue, expectedValue)
		}
	}
}

// valuesMatch compares actual and expected values, handling type differences
func valuesMatch(actualValue interface{}, expectedValue string) bool {
	actualStr := fmt.Sprintf("%v", actualValue)
	return actualStr == expectedValue
}

func TestRenderManifest(t *testing.T) {
	// Create a temporary directory for the template
	tempDir := t.TempDir()

	// Create a simple Jinja template
	templateContent := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ stack_name }}
  labels:
    app.kubernetes.io/name: {{ app }}
    app.kubernetes.io/version: {{ version }}
    app.kubernetes.io/managed-by: frank
spec:
  replicas: {% if context == "prod" %}5{% else %}2{% endif %}
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
        env:
        - name: ENVIRONMENT
          value: {{ context }}
        - name: PROJECT_CODE
          value: {{ project_code }}
        - name: NAMESPACE
          value: {{ k8s_namespace }}`

	templatePath := filepath.Join(tempDir, "deployment.jinja")
	err := os.WriteFile(templatePath, []byte(templateContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	renderer := NewRenderer(nil)
	context := renderer.BuildTemplateContext(
		"frank-dev-app", "dev", "frank", "dev-namespace", "app", "1.2.3", nil,
	)

	rendered, err := renderer.RenderManifest(templatePath, context)
	if err != nil {
		t.Fatalf("RenderManifest() unexpected error: %v", err)
	}

	renderedStr := string(rendered)

	// Check that template variables were replaced
	expectedStrings := []string{
		"name: frank-dev-app",
		"app.kubernetes.io/name: app",
		"app.kubernetes.io/version: 1.2.3",
		"app.kubernetes.io/managed-by: frank",
		"replicas: 2", // dev context, not prod
		"image: nginx:1.2.3",
		"value: dev",
		"value: frank",
		"value: dev-namespace",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(renderedStr, expected) {
			t.Errorf("Rendered template missing expected string: %s", expected)
		}
	}

	// Check that template syntax is not present in output
	unexpectedStrings := []string{
		"{{",
		"}}",
		"{%",
		"%}",
	}

	for _, unexpected := range unexpectedStrings {
		if strings.Contains(renderedStr, unexpected) {
			t.Errorf("Rendered template contains unexpected template syntax: %s", unexpected)
		}
	}
}

func TestRenderManifestWithConditionals(t *testing.T) {
	// Create a temporary directory for the template
	tempDir := t.TempDir()

	// Create a template with conditionals
	templateContent := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ stack_name }}
spec:
  replicas: {% if context == "prod" %}5{% else %}2{% endif %}
  template:
    spec:
      containers:
      - name: {{ app }}
        image: nginx:{% if version %}{{ version }}{% else %}latest{% endif %}
        resources:
          {% if context == "prod" %}
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
          {% else %}
          requests:
            memory: "256Mi"
            cpu: "250m"
          {% endif %}`

	templatePath := filepath.Join(tempDir, "deployment.jinja")
	err := os.WriteFile(templatePath, []byte(templateContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	renderer := NewRenderer(nil)

	// Test prod context
	prodContext := renderer.BuildTemplateContext(
		"frank-prod-app", "prod", "frank", "prod-namespace", "app", "2.0.0", nil,
	)

	prodRendered, err := renderer.RenderManifest(templatePath, prodContext)
	if err != nil {
		t.Fatalf("RenderManifest() unexpected error: %v", err)
	}

	prodStr := string(prodRendered)
	if !strings.Contains(prodStr, "replicas: 5") {
		t.Errorf("Prod context should have 5 replicas")
	}
	if !strings.Contains(prodStr, "image: nginx:2.0.0") {
		t.Errorf("Prod context should use version 2.0.0")
	}
	if !strings.Contains(prodStr, "memory: \"512Mi\"") {
		t.Errorf("Prod context should have higher memory requests")
	}

	// Test dev context
	devContext := renderer.BuildTemplateContext(
		"frank-dev-app", "dev", "frank", "dev-namespace", "app", "", nil,
	)

	devRendered, err := renderer.RenderManifest(templatePath, devContext)
	if err != nil {
		t.Fatalf("RenderManifest() unexpected error: %v", err)
	}

	devStr := string(devRendered)
	if !strings.Contains(devStr, "replicas: 2") {
		t.Errorf("Dev context should have 2 replicas")
	}
	if !strings.Contains(devStr, "image: nginx:latest") {
		t.Errorf("Dev context should use latest when no version specified")
	}
	if !strings.Contains(devStr, "memory: \"256Mi\"") {
		t.Errorf("Dev context should have lower memory requests")
	}
}

func TestIsJinjaTemplate(t *testing.T) {
	renderer := NewRenderer(nil)

	tests := []struct {
		name     string
		filePath string
		expected bool
	}{
		{
			name:     "jinja file",
			filePath: "deployment.jinja",
			expected: true,
		},
		{
			name:     "j2 file",
			filePath: "config.j2",
			expected: true,
		},
		{
			name:     "hcl file",
			filePath: "deployment.hcl",
			expected: false,
		},
		{
			name:     "yaml file",
			filePath: "deployment.yaml",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderer.IsJinjaTemplate(tt.filePath)
			if result != tt.expected {
				t.Errorf("IsJinjaTemplate() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsHCLTemplate(t *testing.T) {
	renderer := NewRenderer(nil)

	tests := []struct {
		name     string
		filePath string
		expected bool
	}{
		{
			name:     "hcl file",
			filePath: "deployment.hcl",
			expected: true,
		},
		{
			name:     "tf file",
			filePath: "config.tf",
			expected: true,
		},
		{
			name:     "jinja file",
			filePath: "deployment.jinja",
			expected: false,
		},
		{
			name:     "yaml file",
			filePath: "deployment.yaml",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderer.IsHCLTemplate(tt.filePath)
			if result != tt.expected {
				t.Errorf("IsHCLTemplate() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRenderManifestWithVars(t *testing.T) {
	// Create a temporary directory for the template
	tempDir := t.TempDir()

	// Create a Jinja template that uses vars
	templateContent := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ stack_name }}
  labels:
    app.kubernetes.io/name: {{ app }}
    app.kubernetes.io/version: {{ version }}
    app.kubernetes.io/managed-by: frank
spec:
  replicas: {{ replicas | default(3) }}
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
        image: {{ image_name }}:{{ version }}
        ports:
        - containerPort: {{ port | default(80) }}
        env:
        - name: ENVIRONMENT
          value: {{ environment }}`

	templatePath := filepath.Join(tempDir, "deployment.jinja")
	err := os.WriteFile(templatePath, []byte(templateContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	renderer := NewRenderer(nil)
	context := renderer.BuildTemplateContext(
		"frank-dev-app", "dev", "frank", "dev-namespace", "app", "1.2.3",
		map[string]interface{}{
			"replicas":    5,
			"image_name":  "nginx",
			"port":        8080,
			"environment": "development",
		},
	)

	rendered, err := renderer.RenderManifest(templatePath, context)
	if err != nil {
		t.Fatalf("RenderManifest() unexpected error: %v", err)
	}

	renderedStr := string(rendered)

	// Check that vars were used
	expectedStrings := []string{
		"replicas: 5",         // from vars
		"image: nginx:1.2.3",  // image_name from vars, version from context
		"containerPort: 8080", // port from vars
		"value: development",  // environment from vars
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(renderedStr, expected) {
			t.Errorf("Rendered template missing expected string: %s", expected)
		}
	}
}

func TestRenderHCLManifest(t *testing.T) {
	// Create a temporary directory for the template
	tempDir := t.TempDir()

	// Create a simple HCL template
	templateContent := `resource "kubernetes_deployment" "app" {
  metadata {
    name = "${stack_name}"
    labels = {
      "app.kubernetes.io/name" = "${app}"
      "app.kubernetes.io/version" = "${version}"
      "app.kubernetes.io/managed-by" = "frank"
    }
  }
  
  spec {
    replicas = ${replicas}
    
    selector {
      match_labels = {
        "app.kubernetes.io/name" = "${app}"
      }
    }
    
    template {
      metadata {
        labels = {
          "app.kubernetes.io/name" = "${app}"
        }
      }
      
      spec {
        container {
          name  = "${app}"
          image = "${image_name}:${version}"
          port {
            container_port = ${port}
          }
        }
      }
    }
  }
}`

	templatePath := filepath.Join(tempDir, "deployment.hcl")
	err := os.WriteFile(templatePath, []byte(templateContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	renderer := NewRenderer(nil)
	context := renderer.BuildTemplateContext(
		"frank-dev-app", "dev", "frank", "dev-namespace", "app", "1.2.3",
		map[string]interface{}{
			"replicas":   5,
			"image_name": "nginx",
			"port":       8080,
		},
	)

	rendered, err := renderer.RenderHCLManifest(templatePath, context)
	if err != nil {
		t.Fatalf("RenderHCLManifest() unexpected error: %v", err)
	}

	renderedStr := string(rendered)

	// Check that variables were substituted
	expectedStrings := []string{
		"name = \"frank-dev-app\"",
		"\"app.kubernetes.io/name\" = \"app\"",
		"\"app.kubernetes.io/version\" = \"1.2.3\"",
		"replicas = 5",
		"image = \"nginx:1.2.3\"",
		"container_port = 8080",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(renderedStr, expected) {
			t.Errorf("Rendered HCL template missing expected string: %s", expected)
		}
	}
}
