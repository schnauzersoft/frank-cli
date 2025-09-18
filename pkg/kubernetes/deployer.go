/*
Copyright © 2025 Ben Sapp ya.bsapp.ru
*/

package kubernetes

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log/slog"
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
func (d *Deployer) DeployManifest(manifestPath string, stackName string, timeout time.Duration) (*DeployResult, error) {
	// Read the manifest file
	manifestData, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("error reading manifest file: %v", err)
	}

	// Parse the YAML content
	decoder := k8syaml.NewYAMLOrJSONDecoder(bytes.NewReader(manifestData), 4096)
	var obj unstructured.Unstructured
	if err := decoder.Decode(&obj); err != nil {
		return nil, fmt.Errorf("error parsing YAML: %v", err)
	}

	apiVersion := obj.GetAPIVersion()
	kind := obj.GetKind()
	name := obj.GetName()
	namespace := obj.GetNamespace()

	if namespace == "" {
		namespace = "default"
	}

	// Add stack name annotation
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations["frankthetank.cloud/stack-name"] = stackName
	obj.SetAnnotations(annotations)

	d.logger.Debug("Starting apply operation",
		"stack", stackName,
		"apiVersion", apiVersion,
		"kind", kind,
		"name", name,
		"namespace", namespace)

	// Get the GVR (GroupVersionResource) for the resource
	gvr, err := d.getGVR(apiVersion, kind)
	if err != nil {
		return nil, fmt.Errorf("failed to get GVR for %s/%s: %v", apiVersion, kind, err)
	}

	// Check if resource already exists
	existing, err := d.dynamicClient.Resource(gvr).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})

	var operation string
	var result *unstructured.Unstructured

	if err != nil {
		// Resource doesn't exist, create it
		d.logger.Warn("Resource does not exist, creating", "stack", stackName, "name", name, "namespace", namespace)
		result, err = d.dynamicClient.Resource(gvr).Namespace(namespace).Create(context.TODO(), &obj, metav1.CreateOptions{})
		operation = "created"
	} else {
		// Resource exists, check if it needs applying
		if d.needsUpdate(existing, &obj) {
			d.logger.Warn("Updating existing resource", "stack", stackName, "name", name, "namespace", namespace)
			obj.SetResourceVersion(existing.GetResourceVersion()) // Set resource version for update
			result, err = d.dynamicClient.Resource(gvr).Namespace(namespace).Update(context.TODO(), &obj, metav1.UpdateOptions{})
			operation = "applied"
		} else {
			result = existing
			operation = "no-change"
		}
	}

	if err != nil {
		return &DeployResult{
			Resource:  &obj,
			Operation: operation,
			Status:    "failed",
			Error:     err,
			Timestamp: time.Now(),
		}, nil
	}

	// Poll for completion only if we made changes
	var status string
	if operation == "created" || operation == "applied" {
		status, err = d.pollForCompletion(gvr, namespace, name, stackName, result, timeout)
		if err != nil {
			d.logger.Warn("Error polling for completion", "stack", stackName, "error", err)
		}
	} else {
		// No changes made, resource is already up to date
		status = "ready"
		d.logger.Info("Resource is already up to date", "stack", stackName, "name", name, "namespace", namespace)
	}

	return &DeployResult{
		Resource:  result,
		Operation: operation,
		Status:    status,
		Error:     err,
		Timestamp: time.Now(),
	}, nil
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
		"Deployment":    "deployments",
		"StatefulSet":   "statefulsets",
		"DaemonSet":     "daemonsets",
		"Service":       "services",
		"ConfigMap":     "configmaps",
		"Secret":        "secrets",
		"Pod":           "pods",
		"Job":           "jobs",
		"CronJob":       "cronjobs",
		"Ingress":       "ingresses",
		"PersistentVolume": "persistentvolumes",
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