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
func (d *Deployer) pollForCompletion(gvr schema.GroupVersionResource, namespace, name, stackName string, timeout time.Duration) (string, error) {
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
			status, err := d.checkResourceStatus(gvr, namespace, name, stackName)
			if err != nil {
				return status, err
			}
			if status != "" {
				return status, nil
			}
		}
	}
}

// checkResourceStatus checks the current status of a resource
func (d *Deployer) checkResourceStatus(gvr schema.GroupVersionResource, namespace, name, stackName string) (string, error) {
	// Get the current state of the resource
	current, err := d.dynamicClient.Resource(gvr).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		d.logger.Warn("Error getting resource during polling", "stack", stackName, "error", err)
		return "", nil // Continue polling
	}

	// Check the status based on resource type
	status := d.getResourceStatus(current)
	return d.handleResourceStatus(status, stackName, name, namespace)
}

// handleResourceStatus handles the resource status and returns appropriate response
func (d *Deployer) handleResourceStatus(status, stackName, name, namespace string) (string, error) {
	if d.isResourceReady(status) {
		d.logger.Info("Resource is ready", "stack", stackName, "name", name, "namespace", namespace, "status", status)
		return status, nil
	}

	if d.isResourceFailed(status) {
		d.logger.Error("Resource failed", "stack", stackName, "name", name, "namespace", namespace, "status", status)
		return status, fmt.Errorf("resource failed with status: %s", status)
	}

	// Still progressing, continue polling
	d.logger.Debug("Resource still progressing", "stack", stackName, "name", name, "namespace", namespace, "status", status)
	return "", nil
}

// isResourceReady checks if the resource is in a ready state
func (d *Deployer) isResourceReady(status string) bool {
	readyStatuses := []string{"Ready", "Available", "Complete"}
	for _, readyStatus := range readyStatuses {
		if status == readyStatus {
			return true
		}
	}
	return false
}

// isResourceFailed checks if the resource is in a failed state
func (d *Deployer) isResourceFailed(status string) bool {
	failedStatuses := []string{"Failed", "ReplicaFailure"}
	for _, failedStatus := range failedStatuses {
		if status == failedStatus {
			return true
		}
	}
	return false
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

	return d.checkDeploymentConditions(status)
}

// checkDeploymentConditions checks deployment-specific conditions
func (d *Deployer) checkDeploymentConditions(status map[string]any) string {
	conditions, _, _ := unstructured.NestedSlice(status, "conditions")

	for _, cond := range conditions {
		condMap, ok := cond.(map[string]any)
		if !ok {
			continue
		}

		condType, _, _ := unstructured.NestedString(condMap, "type")
		condStatus, _, _ := unstructured.NestedString(condMap, "status")

		if d.isDeploymentAvailable(condType, condStatus) {
			return "Available"
		}
		if d.isDeploymentFailed(condType, condStatus) {
			return "Failed"
		}
	}

	return "Progressing"
}

// isDeploymentAvailable checks if deployment is available
func (d *Deployer) isDeploymentAvailable(condType, condStatus string) bool {
	return condType == "Available" && condStatus == "True"
}

// isDeploymentFailed checks if deployment failed
func (d *Deployer) isDeploymentFailed(condType, condStatus string) bool {
	return condType == "Progressing" && condStatus == "False"
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
		condMap, ok := cond.(map[string]any)
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
		condMap, ok := cond.(map[string]any)
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

	return d.checkJobConditions(status)
}

// checkJobConditions checks job-specific conditions
func (d *Deployer) checkJobConditions(status map[string]any) string {
	conditions, _, _ := unstructured.NestedSlice(status, "conditions")

	for _, cond := range conditions {
		condMap, ok := cond.(map[string]any)
		if !ok {
			continue
		}

		condType, _, _ := unstructured.NestedString(condMap, "type")
		condStatus, _, _ := unstructured.NestedString(condMap, "status")

		if d.isJobComplete(condType, condStatus) {
			return "Complete"
		}
		if d.isJobFailed(condType, condStatus) {
			return "Failed"
		}
	}

	return "Progressing"
}

// isJobComplete checks if job is complete
func (d *Deployer) isJobComplete(condType, condStatus string) bool {
	return condType == "Complete" && condStatus == "True"
}

// isJobFailed checks if job failed
func (d *Deployer) isJobFailed(condType, condStatus string) bool {
	return condType == "Failed" && condStatus == "True"
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

	return d.checkGenericConditions(status)
}

// checkGenericConditions checks generic resource conditions
func (d *Deployer) checkGenericConditions(status map[string]any) string {
	conditions, _, _ := unstructured.NestedSlice(status, "conditions")

	for _, cond := range conditions {
		condMap, ok := cond.(map[string]any)
		if !ok {
			continue
		}

		condType, _, _ := unstructured.NestedString(condMap, "type")
		condStatus, _, _ := unstructured.NestedString(condMap, "status")

		if d.isGenericComplete(condType, condStatus) {
			return "Complete"
		}
		if d.isGenericFailed(condType, condStatus) {
			return "Failed"
		}
	}

	return "Progressing"
}

// isGenericComplete checks if generic resource is complete
func (d *Deployer) isGenericComplete(condType, condStatus string) bool {
	return condType == "Complete" && condStatus == "True"
}

// isGenericFailed checks if generic resource failed
func (d *Deployer) isGenericFailed(condType, condStatus string) bool {
	return condType == "Failed" && condStatus == "True"
}
