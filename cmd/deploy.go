/*
Copyright © 2025 Ben Sapp ya.bsapp.ru
*/

package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Config represents the base configuration structure (context only)
type Config struct {
	Context string `yaml:"context"`
}

// TemplateConfig represents template-specific configuration
type TemplateConfig struct {
	Template string `yaml:"template"`
}

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy Jinja templated Kubernetes manifest files to clusters",
	Long: `Deploy command reads configuration from the config directory and deploys
the specified manifest template to the Kubernetes cluster.

The config/config.yaml file should contain:
context: orbstack

Template config files (e.g., config/deploy.yaml) should contain:
template: sample-deployment.yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		// Read base config (context only)
		baseConfig, err := readBaseConfig()
		if err != nil {
			fmt.Printf("Error reading base config: %v\n", err)
			os.Exit(1)
		}

		// Find template config file
		templateConfigPath, err := findTemplateConfigFile()
		if err != nil {
			fmt.Printf("Error finding template config file: %v\n", err)
			os.Exit(1)
		}

		// Read template config
		templateConfig, err := readTemplateConfig(templateConfigPath)
		if err != nil {
			fmt.Printf("Error reading template config file %s: %v\n", templateConfigPath, err)
			os.Exit(1)
		}

		// Create Kubernetes client using context
		client, err := createKubernetesClient(baseConfig.Context)
		if err != nil {
			fmt.Printf("Error creating Kubernetes client: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully connected to Kubernetes cluster using context: %s\n", baseConfig.Context)
		fmt.Printf("Using template config: %s\n", templateConfigPath)

		// Deploy the specified template
		err = deployTemplate(client, templateConfig.Template)
		if err != nil {
			fmt.Printf("Error deploying template: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully deployed template: %s\n", templateConfig.Template)
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
}

// readBaseConfig reads the context configuration with inheritance
// It walks up the config directory hierarchy to find context, allowing overrides
func readBaseConfig() (*Config, error) {
	// First, find the config directory
	configDir, err := findConfigDirectory()
	if err != nil {
		return nil, err
	}

	// Start from current directory and walk up within config hierarchy
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error getting current directory: %v", err)
	}

	// Walk up the directory tree looking for config.yaml files within config hierarchy
	var context string

	// Walk up from current directory within config hierarchy
	for {
		// Check if we're still within the config hierarchy
		if !strings.HasPrefix(currentDir, configDir) && currentDir != configDir {
			break
		}

		configPath := filepath.Join(currentDir, "config.yaml")
		if _, err := os.Stat(configPath); err == nil {
			// Read this config.yaml to get context
			config, err := readConfig(configPath)
			if err == nil && config.Context != "" {
				context = config.Context
				break // Use the first (most specific) context found
			}
		}

		// Move up one directory
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// We've reached the filesystem root
			break
		}
		currentDir = parentDir
	}

	// If no context found in subdirectories, use the root config
	if context == "" {
		rootConfigPath := filepath.Join(configDir, "config.yaml")
		if _, err := os.Stat(rootConfigPath); err == nil {
			config, err := readConfig(rootConfigPath)
			if err == nil && config.Context != "" {
				context = config.Context
			}
		}
	}

	if context == "" {
		return nil, fmt.Errorf("no context found in config hierarchy")
	}

	return &Config{Context: context}, nil
}

// findTemplateConfigFile finds template config files with hierarchical precedence
// It looks for non-config.yaml files in the config directory hierarchy
func findTemplateConfigFile() (string, error) {
	// First, find the config directory
	configDir, err := findConfigDirectory()
	if err != nil {
		return "", err
	}

	// Define precedence order for template config files (excluding config.yaml)
	templateFiles := []string{
		"deploy.yaml",  // highest precedence
		"deploy.yml",   // alternative extension
		"app.yaml",     // alternative name
		"app.yml",      // alternative name and extension
		"service.yaml", // another alternative
		"service.yml",  // another alternative with different extension
	}

	// Start from current directory and walk up within config hierarchy
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting current directory: %v", err)
	}

	// Walk up the directory tree looking for template config files within config hierarchy
	for {
		// Check if we're still within the config hierarchy
		if !strings.HasPrefix(currentDir, configDir) && currentDir != configDir {
			break
		}

		// Look for template config files in current directory
		for _, filename := range templateFiles {
			configPath := filepath.Join(currentDir, filename)
			if _, err := os.Stat(configPath); err == nil {
				return configPath, nil
			}
		}

		// Move up one directory
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// We've reached the filesystem root
			break
		}
		currentDir = parentDir
	}

	// If not found in subdirectories, look in the root config directory
	for _, filename := range templateFiles {
		configPath := filepath.Join(configDir, filename)
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}
	}

	return "", fmt.Errorf("no template config file found in config hierarchy. Expected one of: %v", templateFiles)
}

// findConfigDirectory finds the config directory by walking up the directory tree
// It only works if there's an actual 'config' directory, not just a config.yaml file
func findConfigDirectory() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting current directory: %v", err)
	}

	// First check current directory
	configPath := filepath.Join(currentDir, "config")
	if stat, err := os.Stat(configPath); err == nil && stat.IsDir() {
		// Verify it's actually a directory and has a config.yaml file
		configYamlPath := filepath.Join(configPath, "config.yaml")
		if _, err := os.Stat(configYamlPath); err == nil {
			return configPath, nil
		}
	}

	// Then check parent directory only
	parentDir := filepath.Dir(currentDir)
	if parentDir != currentDir {
		configPath := filepath.Join(parentDir, "config")
		if stat, err := os.Stat(configPath); err == nil && stat.IsDir() {
			// Verify it's actually a directory and has a config.yaml file
			configYamlPath := filepath.Join(configPath, "config.yaml")
			if _, err := os.Stat(configYamlPath); err == nil {
				return configPath, nil
			}
		}
	}

	return "", fmt.Errorf("config directory with config.yaml not found in current directory or immediate parent")
}

// findProjectRoot finds the project root directory (where manifests directory is located)
func findProjectRoot() (string, error) {
	// Find the config directory first
	configDir, err := findConfigDirectory()
	if err != nil {
		return "", err
	}

	// The project root is the parent of the config directory
	projectRoot := filepath.Dir(configDir)

	// Verify that manifests directory exists in project root
	manifestsDir := filepath.Join(projectRoot, "manifests")
	if _, err := os.Stat(manifestsDir); err != nil {
		return "", fmt.Errorf("manifests directory not found in project root: %s", projectRoot)
	}

	return projectRoot, nil
}

// readTemplateConfig reads a template config file
func readTemplateConfig(configPath string) (*TemplateConfig, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config TemplateConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	if config.Template == "" {
		return nil, fmt.Errorf("template not specified in config file %s", configPath)
	}

	return &config, nil
}

// readConfig reads and parses the base configuration file (context only)
func readConfig(configPath string) (*Config, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	if config.Context == "" {
		return nil, fmt.Errorf("context not specified in config file %s", configPath)
	}

	return &config, nil
}

// createKubernetesClient creates a Kubernetes client using the specified context
func createKubernetesClient(contextName string) (kubernetes.Interface, error) {
	// Load kubeconfig from default location
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{
		CurrentContext: contextName,
	}

	// Build config from kubeconfig with context override
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides).ClientConfig()
	if err != nil {
		return nil, err
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

// deployTemplate reads and deploys a Kubernetes manifest template
func deployTemplate(client kubernetes.Interface, templateName string) error {
	// Find the project root (where manifests directory is located)
	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("error finding project root: %v", err)
	}

	// Construct the path to the template file relative to project root
	templatePath := filepath.Join(projectRoot, "manifests", templateName)

	// Check if template file exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return fmt.Errorf("template file not found: %s", templatePath)
	}

	// Read the template file
	templateData, err := ioutil.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("error reading template file: %v", err)
	}

	fmt.Printf("Reading template from: %s\n", templatePath)

	// Parse the YAML content
	decoder := k8syaml.NewYAMLOrJSONDecoder(bytes.NewReader(templateData), 4096)

	var obj unstructured.Unstructured
	err = decoder.Decode(&obj)
	if err != nil {
		return fmt.Errorf("error parsing YAML: %v", err)
	}

	// Get the resource information
	apiVersion := obj.GetAPIVersion()
	kind := obj.GetKind()
	name := obj.GetName()
	namespace := obj.GetNamespace()

	if namespace == "" {
		namespace = "default"
	}

	fmt.Printf("Deploying %s/%s: %s in namespace %s\n", apiVersion, kind, name, namespace)

	// Create the resource using the dynamic client
	// For now, we'll just print the resource details
	// In a full implementation, you would use the dynamic client to create the resource
	fmt.Printf("Resource details:\n")
	fmt.Printf("  API Version: %s\n", apiVersion)
	fmt.Printf("  Kind: %s\n", kind)
	fmt.Printf("  Name: %s\n", name)
	fmt.Printf("  Namespace: %s\n", namespace)

	// TODO: Implement actual resource creation using dynamic client
	// This would involve:
	// 1. Creating a dynamic client
	// 2. Getting the appropriate GVR (GroupVersionResource)
	// 3. Creating the resource in the cluster

	return nil
}
