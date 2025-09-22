/*
Copyright © 2025 Ben Sapp ya.bsapp.ru
*/

package stack

import (
	"testing"
)

func TestResolveDependencies_NoDependencies(t *testing.T) {
	stacks := []StackWithDependencies{
		{StackInfo: &StackInfo{Name: "stack1"}, DependsOn: []string{}},
		{StackInfo: &StackInfo{Name: "stack2"}, DependsOn: []string{}},
		{StackInfo: &StackInfo{Name: "stack3"}, DependsOn: []string{}},
	}
	expectedOrder := []string{"stack1", "stack2", "stack3"}

	result, err := ResolveDependencies(stacks)
	if err != nil {
		t.Errorf("unexpected error: %v", err)

		return
	}

	verifyResult(t, result, expectedOrder, stacks)
}

func TestResolveDependencies_SimpleChain(t *testing.T) {
	stacks := []StackWithDependencies{
		{StackInfo: &StackInfo{Name: "stack1"}, DependsOn: []string{}},
		{StackInfo: &StackInfo{Name: "stack2"}, DependsOn: []string{"stack1"}},
		{StackInfo: &StackInfo{Name: "stack3"}, DependsOn: []string{"stack2"}},
	}
	expectedOrder := []string{"stack1", "stack2", "stack3"}

	result, err := ResolveDependencies(stacks)
	if err != nil {
		t.Errorf("unexpected error: %v", err)

		return
	}

	verifyResult(t, result, expectedOrder, stacks)
}

func TestResolveDependencies_MultipleDependencies(t *testing.T) {
	stacks := []StackWithDependencies{
		{StackInfo: &StackInfo{Name: "stack1"}, DependsOn: []string{}},
		{StackInfo: &StackInfo{Name: "stack2"}, DependsOn: []string{}},
		{StackInfo: &StackInfo{Name: "stack3"}, DependsOn: []string{"stack1", "stack2"}},
	}
	expectedOrder := []string{"stack1", "stack2", "stack3"}

	result, err := ResolveDependencies(stacks)
	if err != nil {
		t.Errorf("unexpected error: %v", err)

		return
	}

	verifyResult(t, result, expectedOrder, stacks)
}

func TestResolveDependencies_ComplexGraph(t *testing.T) {
	stacks := []StackWithDependencies{
		{StackInfo: &StackInfo{Name: "database"}, DependsOn: []string{}},
		{StackInfo: &StackInfo{Name: "redis"}, DependsOn: []string{}},
		{StackInfo: &StackInfo{Name: "api"}, DependsOn: []string{"database", "redis"}},
		{StackInfo: &StackInfo{Name: "worker"}, DependsOn: []string{"database", "redis"}},
		{StackInfo: &StackInfo{Name: "web"}, DependsOn: []string{"api"}},
	}
	expectedOrder := []string{"database", "redis", "api", "worker", "web"}

	result, err := ResolveDependencies(stacks)
	if err != nil {
		t.Errorf("unexpected error: %v", err)

		return
	}

	verifyResult(t, result, expectedOrder, stacks)
}

func TestResolveDependencies_ConfigPathDependencies(t *testing.T) {
	stacks := []StackWithDependencies{
		{StackInfo: &StackInfo{Name: "frank-dev-database", ConfigPath: "/path/to/config/dev/database.yaml"}, DependsOn: []string{}},
		{StackInfo: &StackInfo{Name: "frank-dev-api", ConfigPath: "/path/to/config/dev/api.yaml"}, DependsOn: []string{"dev/database.yaml"}},
		{StackInfo: &StackInfo{Name: "frank-dev-web", ConfigPath: "/path/to/config/dev/web.yaml"}, DependsOn: []string{"dev/api.yaml"}},
	}
	expectedOrder := []string{"frank-dev-database", "frank-dev-api", "frank-dev-web"}

	result, err := ResolveDependencies(stacks)
	if err != nil {
		t.Errorf("unexpected error: %v", err)

		return
	}

	verifyResult(t, result, expectedOrder, stacks)
}

