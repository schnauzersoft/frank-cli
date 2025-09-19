/*
Copyright © 2025 Ben Sapp ya.bsapp.ru
*/

package kubernetes

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// NewDeployer creates a new Kubernetes applier
func NewDeployer(config *rest.Config, logger *slog.Logger) (*Deployer, error) {
	// Create dynamic client for generic resource operations
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %v", err)
	}

	// Create typed client for specific operations
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %v", err)
	}

	return &Deployer{
		dynamicClient: dynamicClient,
		clientset:     clientset,
		logger:        logger,
	}, nil
}

// DeployManifest applies a single manifest file to Kubernetes
func (d *Deployer) DeployManifest(manifestPath string, stackName string, configNamespace string, timeout time.Duration) (*DeployResult, error) {
	// Parse and prepare the manifest
	obj, gvr, err := d.parseAndPrepareManifest(manifestPath, stackName, configNamespace)
	if err != nil {
		return nil, err
	}

	// Apply the resource to Kubernetes
	operation, result, err := d.applyResource(obj, gvr, stackName)
	if err != nil {
		return &DeployResult{
			Resource:  obj,
			Operation: operation,
			Status:    "failed",
			Error:     err,
			Timestamp: time.Now(),
		}, nil
	}

	// Poll for completion and return result
	status := d.determineStatus(operation, gvr, result, stackName, timeout)
	return &DeployResult{
		Resource:  result,
		Operation: operation,
		Status:    status,
		Error:     err,
		Timestamp: time.Now(),
	}, nil
}

// DeployManifestContent applies manifest content from memory to Kubernetes
func (d *Deployer) DeployManifestContent(manifestContent []byte, stackName string, configNamespace string, timeout time.Duration) (*DeployResult, error) {
	// Parse and prepare the manifest content
	obj, gvr, err := d.parseAndPrepareManifestContent(manifestContent, stackName, configNamespace)
	if err != nil {
		return nil, err
	}

	// Apply the resource to Kubernetes
	operation, result, err := d.applyResource(obj, gvr, stackName)
	if err != nil {
		return &DeployResult{
			Resource:  obj,
			Operation: operation,
			Status:    "failed",
			Error:     err,
			Timestamp: time.Now(),
		}, nil
	}

	// Poll for completion and return result
	status := d.determineStatus(operation, gvr, result, stackName, timeout)
	return &DeployResult{
		Resource:  result,
		Operation: operation,
		Status:    status,
		Error:     nil,
		Timestamp: time.Now(),
	}, nil
}

// parseAndPrepareManifest reads, parses, and prepares a manifest file
func (d *Deployer) parseAndPrepareManifest(manifestPath, stackName, configNamespace string) (*unstructured.Unstructured, schema.GroupVersionResource, error) {
	// Read the manifest file
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, schema.GroupVersionResource{}, fmt.Errorf("error reading manifest file: %v", err)
	}

	// Parse the YAML content
	decoder := k8syaml.NewYAMLOrJSONDecoder(bytes.NewReader(manifestData), 4096)
	var obj unstructured.Unstructured
	if err := decoder.Decode(&obj); err != nil {
		return nil, schema.GroupVersionResource{}, fmt.Errorf("error parsing YAML: %v", err)
	}

	// Set namespace
	namespace := d.determineNamespace(obj.GetNamespace(), configNamespace)
	obj.SetNamespace(namespace)

	// Add stack name annotation
	d.addStackAnnotation(&obj, stackName)

	// Get the GVR (GroupVersionResource) for the resource
	gvr, err := d.getGVR(obj.GetAPIVersion(), obj.GetKind())
	if err != nil {
		return nil, schema.GroupVersionResource{}, fmt.Errorf("failed to get GVR for %s/%s: %v", obj.GetAPIVersion(), obj.GetKind(), err)
	}

	d.logger.Debug("Starting apply operation",
		"stack", stackName,
		"apiVersion", obj.GetAPIVersion(),
		"kind", obj.GetKind(),
		"name", obj.GetName(),
		"namespace", namespace)

	return &obj, gvr, nil
}

