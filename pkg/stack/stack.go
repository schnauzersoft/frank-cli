/*
Copyright © 2025 Ben Sapp ya.bsapp.ru
*/

package stack

import (
	"fmt"
	"io/ioutil"
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
	data, err := ioutil.ReadFile(configPath)
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
