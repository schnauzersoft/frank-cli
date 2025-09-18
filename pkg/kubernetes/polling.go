package kubernetes

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// pollForCompletion polls the Kubernetes API until the resource is ready or timeout
func (d *Deployer) pollForCompletion(gvr schema.GroupVersionResource, namespace, name, stackName string, resource *unstructured.Unstructured, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	d.logger.Warn("Waiting for resource to be ready", "stack", stackName, "name", name, "namespace", namespace)

	for {
		select {
		case <-ctx.Done():
			return "timeout", fmt.Errorf("timeout waiting for resource to be ready")
		case <-ticker.C:
			// Get the current state of the resource
			current, err := d.dynamicClient.Resource(gvr).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
			if err != nil {
				d.logger.Warn("Error getting resource during polling", "stack", stackName, "error", err)
				continue
			}

			// Check the status based on resource type
			status := d.getResourceStatus(current)
			if status == "Ready" || status == "Available" || status == "Complete" {
				d.logger.Info("Resource is ready", "stack", stackName, "name", name, "namespace", namespace, "status", status)
				return status, nil
			}
			if status == "Failed" || status == "ReplicaFailure" {
				d.logger.Error("Resource failed", "stack", stackName, "name", name, "namespace", namespace, "status", status)
				return status, fmt.Errorf("resource failed with status: %s", status)
			}

			// Still progressing, continue polling
			d.logger.Debug("Resource still progressing", "stack", stackName, "name", name, "namespace", namespace, "status", status)
		}
	}
}

// getResourceStatus determines the status of a Kubernetes resource
func (d *Deployer) getResourceStatus(resource *unstructured.Unstructured) string {
	kind := resource.GetKind()

	switch kind {
	case "Deployment":
		return d.getDeploymentStatus(resource)
	case "StatefulSet":
		return d.getStatefulSetStatus(resource)
	case "DaemonSet":
		return d.getDaemonSetStatus(resource)
	case "Job":
		return d.getJobStatus(resource)
	case "Service":
		return "Ready" // Services are immediately ready
	case "ConfigMap":
		return "Ready" // ConfigMaps are immediately ready
	case "Secret":
		return "Ready" // Secrets are immediately ready
	case "Pod":
		return d.getPodStatus(resource)
	default:
		// For unknown resources, check if they have a ready condition
		return d.getGenericStatus(resource)
	}
}

// getDeploymentStatus checks the status of a Deployment
func (d *Deployer) getDeploymentStatus(resource *unstructured.Unstructured) string {
	status, _, _ := unstructured.NestedMap(resource.Object, "status")
	if status == nil {
		return "Progressing"
	}

	// Check conditions
	conditions, _, _ := unstructured.NestedSlice(status, "conditions")
	for _, cond := range conditions {
		condMap, ok := cond.(map[string]interface{})
		if !ok {
			continue
		}

		condType, _, _ := unstructured.NestedString(condMap, "type")
		condStatus, _, _ := unstructured.NestedString(condMap, "status")

		if condType == "Available" && condStatus == "True" {
			return "Available"
		}
		if condType == "Progressing" && condStatus == "False" {
			return "Failed"
		}
	}

	return "Progressing"
}

// getStatefulSetStatus checks the status of a StatefulSet
func (d *Deployer) getStatefulSetStatus(resource *unstructured.Unstructured) string {
	status, _, _ := unstructured.NestedMap(resource.Object, "status")
	if status == nil {
		return "Progressing"
	}

	// Check conditions
	conditions, _, _ := unstructured.NestedSlice(status, "conditions")
	for _, cond := range conditions {
		condMap, ok := cond.(map[string]interface{})
		if !ok {
			continue
		}

		condType, _, _ := unstructured.NestedString(condMap, "type")
		condStatus, _, _ := unstructured.NestedString(condMap, "status")

		if condType == "Ready" && condStatus == "True" {
			return "Ready"
		}
	}

	return "Progressing"
}

// getDaemonSetStatus checks the status of a DaemonSet
func (d *Deployer) getDaemonSetStatus(resource *unstructured.Unstructured) string {
	status, _, _ := unstructured.NestedMap(resource.Object, "status")
	if status == nil {
		return "Progressing"
	}

	// Check conditions
	conditions, _, _ := unstructured.NestedSlice(status, "conditions")
	for _, cond := range conditions {
		condMap, ok := cond.(map[string]interface{})
		if !ok {
			continue
		}

		condType, _, _ := unstructured.NestedString(condMap, "type")
		condStatus, _, _ := unstructured.NestedString(condMap, "status")

		if condType == "Ready" && condStatus == "True" {
			return "Ready"
		}
	}

	return "Progressing"
}

// getJobStatus checks the status of a Job
func (d *Deployer) getJobStatus(resource *unstructured.Unstructured) string {
	status, _, _ := unstructured.NestedMap(resource.Object, "status")
	if status == nil {
		return "Progressing"
	}

	// Check conditions
	conditions, _, _ := unstructured.NestedSlice(status, "conditions")
	for _, cond := range conditions {
		condMap, ok := cond.(map[string]interface{})
		if !ok {
			continue
		}

		condType, _, _ := unstructured.NestedString(condMap, "type")
		condStatus, _, _ := unstructured.NestedString(condMap, "status")

		if condType == "Complete" && condStatus == "True" {
			return "Complete"
		}
		if condType == "Failed" && condStatus == "True" {
			return "Failed"
		}
	}

	return "Progressing"
}

// getPodStatus checks the status of a Pod
func (d *Deployer) getPodStatus(resource *unstructured.Unstructured) string {
	status, _, _ := unstructured.NestedMap(resource.Object, "status")
	if status == nil {
		return "Progressing"
	}

	phase, _, _ := unstructured.NestedString(status, "phase")
	switch phase {
	case "Running":
		return "Ready"
	case "Succeeded":
		return "Complete"
	case "Failed":
		return "Failed"
	case "Pending":
		return "Progressing"
	default:
		return "Progressing"
	}
}

// getGenericStatus checks the status of any resource with conditions
func (d *Deployer) getGenericStatus(resource *unstructured.Unstructured) string {
	status, _, _ := unstructured.NestedMap(resource.Object, "status")
	if status == nil {
		return "Progressing"
	}

	// Check conditions
	conditions, _, _ := unstructured.NestedSlice(status, "conditions")
	for _, cond := range conditions {
		condMap, ok := cond.(map[string]interface{})
		if !ok {
			continue
		}

		condType, _, _ := unstructured.NestedString(condMap, "type")
		condStatus, _, _ := unstructured.NestedString(condMap, "status")

		if condType == "Complete" && condStatus == "True" {
			return "Complete"
		}
		if condType == "Failed" && condStatus == "True" {
			return "Failed"
		}
	}

	return "Progressing"
}
