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

// Deployer handles actual Kubernetes applications
type Deployer struct {
	dynamicClient dynamic.Interface
	clientset     kubernetes.Interface
	logger        *slog.Logger
}

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

// DeployResult represents the result of a deployment operation
type DeployResult struct {
	Resource  *unstructured.Unstructured
	Operation string // "created", "updated", "no-change"
	Status    string
	Error     error
	Timestamp time.Time
}

// DeployManifest applies a single manifest file to Kubernetes
func (d *Deployer) DeployManifest(manifestPath string, timeout time.Duration) (*DeployResult, error) {
	// Read the manifest file
	manifestData, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("error reading manifest file: %v", err)
	}

	// Parse the YAML content
	decoder := k8syaml.NewYAMLOrJSONDecoder(bytes.NewReader(manifestData), 4096)

	var obj unstructured.Unstructured
	err = decoder.Decode(&obj)
	if err != nil {
		return nil, fmt.Errorf("error parsing YAML: %v", err)
	}

	// Get resource information
	apiVersion := obj.GetAPIVersion()
	kind := obj.GetKind()
	name := obj.GetName()
	namespace := obj.GetNamespace()

	if namespace == "" {
		namespace = "default"
	}

	d.logger.Debug("Starting apply operation",
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
		d.logger.Warn("Resource does not exist, creating", "name", name, "namespace", namespace)
		result, err = d.dynamicClient.Resource(gvr).Namespace(namespace).Create(context.TODO(), &obj, metav1.CreateOptions{})
		operation = "created"
	} else {
		// Resource exists, check if it needs applying
		if d.needsUpdate(existing, &obj) {
			d.logger.Warn("Updating existing resource", "name", name, "namespace", namespace)
			result, err = d.dynamicClient.Resource(gvr).Namespace(namespace).Update(context.TODO(), &obj, metav1.UpdateOptions{})
			operation = "applied"
		} else {
			d.logger.Info("Resource is already up to date", "name", name, "namespace", namespace)
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

	// Poll for completion if it's a deployment or similar workload
	status, err := d.pollForCompletion(gvr, namespace, name, result, timeout)
	if err != nil {
		d.logger.Warn("Error polling for completion", "error", err)
	}

	return &DeployResult{
		Resource:  result,
		Operation: operation,
		Status:    status,
		Error:     nil,
		Timestamp: time.Now(),
	}, nil
}

// getGVR converts apiVersion and kind to GroupVersionResource
func (d *Deployer) getGVR(apiVersion, kind string) (schema.GroupVersionResource, error) {
	// Parse apiVersion
	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return schema.GroupVersionResource{}, err
	}

	// Map common kinds to their resource names
	resourceMap := map[string]string{
		"Deployment":            "deployments",
		"Service":               "services",
		"ConfigMap":             "configmaps",
		"Secret":                "secrets",
		"PersistentVolumeClaim": "persistentvolumeclaims",
		"Ingress":               "ingresses",
		"Job":                   "jobs",
		"CronJob":               "cronjobs",
		"StatefulSet":           "statefulsets",
		"DaemonSet":             "daemonsets",
		"ReplicaSet":            "replicasets",
		"Pod":                   "pods",
	}

	resource, exists := resourceMap[kind]
	if !exists {
		return schema.GroupVersionResource{}, fmt.Errorf("unsupported resource kind: %s", kind)
	}

	return schema.GroupVersionResource{
		Group:    gv.Group,
		Version:  gv.Version,
		Resource: resource,
	}, nil
}

// needsUpdate checks if a resource needs updating
func (d *Deployer) needsUpdate(existing, desired *unstructured.Unstructured) bool {
	// Compare only the spec fields that actually matter for deployments
	existingSpec, _, _ := unstructured.NestedMap(existing.Object, "spec")
	desiredSpec, _, _ := unstructured.NestedMap(desired.Object, "spec")

	if existingSpec == nil || desiredSpec == nil {
		return (existingSpec == nil) != (desiredSpec == nil)
	}

	// For deployments, compare key fields that matter
	kind := desired.GetKind()

	switch kind {
	case "Deployment":
		return d.deploymentNeedsUpdate(existingSpec, desiredSpec)
	case "StatefulSet":
		return d.statefulSetNeedsUpdate(existingSpec, desiredSpec)
	case "DaemonSet":
		return d.daemonSetNeedsUpdate(existingSpec, desiredSpec)
	case "Job":
		return d.jobNeedsUpdate(existingSpec, desiredSpec)
	default:
		// For other resources, do a simple spec comparison
		return !d.mapsEqual(existingSpec, desiredSpec)
	}
}

