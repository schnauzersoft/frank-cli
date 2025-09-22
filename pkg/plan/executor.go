/*
Copyright © 2025 Ben Sapp ya.bsapp.ru
*/

package plan

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/schnauzersoft/frank-cli/pkg/kubernetes"
	"github.com/schnauzersoft/frank-cli/pkg/stack"
	"github.com/schnauzersoft/frank-cli/pkg/template"

	"gopkg.in/yaml.v3"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Executor handles planning operations for multiple configurations.
type Executor struct {
	configDir        string
	logger           *slog.Logger
	k8sDeployer      *kubernetes.Deployer
	templateRenderer *template.Renderer
	planner          *Planner
}

// NewExecutor creates a new plan executor.
func NewExecutor(configDir string, logger *slog.Logger) (*Executor, error) {
	// Create Kubernetes config
	config, err := createKubernetesConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes config: %w", err)
	}

	// Create Kubernetes deployer
	k8sDeployer, err := kubernetes.NewDeployer(config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes deployer: %w", err)
	}

	return NewExecutorWithDeployer(configDir, logger, k8sDeployer), nil
}

// NewExecutorWithDeployer creates a new plan executor with a deployer (to enable test mocking).
func NewExecutorWithDeployer(configDir string, logger *slog.Logger, k8sDeployer *kubernetes.Deployer) *Executor {
	// Create template renderer
	templateRenderer := template.NewRenderer(logger)

	// Create planner
	planner := NewPlanner(k8sDeployer, templateRenderer, logger)

	return &Executor{
		configDir:        configDir,
		logger:           logger,
		k8sDeployer:      k8sDeployer,
		templateRenderer: templateRenderer,
		planner:          planner,
	}
}

// PlanAll plans all configurations without applying them in dependency order.
func (e *Executor) PlanAll(stackFilter string) ([]PlanResult, error) {
	// Find all YAML config files
	configFiles, err := e.findAllConfigFiles()
	if err != nil {
		return nil, fmt.Errorf("error finding config files: %w", err)
	}

	if len(configFiles) == 0 {
		return nil, errors.New("no config files found")
	}

	// Filter config files by stack if filter is provided
	if stackFilter != "" {
		configFiles = e.filterConfigFilesByStack(configFiles, stackFilter)
		if len(configFiles) == 0 {
			return nil, fmt.Errorf("no config files found matching stack filter: %s", stackFilter)
		}
	}

	e.logger.Debug("Found config files for plan", "count", len(configFiles), "files", configFiles, "filter", stackFilter)

	// Collect stack info and dependencies for all config files
	stacksWithDeps, err := e.collectStacksWithDependencies(configFiles)
	if err != nil {
		return nil, fmt.Errorf("error collecting stack info and dependencies: %w", err)
	}

	// Resolve dependencies and get execution order
	orderedStacks, err := stack.ResolveDependencies(stacksWithDeps)
	if err != nil {
		return nil, fmt.Errorf("error resolving dependencies: %w", err)
	}

	e.logger.Debug("Resolved execution order for plan", "stacks", len(orderedStacks))

	// Plan stacks in dependency order
	planResults := make([]PlanResult, 0, len(orderedStacks))

	for _, stackInfo := range orderedStacks {
		e.logger.Debug("Starting plan", "config_file", stackInfo.ConfigPath, "stack", stackInfo.Name)
		result := e.planSingleConfig(stackInfo.ConfigPath)
		planResults = append(planResults, result)
	}

	e.logger.Debug("All plans completed", "total", len(planResults))

	return planResults, nil
}

// planSingleConfig plans a single configuration without applying it.
func (e *Executor) planSingleConfig(configPath string) PlanResult {
	// Read manifest config
	manifestConfig, stackInfo, err := e.readConfigAndStackInfoForPlan(configPath)
	if err != nil {
		return PlanResult{
			Context:   "unknown",
			StackName: "unknown",
			Manifest:  filepath.Base(configPath),
			Error:     err,
		}
	}

	// Find and prepare manifest file
	manifestData, err := e.findAndPrepareManifestForPlan(manifestConfig, stackInfo)
	if err != nil {
		return PlanResult{
			Context:   stackInfo.Context,
			StackName: stackInfo.Name,
			Manifest:  manifestConfig.Manifest,
			Error:     err,
		}
	}

	// Plan the manifest (compare current vs desired state)
	return e.planner.PlanManifest(manifestData, manifestConfig, stackInfo)
}

// findAllConfigFiles finds all YAML config files in the config directory and subdirectories.
func (e *Executor) findAllConfigFiles() ([]string, error) {
	var configFiles []string

	err := filepath.Walk(e.configDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if e.isConfigFile(path) {
			configFiles = append(configFiles, path)
		}

		return nil
	})

	return configFiles, err
}

// isConfigFile checks if a file is a config file.
func (e *Executor) isConfigFile(path string) bool {
	return e.isYAMLFile(path) && e.isConfigYAML(path)
}

// isYAMLFile checks if a file is a YAML file.
func (e *Executor) isYAMLFile(path string) bool {
	return strings.HasSuffix(strings.ToLower(path), ".yaml") || strings.HasSuffix(strings.ToLower(path), ".yml")
}

