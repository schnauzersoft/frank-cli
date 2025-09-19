/*
Copyright © 2025 Ben Sapp ya.bsapp.ru
*/

package plan

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/schnauzersoft/frank-cli/pkg/stack"
	"github.com/schnauzersoft/frank-cli/pkg/template"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestPlanner_PlanManifest(t *testing.T) {
	planner := createTestPlanner()

	tests := createPlanManifestTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := planner.PlanManifest(tt.manifestData, tt.manifestConfig, tt.stackInfo)
			validatePlanManifestResult(t, result, tt.expectError, tt.expectOperation)
		})
	}
}

func createPlanManifestTestCases() []planManifestTestCase {
	return []planManifestTestCase{
		{
			name:         "string manifest data",
			manifestData: "test-manifest.yaml",
			manifestConfig: &ManifestConfig{
				Manifest: "test-manifest.yaml",
			},
			stackInfo: &stack.StackInfo{
				Context: "test",
				Name:    "test-stack",
			},
			expectError:     true, // Will fail because file doesn't exist
			expectOperation: "",
		},
		{
			name:         "byte manifest data",
			manifestData: []byte("apiVersion: v1\nkind: Pod\nmetadata:\n  name: test"),
			manifestConfig: &ManifestConfig{
				Manifest: "test-pod.yaml",
			},
			stackInfo: &stack.StackInfo{
				Context: "test",
				Name:    "test-stack",
			},
			expectError:     false,
			expectOperation: "create", // Should be create since resource doesn't exist
		},
		{
			name:         "invalid manifest data type",
			manifestData: 123,
			manifestConfig: &ManifestConfig{
				Manifest: "test.yaml",
			},
			stackInfo: &stack.StackInfo{
				Context: "test",
				Name:    "test-stack",
			},
			expectError:     true,
			expectOperation: "",
		},
	}
}

type planManifestTestCase struct {
	name            string
	manifestData    interface{}
	manifestConfig  *ManifestConfig
	stackInfo       *stack.StackInfo
	expectError     bool
	expectOperation string
}

func validatePlanManifestResult(t *testing.T, result PlanResult, expectError bool, expectOperation string) {
	if expectError && result.Error == nil {
		t.Errorf("Expected error but got none")
	}
	if !expectError && result.Error != nil {
		t.Errorf("Unexpected error: %v", result.Error)
	}
	if !expectError && result.Operation != expectOperation {
		t.Errorf("Expected operation %s, got %s", expectOperation, result.Operation)
	}
}

func TestPlanner_convertManifestData(t *testing.T) {
	planner := createTestPlanner()

	tests := []struct {
		name         string
		manifestData interface{}
		expectError  bool
	}{
		{
			name:         "string data",
			manifestData: "test.yaml",
			expectError:  true, // File doesn't exist
		},
		{
			name:         "byte data",
			manifestData: []byte("test content"),
			expectError:  false,
		},
		{
			name:         "invalid type",
			manifestData: 123,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := planner.convertManifestData(tt.manifestData)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestPlanner_basicDiff(t *testing.T) {
	planner := createTestPlanner()

	tests := []struct {
		name     string
		old      string
		new      string
		expected string
	}{
		{
			name:     "identical strings",
			old:      "line1\nline2\nline3",
			new:      "line1\nline2\nline3",
			expected: " line1\n line2\n line3\n",
		},
		{
			name:     "different strings",
			old:      "line1\nline2\nline3",
			new:      "line1\nmodified\nline3",
			expected: " line1\n-line2\n+modified\n line3\n",
		},
		{
			name:     "added lines",
			old:      "line1\nline2",
			new:      "line1\nline2\nline3",
			expected: " line1\n line2\n+line3\n",
		},
		{
			name:     "removed lines",
			old:      "line1\nline2\nline3",
			new:      "line1\nline2",
			expected: " line1\n line2\n-line3\n",
		},
		{
			name:     "empty old",
			old:      "",
			new:      "line1\nline2",
			expected: "+line1\n+line2\n",
		},
		{
			name:     "empty new",
			old:      "line1\nline2",
			new:      "",
			expected: "-line1\n-line2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := planner.basicDiff(tt.old, tt.new)
			if result != tt.expected {
				t.Errorf("Expected:\n%q\nGot:\n%q", tt.expected, result)
			}
		})
	}
}

func TestPlanner_createUnifiedDiff(t *testing.T) {
	planner := createTestPlanner()

	tests := []struct {
		name     string
		old      string
		new      string
		expected string
	}{
		{
			name:     "identical strings",
			old:      "same content",
			new:      "same content",
			expected: "",
		},
		{
			name:     "different strings",
			old:      "old content",
			new:      "new content",
			expected: "--- current\n+++ desired\n-old content\n+new content\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := planner.createUnifiedDiff(tt.old, tt.new)
			if result != tt.expected {
				t.Errorf("Expected:\n%q\nGot:\n%q", tt.expected, result)
			}
		})
	}
}

func TestPlanner_generateDiff(t *testing.T) {
	planner := createTestPlanner()

	tests := createGenerateDiffTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff, operation := planner.generateDiff(tt.currentState, tt.desiredState)
			validateGenerateDiffResult(t, diff, operation, tt.expectDiff, tt.expectOperation)
		})
	}
}