// deploymentNeedsUpdate checks if a deployment needs updating by comparing key fields
func (d *Deployer) deploymentNeedsUpdate(existing, desired map[string]interface{}) bool {
	// Compare replicas
	if !d.compareField(existing, desired, "replicas") {
		d.logger.Debug("Deployment needs update: replicas differ", "existing", existing["replicas"], "desired", desired["replicas"])
		return true
	}

	// Compare template (which contains the actual pod spec)
	existingTemplate, _, _ := unstructured.NestedMap(existing, "template")
	desiredTemplate, _, _ := unstructured.NestedMap(desired, "template")

	if !d.templateEqual(existingTemplate, desiredTemplate) {
		d.logger.Debug("Deployment needs update: template differs")
		return true
	}

	d.logger.Debug("Deployment is up to date")
	return false
}

// statefulSetNeedsUpdate checks if a statefulset needs updating
func (d *Deployer) statefulSetNeedsUpdate(existing, desired map[string]interface{}) bool {
	// Compare replicas
	if !d.compareField(existing, desired, "replicas") {
		return true
	}

	// Compare template
	existingTemplate, _, _ := unstructured.NestedMap(existing, "template")
	desiredTemplate, _, _ := unstructured.NestedMap(desired, "template")

	if !d.templateEqual(existingTemplate, desiredTemplate) {
		return true
	}

	return false
}

// daemonSetNeedsUpdate checks if a daemonset needs updating
func (d *Deployer) daemonSetNeedsUpdate(existing, desired map[string]interface{}) bool {
	// Compare template
	existingTemplate, _, _ := unstructured.NestedMap(existing, "template")
	desiredTemplate, _, _ := unstructured.NestedMap(desired, "template")

	if !d.templateEqual(existingTemplate, desiredTemplate) {
		return true
	}

	return false
}

// jobNeedsUpdate checks if a job needs updating
func (d *Deployer) jobNeedsUpdate(existing, desired map[string]interface{}) bool {
	// Compare template
	existingTemplate, _, _ := unstructured.NestedMap(existing, "template")
	desiredTemplate, _, _ := unstructured.NestedMap(desired, "template")

	if !d.templateEqual(existingTemplate, desiredTemplate) {
		return true
	}

	return false
}

// templateEqual compares pod templates for equality
func (d *Deployer) templateEqual(existing, desired map[string]interface{}) bool {
	if existing == nil || desired == nil {
		return (existing == nil) != (desired == nil)
	}

	// Compare spec within template
	existingSpec, _, _ := unstructured.NestedMap(existing, "spec")
	desiredSpec, _, _ := unstructured.NestedMap(desired, "spec")

	if existingSpec == nil || desiredSpec == nil {
		return (existingSpec == nil) != (desiredSpec == nil)
	}

	// Compare containers
	existingContainers, _, _ := unstructured.NestedSlice(existingSpec, "containers")
	desiredContainers, _, _ := unstructured.NestedSlice(desiredSpec, "containers")

	if !d.containersEqual(existingContainers, desiredContainers) {
		d.logger.Debug("Template differs: containers differ")
		return false
	}

	return true
}

// containersEqual compares container specs
func (d *Deployer) containersEqual(existing, desired []interface{}) bool {
	if len(existing) != len(desired) {
		d.logger.Debug("Containers differ: length", "existing", len(existing), "desired", len(desired))
		return false
	}

	for i, existingContainer := range existing {
		existingMap, ok := existingContainer.(map[string]interface{})
		if !ok {
			d.logger.Debug("Containers differ: existing container not a map", "index", i)
			return false
		}

		desiredMap, ok := desired[i].(map[string]interface{})
		if !ok {
			d.logger.Debug("Containers differ: desired container not a map", "index", i)
			return false
		}

		// Compare key container fields
		if !d.compareField(existingMap, desiredMap, "name") {
			d.logger.Debug("Containers differ: name", "existing", existingMap["name"], "desired", desiredMap["name"])
			return false
		}
		if !d.compareField(existingMap, desiredMap, "image") {
			d.logger.Debug("Containers differ: image", "existing", existingMap["image"], "desired", desiredMap["image"])
			return false
		}
		if !d.comparePorts(existingMap, desiredMap) {
			d.logger.Debug("Containers differ: ports", "existing", existingMap["ports"], "desired", desiredMap["ports"])
			return false
		}
	}

	return true
}

