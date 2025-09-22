/*
Copyright © 2025 Ben Sapp ya.bsapp.ru
*/

package stack

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the base configuration structure
type Config struct {
	Context     string `yaml:"context"`
	ProjectCode string `yaml:"project_code"`
	Namespace   string `yaml:"namespace"`
	App         string `yaml:"app"`
	Version     string `yaml:"version"`
}

// StackInfo represents information about a stack
type StackInfo struct {
	Name        string
	Context     string
	ProjectCode string
	Namespace   string
	App         string
	Version     string
	ConfigPath  string
}

// GenerateStackName creates a stack name from project_code, context, and config file name
func GenerateStackName(projectCode, context, configFilePath string) string {
	// Get the file name without extension
	fileName := filepath.Base(configFilePath)
	fileName = strings.TrimSuffix(fileName, filepath.Ext(fileName))

	// Generate stack name: project_code-context-filename
	stackName := fmt.Sprintf("%s-%s-%s", projectCode, context, fileName)

	// Clean up the stack name (remove any invalid characters)
	stackName = strings.ReplaceAll(stackName, "_", "-")
	stackName = strings.ToLower(stackName)

	return stackName
}

// GenerateFallbackStackName creates a fallback stack name when config reading fails
func GenerateFallbackStackName(configFilePath string) string {
	// Get the file name without extension
	fileName := filepath.Base(configFilePath)
	fileName = strings.TrimSuffix(fileName, filepath.Ext(fileName))

	// Get the directory name as context
	dirName := filepath.Base(filepath.Dir(configFilePath))
	if dirName == "." {
		dirName = "unknown"
	}

	// Generate fallback stack name: unknown-context-filename
	stackName := fmt.Sprintf("unknown-%s-%s", dirName, fileName)

	// Clean up the stack name
	stackName = strings.ReplaceAll(stackName, "_", "-")
	stackName = strings.ToLower(stackName)

	return stackName
}

// ReadConfigForFile reads the context configuration with inheritance support for a specific file
func ReadConfigForFile(configFilePath string) (*Config, error) {
	// Determine the config directory for this file
	configDir := filepath.Dir(configFilePath)

	// Start with the config in the same directory as the file
	configPath := filepath.Join(configDir, "config.yaml")
	config, err := readConfigFile(configPath)
	if err != nil {
		return nil, err
	}

	// Check if we're in a subdirectory and need to inherit from parent
	currentDir := configDir
	parentDir := filepath.Dir(currentDir)

	// If we're in a subdirectory (like config/dev/), try to read parent config
	if filepath.Base(currentDir) != "config" && filepath.Base(currentDir) != "." {
		parentConfigPath := filepath.Join(parentDir, "config.yaml")
		parentConfig, err := readConfigFile(parentConfigPath)
		if err == nil {
			// Merge parent config with child config (child overrides parent)
			config = mergeConfigs(parentConfig, config)
		}
	}

	if config.Context == "" {
		return nil, fmt.Errorf("context not specified in config files")
	}

	if config.ProjectCode == "" {
		return nil, fmt.Errorf("project_code not specified in config files")
	}

	return config, nil
}

// extractAppNameFromFilename extracts the app name from a config file path
func extractAppNameFromFilename(configFilePath string) string {
	// Get the file name without extension
	fileName := filepath.Base(configFilePath)
	fileName = strings.TrimSuffix(fileName, ".yaml")
	fileName = strings.TrimSuffix(fileName, ".yml")
	fileName = strings.TrimSuffix(fileName, ".jinja")
	fileName = strings.TrimSuffix(fileName, ".j2")

	return fileName
}

