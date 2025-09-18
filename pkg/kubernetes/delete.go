package kubernetes

import (
	"context"
	"fmt"
	"log/slog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
func (d *Deployer) DeleteAllManagedResources() ([]DeleteResult, error) {
	var results []DeleteResult

	// List of resource types to check for frank-managed resources
	resourceTypes := []struct {
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

	for _, rt := range resourceTypes {
		gvr := schema.GroupVersionResource{
			Group:    rt.Group,
			Version:  rt.Version,
			Resource: rt.Resource,
		}

		// List resources across all namespaces
		resourceList, err := d.dynamicClient.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			d.logger.Warn("Failed to list resources", "resource", rt.Resource, "error", err)
			continue
		}

		// Check each resource for frank annotation
		for _, item := range resourceList.Items {
			annotations := item.GetAnnotations()
			if annotations == nil {
				continue
			}

			stackName, hasStackAnnotation := annotations["frankthetank.cloud/stack-name"]
			if !hasStackAnnotation {
				continue
			}

			// This is a frank-managed resource, delete it
			namespace := item.GetNamespace()
			name := item.GetName()

			d.logger.Warn("Deleting frank-managed resource",
				"stack", stackName,
				"resource", rt.Kind,
				"name", name,
				"namespace", namespace)

			err = d.dynamicClient.Resource(gvr).Namespace(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
			
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

			results = append(results, result)
		}
	}

	return results, nil
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