func TestResolveDependencies_MixedDependencies(t *testing.T) {
	stacks := []StackWithDependencies{
		{StackInfo: &StackInfo{Name: "database", ConfigPath: "/path/to/config/dev/database.yaml"}, DependsOn: []string{}},
		{StackInfo: &StackInfo{Name: "api", ConfigPath: "/path/to/config/dev/api.yaml"}, DependsOn: []string{"database"}},     // Stack name
		{StackInfo: &StackInfo{Name: "web", ConfigPath: "/path/to/config/dev/web.yaml"}, DependsOn: []string{"dev/api.yaml"}}, // Config path
	}
	expectedOrder := []string{"database", "api", "web"}

	result, err := ResolveDependencies(stacks)
	if err != nil {
		t.Errorf("unexpected error: %v", err)

		return
	}

	verifyResult(t, result, expectedOrder, stacks)
}

func TestResolveDependencies_CircularDependency(t *testing.T) {
	stacks := []StackWithDependencies{
		{StackInfo: &StackInfo{Name: "stack1"}, DependsOn: []string{"stack2"}},
		{StackInfo: &StackInfo{Name: "stack2"}, DependsOn: []string{"stack1"}},
	}

	_, err := ResolveDependencies(stacks)
	if err == nil {
		t.Errorf("expected error but got none")

		return
	}

	if !containsString(err.Error(), "circular dependency detected") {
		t.Errorf("expected error to contain 'circular dependency detected', got '%s'", err.Error())
	}
}

func TestResolveDependencies_MissingDependency(t *testing.T) {
	stacks := []StackWithDependencies{
		{StackInfo: &StackInfo{Name: "stack1"}, DependsOn: []string{"nonexistent"}},
	}

	_, err := ResolveDependencies(stacks)
	if err == nil {
		t.Errorf("expected error but got none")

		return
	}

	if !containsString(err.Error(), "depends on 'nonexistent' which does not exist") {
		t.Errorf("expected error to contain 'depends on 'nonexistent' which does not exist', got '%s'", err.Error())
	}
}

func TestResolveDependencies_SelfDependency(t *testing.T) {
	stacks := []StackWithDependencies{
		{StackInfo: &StackInfo{Name: "stack1"}, DependsOn: []string{"stack1"}},
	}

	_, err := ResolveDependencies(stacks)
	if err == nil {
		t.Errorf("expected error but got none")

		return
	}

	if !containsString(err.Error(), "circular dependency detected") {
		t.Errorf("expected error to contain 'circular dependency detected', got '%s'", err.Error())
	}
}

// verifyResult verifies that the result matches the expected order and satisfies dependency constraints.
func verifyResult(t *testing.T, result []*StackInfo, expectedOrder []string, originalStacks []StackWithDependencies) {
	if len(result) != len(expectedOrder) {
		t.Errorf("expected %d stacks, got %d", len(expectedOrder), len(result))

		return
	}

	verifyStackPresence(t, result, expectedOrder)
	verifyDependencyConstraintsWithDebug(t, result, originalStacks)
}

// verifyStackPresence verifies that all expected stacks are present and no unexpected stacks exist.
func verifyStackPresence(t *testing.T, result []*StackInfo, expectedOrder []string) {
	expectedSet := make(map[string]bool)
	for _, name := range expectedOrder {
		expectedSet[name] = true
	}

	resultSet := make(map[string]bool)
	for _, stack := range result {
		resultSet[stack.Name] = true
	}

	for expected := range expectedSet {
		if !resultSet[expected] {
			t.Errorf("expected stack '%s' not found in result", expected)
		}
	}

	for result := range resultSet {
		if !expectedSet[result] {
			t.Errorf("unexpected stack '%s' found in result", result)
		}
	}
}

// verifyDependencyConstraintsWithDebug verifies dependency constraints and provides debug info if they fail.
func verifyDependencyConstraintsWithDebug(t *testing.T, result []*StackInfo, originalStacks []StackWithDependencies) {
	if !verifyDependencyConstraints(result, originalStacks) {
		t.Errorf("dependency constraints not satisfied")
		// Debug: print the execution order
		t.Logf("Execution order:")

		for i, stack := range result {
			t.Logf("  %d: %s", i, stack.Name)
		}
		// Debug: print the original dependencies
		t.Logf("Original dependencies:")

		for _, stack := range originalStacks {
			t.Logf("  %s depends on: %v", stack.StackInfo.Name, stack.DependsOn)
		}
	}
}