// comparePorts compares port specifications, handling default protocol values
func (d *Deployer) comparePorts(existing, desired map[string]interface{}) bool {
	existingPorts, existingExists := existing["ports"]
	desiredPorts, desiredExists := desired["ports"]

	if existingExists != desiredExists {
		return false
	}

	if !existingExists {
		return true
	}

	// Handle slices of ports
	existingSlice, ok := existingPorts.([]interface{})
	if !ok {
		return existingPorts == desiredPorts
	}

	desiredSlice, ok := desiredPorts.([]interface{})
	if !ok {
		return false
	}

	if len(existingSlice) != len(desiredSlice) {
		return false
	}

	for i, existingPort := range existingSlice {
		existingPortMap, ok := existingPort.(map[string]interface{})
		if !ok {
			return false
		}

		desiredPortMap, ok := desiredSlice[i].(map[string]interface{})
		if !ok {
			return false
		}

		// Compare port fields, handling default protocol
		if !d.comparePortFields(existingPortMap, desiredPortMap) {
			return false
		}
	}

	return true
}

// comparePortFields compares individual port fields
func (d *Deployer) comparePortFields(existing, desired map[string]interface{}) bool {
	// Compare containerPort
	if existing["containerPort"] != desired["containerPort"] {
		return false
	}

	// Compare protocol, but treat missing protocol as "TCP" (Kubernetes default)
	existingProtocol := existing["protocol"]
	desiredProtocol := desired["protocol"]

	// If either is missing, treat as "TCP"
	if existingProtocol == nil {
		existingProtocol = "TCP"
	}
	if desiredProtocol == nil {
		desiredProtocol = "TCP"
	}

	return existingProtocol == desiredProtocol
}

// compareField compares a specific field between two maps
func (d *Deployer) compareField(existing, desired map[string]interface{}, field string) bool {
	existingVal, existingExists := existing[field]
	desiredVal, desiredExists := desired[field]

	if existingExists != desiredExists {
		return false
	}

	if !existingExists {
		return true
	}

	// Handle slices specially
	if existingSlice, ok := existingVal.([]interface{}); ok {
		desiredSlice, ok := desiredVal.([]interface{})
		if !ok {
			return false
		}
		return d.slicesEqual(existingSlice, desiredSlice)
	}

	return existingVal == desiredVal
}

// slicesEqual compares two slices for equality
func (d *Deployer) slicesEqual(existing, desired []interface{}) bool {
	if len(existing) != len(desired) {
		return false
	}

	for i, existingItem := range existing {
		desiredItem := desired[i]

		// Handle nested maps
		if existingMap, ok := existingItem.(map[string]interface{}); ok {
			desiredMap, ok := desiredItem.(map[string]interface{})
			if !ok {
				return false
			}
			if !d.mapsEqual(existingMap, desiredMap) {
				return false
			}
		} else {
			// Direct comparison for primitives
			if existingItem != desiredItem {
				return false
			}
		}
	}

	return true
}

// mapsEqual compares two maps for equality
func (d *Deployer) mapsEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for k, v := range a {
		if b[k] != v {
			return false
		}
	}

	return true
}

// pollForCompletion polls a resource until it reaches a stable state
func (d *Deployer) pollForCompletion(gvr schema.GroupVersionResource, namespace, name string, obj *unstructured.Unstructured, timeout time.Duration) (string, error) {
	kind := obj.GetKind()

	// Only poll for certain resource types
	if kind != "Deployment" && kind != "StatefulSet" && kind != "DaemonSet" && kind != "Job" {
		return "ready", nil
	}

	d.logger.Warn("Waiting for resource to be ready", "kind", kind, "name", name, "namespace", namespace, "timeout", timeout)

	timeoutDuration := time.After(timeout)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutDuration:
			return "timeout", fmt.Errorf("timeout waiting for %s/%s to be ready", kind, name)
		case <-ticker.C:
			resource, err := d.dynamicClient.Resource(gvr).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
			if err != nil {
				d.logger.Warn("Error getting resource during polling", "error", err)
				continue
			}

			status := d.getResourceStatus(resource, kind)

			if status == "Available" || status == "Ready" || status == "Complete" {
				d.logger.Info("Resource is ready", "kind", kind, "name", name, "status", status)
				return status, nil
			}

			if status == "Failed" || status == "ReplicaFailure" {
				d.logger.Error("Resource failed", "kind", kind, "name", name, "status", status)
				return status, fmt.Errorf("resource %s/%s failed with status: %s", kind, name, status)
			}

			// Only log if status is still progressing
			d.logger.Warn("Waiting for resource to be ready", "kind", kind, "name", name, "status", status)
		}
	}
}

// getResourceStatus determines the status of a resource
func (d *Deployer) getResourceStatus(obj *unstructured.Unstructured, kind string) string {
	switch kind {
	case "Deployment":
		return d.getDeploymentStatus(obj)
	case "StatefulSet":
		return d.getStatefulSetStatus(obj)
	case "DaemonSet":
		return d.getDaemonSetStatus(obj)
	case "Job":
		return d.getJobStatus(obj)
	default:
		return "ready"
	}
}

