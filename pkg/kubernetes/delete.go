package kubernetes

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// NewDeployerForDelete creates a new Kubernetes deployer specifically for delete operations
func NewDeployerForDelete(logger *slog.Logger) (*Deployer, error) {
	config, err := createKubernetesConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes config: %v", err)
	}

	return NewDeployer(config, logger)
}

// DeleteAllManagedResources finds and deletes all resources with frankthetank.cloud/stack-name annotation
func (d *Deployer) DeleteAllManagedResources(stackFilter string) ([]DeleteResult, error) {
	var results []DeleteResult

	resourceTypes := d.getResourceTypesToDelete()

	for _, rt := range resourceTypes {
		resourceResults := d.deleteResourcesOfType(rt, stackFilter)
		results = append(results, resourceResults...)
	}

	return results, nil
}

// getResourceTypesToDelete returns the list of resource types to check for frank-managed resources
func (d *Deployer) getResourceTypesToDelete() []struct {
	Group    string
	Version  string
	Resource string
	Kind     string
} {
	return []struct {
		Group    string
		Version  string
		Resource string
		Kind     string
	}{
		{"apps", "v1", "deployments", "Deployment"},
		{"apps", "v1", "statefulsets", "StatefulSet"},
		{"apps", "v1", "daemonsets", "DaemonSet"},
		{"", "v1", "services", "Service"},
		{"", "v1", "configmaps", "ConfigMap"},
		{"", "v1", "secrets", "Secret"},
		{"", "v1", "pods", "Pod"},
		{"batch", "v1", "jobs", "Job"},
		{"batch", "v1", "cronjobs", "CronJob"},
		{"networking.k8s.io", "v1", "ingresses", "Ingress"},
	}
}

// deleteResourcesOfType deletes all frank-managed resources of a specific type
func (d *Deployer) deleteResourcesOfType(rt struct {
	Group    string
	Version  string
	Resource string
	Kind     string
}, stackFilter string,
) []DeleteResult {
	var results []DeleteResult

	gvr := schema.GroupVersionResource{
		Group:    rt.Group,
		Version:  rt.Version,
		Resource: rt.Resource,
	}

	// List resources across all namespaces
	resourceList, err := d.dynamicClient.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		d.logger.Warn("Failed to list resources", "resource", rt.Resource, "error", err)
		return results
	}

	// Check each resource for frank annotation and delete if matches
	for _, item := range resourceList.Items {
		if d.shouldDeleteResource(item, stackFilter) {
			result := d.deleteResource(item, rt, gvr)
			results = append(results, result)
		}
	}

	return results
}

// shouldDeleteResource checks if a resource should be deleted based on annotations and filter
func (d *Deployer) shouldDeleteResource(item unstructured.Unstructured, stackFilter string) bool {
	annotations := item.GetAnnotations()
	if annotations == nil {
		return false
	}

	stackName, hasStackAnnotation := annotations["frankthetank.cloud/stack-name"]
	if !hasStackAnnotation {
		return false
	}

	// Apply stack filter if provided
	if stackFilter != "" && !d.matchesStackFilter(stackName, stackFilter) {
		return false
	}

	return true
}

// deleteResource deletes a single resource and returns the result
func (d *Deployer) deleteResource(item unstructured.Unstructured, rt struct {
	Group    string
	Version  string
	Resource string
	Kind     string
}, gvr schema.GroupVersionResource,
) DeleteResult {
	annotations := item.GetAnnotations()
	stackName := annotations["frankthetank.cloud/stack-name"]
	namespace := item.GetNamespace()
	name := item.GetName()

	d.logger.Warn("Deleting frank-managed resource",
		"stack", stackName,
		"resource", rt.Kind,
		"name", name,
		"namespace", namespace)

	err := d.dynamicClient.Resource(gvr).Namespace(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})

	result := DeleteResult{
		StackName:    stackName,
		ResourceType: rt.Kind,
		ResourceName: name,
		Namespace:    namespace,
		Error:        err,
	}

	if err != nil {
		d.logger.Error("Failed to delete resource", "error", err)
	} else {
		d.logger.Info("Successfully deleted resource")
	}

	return result
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

// matchesStackFilter checks if a stack name matches the given filter
func (d *Deployer) matchesStackFilter(stackName, filter string) bool {
	// Exact match
	if stackName == filter {
		return true
	}

	// Check if stack name starts with filter (for partial matching)
	if strings.HasPrefix(stackName, filter) {
		return true
	}

	// Check if filter is a directory pattern that matches
	// e.g., "dev" should match "dev-app", "dev-web", etc.
	if strings.HasPrefix(stackName, filter+"-") {
		return true
	}

	// Check if filter is a path pattern that matches
	// e.g., "dev/app" should match "dev-app", "dev-app-1", etc.
	if strings.HasPrefix(stackName, strings.ReplaceAll(filter, "/", "-")) {
		return true
	}

	return false
}
