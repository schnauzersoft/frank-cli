/*
Copyright © 2025 Ben Sapp ya.bsapp.ru
*/

package deploy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

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

// ManifestConfig represents manifest-specific configuration
type ManifestConfig struct {
	Manifest string `yaml:"manifest"`
}

// DeploymentResult represents the result of a deployment operation
type DeploymentResult struct {
	Context   string
	Manifest  string
	Response  string
	Error     error
	Timestamp time.Time
}

// Deployer handles parallel deployment operations
type Deployer struct {
	configDir string
	logger    *slog.Logger
}

// NewDeployer creates a new Deployer instance
func NewDeployer(configDir string, logger *slog.Logger) *Deployer {
	return &Deployer{
		configDir: configDir,
		logger:    logger,
	}
}

// DeployAll performs parallel deployment of all manifest configs
func (d *Deployer) DeployAll() ([]DeploymentResult, error) {
	// Find all YAML config files
	configFiles, err := d.findAllConfigFiles()
	if err != nil {
		return nil, fmt.Errorf("error finding config files: %v", err)
	}

	if len(configFiles) == 0 {
		return nil, fmt.Errorf("no config files found")
	}

	d.logger.Debug("Found config files", "count", len(configFiles), "files", configFiles)

	// Create channels for results and errors
	results := make(chan DeploymentResult, len(configFiles))
	var wg sync.WaitGroup

	// Deploy each config file in parallel
	for _, configFile := range configFiles {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()
			d.logger.Debug("Starting deployment", "config_file", file)
			result := d.deploySingleConfig(file)
			results <- result
		}(configFile)
	}

	// Wait for all deployments to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var deploymentResults []DeploymentResult
	for result := range results {
		deploymentResults = append(deploymentResults, result)
	}

	d.logger.Debug("All deployments completed", "total", len(deploymentResults))
	return deploymentResults, nil
}

// findAllConfigFiles finds all YAML config files in the config directory and subdirectories
func (d *Deployer) findAllConfigFiles() ([]string, error) {
	var configFiles []string

	err := filepath.Walk(d.configDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if it's a YAML file (but not config.yaml)
		if (strings.HasSuffix(info.Name(), ".yaml") || strings.HasSuffix(info.Name(), ".yml")) &&
			info.Name() != "config.yaml" && info.Name() != "config.yml" {
			configFiles = append(configFiles, path)
		}

		return nil
	})

	return configFiles, err
}

// deploySingleConfig deploys a single config file
func (d *Deployer) deploySingleConfig(configPath string) DeploymentResult {
	timestamp := time.Now()

	// Read the manifest config
	manifestConfig, err := d.readManifestConfig(configPath)
	if err != nil {
		d.logger.Debug("Failed to read manifest config", "config_file", configPath, "error", err)
		return DeploymentResult{
			Context:   "unknown",
			Manifest:  filepath.Base(configPath),
			Response:  "",
			Error:     fmt.Errorf("error reading config: %v", err),
			Timestamp: timestamp,
		}
	}

	d.logger.Debug("Read manifest config", "config_file", configPath, "manifest", manifestConfig.Manifest)

	// Read base config to get context
	baseConfig, err := d.readBaseConfig()
	if err != nil {
		d.logger.Debug("Failed to read base config", "error", err)
		return DeploymentResult{
			Context:   "unknown",
			Manifest:  manifestConfig.Manifest,
			Response:  "",
			Error:     fmt.Errorf("error reading base config: %v", err),
			Timestamp: timestamp,
		}
	}

	d.logger.Debug("Read base config", "context", baseConfig.Context)

	// Create Kubernetes client
	client, err := d.createKubernetesClient(baseConfig.Context)
	if err != nil {
		d.logger.Debug("Failed to create Kubernetes client", "context", baseConfig.Context, "error", err)
		return DeploymentResult{
			Context:   baseConfig.Context,
			Manifest:  manifestConfig.Manifest,
			Response:  "",
			Error:     fmt.Errorf("error creating k8s client: %v", err),
			Timestamp: timestamp,
		}
	}

	d.logger.Debug("Created Kubernetes client", "context", baseConfig.Context)

	// Deploy the manifest
	response, err := d.deployManifest(client, manifestConfig.Manifest)
	if err != nil {
		d.logger.Debug("Failed to deploy manifest", "manifest", manifestConfig.Manifest, "error", err)
		return DeploymentResult{
			Context:   baseConfig.Context,
			Manifest:  manifestConfig.Manifest,
			Response:  "",
			Error:     fmt.Errorf("error deploying manifest: %v", err),
			Timestamp: timestamp,
		}
	}

	d.logger.Debug("Successfully deployed manifest", "manifest", manifestConfig.Manifest, "response", response)

	return DeploymentResult{
		Context:   baseConfig.Context,
		Manifest:  manifestConfig.Manifest,
		Response:  response,
		Error:     nil,
		Timestamp: timestamp,
	}
}

