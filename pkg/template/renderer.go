/*
Copyright © 2025 Ben Sapp ya.bsapp.ru
*/

package template

import (
	"fmt"
	"log/slog"
	"os"
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

// IsTemplateFile checks if a file is a template (Jinja or HCL)
func (r *Renderer) IsTemplateFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".jinja" || ext == ".j2" || ext == ".hcl" || ext == ".tf"
}

// IsJinjaTemplate checks if a file is a Jinja template
func (r *Renderer) IsJinjaTemplate(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".jinja" || ext == ".j2"
}

// IsHCLTemplate checks if a file is an HCL template
func (r *Renderer) IsHCLTemplate(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".hcl" || ext == ".tf"
}

// RenderMultiDocYAML renders a multi-document YAML template
func (r *Renderer) RenderMultiDocYAML(templatePath string, context map[string]interface{}) ([]byte, error) {
	// Render the template first
	rendered, err := r.RenderManifest(templatePath, context)
	if err != nil {
		return nil, err
	}

	// Parse the rendered content as multi-doc YAML
	documents, err := r.parseMultiDocYAML(rendered)
	if err != nil {
		return nil, err
	}

	// Re-encode as single YAML with document separators
	return r.encodeMultiDocYAML(documents)
}

// parseMultiDocYAML parses multi-document YAML content
func (r *Renderer) parseMultiDocYAML(rendered []byte) ([]interface{}, error) {
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

	return documents, nil
}

// encodeMultiDocYAML encodes documents as multi-document YAML
func (r *Renderer) encodeMultiDocYAML(documents []interface{}) ([]byte, error) {
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
func (r *Renderer) BuildTemplateContext(stackName, context, projectCode, namespace, app, version string, vars map[string]interface{}) map[string]interface{} {
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

	// Add app_name for backward compatibility
	templateContext["app_name"] = app

	// Add vars from manifest config
	for key, value := range vars {
		templateContext[key] = value
	}

	return templateContext
}

// RenderHCLManifest renders an HCL template file to Kubernetes manifests
func (r *Renderer) RenderHCLManifest(templatePath string, context map[string]interface{}) ([]byte, error) {
	// Read the template file
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %v", err)
	}

	// For now, we'll do simple string substitution
	// This is a simplified implementation - in practice, you'd want more sophisticated HCL parsing
	rendered := r.substituteHCLVariables(string(templateContent), context)

	return []byte(rendered), nil
}

// substituteHCLVariables performs simple variable substitution in HCL content
func (r *Renderer) substituteHCLVariables(content string, context map[string]interface{}) string {
	// Simple variable substitution for ${var.name} patterns
	result := content
	for varName, value := range context {
		placeholder := fmt.Sprintf("${%s}", varName)
		substitutedValue := fmt.Sprintf("%v", value)
		result = strings.ReplaceAll(result, placeholder, substitutedValue)
	}
	return result
}

// RenderManifest renders a template file (Jinja or HCL) to Kubernetes manifests
func (r *Renderer) RenderManifest(templatePath string, context map[string]interface{}) ([]byte, error) {
	if r.IsJinjaTemplate(templatePath) {
		return r.RenderJinjaManifest(templatePath, context)
	} else if r.IsHCLTemplate(templatePath) {
		return r.RenderHCLManifest(templatePath, context)
	}
	return nil, fmt.Errorf("unsupported template type: %s", filepath.Ext(templatePath))
}

// RenderJinjaManifest renders a Jinja template file to Kubernetes manifests
func (r *Renderer) RenderJinjaManifest(templatePath string, context map[string]interface{}) ([]byte, error) {
	// Read the template file
	templateContent, err := os.ReadFile(templatePath)
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
