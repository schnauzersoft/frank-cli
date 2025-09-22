/*
Copyright © 2025 Ben Sapp ya.bsapp.ru
*/

package deploy

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/schnauzersoft/frank-cli/pkg/kubernetes"
	"github.com/schnauzersoft/frank-cli/pkg/stack"
	"github.com/schnauzersoft/frank-cli/pkg/template"

	"gopkg.in/yaml.v3"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ManifestConfig represents manifest-specific configuration
type ManifestConfig struct {
	Manifest  string         `yaml:"manifest"`
	Timeout   time.Duration  `yaml:"timeout"`
	Version   string         `yaml:"version"`
	Vars      map[string]any `yaml:"vars"`
	DependsOn []string       `yaml:"depends_on"`
}

// DeploymentResult represents the result of a deployment operation
type DeploymentResult struct {
	Context   string
	StackName string
	Manifest  string
	Response  string
	Error     error
	Timestamp time.Time
}

// Deployer handles parallel application operations
type Deployer struct {
	configDir        string
	logger           *slog.Logger
	k8sDeployer      *kubernetes.Deployer
	templateRenderer *template.Renderer
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
		configDir:        configDir,
		logger:           logger,
		k8sDeployer:      k8sDeployer,
		templateRenderer: template.NewRenderer(logger),
	}, nil
}

// DeployAll performs application of all manifest configs in dependency order
func (d *Deployer) DeployAll(stackFilter string) ([]DeploymentResult, error) {
	// Find all YAML config files
	configFiles, err := d.findAllConfigFiles()
	if err != nil {
		return nil, fmt.Errorf("error finding config files: %v", err)
	}

	if len(configFiles) == 0 {
		return nil, fmt.Errorf("no config files found")
	}

	// Filter config files by stack if filter is provided
	if stackFilter != "" {
		configFiles = d.filterConfigFilesByStack(configFiles, stackFilter)
		if len(configFiles) == 0 {
			return nil, fmt.Errorf("no config files found matching stack filter: %s", stackFilter)
		}
	}

	d.logger.Debug("Found config files", "count", len(configFiles), "files", configFiles, "filter", stackFilter)

	// Collect stack info and dependencies for all config files
	stacksWithDeps, err := d.collectStacksWithDependencies(configFiles)
	if err != nil {
		return nil, fmt.Errorf("error collecting stack info and dependencies: %v", err)
	}

	// Resolve dependencies and get execution order
	orderedStacks, err := stack.ResolveDependencies(stacksWithDeps)
	if err != nil {
		return nil, fmt.Errorf("error resolving dependencies: %v", err)
	}

	d.logger.Debug("Resolved execution order", "stacks", len(orderedStacks))

	// Execute stacks in dependency order
	var deploymentResults []DeploymentResult
	for _, stackInfo := range orderedStacks {
		d.logger.Debug("Starting apply", "config_file", stackInfo.ConfigPath, "stack", stackInfo.Name)
		result := d.deploySingleConfig(stackInfo.ConfigPath)
		deploymentResults = append(deploymentResults, result)

		// If deployment failed, we might want to stop or continue depending on requirements
		if result.Error != nil {
			d.logger.Error("Deployment failed", "stack", stackInfo.Name, "error", result.Error)
			// For now, continue with other deployments, but this could be configurable
		}
	}

	d.logger.Debug("All applies completed", "total", len(deploymentResults))
	return deploymentResults, nil
}

// collectStacksWithDependencies collects stack information and dependencies for all config files
func (d *Deployer) collectStacksWithDependencies(configFiles []string) ([]stack.StackWithDependencies, error) {
	var stacksWithDeps []stack.StackWithDependencies

	for _, configFile := range configFiles {
		// Get stack info
		stackInfo, err := stack.GetStackInfo(configFile)
		if err != nil {
			d.logger.Warn("Failed to get stack info", "config_file", configFile, "error", err)
			continue
		}

		// Get manifest config to extract dependencies
		manifestConfig, err := d.readManifestConfig(configFile)
		if err != nil {
			d.logger.Warn("Failed to read manifest config", "config_file", configFile, "error", err)
			continue
		}

		stacksWithDeps = append(stacksWithDeps, stack.StackWithDependencies{
			StackInfo: stackInfo,
			DependsOn: manifestConfig.DependsOn,
		})
	}

	return stacksWithDeps, nil
}

