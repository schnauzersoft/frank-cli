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

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/nikolalohinski/gonja"
	"github.com/zclconf/go-cty/cty"
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
func (r *Renderer) RenderMultiDocYAML(templatePath string, context map[string]any) ([]byte, error) {
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
func (r *Renderer) parseMultiDocYAML(rendered []byte) ([]any, error) {
	var documents []any
	decoder := yaml.NewDecoder(strings.NewReader(string(rendered)))

	for {
		var doc any
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
func (r *Renderer) encodeMultiDocYAML(documents []any) ([]byte, error) {
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
func (r *Renderer) BuildTemplateContext(stackName, context, projectCode, namespace, app, version string, vars map[string]any) map[string]any {
	templateContext := map[string]any{
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
func (r *Renderer) RenderHCLManifest(templatePath string, context map[string]any) ([]byte, error) {
	// Read the template file
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %v", err)
	}

	// First, substitute variables in the HCL content
	substitutedContent := r.substituteHCLVariables(string(templateContent), context)

	// Parse the HCL content
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL([]byte(substitutedContent), "template.hcl")
	if diags.HasErrors() {
		return nil, fmt.Errorf("HCL parsing errors: %v", diags)
	}

	// Convert HCL to Kubernetes YAML
	kubernetesYAML, err := r.convertHCLToKubernetesYAML(file.Body, context)
	if err != nil {
		return nil, fmt.Errorf("failed to convert HCL to Kubernetes YAML: %v", err)
	}

	return []byte(kubernetesYAML), nil
}

// convertHCLToKubernetesYAML converts HCL body to Kubernetes YAML
func (r *Renderer) convertHCLToKubernetesYAML(body hcl.Body, context map[string]any) (string, error) {
	// Parse the HCL body to extract resource blocks
	content, diags := body.Content(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "resource", LabelNames: []string{"type", "name"}},
		},
	})
	if diags.HasErrors() {
		return "", fmt.Errorf("failed to parse HCL body: %v", diags)
	}

	var kubernetesResources []string

	// Process each resource block
	for _, block := range content.Blocks {
		if block.Type == "resource" {
			resourceYAML, err := r.convertResourceBlockToYAML(block)
			if err != nil {
				return "", fmt.Errorf("failed to convert resource block: %v", err)
			}
			kubernetesResources = append(kubernetesResources, resourceYAML)
		}
	}

	// Join all resources with document separators
	return strings.Join(kubernetesResources, "\n---\n"), nil
}

// convertResourceBlockToYAML converts a single HCL resource block to Kubernetes YAML
func (r *Renderer) convertResourceBlockToYAML(block *hcl.Block) (string, error) {
	resourceType := block.Labels[0]
	_ = block.Labels[1] // resourceName - not used in this simplified implementation

	// Map HCL resource types to Kubernetes API versions and kinds
	resourceMappings := map[string]struct {
		apiVersion string
		kind       string
	}{
		"kubernetes_deployment": {
			apiVersion: "apps/v1",
			kind:       "Deployment",
		},
		"kubernetes_service": {
			apiVersion: "v1",
			kind:       "Service",
		},
		"kubernetes_config_map": {
			apiVersion: "v1",
			kind:       "ConfigMap",
		},
		"kubernetes_secret": {
			apiVersion: "v1",
			kind:       "Secret",
		},
		"kubernetes_ingress": {
			apiVersion: "networking.k8s.io/v1",
			kind:       "Ingress",
		},
		"kubernetes_persistent_volume": {
			apiVersion: "v1",
			kind:       "PersistentVolume",
		},
		"kubernetes_persistent_volume_claim": {
			apiVersion: "v1",
			kind:       "PersistentVolumeClaim",
		},
	}

	mapping, exists := resourceMappings[resourceType]
	if !exists {
		return "", fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	// Convert HCL attributes to Kubernetes YAML
	kubernetesObj := map[string]any{
		"apiVersion": mapping.apiVersion,
		"kind":       mapping.kind,
	}

	// Parse the resource body
	resourceContent, diags := block.Body.Content(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{Name: "metadata", Required: false},
			{Name: "spec", Required: false},
		},
	})
	if diags.HasErrors() {
		return "", fmt.Errorf("failed to parse resource body: %v", diags)
	}

	// Convert metadata
	if metadataAttr, exists := resourceContent.Attributes["metadata"]; exists {
		metadata, err := r.convertHCLValueToGo(metadataAttr.Expr)
		if err != nil {
			return "", fmt.Errorf("failed to convert metadata: %v", err)
		}
		kubernetesObj["metadata"] = metadata
	}

	// Convert spec
	if specAttr, exists := resourceContent.Attributes["spec"]; exists {
		spec, err := r.convertHCLValueToGo(specAttr.Expr)
		if err != nil {
			return "", fmt.Errorf("failed to convert spec: %v", err)
		}
		kubernetesObj["spec"] = spec
	}

	// Convert to YAML
	yamlBytes, err := yaml.Marshal(kubernetesObj)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to YAML: %v", err)
	}

	return string(yamlBytes), nil
}