// getDeploymentStatus checks if a deployment is ready using official Kubernetes conditions
func (d *Deployer) getDeploymentStatus(obj *unstructured.Unstructured) string {
	status, _, _ := unstructured.NestedMap(obj.Object, "status")
	if status == nil {
		return "Progressing"
	}

	conditions, _, _ := unstructured.NestedSlice(status, "conditions")
	for _, condition := range conditions {
		cond, ok := condition.(map[string]interface{})
		if !ok {
			continue
		}

		condType, _, _ := unstructured.NestedString(cond, "type")
		condStatus, _, _ := unstructured.NestedString(cond, "status")

		// Check for Available condition (deployment is ready)
		if condType == "Available" && condStatus == "True" {
			return "Available"
		}

		// Check for Progressing condition
		if condType == "Progressing" {
			if condStatus == "True" {
				reason, _, _ := unstructured.NestedString(cond, "reason")
				if reason == "NewReplicaSetAvailable" {
					return "Progressing"
				}
			}
		}

		// Check for ReplicaFailure condition
		if condType == "ReplicaFailure" && condStatus == "True" {
			return "ReplicaFailure"
		}
	}

	// Fallback to replica counts if conditions aren't available
	replicas, _, _ := unstructured.NestedInt64(status, "replicas")
	readyReplicas, _, _ := unstructured.NestedInt64(status, "readyReplicas")
	availableReplicas, _, _ := unstructured.NestedInt64(status, "availableReplicas")

	if replicas > 0 && readyReplicas == replicas && availableReplicas == replicas {
		return "Available"
	}

	return "Progressing"
}

// getStatefulSetStatus checks if a statefulset is ready using official Kubernetes conditions
func (d *Deployer) getStatefulSetStatus(obj *unstructured.Unstructured) string {
	status, _, _ := unstructured.NestedMap(obj.Object, "status")
	if status == nil {
		return "Progressing"
	}

	conditions, _, _ := unstructured.NestedSlice(status, "conditions")
	for _, condition := range conditions {
		cond, ok := condition.(map[string]interface{})
		if !ok {
			continue
		}

		condType, _, _ := unstructured.NestedString(cond, "type")
		condStatus, _, _ := unstructured.NestedString(cond, "status")

		// Check for Ready condition
		if condType == "Ready" && condStatus == "True" {
			return "Ready"
		}

		// Check for Progressing condition
		if condType == "Progressing" {
			return "Progressing"
		}
	}

	// Fallback to replica counts
	replicas, _, _ := unstructured.NestedInt64(status, "replicas")
	readyReplicas, _, _ := unstructured.NestedInt64(status, "readyReplicas")

	if replicas > 0 && readyReplicas == replicas {
		return "Ready"
	}

	return "Progressing"
}

// getDaemonSetStatus checks if a daemonset is ready using official Kubernetes conditions
func (d *Deployer) getDaemonSetStatus(obj *unstructured.Unstructured) string {
	status, _, _ := unstructured.NestedMap(obj.Object, "status")
	if status == nil {
		return "Progressing"
	}

	conditions, _, _ := unstructured.NestedSlice(status, "conditions")
	for _, condition := range conditions {
		cond, ok := condition.(map[string]interface{})
		if !ok {
			continue
		}

		condType, _, _ := unstructured.NestedString(cond, "type")
		condStatus, _, _ := unstructured.NestedString(cond, "status")

		// Check for Ready condition
		if condType == "Ready" && condStatus == "True" {
			return "Ready"
		}

		// Check for Progressing condition
		if condType == "Progressing" {
			return "Progressing"
		}
	}

	// Fallback to replica counts
	desiredNumberScheduled, _, _ := unstructured.NestedInt64(status, "desiredNumberScheduled")
	numberReady, _, _ := unstructured.NestedInt64(status, "numberReady")

	if desiredNumberScheduled > 0 && numberReady == desiredNumberScheduled {
		return "Ready"
	}

	return "Progressing"
}

// getJobStatus checks if a job is complete using official Kubernetes conditions
func (d *Deployer) getJobStatus(obj *unstructured.Unstructured) string {
	status, _, _ := unstructured.NestedMap(obj.Object, "status")
	if status == nil {
		return "Progressing"
	}

	conditions, _, _ := unstructured.NestedSlice(status, "conditions")
	for _, condition := range conditions {
		cond, ok := condition.(map[string]interface{})
		if !ok {
			continue
		}

		condType, _, _ := unstructured.NestedString(cond, "type")
		condStatus, _, _ := unstructured.NestedString(cond, "status")

		if condType == "Complete" && condStatus == "True" {
			return "Complete"
		}
		if condType == "Failed" && condStatus == "True" {
			return "Failed"
		}
	}

	return "Progressing"
}