// findAllConfigFiles finds all YAML config files in the config directory and subdirectories
func (d *Deployer) findAllConfigFiles() ([]string, error) {
	var configFiles []string

	err := filepath.Walk(d.configDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if d.isConfigFile(info.Name()) {
			configFiles = append(configFiles, path)
		}

		return nil
	})

	return configFiles, err
}

// isConfigFile checks if a file is a valid config file
func (d *Deployer) isConfigFile(filename string) bool {
	// Check for YAML files (but not config.yaml)
	if d.isYAMLFile(filename) && !d.isConfigYAML(filename) {
		return true
	}

	// Check for Jinja templates
	return d.isJinjaFile(filename)
}

// isYAMLFile checks if file is a YAML file
func (d *Deployer) isYAMLFile(filename string) bool {
	return strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml")
}

// isConfigYAML checks if file is a config.yaml file
func (d *Deployer) isConfigYAML(filename string) bool {
	return filename == "config.yaml" || filename == "config.yml"
}

// isJinjaFile checks if file is a Jinja template
func (d *Deployer) isJinjaFile(filename string) bool {
	return strings.HasSuffix(filename, ".jinja") || strings.HasSuffix(filename, ".j2")
}

// filterConfigFilesByStack filters config files based on stack pattern
func (d *Deployer) filterConfigFilesByStack(configFiles []string, stackFilter string) []string {
	var filteredFiles []string

	for _, configFile := range configFiles {
		// Convert file path to stack pattern for matching
		// Remove config directory prefix and .yaml/.yml extension
		relativePath := strings.TrimPrefix(configFile, d.configDir+"/")
		stackPattern := strings.TrimSuffix(relativePath, ".yaml")
		stackPattern = strings.TrimSuffix(stackPattern, ".yml")

		// Check if the stack pattern matches the filter
		if d.matchesStackFilter(stackPattern, stackFilter) {
			filteredFiles = append(filteredFiles, configFile)
		}
	}

	return filteredFiles
}

// matchesStackFilter checks if a stack pattern matches the given filter
func (d *Deployer) matchesStackFilter(stackPattern, filter string) bool {
	// Empty filter should not match anything
	if filter == "" {
		return false
	}

	// Check exact match
	if stackPattern == filter {
		return true
	}

	// Check prefix matches
	if d.matchesPrefixPattern(stackPattern, filter) {
		return true
	}

	// Check path pattern with dashes
	if d.matchesPathPattern(stackPattern, filter) {
		return true
	}

	return false
}

// matchesPrefixPattern checks if stack pattern starts with filter
func (d *Deployer) matchesPrefixPattern(stackPattern, filter string) bool {
	// Directory matching: "dev" matches "dev/app"
	if strings.HasPrefix(stackPattern, filter+"/") {
		return true
	}

	// Partial matching: "dev" matches "dev-app"
	if strings.HasPrefix(stackPattern, filter) {
		return true
	}

	return false
}

// matchesPathPattern checks if filter matches as a path pattern with dashes
func (d *Deployer) matchesPathPattern(stackPattern, filter string) bool {
	// Convert "dev/app" to "dev-app" and check if stack starts with it
	filterWithDashes := strings.ReplaceAll(filter, "/", "-")
	if strings.HasPrefix(stackPattern, filterWithDashes) {
		return true
	}

	// Check file pattern matching: "dev/app" matches "dev/app.yaml" -> "dev/app"
	if strings.HasPrefix(stackPattern, filter) && len(stackPattern) > len(filter) {
		nextChar := stackPattern[len(filter) : len(filter)+1]
		return nextChar == "/" || nextChar == "-" || nextChar == "_"
	}

	return false
}

// deploySingleConfig deploys a single config file
func (d *Deployer) deploySingleConfig(configPath string) DeploymentResult {
	timestamp := time.Now()

	// Read manifest config
	manifestConfig, stackInfo, result := d.readConfigAndStackInfo(configPath, timestamp)
	if result.Error != nil {
		return result
	}

	// Find and prepare manifest file
	manifestData, result := d.findAndPrepareManifest(manifestConfig, stackInfo, timestamp)
	if result.Error != nil {
		return result
	}

	// Validate and apply manifest
	return d.validateAndApplyManifest(manifestData, manifestConfig, stackInfo, timestamp)
}

