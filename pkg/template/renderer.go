/*
Copyright © 2025 Ben Sapp ya.bsapp.ru
*/

package template

import (
	"fmt"
	"io/ioutil"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/nikolalohinski/gonja"
	"gopkg.in/yaml.v3"
)

// Renderer handles Jinja template rendering
type Renderer struct {
	logger *slog.Logger
}

// NewRenderer creates a new template renderer
func NewRenderer(logger *slog.Logger) *Renderer {
	return &Renderer{
		logger: logger,
	}
}

// RenderManifest renders a Jinja template file to Kubernetes manifests
func (r *Renderer) RenderManifest(templatePath string, context map[string]interface{}) ([]byte, error) {
	// Read the template file
	templateContent, err := ioutil.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %v", err)
	}

	// Create gonja template
	template, err := gonja.FromString(string(templateContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %v", err)
	}

	// Render the template
	rendered, err := template.Execute(context)
	if err != nil {
		return nil, fmt.Errorf("failed to render template: %v", err)
	}

	return []byte(rendered), nil
}

// IsTemplateFile checks if a file is a Jinja template
func (r *Renderer) IsTemplateFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".jinja" || ext == ".j2"
}

// RenderMultiDocYAML renders a multi-document YAML template
func (r *Renderer) RenderMultiDocYAML(templatePath string, context map[string]interface{}) ([]byte, error) {
	// Render the template first
	rendered, err := r.RenderManifest(templatePath, context)
	if err != nil {
		return nil, err
	}

	// Parse the rendered content as multi-doc YAML
	var documents []interface{}
	decoder := yaml.NewDecoder(strings.NewReader(string(rendered)))

	for {
		var doc interface{}
		if err := decoder.Decode(&doc); err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, fmt.Errorf("failed to parse multi-doc YAML: %v", err)
		}
		documents = append(documents, doc)
	}

	// Re-encode as single YAML with document separators
	var result strings.Builder
	encoder := yaml.NewEncoder(&result)

	for i, doc := range documents {
		if i > 0 {
			result.WriteString("---\n")
		}
		if err := encoder.Encode(doc); err != nil {
			return nil, fmt.Errorf("failed to encode document %d: %v", i, err)
		}
	}

	return []byte(result.String()), nil
}

// BuildTemplateContext builds the template context from stack info and config
func (r *Renderer) BuildTemplateContext(stackName, context, projectCode, namespace, app, version string) map[string]interface{} {
	templateContext := map[string]interface{}{
		"stack_name":   stackName,
		"context":      context,
		"project_code": projectCode,
		"namespace":    namespace,
		"app":          app,
		"version":      version,
	}

	// Add common Kubernetes context
	if namespace == "" {
		namespace = "default"
	}
	templateContext["k8s_namespace"] = namespace

	return templateContext
}
