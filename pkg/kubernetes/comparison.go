package kubernetes

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

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
		return false
	}

	// Compare other important fields
	fieldsToCompare := []string{"restartPolicy", "terminationGracePeriodSeconds", "dnsPolicy", "hostNetwork", "hostPID", "hostIPC"}
	for _, field := range fieldsToCompare {
		if !d.compareField(existingSpec, desiredSpec, field) {
			return false
		}
	}

	return true
}

// containersEqual compares container slices for equality
func (d *Deployer) containersEqual(existing, desired []interface{}) bool {
	if len(existing) != len(desired) {
		return false
	}

	for i, existingContainer := range existing {
		existingMap, ok := existingContainer.(map[string]interface{})
		if !ok {
			return false
		}

		desiredMap, ok := desired[i].(map[string]interface{})
		if !ok {
			return false
		}

		// Compare container name
		if !d.compareField(existingMap, desiredMap, "name") {
			return false
		}

		// Compare image
		if !d.compareField(existingMap, desiredMap, "image") {
			return false
		}

		// Compare ports (handle default protocol)
		existingPorts, _, _ := unstructured.NestedSlice(existingMap, "ports")
		desiredPorts, _, _ := unstructured.NestedSlice(desiredMap, "ports")

		if !d.comparePorts(existingPorts, desiredPorts) {
			return false
		}

		// Compare other important fields
		fieldsToCompare := []string{"command", "args", "workingDir", "env", "resources", "volumeMounts"}
		for _, field := range fieldsToCompare {
			if !d.compareField(existingMap, desiredMap, field) {
				return false
			}
		}
	}

	return true
}

// comparePorts compares port slices, handling default TCP protocol
func (d *Deployer) comparePorts(existing, desired []interface{}) bool {
	if len(existing) != len(desired) {
		return false
	}

	for i, existingPort := range existing {
		existingMap, ok := existingPort.(map[string]interface{})
		if !ok {
			return false
		}

		desiredMap, ok := desired[i].(map[string]interface{})
		if !ok {
			return false
		}

		// Handle default protocol - if not specified, assume TCP
		existingProtocol := existingMap["protocol"]
		desiredProtocol := desiredMap["protocol"]

		if existingProtocol == nil {
			existingProtocol = "TCP"
		}
		if desiredProtocol == nil {
			desiredProtocol = "TCP"
		}

		if existingProtocol != desiredProtocol {
			return false
		}

		// Compare other port fields
		fieldsToCompare := []string{"containerPort", "name", "hostPort"}
		for _, field := range fieldsToCompare {
			if !d.compareField(existingMap, desiredMap, field) {
				return false
			}
		}
	}

	return true
}

// compareField compares a specific field between two maps
func (d *Deployer) compareField(existing, desired map[string]interface{}, field string) bool {
	existingVal := existing[field]
	desiredVal := desired[field]

	// Handle nil values
	if existingVal == nil && desiredVal == nil {
		return true
	}
	if existingVal == nil || desiredVal == nil {
		return false
	}

	// For slices and maps, use deep comparison
	switch existingVal.(type) {
	case []interface{}:
		existingSlice, _ := existingVal.([]interface{})
		desiredSlice, ok := desiredVal.([]interface{})
		if !ok {
			return false
		}
		return d.slicesEqual(existingSlice, desiredSlice)
	case map[string]interface{}:
		existingMap, _ := existingVal.(map[string]interface{})
		desiredMap, ok := desiredVal.(map[string]interface{})
		if !ok {
			return false
		}
		return d.mapsEqual(existingMap, desiredMap)
	default:
		return existingVal == desiredVal
	}
}

// mapsEqual performs deep comparison of two maps
func (d *Deployer) mapsEqual(existing, desired map[string]interface{}) bool {
	if len(existing) != len(desired) {
		return false
	}

	for key, existingVal := range existing {
		desiredVal, exists := desired[key]
		if !exists {
			return false
		}

		if !d.valuesEqual(existingVal, desiredVal) {
			return false
		}
	}

	return true
}

// slicesEqual performs deep comparison of two slices
func (d *Deployer) slicesEqual(existing, desired []interface{}) bool {
	if len(existing) != len(desired) {
		return false
	}

	for i, existingVal := range existing {
		if !d.valuesEqual(existingVal, desired[i]) {
			return false
		}
	}

	return true
}

// valuesEqual performs deep comparison of two values
func (d *Deployer) valuesEqual(existing, desired interface{}) bool {
	switch existingVal := existing.(type) {
	case map[string]interface{}:
		desiredMap, ok := desired.(map[string]interface{})
		if !ok {
			return false
		}
		return d.mapsEqual(existingVal, desiredMap)
	case []interface{}:
		desiredSlice, ok := desired.([]interface{})
		if !ok {
			return false
		}
		return d.slicesEqual(existingVal, desiredSlice)
	default:
		return existingVal == desired
	}
}