// readConfigAndStackInfo reads the manifest config and gets stack info
func (d *Deployer) readConfigAndStackInfo(configPath string, timestamp time.Time) (*ManifestConfig, *stack.StackInfo, DeploymentResult) {
	// Read the manifest config
	manifestConfig, err := d.readManifestConfig(configPath)
	if err != nil {
		d.logger.Debug("Failed to read manifest config", "config_file", configPath, "error", err)
		return nil, nil, DeploymentResult{
			Context:   "unknown",
			StackName: "unknown",
			Manifest:  filepath.Base(configPath),
			Response:  "",
			Error:     fmt.Errorf("error reading config: %v", err),
			Timestamp: timestamp,
		}
	}

	d.logger.Debug("Read manifest config", "config_file", configPath, "manifest", manifestConfig.Manifest)

	// Get stack information
	stackInfo, err := stack.GetStackInfo(configPath)
	if err != nil {
		d.logger.Debug("Failed to get stack info", "error", err)
		return nil, nil, DeploymentResult{
			Context:   "unknown",
			StackName: stack.GenerateFallbackStackName(configPath),
			Manifest:  manifestConfig.Manifest,
			Response:  "",
			Error:     fmt.Errorf("error getting stack info: %v", err),
			Timestamp: timestamp,
		}
	}

	d.logger.Debug("Generated stack info", "stack_name", stackInfo.Name, "context", stackInfo.Context, "project_code", stackInfo.ProjectCode, "namespace", stackInfo.Namespace, "app", stackInfo.App, "version", stackInfo.Version)
	d.logger.Debug("Using Kubernetes client", "context", stackInfo.Context)

	return manifestConfig, stackInfo, DeploymentResult{}
}

// findAndPrepareManifest finds the manifest file and renders it if it's a template
func (d *Deployer) findAndPrepareManifest(manifestConfig *ManifestConfig, stackInfo *stack.StackInfo, timestamp time.Time) (any, DeploymentResult) {
	// Find the manifest file
	manifestPath, err := d.findManifestFile(manifestConfig.Manifest)
	if err != nil {
		d.logger.Debug("Failed to find manifest file", "manifest", manifestConfig.Manifest, "error", err)
		return "", DeploymentResult{
			Context:   stackInfo.Context,
			StackName: stackInfo.Name,
			Manifest:  manifestConfig.Manifest,
			Response:  "",
			Error:     fmt.Errorf("error finding manifest file: %v", err),
			Timestamp: timestamp,
		}
	}

	// Check if this is a template file and render it
	if d.templateRenderer.IsTemplateFile(manifestPath) {
		content, result := d.renderTemplate(manifestPath, stackInfo, manifestConfig, timestamp)
		if result.Error != nil {
			// If template rendering fails, fall back to treating it as a regular file
			d.logger.Warn("Template rendering failed, treating as regular file", "template", manifestPath, "error", result.Error)
			return manifestPath, DeploymentResult{}
		}
		return content, DeploymentResult{}
	}

	return manifestPath, DeploymentResult{}
}

// renderTemplate renders a Jinja template to memory
func (d *Deployer) renderTemplate(manifestPath string, stackInfo *stack.StackInfo, manifestConfig *ManifestConfig, timestamp time.Time) ([]byte, DeploymentResult) {
	d.logger.Debug("Rendering template", "template", manifestPath)

	// Build template context
	// Use manifest config version if specified, otherwise fall back to stack version
	version := manifestConfig.Version
	if version == "" {
		version = stackInfo.Version
	}

	templateContext := d.templateRenderer.BuildTemplateContext(
		stackInfo.Name,
		stackInfo.Context,
		stackInfo.ProjectCode,
		stackInfo.Namespace,
		stackInfo.App,
		version,
		manifestConfig.Vars,
	)

	// Render the template to memory
	renderedContent, err := d.templateRenderer.RenderManifest(manifestPath, templateContext)
	if err != nil {
		d.logger.Debug("Failed to render template", "template", manifestPath, "error", err)
		return nil, DeploymentResult{
			Context:   stackInfo.Context,
			StackName: stackInfo.Name,
			Manifest:  manifestConfig.Manifest,
			Response:  "",
			Error:     fmt.Errorf("error rendering template: %v", err),
			Timestamp: timestamp,
		}
	}

	d.logger.Debug("Template rendered successfully", "template", manifestPath, "size", len(renderedContent))
	return renderedContent, DeploymentResult{}
}

