/*
Copyright © 2025 Ben Sapp ya.bsapp.ru
*/

package plan

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/schnauzersoft/frank-cli/pkg/stack"
	"github.com/schnauzersoft/frank-cli/pkg/template"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
)

// KubernetesDeployer interface for planning operations
type KubernetesDeployer interface {
	GetGVR(apiVersion, kind string) (schema.GroupVersionResource, error)
	GetResource(gvr schema.GroupVersionResource, namespace, name string) (*unstructured.Unstructured, error)
}

// Planner handles planning operations
type Planner struct {
	k8sDeployer      KubernetesDeployer
	templateRenderer *template.Renderer
	logger           *slog.Logger
}

// PlanResult represents the result of a plan operation
type PlanResult struct {
	Context         string
	StackName       string
	Manifest        string
	Operation       string
	Diff            string
	ManifestContent string
	Error           error
}

// NewPlanner creates a new planner instance
func NewPlanner(k8sDeployer KubernetesDeployer, templateRenderer *template.Renderer, logger *slog.Logger) *Planner {
	return &Planner{
		k8sDeployer:      k8sDeployer,
		templateRenderer: templateRenderer,
		logger:           logger,
	}
}

// PlanManifest plans a manifest by comparing current vs desired state
func (p *Planner) PlanManifest(manifestData interface{}, manifestConfig *ManifestConfig, stackInfo *stack.StackInfo) PlanResult {
	// Convert manifest data to bytes for processing
	manifestContent, err := p.convertManifestData(manifestData)
	if err != nil {
		return PlanResult{
			Context:   stackInfo.Context,
			StackName: stackInfo.Name,
			Manifest:  manifestConfig.Manifest,
			Error:     err,
		}
	}

	// Get the current state from Kubernetes
	currentState, err := p.getCurrentState(stackInfo, manifestContent)
	if err != nil {
		return PlanResult{
			Context:   stackInfo.Context,
			StackName: stackInfo.Name,
			Manifest:  manifestConfig.Manifest,
			Error:     fmt.Errorf("error getting current state: %v", err),
		}
	}

	// Generate diff
	var currentStateBytes []byte
	if currentState != "" {
		currentStateBytes = []byte(currentState)
	}
	diff, operation := p.generateDiff(currentStateBytes, manifestContent)

	return PlanResult{
		Context:         stackInfo.Context,
		StackName:       stackInfo.Name,
		Manifest:        manifestConfig.Manifest,
		Operation:       operation,
		Diff:            diff,
		ManifestContent: string(manifestContent),
		Error:           nil,
	}
}

// convertManifestData converts manifest data to bytes
func (p *Planner) convertManifestData(manifestData interface{}) ([]byte, error) {
	switch data := manifestData.(type) {
	case string:
		// It's a file path, read the file
		return p.readManifestFile(data)
	case []byte:
		// It's already rendered content
		return data, nil
	default:
		return nil, fmt.Errorf("unexpected manifest data type: %T", data)
	}
}

// readManifestFile reads a manifest file
func (p *Planner) readManifestFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}

// getCurrentState gets the current state of the resource from Kubernetes
func (p *Planner) getCurrentState(stackInfo *stack.StackInfo, manifestContent []byte) (string, error) {
	// Parse the manifest to get resource info
	decoder := k8syaml.NewYAMLOrJSONDecoder(strings.NewReader(string(manifestContent)), 4096)
	var obj unstructured.Unstructured
	if err := decoder.Decode(&obj); err != nil {
		return "", fmt.Errorf("error parsing manifest: %v", err)
	}

	// Get the GVR for the resource
	gvr, err := p.k8sDeployer.GetGVR(obj.GetAPIVersion(), obj.GetKind())
	if err != nil {
		return "", fmt.Errorf("failed to get GVR: %v", err)
	}

	// Try to get the current resource
	namespace := obj.GetNamespace()
	if namespace == "" {
		namespace = stackInfo.Namespace
		if namespace == "" {
			namespace = "default"
		}
	}

	existing, err := p.k8sDeployer.GetResource(gvr, namespace, obj.GetName())
	if err != nil {
		// Resource doesn't exist
		return "", nil
	}

	// Convert to YAML
	existingYAML, err := yaml.Marshal(existing.Object)
	if err != nil {
		return "", fmt.Errorf("error marshaling existing resource: %v", err)
	}

	return string(existingYAML), nil
}