func TestValidateDependencies_Valid(t *testing.T) {
	stacks := []StackWithDependencies{
		{StackInfo: &StackInfo{Name: "stack1"}, DependsOn: []string{}},
		{StackInfo: &StackInfo{Name: "stack2"}, DependsOn: []string{"stack1"}},
	}

	stackMap := make(map[string]*StackInfo)
	for _, stack := range stacks {
		stackMap[stack.StackInfo.Name] = stack.StackInfo
	}

	err := validateDependencies(stackMap, stacks)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateDependencies_MissingDependency(t *testing.T) {
	stacks := []StackWithDependencies{
		{StackInfo: &StackInfo{Name: "stack1"}, DependsOn: []string{"nonexistent"}},
	}

	stackMap := make(map[string]*StackInfo)
	for _, stack := range stacks {
		stackMap[stack.StackInfo.Name] = stack.StackInfo
	}

	err := validateDependencies(stackMap, stacks)
	if err == nil {
		t.Errorf("expected error but got none")

		return
	}

	if !containsString(err.Error(), "depends on 'nonexistent' which does not exist") {
		t.Errorf("expected error to contain 'depends on 'nonexistent' which does not exist', got '%s'", err.Error())
	}
}

func TestValidateDependencies_MultipleMissingDependencies(t *testing.T) {
	stacks := []StackWithDependencies{
		{StackInfo: &StackInfo{Name: "stack1"}, DependsOn: []string{"missing1", "missing2"}},
	}

	stackMap := make(map[string]*StackInfo)
	for _, stack := range stacks {
		stackMap[stack.StackInfo.Name] = stack.StackInfo
	}

	err := validateDependencies(stackMap, stacks)
	if err == nil {
		t.Errorf("expected error but got none")

		return
	}

	if !containsString(err.Error(), "depends on 'missing1' which does not exist") {
		t.Errorf("expected error to contain 'depends on 'missing1' which does not exist', got '%s'", err.Error())
	}
}

func TestDetectCircularDependencies_NoCircular(t *testing.T) {
	graph := map[string][]string{
		"stack1": {"stack2"},
		"stack2": {"stack3"},
		"stack3": {},
	}

	err := detectCircularDependencies(graph)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDetectCircularDependencies_SimpleCircular(t *testing.T) {
	graph := map[string][]string{
		"stack1": {"stack2"},
		"stack2": {"stack1"},
	}

	err := detectCircularDependencies(graph)
	if err == nil {
		t.Errorf("expected error but got none")

		return
	}

	if !containsString(err.Error(), "circular dependency detected") {
		t.Errorf("expected error to contain 'circular dependency detected', got '%s'", err.Error())
	}
}

func TestDetectCircularDependencies_SelfDependency(t *testing.T) {
	graph := map[string][]string{
		"stack1": {"stack1"},
	}

	err := detectCircularDependencies(graph)
	if err == nil {
		t.Errorf("expected error but got none")

		return
	}

	if !containsString(err.Error(), "circular dependency detected") {
		t.Errorf("expected error to contain 'circular dependency detected', got '%s'", err.Error())
	}
}

func TestDetectCircularDependencies_ComplexCircular(t *testing.T) {
	graph := map[string][]string{
		"stack1": {"stack2"},
		"stack2": {"stack3"},
		"stack3": {"stack1"},
	}

	err := detectCircularDependencies(graph)
	if err == nil {
		t.Errorf("expected error but got none")

		return
	}

	if !containsString(err.Error(), "circular dependency detected") {
		t.Errorf("expected error to contain 'circular dependency detected', got '%s'", err.Error())
	}
}

func TestDetectCircularDependencies_MultipleComponents(t *testing.T) {
	graph := map[string][]string{
		"stack1": {"stack2"},
		"stack2": {"stack1"},
		"stack3": {},
		"stack4": {"stack3"},
	}

	err := detectCircularDependencies(graph)
	if err == nil {
		t.Errorf("expected error but got none")

		return
	}

	if !containsString(err.Error(), "circular dependency detected") {
		t.Errorf("expected error to contain 'circular dependency detected', got '%s'", err.Error())
	}
}

func TestTopologicalSort_SimpleLinear(t *testing.T) {
	graph := map[string][]string{
		"stack1": {"stack2"},
		"stack2": {"stack3"},
		"stack3": {},
	}
	expectedOrder := []string{"stack1", "stack2", "stack3"}

	result, err := topologicalSort(graph)
	if err != nil {
		t.Errorf("unexpected error: %v", err)

		return
	}

	verifyTopologicalResult(t, result, expectedOrder)
}

func TestTopologicalSort_MultipleRoots(t *testing.T) {
	graph := map[string][]string{
		"stack1": {},
		"stack2": {},
		"stack3": {"stack1", "stack2"},
	}
	expectedOrder := []string{"stack1", "stack2", "stack3"}

	result, err := topologicalSort(graph)
	if err != nil {
		t.Errorf("unexpected error: %v", err)

		return
	}

	verifyTopologicalResult(t, result, expectedOrder)
}

func TestTopologicalSort_ComplexGraph(t *testing.T) {
	graph := map[string][]string{
		"database": {},
		"redis":    {},
		"api":      {"database", "redis"},
		"worker":   {"database", "redis"},
		"web":      {"api"},
	}
	expectedOrder := []string{"database", "redis", "api", "worker", "web"}

	result, err := topologicalSort(graph)
	if err != nil {
		t.Errorf("unexpected error: %v", err)

		return
	}

	verifyTopologicalResult(t, result, expectedOrder)
}

func TestTopologicalSort_CircularGraph(t *testing.T) {
	graph := map[string][]string{
		"stack1": {"stack2"},
		"stack2": {"stack1"},
	}

	_, err := topologicalSort(graph)
	if err == nil {
		t.Errorf("expected error but got none")

		return
	}

	if !containsString(err.Error(), "graph contains cycles") {
		t.Errorf("expected error to contain 'graph contains cycles', got '%s'", err.Error())
	}
}

func TestTopologicalSort_EmptyGraph(t *testing.T) {
	graph := map[string][]string{}
	expectedOrder := []string{}

	result, err := topologicalSort(graph)
	if err != nil {
		t.Errorf("unexpected error: %v", err)

		return
	}

	verifyTopologicalResult(t, result, expectedOrder)
}

// verifyTopologicalResult verifies that the topological sort result contains all expected items.
func verifyTopologicalResult(t *testing.T, result, expectedOrder []string) {
	if len(result) != len(expectedOrder) {
		t.Errorf("expected %d items, got %d", len(expectedOrder), len(result))

		return
	}

	// Check that all expected items are present (order may vary for items at same level)
	expectedSet := make(map[string]bool)
	for _, item := range expectedOrder {
		expectedSet[item] = true
	}

	resultSet := make(map[string]bool)
	for _, item := range result {
		resultSet[item] = true
	}

	for expected := range expectedSet {
		if !resultSet[expected] {
			t.Errorf("expected item '%s' not found in result", expected)
		}
	}

	for result := range resultSet {
		if !expectedSet[result] {
			t.Errorf("unexpected item '%s' found in result", result)
		}
	}
}

// verifyDependencyConstraints checks that all dependency constraints are satisfied
// in the given execution order.
func verifyDependencyConstraints(executionOrder []*StackInfo, originalStacks []StackWithDependencies) bool {
	dependencyMap := buildDependencyMap(originalStacks)
	positionMap := buildPositionMap(executionOrder)

	return checkAllDependencyConstraints(executionOrder, dependencyMap, positionMap)
}

// buildDependencyMap creates a map of stack names to their dependencies.
func buildDependencyMap(originalStacks []StackWithDependencies) map[string][]string {
	dependencyMap := make(map[string][]string)
	for _, stack := range originalStacks {
		dependencyMap[stack.StackInfo.Name] = stack.DependsOn
	}

	return dependencyMap
}

// buildPositionMap creates a map of stack names to their position in execution order.
func buildPositionMap(executionOrder []*StackInfo) map[string]int {
	positionMap := make(map[string]int)
	for i, stack := range executionOrder {
		positionMap[stack.Name] = i
	}

	return positionMap
}

// checkAllDependencyConstraints verifies that all dependency constraints are satisfied.
func checkAllDependencyConstraints(executionOrder []*StackInfo, dependencyMap map[string][]string, positionMap map[string]int) bool {
	for _, stack := range executionOrder {
		if !checkStackDependencyConstraints(stack, dependencyMap, positionMap) {
			return false
		}
	}

	return true
}

// checkStackDependencyConstraints checks dependency constraints for a single stack.
func checkStackDependencyConstraints(stack *StackInfo, dependencyMap map[string][]string, positionMap map[string]int) bool {
	dependencies := dependencyMap[stack.Name]
	stackPosition := positionMap[stack.Name]

	for _, dep := range dependencies {
		if !checkSingleDependencyConstraint(dep, stackPosition, positionMap) {
			return false
		}
	}

	return true
}

// checkSingleDependencyConstraint checks a single dependency constraint.
func checkSingleDependencyConstraint(dep string, stackPosition int, positionMap map[string]int) bool {
	depPosition, exists := positionMap[dep]
	if !exists {
		// Dependency not found in execution order - this should be caught by validation
		return true
	}
	// Dependency must appear before the current stack
	return depPosition < stackPosition
}

// Helper function to check if a string contains a substring.
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsStringHelper(s, substr))))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}