// validateAndApplyManifest validates namespace and applies the manifest
func (d *Deployer) validateAndApplyManifest(manifestData any, manifestConfig *ManifestConfig, stackInfo *stack.StackInfo, timestamp time.Time) DeploymentResult {
	// Set default timeout if not specified
	timeout := manifestConfig.Timeout
	if timeout == 0 {
		timeout = 10 * time.Minute // Default 10 minutes
	}

	// Validate namespace configuration
	d.logger.Debug("Validating namespace", "config_namespace", stackInfo.Namespace, "manifest", manifestConfig.Manifest)
	if err := d.validateNamespaceConfiguration(manifestData, stackInfo.Namespace); err != nil {
		d.logger.Error("Namespace validation failed", "manifest", manifestConfig.Manifest, "error", err)
		return DeploymentResult{
			Context:   stackInfo.Context,
			StackName: stackInfo.Name,
			Manifest:  manifestConfig.Manifest,
			Response:  "",
			Error:     fmt.Errorf("namespace validation failed: %v", err),
			Timestamp: timestamp,
		}
	}

	// Apply the manifest using the real Kubernetes deployer
	var result *kubernetes.DeployResult
	var err error

	if manifestPath, ok := manifestData.(string); ok {
		// It's a file path
		result, err = d.k8sDeployer.DeployManifest(manifestPath, stackInfo.Name, stackInfo.Namespace, timeout)
	} else if manifestContent, ok := manifestData.([]byte); ok {
		// It's content in memory
		result, err = d.k8sDeployer.DeployManifestContent(manifestContent, stackInfo.Name, stackInfo.Namespace, timeout)
	} else {
		return DeploymentResult{
			Context:   stackInfo.Context,
			StackName: stackInfo.Name,
			Manifest:  manifestConfig.Manifest,
			Response:  "",
			Error:     fmt.Errorf("invalid manifest data type: %T", manifestData),
			Timestamp: timestamp,
		}
	}

	if err != nil {
		d.logger.Debug("Failed to apply manifest", "manifest", manifestConfig.Manifest, "error", err)
		return DeploymentResult{
			Context:   stackInfo.Context,
			StackName: stackInfo.Name,
			Manifest:  manifestConfig.Manifest,
			Response:  "",
			Error:     fmt.Errorf("error applying manifest: %v", err),
			Timestamp: timestamp,
		}
	}

	// Check if the deployment result contains an error
	if result.Error != nil {
		d.logger.Debug("Deployment failed", "manifest", manifestConfig.Manifest, "error", result.Error)
		return DeploymentResult{
			Context:   stackInfo.Context,
			StackName: stackInfo.Name,
			Manifest:  manifestConfig.Manifest,
			Response:  d.formatResponse(result),
			Error:     fmt.Errorf("deployment failed: %v", result.Error),
			Timestamp: timestamp,
		}
	}

	// Format response based on the result
	response := d.formatResponse(result)

	d.logger.Debug("Apply completed", "manifest", manifestConfig.Manifest, "response", response)

	return DeploymentResult{
		Context:   stackInfo.Context,
		StackName: stackInfo.Name,
		Manifest:  manifestConfig.Manifest,
		Response:  response,
		Error:     nil,
		Timestamp: timestamp,
	}
}

// formatResponse formats the response string based on the deployment result
func (d *Deployer) formatResponse(result *kubernetes.DeployResult) string {
	if result.Error != nil {
		return fmt.Sprintf("Apply failed: %v", result.Error)
	}
	return fmt.Sprintf("Applied %s/%s: %s in namespace %s (operation: %s, status: %s)",
		result.Resource.GetAPIVersion(),
		result.Resource.GetKind(),
		result.Resource.GetName(),
		result.Resource.GetNamespace(),
		result.Operation,
		result.Status)
}

// readManifestConfig reads a manifest config file
func (d *Deployer) readManifestConfig(configPath string) (*ManifestConfig, error) {
	data, err := os.ReadFile(configPath)
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

	// Check for Jinja template versions
	jinjaExtensions := []string{".jinja", ".j2"}
	for _, ext := range jinjaExtensions {
		jinjaPath := strings.TrimSuffix(manifestPath, filepath.Ext(manifestPath)) + ext
		if _, err := os.Stat(jinjaPath); err == nil {
			return jinjaPath, nil
		}
	}

	// If not found, search in subdirectories
	return d.findManifestInSubdirectories(manifestsDir, manifestName)
}

// findManifestInSubdirectories recursively searches for a manifest file in subdirectories
func (d *Deployer) findManifestInSubdirectories(dir, manifestName string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			subdirPath := filepath.Join(dir, entry.Name())
			if found := d.searchInSubdirectory(subdirPath, manifestName); found != "" {
				return found, nil
			}
		}
	}

	return "", fmt.Errorf("manifest file not found: %s (searched in manifests directory and subdirectories)", manifestName)
}