// isConfigYAML checks if a YAML file is a config file (not a Jinja template).
func (e *Executor) isConfigYAML(path string) bool {
	return !e.templateRenderer.IsTemplateFile(path)
}

// filterConfigFilesByStack filters config files by stack name.
func (e *Executor) filterConfigFilesByStack(configFiles []string, stackFilter string) []string {
	var filtered []string

	for _, configFile := range configFiles {
		// Get stack info for this config file
		stackInfo, err := stack.GetStackInfo(configFile)
		if err != nil {
			e.logger.Debug("Failed to get stack info for filtering", "config_file", configFile, "error", err)

			continue
		}

		// Check if this stack matches the filter
		if e.matchesStackFilter(stackInfo.Name, stackFilter) {
			filtered = append(filtered, configFile)
		}
	}

	return filtered
}

// matchesStackFilter checks if a stack name matches the given filter.
func (e *Executor) matchesStackFilter(stackName, filter string) bool {
	// Remove "config/" prefix from filter if present
	filter = strings.TrimPrefix(filter, "config/")

	// Remove file extension from filter if present
	filter = strings.TrimSuffix(filter, ".yaml")
	filter = strings.TrimSuffix(filter, ".yml")

	// Check if stack name matches filter
	return strings.Contains(stackName, filter)
}

// readConfigAndStackInfoForPlan reads the manifest config and gets stack info for planning.
func (e *Executor) readConfigAndStackInfoForPlan(configPath string) (*ManifestConfig, *stack.StackInfo, error) {
	// Read the manifest config
	manifestConfig, err := e.readManifestConfig(configPath)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading config: %w", err)
	}

	// Get stack information
	stackInfo, err := stack.GetStackInfo(configPath)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting stack info: %w", err)
	}

	return manifestConfig, stackInfo, nil
}

// readManifestConfig reads a manifest configuration file.
func (e *Executor) readManifestConfig(configPath string) (*ManifestConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config ManifestConfig

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// findAndPrepareManifestForPlan finds the manifest file and renders it if it's a template for planning.
func (e *Executor) findAndPrepareManifestForPlan(manifestConfig *ManifestConfig, stackInfo *stack.StackInfo) (any, error) {
	// Find the manifest file
	manifestPath, err := e.findManifestFile(manifestConfig.Manifest)
	if err != nil {
		return nil, fmt.Errorf("error finding manifest file: %w", err)
	}

	// Check if this is a template file and render it
	if e.templateRenderer.IsTemplateFile(manifestPath) {
		content, err := e.renderTemplateForPlan(manifestPath, stackInfo, manifestConfig)
		if err != nil {
			return nil, fmt.Errorf("error rendering template: %w", err)
		}

		return content, nil
	}

	return manifestPath, nil
}

// findManifestFile finds the manifest file in the manifests directory.
func (e *Executor) findManifestFile(manifestName string) (string, error) {
	// Look for the manifest file in the manifests directory
	manifestsDir := filepath.Join(filepath.Dir(e.configDir), "manifests")
	manifestPath := filepath.Join(manifestsDir, manifestName)

	var err error

	_, err = os.Stat(manifestPath)
	if err != nil {
		return "", fmt.Errorf("manifest file not found: %s", manifestName)
	}

	return manifestPath, nil
}

// renderTemplateForPlan renders a template for planning.
func (e *Executor) renderTemplateForPlan(manifestPath string, stackInfo *stack.StackInfo, manifestConfig *ManifestConfig) ([]byte, error) {
	// Build template context
	version := manifestConfig.Version
	if version == "" {
		version = stackInfo.Version
	}

	templateContext := e.templateRenderer.BuildTemplateContext(
		stackInfo.Name,
		stackInfo.Context,
		stackInfo.ProjectCode,
		stackInfo.Namespace,
		stackInfo.App,
		version,
		manifestConfig.Vars,
	)

	// Render the template
	rendered, err := e.templateRenderer.RenderManifest(manifestPath, templateContext)
	if err != nil {
		return nil, fmt.Errorf("error rendering template: %w", err)
	}

	return rendered, nil
}

// createKubernetesConfig creates a Kubernetes REST config.
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

// collectStacksWithDependencies collects stack information and dependencies for all config files.
func (e *Executor) collectStacksWithDependencies(configFiles []string) ([]stack.StackWithDependencies, error) {
	stacksWithDeps := make([]stack.StackWithDependencies, 0, len(configFiles))

	for _, configFile := range configFiles {
		// Get stack info
		stackInfo, err := stack.GetStackInfo(configFile)
		if err != nil {
			e.logger.Warn("Failed to get stack info", "config_file", configFile, "error", err)

			continue
		}

		// Get manifest config to extract dependencies
		manifestConfig, err := e.readManifestConfig(configFile)
		if err != nil {
			e.logger.Warn("Failed to read manifest config", "config_file", configFile, "error", err)

			continue
		}

		stacksWithDeps = append(stacksWithDeps, stack.StackWithDependencies{
			StackInfo: stackInfo,
			DependsOn: manifestConfig.DependsOn,
		})
	}

	return stacksWithDeps, nil
}