// convertHCLValueToGo converts an HCL expression to a Go value
func (r *Renderer) convertHCLValueToGo(expr hcl.Expression) (any, error) {
	// This is a simplified implementation that handles basic HCL expressions
	// We'll use the HCL evaluator to convert expressions to Go values

	// Create a simple evaluation context
	ctx := &hcl.EvalContext{
		Variables: map[string]cty.Value{},
	}

	// Evaluate the expression
	val, diags := expr.Value(ctx)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to evaluate HCL expression: %v", diags)
	}

	// Convert cty.Value to Go any
	return r.ctyValueToGo(val)
}

// ctyValueToGo converts a cty.Value to a Go any
func (r *Renderer) ctyValueToGo(val cty.Value) (any, error) {
	if val.IsNull() {
		return nil, nil
	}

	if !val.IsKnown() {
		return nil, fmt.Errorf("value is not known")
	}

	switch val.Type() {
	case cty.String:
		return val.AsString(), nil
	case cty.Number:
		return r.convertNumber(val)
	case cty.Bool:
		return val.True(), nil
	case cty.List(cty.String):
		return r.convertStringList(val)
	case cty.Map(cty.String):
		return r.convertStringMap(val)
	default:
		return r.convertComplexType(val)
	}
}

// convertNumber converts a cty.Number to Go int or float64
func (r *Renderer) convertNumber(val cty.Value) (any, error) {
	bigFloat := val.AsBigFloat()
	if bigFloat.IsInt() {
		intVal, _ := bigFloat.Int64()
		return intVal, nil
	}
	floatVal, _ := bigFloat.Float64()
	return floatVal, nil
}

// convertStringList converts a cty.List(cty.String) to []string
func (r *Renderer) convertStringList(val cty.Value) ([]string, error) {
	var result []string
	for it := val.ElementIterator(); it.Next(); {
		_, elemVal := it.Element()
		result = append(result, elemVal.AsString())
	}
	return result, nil
}

// convertStringMap converts a cty.Map(cty.String) to map[string]string
func (r *Renderer) convertStringMap(val cty.Value) (map[string]string, error) {
	result := make(map[string]string)
	for it := val.ElementIterator(); it.Next(); {
		k, v := it.Element()
		result[k.AsString()] = v.AsString()
	}
	return result, nil
}

// convertComplexType handles objects, tuples, and other complex types
func (r *Renderer) convertComplexType(val cty.Value) (any, error) {
	if val.Type().IsObjectType() {
		return r.convertObject(val)
	}

	if val.Type().IsTupleType() {
		return r.convertTuple(val)
	}

	return nil, fmt.Errorf("unsupported HCL type: %s", val.Type().FriendlyName())
}

// convertObject converts a cty.Object to map[string]any
func (r *Renderer) convertObject(val cty.Value) (map[string]any, error) {
	result := make(map[string]any)
	for k, v := range val.AsValueMap() {
		goVal, err := r.ctyValueToGo(v)
		if err != nil {
			return nil, fmt.Errorf("failed to convert object field %s: %v", k, err)
		}
		result[k] = goVal
	}
	return result, nil
}

// convertTuple converts a cty.Tuple to []any
func (r *Renderer) convertTuple(val cty.Value) ([]any, error) {
	var result []any
	for it := val.ElementIterator(); it.Next(); {
		_, elemVal := it.Element()
		goVal, err := r.ctyValueToGo(elemVal)
		if err != nil {
			return nil, fmt.Errorf("failed to convert tuple element: %v", err)
		}
		result = append(result, goVal)
	}
	return result, nil
}

// substituteHCLVariables performs simple variable substitution in HCL content
func (r *Renderer) substituteHCLVariables(content string, context map[string]any) string {
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
func (r *Renderer) RenderManifest(templatePath string, context map[string]any) ([]byte, error) {
	if r.IsJinjaTemplate(templatePath) {
		return r.RenderJinjaManifest(templatePath, context)
	} else if r.IsHCLTemplate(templatePath) {
		return r.RenderHCLManifest(templatePath, context)
	}
	return nil, fmt.Errorf("unsupported template type: %s", filepath.Ext(templatePath))
}

// RenderJinjaManifest renders a Jinja template file to Kubernetes manifests
func (r *Renderer) RenderJinjaManifest(templatePath string, context map[string]any) ([]byte, error) {
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