// searchInSubdirectory searches for a manifest in a specific subdirectory
func (d *Deployer) searchInSubdirectory(subdirPath, manifestName string) string {
	// Check for regular manifest file
	manifestPath := filepath.Join(subdirPath, manifestName)
	if d.fileExists(manifestPath) {
		return manifestPath
	}

	// Check for Jinja template versions
	if jinjaPath := d.findJinjaTemplate(manifestPath); jinjaPath != "" {
		return jinjaPath
	}

	// Recursively search in deeper subdirectories
	if found, err := d.findManifestInSubdirectories(subdirPath, manifestName); err == nil {
		return found
	}

	return ""
}

// fileExists checks if a file exists
func (d *Deployer) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// findJinjaTemplate looks for Jinja template versions of a manifest file
func (d *Deployer) findJinjaTemplate(manifestPath string) string {
	jinjaExtensions := []string{".jinja", ".j2"}
	for _, ext := range jinjaExtensions {
		jinjaPath := strings.TrimSuffix(manifestPath, filepath.Ext(manifestPath)) + ext
		if d.fileExists(jinjaPath) {
			return jinjaPath
		}
	}
	return ""
}

// validateNamespaceConfiguration checks for namespace conflicts between config and manifest
func (d *Deployer) validateNamespaceConfiguration(manifestData any, configNamespace string) error {
	manifestContent, err := d.extractManifestContent(manifestData)
	if err != nil {
		return err
	}

	manifestNamespace, err := d.extractManifestNamespace(manifestContent)
	if err != nil {
		return err
	}

	return d.checkNamespaceConflict(configNamespace, manifestNamespace)
}

// extractManifestContent extracts content from either file path or byte content
func (d *Deployer) extractManifestContent(manifestData any) ([]byte, error) {
	if manifestPath, ok := manifestData.(string); ok {
		// It's a file path
		content, err := os.ReadFile(manifestPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read manifest file: %v", err)
		}
		return content, nil
	}

	if content, ok := manifestData.([]byte); ok {
		// It's content in memory
		return content, nil
	}

	return nil, fmt.Errorf("invalid manifest data type: %T", manifestData)
}

// extractManifestNamespace extracts namespace from manifest YAML
func (d *Deployer) extractManifestNamespace(manifestContent []byte) (string, error) {
	// Handle multi-document YAML (for HCL templates that generate multiple resources)
	decoder := yaml.NewDecoder(bytes.NewReader(manifestContent))

	for {
		var manifest map[string]any
		err := decoder.Decode(&manifest)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", d.handleYAMLParsingError(err, manifestContent)
		}

		// Check if this document has a namespace
		if namespace := d.extractNamespaceFromManifest(manifest); namespace != "" {
			return namespace, nil
		}
	}

	return "", nil
}

// handleYAMLParsingError handles YAML parsing errors with smart error detection
func (d *Deployer) handleYAMLParsingError(err error, manifestContent []byte) error {
	// Check if this looks like raw HCL content
	contentStr := string(manifestContent)
	if d.isHCLContent(contentStr) {
		return fmt.Errorf("manifest appears to be an HCL template that failed to render - check template syntax and variables")
	}
	return fmt.Errorf("invalid YAML format in manifest: %v", err)
}

// isHCLContent checks if content looks like HCL
func (d *Deployer) isHCLContent(content string) bool {
	return strings.Contains(content, "resource \"") || strings.Contains(content, "kubernetes_")
}

// extractNamespaceFromManifest extracts namespace from a single manifest document
func (d *Deployer) extractNamespaceFromManifest(manifest map[string]any) string {
	metadata, hasMetadata := manifest["metadata"].(map[string]any)
	if !hasMetadata {
		return ""
	}

	manifestNamespace, hasManifestNamespace := metadata["namespace"].(string)
	if hasManifestNamespace && manifestNamespace != "" {
		return manifestNamespace
	}

	return ""
}

// checkNamespaceConflict checks for namespace conflicts
func (d *Deployer) checkNamespaceConflict(configNamespace, manifestNamespace string) error {
	if manifestNamespace == "" {
		return nil
	}

	d.logger.Debug("Namespace validation",
		"config_namespace", configNamespace,
		"manifest_namespace", manifestNamespace,
		"has_manifest_namespace", true)

	if configNamespace != "" {
		return fmt.Errorf("namespace specified in both config file (%s) and manifest file (%s) - specify namespace in only one place",
			configNamespace, manifestNamespace)
	}

	return nil
}