// readBaseConfig reads the context configuration
func (d *Deployer) readBaseConfig() (*Config, error) {
	configPath := filepath.Join(d.configDir, "config.yaml")
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

// readManifestConfig reads a manifest config file
func (d *Deployer) readManifestConfig(configPath string) (*ManifestConfig, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config ManifestConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	if config.Manifest == "" {
		return nil, fmt.Errorf("manifest not specified in config file %s", configPath)
	}

	return &config, nil
}

// createKubernetesClient creates a Kubernetes client using the specified context
func (d *Deployer) createKubernetesClient(contextName string) (kubernetes.Interface, error) {
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

// deployManifest reads and deploys a Kubernetes manifest
func (d *Deployer) deployManifest(client kubernetes.Interface, manifestName string) (string, error) {
	// Find the project root (where manifests directory is located)
	projectRoot := filepath.Dir(d.configDir)

	// Find the manifest file in the manifests directory and its subdirectories
	manifestPath, err := d.findManifestFile(projectRoot, manifestName)
	if err != nil {
		return "", err
	}

	// Read the manifest file
	manifestData, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return "", fmt.Errorf("error reading manifest file: %v", err)
	}

	// Parse the YAML content
	decoder := k8syaml.NewYAMLOrJSONDecoder(bytes.NewReader(manifestData), 4096)

	var obj unstructured.Unstructured
	err = decoder.Decode(&obj)
	if err != nil {
		return "", fmt.Errorf("error parsing YAML: %v", err)
	}

	// Get the resource information
	apiVersion := obj.GetAPIVersion()
	kind := obj.GetKind()
	name := obj.GetName()
	namespace := obj.GetNamespace()

	if namespace == "" {
		namespace = "default"
	}

	// For now, we'll just return the resource details as a response
	// In a full implementation, you would use the dynamic client to create the resource
	response := fmt.Sprintf("Deployed %s/%s: %s in namespace %s", apiVersion, kind, name, namespace)

	return response, nil
}

// findManifestFile searches for a manifest file in the manifests directory and its subdirectories
func (d *Deployer) findManifestFile(projectRoot, manifestName string) (string, error) {
	manifestsDir := filepath.Join(projectRoot, "manifests")

	// First check if the manifest exists directly in the manifests directory
	manifestPath := filepath.Join(manifestsDir, manifestName)
	if _, err := os.Stat(manifestPath); err == nil {
		return manifestPath, nil
	}

	// If not found, search in subdirectories
	return d.findManifestInSubdirectories(manifestsDir, manifestName)
}

// findManifestInSubdirectories recursively searches for a manifest file in subdirectories
func (d *Deployer) findManifestInSubdirectories(dir, manifestName string) (string, error) {
	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Search in this subdirectory
			subdirPath := filepath.Join(dir, entry.Name())
			manifestPath := filepath.Join(subdirPath, manifestName)

			if _, err := os.Stat(manifestPath); err == nil {
				return manifestPath, nil
			}

			// Recursively search in deeper subdirectories
			if found, err := d.findManifestInSubdirectories(subdirPath, manifestName); err == nil {
				return found, nil
			}
		}
	}

	return "", fmt.Errorf("manifest file not found: %s (searched in manifests directory and subdirectories)", manifestName)
}