// parseAndPrepareManifestContent parses and prepares manifest content from memory
func (d *Deployer) parseAndPrepareManifestContent(manifestContent []byte, stackName, configNamespace string) (*unstructured.Unstructured, schema.GroupVersionResource, error) {
	// Parse the YAML content
	decoder := k8syaml.NewYAMLOrJSONDecoder(bytes.NewReader(manifestContent), 4096)
	var obj unstructured.Unstructured
	if err := decoder.Decode(&obj); err != nil {
		return nil, schema.GroupVersionResource{}, fmt.Errorf("error parsing YAML: %v", err)
	}

	// Set namespace
	namespace := d.determineNamespace(obj.GetNamespace(), configNamespace)
	obj.SetNamespace(namespace)

	// Add stack name annotation
	d.addStackAnnotation(&obj, stackName)

	// Get the GVR (GroupVersionResource) for the resource
	gvr, err := d.getGVR(obj.GetAPIVersion(), obj.GetKind())
	if err != nil {
		return nil, schema.GroupVersionResource{}, fmt.Errorf("failed to get GVR for %s/%s: %v", obj.GetAPIVersion(), obj.GetKind(), err)
	}

	d.logger.Debug("Starting apply operation",
		"stack", stackName,
		"apiVersion", obj.GetAPIVersion(),
		"kind", obj.GetKind(),
		"name", obj.GetName(),
		"namespace", namespace)

	return &obj, gvr, nil
}

// determineNamespace determines the namespace to use
func (d *Deployer) determineNamespace(manifestNamespace, configNamespace string) string {
	if manifestNamespace != "" {
		return manifestNamespace
	}
	if configNamespace != "" {
		return configNamespace
	}
	return "default"
}

// addStackAnnotation adds the stack name annotation to the resource
func (d *Deployer) addStackAnnotation(obj *unstructured.Unstructured, stackName string) {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations["frankthetank.cloud/stack-name"] = stackName
	obj.SetAnnotations(annotations)
}

// applyResource applies the resource to Kubernetes
func (d *Deployer) applyResource(obj *unstructured.Unstructured, gvr schema.GroupVersionResource, stackName string) (string, *unstructured.Unstructured, error) {
	namespace := obj.GetNamespace()
	name := obj.GetName()

	// Check if resource already exists
	existing, err := d.dynamicClient.Resource(gvr).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})

	if err != nil {
		// Resource doesn't exist, create it
		d.logger.Warn("Resource does not exist, creating", "stack", stackName, "name", name, "namespace", namespace)
		result, err := d.dynamicClient.Resource(gvr).Namespace(namespace).Create(context.TODO(), obj, metav1.CreateOptions{})
		return "created", result, err
	}

	// Resource exists, check if it needs applying
	if d.needsUpdate(existing, obj) {
		d.logger.Warn("Updating existing resource", "stack", stackName, "name", name, "namespace", namespace)
		obj.SetResourceVersion(existing.GetResourceVersion()) // Set resource version for update
		result, err := d.dynamicClient.Resource(gvr).Namespace(namespace).Update(context.TODO(), obj, metav1.UpdateOptions{})
		return "applied", result, err
	}

	// No changes needed
	return "no-change", existing, nil
}

// determineStatus determines the final status of the deployment
func (d *Deployer) determineStatus(operation string, gvr schema.GroupVersionResource, result *unstructured.Unstructured, stackName string, timeout time.Duration) string {
	if operation == "created" || operation == "applied" {
		status, err := d.pollForCompletion(gvr, result.GetNamespace(), result.GetName(), stackName, result, timeout)
		if err != nil {
			d.logger.Warn("Error polling for completion", "stack", stackName, "error", err)
		}
		return status
	}

	// No changes made, resource is already up to date
	d.logger.Info("Resource is already up to date", "stack", stackName, "name", result.GetName(), "namespace", result.GetNamespace())
	return "ready"
}

// getGVR converts an API version and kind to a GroupVersionResource
func (d *Deployer) getGVR(apiVersion, kind string) (schema.GroupVersionResource, error) {
	// Parse the API version
	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return schema.GroupVersionResource{}, fmt.Errorf("invalid API version: %v", err)
	}

	// Map common kinds to their resource names
	resourceMap := map[string]string{
		"Deployment":            "deployments",
		"StatefulSet":           "statefulsets",
		"DaemonSet":             "daemonsets",
		"Service":               "services",
		"ConfigMap":             "configmaps",
		"Secret":                "secrets",
		"Pod":                   "pods",
		"Job":                   "jobs",
		"CronJob":               "cronjobs",
		"Ingress":               "ingresses",
		"PersistentVolume":      "persistentvolumes",
		"PersistentVolumeClaim": "persistentvolumeclaims",
	}

	resource, exists := resourceMap[kind]
	if !exists {
		// For unknown resources, try to guess the resource name
		// This is a simple heuristic and might not work for all resources
		resource = kind + "s"
	}

	return schema.GroupVersionResource{
		Group:    gv.Group,
		Version:  gv.Version,
		Resource: resource,
	}, nil
}