func TestGetRelativeConfigPath(t *testing.T) {
	tests := []struct {
		name     string
		fullPath string
		expected string
	}{
		{
			name:     "path with config directory",
			fullPath: "/path/to/config/dev/database.yaml",
			expected: "dev/database.yaml",
		},
		{
			name:     "path with nested config directory",
			fullPath: "/home/user/project/config/prod/api.yaml",
			expected: "prod/api.yaml",
		},
		{
			name:     "path without config directory",
			fullPath: "/path/to/database.yaml",
			expected: "database.yaml",
		},
		{
			name:     "relative path with config",
			fullPath: "config/dev/database.yaml",
			expected: "dev/database.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getRelativeConfigPath(tt.fullPath)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestResolveDependencyReferences_ConfigPaths(t *testing.T) {
	stackMap := map[string]*StackInfo{
		"stack1": {Name: "stack1"},
		"stack2": {Name: "stack2"},
	}
	configPathMap := map[string]*StackInfo{
		"dev/database.yaml": {Name: "stack1"},
		"dev/api.yaml":      {Name: "stack2"},
	}

	deps := []string{"dev/database.yaml", "dev/api.yaml"}
	expected := []string{"stack1", "stack2"}

	result := resolveDependencyReferences(deps, stackMap, configPathMap)
	verifyDependencyReferences(t, result, expected)
}

func TestResolveDependencyReferences_StackNames(t *testing.T) {
	stackMap := map[string]*StackInfo{
		"stack1": {Name: "stack1"},
		"stack2": {Name: "stack2"},
	}
	configPathMap := map[string]*StackInfo{}

	deps := []string{"stack1", "stack2"}
	expected := []string{"stack1", "stack2"}

	result := resolveDependencyReferences(deps, stackMap, configPathMap)
	verifyDependencyReferences(t, result, expected)
}

func TestResolveDependencyReferences_Mixed(t *testing.T) {
	stackMap := map[string]*StackInfo{
		"stack1": {Name: "stack1"},
		"stack2": {Name: "stack2"},
	}
	configPathMap := map[string]*StackInfo{
		"dev/database.yaml": {Name: "stack1"},
	}

	deps := []string{"dev/database.yaml", "stack2"}
	expected := []string{"stack1", "stack2"}

	result := resolveDependencyReferences(deps, stackMap, configPathMap)
	verifyDependencyReferences(t, result, expected)
}

func TestResolveDependencyReferences_Unknown(t *testing.T) {
	stackMap := map[string]*StackInfo{}
	configPathMap := map[string]*StackInfo{}

	deps := []string{"unknown"}
	expected := []string{"unknown"}

	result := resolveDependencyReferences(deps, stackMap, configPathMap)
	verifyDependencyReferences(t, result, expected)
}

// verifyDependencyReferences verifies the result of dependency reference resolution.
func verifyDependencyReferences(t *testing.T, result, expected []string) {
	if len(result) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(result))

		return
	}

	for i, exp := range expected {
		if result[i] != exp {
			t.Errorf("Expected %q at index %d, got %q", exp, i, result[i])
		}
	}
}