// GetStackInfo extracts stack information from a config file path
func GetStackInfo(configFilePath string) (*StackInfo, error) {
	config, err := ReadConfigForFile(configFilePath)
	if err != nil {
		// Return fallback stack info if config reading fails
		return &StackInfo{
			Name:        GenerateFallbackStackName(configFilePath),
			Context:     "unknown",
			ProjectCode: "unknown",
			Namespace:   "",
			App:         extractAppNameFromFilename(configFilePath),
			Version:     "",
			ConfigPath:  configFilePath,
		}, nil
	}

	stackName := GenerateStackName(config.ProjectCode, config.Context, configFilePath)

	// Extract app name from filename if not specified in config
	appName := config.App
	if appName == "" {
		appName = extractAppNameFromFilename(configFilePath)
	}

	return &StackInfo{
		Name:        stackName,
		Context:     config.Context,
		ProjectCode: config.ProjectCode,
		Namespace:   config.Namespace,
		App:         appName,
		Version:     config.Version,
		ConfigPath:  configFilePath,
	}, nil
}

// readConfigFile reads a single config file
func readConfigFile(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// mergeConfigs merges parent and child configs (child overrides parent)
func mergeConfigs(parent, child *Config) *Config {
	result := &Config{
		Context:     parent.Context,
		ProjectCode: parent.ProjectCode,
		Namespace:   parent.Namespace,
		App:         parent.App,
		Version:     parent.Version,
	}

	// Child overrides parent if set
	if child.Context != "" {
		result.Context = child.Context
	}
	if child.ProjectCode != "" {
		result.ProjectCode = child.ProjectCode
	}
	if child.Namespace != "" {
		result.Namespace = child.Namespace
	}
	if child.App != "" {
		result.App = child.App
	}
	if child.Version != "" {
		result.Version = child.Version
	}

	return result
}

// StackWithDependencies represents a stack with its dependencies
type StackWithDependencies struct {
	StackInfo *StackInfo
	DependsOn []string
}

// ResolveDependencies resolves the execution order for stacks based on their dependencies
func ResolveDependencies(stacksWithDeps []StackWithDependencies) ([]*StackInfo, error) {
	// Create maps for both stack names and config file paths
	stackMap := make(map[string]*StackInfo)
	configPathMap := make(map[string]*StackInfo)
	graph := make(map[string][]string)

	for _, stackWithDep := range stacksWithDeps {
		stackMap[stackWithDep.StackInfo.Name] = stackWithDep.StackInfo

		// Create a relative config path for easier referencing
		configPath := getRelativeConfigPath(stackWithDep.StackInfo.ConfigPath)
		configPathMap[configPath] = stackWithDep.StackInfo

		// Convert dependencies from config paths to stack names
		resolvedDeps := resolveDependencyReferences(stackWithDep.DependsOn, stackMap, configPathMap)
		graph[stackWithDep.StackInfo.Name] = resolvedDeps
	}

	// Validate dependencies exist
	if err := validateDependencies(stackMap, stacksWithDeps); err != nil {
		return nil, err
	}

	// Check for circular dependencies
	if err := detectCircularDependencies(graph); err != nil {
		return nil, err
	}

	// Topological sort to determine execution order
	executionOrder, err := topologicalSort(graph)
	if err != nil {
		return nil, err
	}

	// Reverse the order so that dependencies come before dependents
	// (topological sort gives us dependencies first, but we want dependents last)
	for i, j := 0, len(executionOrder)-1; i < j; i, j = i+1, j-1 {
		executionOrder[i], executionOrder[j] = executionOrder[j], executionOrder[i]
	}

	// Convert back to StackInfo slice in execution order
	var orderedStacks []*StackInfo
	for _, stackName := range executionOrder {
		if stack, exists := stackMap[stackName]; exists {
			orderedStacks = append(orderedStacks, stack)
		}
	}

	return orderedStacks, nil
}

// validateDependencies checks that all dependencies exist
func validateDependencies(stackMap map[string]*StackInfo, stacksWithDeps []StackWithDependencies) error {
	// Create config path map for validation
	configPathMap := make(map[string]*StackInfo)
	for _, stackWithDep := range stacksWithDeps {
		configPath := getRelativeConfigPath(stackWithDep.StackInfo.ConfigPath)
		configPathMap[configPath] = stackWithDep.StackInfo
	}

	for _, stackWithDep := range stacksWithDeps {
		for _, dep := range stackWithDep.DependsOn {
			// Check if dependency exists as either config path or stack name
			_, existsAsConfigPath := configPathMap[dep]
			_, existsAsStackName := stackMap[dep]

			if !existsAsConfigPath && !existsAsStackName {
				return fmt.Errorf("stack '%s' depends on '%s' which does not exist", stackWithDep.StackInfo.Name, dep)
			}
		}
	}
	return nil
}

// detectCircularDependencies checks for circular dependencies using DFS
func detectCircularDependencies(graph map[string][]string) error {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for node := range graph {
		if !visited[node] {
			if err := detectCircularDependenciesFromNode(graph, node, visited, recStack); err != nil {
				return err
			}
		}
	}

	return nil
}

// detectCircularDependenciesFromNode performs DFS from a specific node to detect cycles
func detectCircularDependenciesFromNode(graph map[string][]string, node string, visited, recStack map[string]bool) error {
	visited[node] = true
	recStack[node] = true

	for _, neighbor := range graph[node] {
		if !visited[neighbor] {
			if err := detectCircularDependenciesFromNode(graph, neighbor, visited, recStack); err != nil {
				return err
			}
		} else if recStack[neighbor] {
			return fmt.Errorf("circular dependency detected: %s -> %s", node, neighbor)
		}
	}

	recStack[node] = false
	return nil
}

// topologicalSort performs topological sorting using Kahn's algorithm
func topologicalSort(graph map[string][]string) ([]string, error) {
	inDegree := calculateInDegrees(graph)
	queue := findNodesWithNoIncomingEdges(inDegree)

	result, err := processTopologicalQueue(graph, inDegree, queue)
	if err != nil {
		return nil, err
	}

	// Check if all nodes were processed
	if len(result) != len(graph) {
		return nil, fmt.Errorf("graph contains cycles or unreachable nodes")
	}

	return result, nil
}

// getRelativeConfigPath creates a relative path from the config directory
func getRelativeConfigPath(fullPath string) string {
	// Find the config directory in the path
	parts := strings.Split(fullPath, string(filepath.Separator))
	configIndex := -1
	for i, part := range parts {
		if part == "config" {
			configIndex = i
			break
		}
	}

	if configIndex == -1 {
		// If no config directory found, return the filename
		return filepath.Base(fullPath)
	}

	// Return the path relative to the config directory
	relativeParts := parts[configIndex+1:]
	return strings.Join(relativeParts, string(filepath.Separator))
}

// resolveDependencyReferences converts dependency references from config paths to stack names
func resolveDependencyReferences(deps []string, stackMap, configPathMap map[string]*StackInfo) []string {
	var resolved []string

	for _, dep := range deps {
		// First try to resolve as a config path
		if stack, exists := configPathMap[dep]; exists {
			resolved = append(resolved, stack.Name)
			continue
		}

		// If not found as config path, try as stack name
		if _, exists := stackMap[dep]; exists {
			resolved = append(resolved, dep)
			continue
		}

		// If neither found, keep the original reference for error reporting
		resolved = append(resolved, dep)
	}

	return resolved
}

// calculateInDegrees calculates the in-degree for each node in the graph
func calculateInDegrees(graph map[string][]string) map[string]int {
	inDegree := make(map[string]int)
	for node := range graph {
		inDegree[node] = 0
	}

	for _, dependencies := range graph {
		for _, dep := range dependencies {
			inDegree[dep]++
		}
	}

	return inDegree
}

// findNodesWithNoIncomingEdges finds all nodes with no incoming edges
func findNodesWithNoIncomingEdges(inDegree map[string]int) []string {
	queue := make([]string, 0)
	for node, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, node)
		}
	}
	return queue
}

// processTopologicalQueue processes the queue for topological sorting
func processTopologicalQueue(graph map[string][]string, inDegree map[string]int, queue []string) ([]string, error) {
	var result []string

	for len(queue) > 0 {
		// Remove a node from the queue
		node := queue[0]
		queue = queue[1:]
		result = append(result, node)

		// Reduce in-degree for all neighbors
		for _, neighbor := range graph[node] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	return result, nil
}