func createGenerateDiffTestCases() []generateDiffTestCase {
	return []generateDiffTestCase{
		{
			name:            "create operation",
			currentState:    nil,
			desiredState:    []byte("apiVersion: v1\nkind: Pod"),
			expectDiff:      true,
			expectOperation: "create",
		},
		{
			name:            "no change",
			currentState:    []byte("apiVersion: v1\nkind: Pod"),
			desiredState:    []byte("apiVersion: v1\nkind: Pod"),
			expectDiff:      false,
			expectOperation: "no-change",
		},
		{
			name:            "update operation",
			currentState:    []byte("apiVersion: v1\nkind: Pod\nmetadata:\n  name: old"),
			desiredState:    []byte("apiVersion: v1\nkind: Pod\nmetadata:\n  name: new"),
			expectDiff:      true,
			expectOperation: "update",
		},
	}
}

type generateDiffTestCase struct {
	name            string
	currentState    []byte
	desiredState    []byte
	expectDiff      bool
	expectOperation string
}

func validateGenerateDiffResult(t *testing.T, diff, operation string, expectDiff bool, expectOperation string) {
	if expectDiff && diff == "" {
		t.Errorf("Expected diff but got empty string")
	}
	if !expectDiff && diff != "" {
		t.Errorf("Expected no diff but got: %s", diff)
	}
	if operation != expectOperation {
		t.Errorf("Expected operation %s, got %s", expectOperation, operation)
	}
}

func TestPlanner_colorizeDiff(t *testing.T) {
	planner := createTestPlanner()

	tests := []struct {
		name     string
		diff     string
		expected string
	}{
		{
			name:     "empty diff",
			diff:     "",
			expected: "",
		},
		{
			name:     "header lines",
			diff:     "--- current\n+++ desired",
			expected: "\033[1m--- current\033[0m\n\033[1m+++ desired\033[0m\n",
		},
		{
			name:     "deleted lines",
			diff:     "-old line",
			expected: "\033[31m-old line\033[0m\n",
		},
		{
			name:     "added lines",
			diff:     "+new line",
			expected: "\033[32m+new line\033[0m\n",
		},
		{
			name:     "context lines",
			diff:     " unchanged line",
			expected: " unchanged line\n",
		},
		{
			name:     "mixed diff",
			diff:     "--- current\n+++ desired\n unchanged\n-old\n+new",
			expected: "\033[1m--- current\033[0m\n\033[1m+++ desired\033[0m\n unchanged\n\033[31m-old\033[0m\n\033[32m+new\033[0m\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := planner.colorizeDiff("", "", tt.diff)
			if result != tt.expected {
				t.Errorf("Expected:\n%q\nGot:\n%q", tt.expected, result)
			}
		})
	}
}

func TestPlanner_getLineAt(t *testing.T) {
	planner := createTestPlanner()

	tests := []struct {
		name     string
		lines    []string
		index    int
		expected string
	}{
		{
			name:     "valid index",
			lines:    []string{"line1", "line2", "line3"},
			index:    1,
			expected: "line2",
		},
		{
			name:     "index out of bounds",
			lines:    []string{"line1", "line2"},
			index:    5,
			expected: "",
		},
		{
			name:     "empty slice",
			lines:    []string{},
			index:    0,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := planner.getLineAt(tt.lines, tt.index)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestPlanner_writeDiffLine(t *testing.T) {
	planner := createTestPlanner()

	tests := []struct {
		name     string
		oldLine  string
		newLine  string
		expected string
	}{
		{
			name:     "identical lines",
			oldLine:  "same line",
			newLine:  "same line",
			expected: " same line\n",
		},
		{
			name:     "different lines",
			oldLine:  "old line",
			newLine:  "new line",
			expected: "-old line\n+new line\n",
		},
		{
			name:     "empty old line",
			oldLine:  "",
			newLine:  "new line",
			expected: "+new line\n",
		},
		{
			name:     "empty new line",
			oldLine:  "old line",
			newLine:  "",
			expected: "-old line\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diff strings.Builder
			planner.writeDiffLine(&diff, tt.oldLine, tt.newLine)
			result := diff.String()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestPlanner_writeChangedLines(t *testing.T) {
	planner := createTestPlanner()

	tests := []struct {
		name     string
		oldLine  string
		newLine  string
		expected string
	}{
		{
			name:     "both lines present",
			oldLine:  "old line",
			newLine:  "new line",
			expected: "-old line\n+new line\n",
		},
		{
			name:     "only old line",
			oldLine:  "old line",
			newLine:  "",
			expected: "-old line\n",
		},
		{
			name:     "only new line",
			oldLine:  "",
			newLine:  "new line",
			expected: "+new line\n",
		},
		{
			name:     "both empty",
			oldLine:  "",
			newLine:  "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diff strings.Builder
			planner.writeChangedLines(&diff, tt.oldLine, tt.newLine)
			result := diff.String()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// mockKubernetesDeployer is a mock implementation for testing
type mockKubernetesDeployer struct{}

func (m *mockKubernetesDeployer) GetGVR(apiVersion, kind string) (schema.GroupVersionResource, error) {
	return schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "deployments",
	}, nil
}

func (m *mockKubernetesDeployer) GetResource(gvr schema.GroupVersionResource, namespace, name string) (*unstructured.Unstructured, error) {
	// Return an error to simulate resource not found (create operation)
	return nil, fmt.Errorf("resource not found")
}

// Helper function to create a test planner
func createTestPlanner() *Planner {
	// Create a mock Kubernetes deployer
	k8sDeployer := &mockKubernetesDeployer{}

	// Create a template renderer
	templateRenderer := template.NewRenderer(nil)

	// Create a logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError, // Only show errors in tests
	}))

	return NewPlanner(k8sDeployer, templateRenderer, logger)
}