// generateDiff generates a colored diff between current and desired state
func (p *Planner) generateDiff(currentState, desiredState []byte) (string, string) {
	if currentState == nil {
		// Resource doesn't exist, will be created
		return p.colorizeDiff("", string(desiredState), "+"), "create"
	}

	// Generate unified diff
	diff := p.createUnifiedDiff(string(currentState), string(desiredState))
	if diff == "" {
		// No changes needed
		return "", "no-change"
	}

	return p.colorizeDiff(string(currentState), string(desiredState), diff), "update"
}

// createUnifiedDiff creates a unified diff between two strings
func (p *Planner) createUnifiedDiff(old, new string) string {
	// Simple diff implementation - in a real implementation you'd use a proper diff library
	// For now, just return a basic comparison
	if old == new {
		return ""
	}
	return "--- current\n+++ desired\n" + p.basicDiff(old, new)
}

// basicDiff creates a basic line-by-line diff
func (p *Planner) basicDiff(old, new string) string {
	oldLines := strings.Split(old, "\n")
	newLines := strings.Split(new, "\n")

	var diff strings.Builder
	maxLines := max(len(oldLines), len(newLines))

	for i := 0; i < maxLines; i++ {
		oldLine := p.getLineAt(oldLines, i)
		newLine := p.getLineAt(newLines, i)
		p.writeDiffLine(&diff, oldLine, newLine)
	}

	return diff.String()
}

// getLineAt safely gets a line at the given index
func (p *Planner) getLineAt(lines []string, index int) string {
	if index < len(lines) {
		return lines[index]
	}
	return ""
}

// writeDiffLine writes a diff line for the given old and new lines
func (p *Planner) writeDiffLine(diff *strings.Builder, oldLine, newLine string) {
	if oldLine != newLine {
		p.writeChangedLines(diff, oldLine, newLine)
	} else {
		diff.WriteString(fmt.Sprintf(" %s\n", oldLine))
	}
}

// writeChangedLines writes the old and new lines when they differ
func (p *Planner) writeChangedLines(diff *strings.Builder, oldLine, newLine string) {
	if oldLine != "" {
		diff.WriteString(fmt.Sprintf("-%s\n", oldLine))
	}
	if newLine != "" {
		diff.WriteString(fmt.Sprintf("+%s\n", newLine))
	}
}

// colorizeDiff adds ANSI color codes to diff output
func (p *Planner) colorizeDiff(_, _, diff string) string {
	if diff == "" {
		return ""
	}

	lines := strings.Split(diff, "\n")
	var colored strings.Builder

	for _, line := range lines {
		if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") {
			// Header lines - bold
			colored.WriteString(fmt.Sprintf("\033[1m%s\033[0m\n", line))
		} else if strings.HasPrefix(line, "-") {
			// Deleted lines - red
			colored.WriteString(fmt.Sprintf("\033[31m%s\033[0m\n", line))
		} else if strings.HasPrefix(line, "+") {
			// Added lines - green
			colored.WriteString(fmt.Sprintf("\033[32m%s\033[0m\n", line))
		} else {
			// Context lines - normal
			colored.WriteString(line + "\n")
		}
	}

	return colored.String()
}

// ManifestConfig represents manifest-specific configuration for planning
type ManifestConfig struct {
	Manifest string                 `yaml:"manifest"`
	Timeout  int                    `yaml:"timeout"`
	Version  string                 `yaml:"version"`
	Vars     map[string]interface{} `yaml:"vars"`
}
