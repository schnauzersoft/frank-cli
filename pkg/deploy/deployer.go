/*
Copyright © 2025 Ben Sapp ya.bsapp.ru
*/

package deploy

import (
	"fmt"
	"io/ioutil"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"frank/pkg/kubernetes"

	"gopkg.in/yaml.v3"
	k8sclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Config represents the base configuration structure (context only)
type Config struct {
	Context string `yaml:"context"`
}

// ManifestConfig represents manifest-specific configuration
type ManifestConfig struct {
	Manifest string        `yaml:"manifest"`
	Timeout  time.Duration `yaml:"timeout"`
}

// DeploymentResult represents the result of a deployment operation
type DeploymentResult struct {
	Context   string
	Manifest  string
	Response  string
	Error     error
	Timestamp time.Time
}

// Deployer handles parallel application operations
type Deployer struct {
	configDir   string
	logger      *slog.Logger
	k8sDeployer *kubernetes.Deployer
}

// NewDeployer creates a new Deployer instance
func NewDeployer(configDir string, logger *slog.Logger) (*Deployer, error) {
	// Create Kubernetes client configuration
	config, err := createKubernetesConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes config: %v", err)
	}

	// Create Kubernetes deployer
	k8sDeployer, err := kubernetes.NewDeployer(config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes deployer: %v", err)
	}

	return &Deployer{
		configDir:   configDir,
		logger:      logger,
		k8sDeployer: k8sDeployer,
	}, nil
}

// DeployAll performs parallel application of all manifest configs
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

	// Apply each config file in parallel
	for _, configFile := range configFiles {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()
			d.logger.Debug("Starting apply", "config_file", file)
			result := d.deploySingleConfig(file)
			results <- result
		}(configFile)
	}

	// Wait for all applies to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var deploymentResults []DeploymentResult
	for result := range results {
		deploymentResults = append(deploymentResults, result)
	}

	d.logger.Debug("All applies completed", "total", len(deploymentResults))
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

	d.logger.Debug("Using Kubernetes client", "context", baseConfig.Context)

	// Find the manifest file
	manifestPath, err := d.findManifestFile(manifestConfig.Manifest)
	if err != nil {
		d.logger.Debug("Failed to find manifest file", "manifest", manifestConfig.Manifest, "error", err)
		return DeploymentResult{
			Context:   baseConfig.Context,
			Manifest:  manifestConfig.Manifest,
			Response:  "",
			Error:     fmt.Errorf("error finding manifest file: %v", err),
			Timestamp: timestamp,
		}
	}

	// Set default timeout if not specified
	timeout := manifestConfig.Timeout
	if timeout == 0 {
		timeout = 10 * time.Minute // Default 10 minutes
	}

	// Apply the manifest using the real Kubernetes deployer
	result, err := d.k8sDeployer.DeployManifest(manifestPath, timeout)
	if err != nil {
		d.logger.Debug("Failed to apply manifest", "manifest", manifestConfig.Manifest, "error", err)
		return DeploymentResult{
			Context:   baseConfig.Context,
			Manifest:  manifestConfig.Manifest,
			Response:  "",
			Error:     fmt.Errorf("error applying manifest: %v", err),
			Timestamp: timestamp,
		}
	}

	// Format response based on the result
	var response string
	if result.Error != nil {
		response = fmt.Sprintf("Apply failed: %v", result.Error)
	} else {
		response = fmt.Sprintf("Applied %s/%s: %s in namespace %s (operation: %s, status: %s)",
			result.Resource.GetAPIVersion(),
			result.Resource.GetKind(),
			result.Resource.GetName(),
			result.Resource.GetNamespace(),
			result.Operation,
			result.Status)
	}

	d.logger.Debug("Apply completed", "manifest", manifestConfig.Manifest, "response", response)

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

// createKubernetesConfig creates a Kubernetes REST config
func createKubernetesConfig() (*rest.Config, error) {
	// Load kubeconfig from default location
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}

	// Build config from kubeconfig
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides).ClientConfig()
	if err != nil {
		return nil, err
	}

	return config, nil
}

// createKubernetesClient creates a Kubernetes client using the specified context
func (d *Deployer) createKubernetesClient(contextName string) (k8sclient.Interface, error) {
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
	clientset, err := k8sclient.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

// findManifestFile searches for a manifest file in the manifests directory and its subdirectories
func (d *Deployer) findManifestFile(manifestName string) (string, error) {
	// Find the project root (where manifests directory is located)
	projectRoot := filepath.Dir(d.configDir)
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
