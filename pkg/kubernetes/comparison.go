package kubernetes

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

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

func (d *Deployer) deploymentNeedsUpdate(existing, desired map[string]any) bool {
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

func (d *Deployer) statefulSetNeedsUpdate(existing, desired map[string]any) bool {
	// Compare replicas
	if !d.compareField(existing, desired, "replicas") {
		return true
	}

	// Compare template
	existingTemplate, _, _ := unstructured.NestedMap(existing, "template")
	desiredTemplate, _, _ := unstructured.NestedMap(desired, "template")

	return !d.templateEqual(existingTemplate, desiredTemplate)
}

func (d *Deployer) daemonSetNeedsUpdate(existing, desired map[string]any) bool {
	// Compare template
	existingTemplate, _, _ := unstructured.NestedMap(existing, "template")
	desiredTemplate, _, _ := unstructured.NestedMap(desired, "template")

	return !d.templateEqual(existingTemplate, desiredTemplate)
}

func (d *Deployer) jobNeedsUpdate(existing, desired map[string]any) bool {
	// Compare template
	existingTemplate, _, _ := unstructured.NestedMap(existing, "template")
	desiredTemplate, _, _ := unstructured.NestedMap(desired, "template")

	return !d.templateEqual(existingTemplate, desiredTemplate)
}

func (d *Deployer) templateEqual(existing, desired map[string]any) bool {
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

func (d *Deployer) containersEqual(existing, desired []any) bool {
	if len(existing) != len(desired) {
		return false
	}

	for i, existingContainer := range existing {
		if !d.compareSingleContainer(existingContainer, desired[i]) {
			return false
		}
	}

	return true
}

func (d *Deployer) compareSingleContainer(existing, desired any) bool {
	existingMap, ok := existing.(map[string]any)
	if !ok {
		return false
	}

	desiredMap, ok := desired.(map[string]any)
	if !ok {
		return false
	}

	// Compare basic fields
	if !d.compareContainerBasicFields(existingMap, desiredMap) {
		return false
	}

	// Compare ports
	if !d.compareContainerPorts(existingMap, desiredMap) {
		return false
	}

	// Compare other important fields
	return d.compareContainerOtherFields(existingMap, desiredMap)
}

func (d *Deployer) compareContainerBasicFields(existing, desired map[string]any) bool {
	// Compare container name
	if !d.compareField(existing, desired, "name") {
		return false
	}

	// Compare image
	if !d.compareField(existing, desired, "image") {
		return false
	}

	return true
}

func (d *Deployer) compareContainerPorts(existing, desired map[string]any) bool {
	existingPorts, _, _ := unstructured.NestedSlice(existing, "ports")
	desiredPorts, _, _ := unstructured.NestedSlice(desired, "ports")

	return d.comparePorts(existingPorts, desiredPorts)
}

func (d *Deployer) compareContainerOtherFields(existing, desired map[string]any) bool {
	fieldsToCompare := []string{"command", "args", "workingDir", "env", "resources", "volumeMounts"}
	for _, field := range fieldsToCompare {
		if !d.compareField(existing, desired, field) {
			return false
		}
	}
	return true
}

// comparePorts compares port slices, handling default TCP protocol
func (d *Deployer) comparePorts(existing, desired []any) bool {
	if len(existing) != len(desired) {
		return false
	}

	for i, existingPort := range existing {
		if !d.compareSinglePort(existingPort, desired[i]) {
			return false
		}
	}

	return true
}

func (d *Deployer) compareSinglePort(existing, desired any) bool {
	existingMap, ok := existing.(map[string]any)
	if !ok {
		return false
	}

	desiredMap, ok := desired.(map[string]any)
	if !ok {
		return false
	}

	// Compare protocol (with default handling)
	if !d.comparePortProtocol(existingMap, desiredMap) {
		return false
	}

	// Compare other port fields
	return d.comparePortOtherFields(existingMap, desiredMap)
}

func (d *Deployer) comparePortProtocol(existing, desired map[string]any) bool {
	existingProtocol := d.getPortProtocol(existing)
	desiredProtocol := d.getPortProtocol(desired)

	return existingProtocol == desiredProtocol
}

func (d *Deployer) getPortProtocol(portMap map[string]any) string {
	protocol := portMap["protocol"]
	if protocol == nil {
		return "TCP"
	}
	return protocol.(string)
}

func (d *Deployer) comparePortOtherFields(existing, desired map[string]any) bool {
	fieldsToCompare := []string{"containerPort", "name", "hostPort"}
	for _, field := range fieldsToCompare {
		if !d.compareField(existing, desired, field) {
			return false
		}
	}
	return true
}

func (d *Deployer) compareField(existing, desired map[string]any, field string) bool {
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
	case []any:
		existingSlice, _ := existingVal.([]any)
		desiredSlice, ok := desiredVal.([]any)
		if !ok {
			return false
		}
		return d.slicesEqual(existingSlice, desiredSlice)
	case map[string]any:
		existingMap, _ := existingVal.(map[string]any)
		desiredMap, ok := desiredVal.(map[string]any)
		if !ok {
			return false
		}
		return d.mapsEqual(existingMap, desiredMap)
	default:
		return existingVal == desiredVal
	}
}

func (d *Deployer) mapsEqual(existing, desired map[string]any) bool {
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

func (d *Deployer) slicesEqual(existing, desired []any) bool {
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

func (d *Deployer) valuesEqual(existing, desired any) bool {
	switch existingVal := existing.(type) {
	case map[string]any:
		desiredMap, ok := desired.(map[string]any)
		if !ok {
			return false
		}
		return d.mapsEqual(existingVal, desiredMap)
	case []any:
		desiredSlice, ok := desired.([]any)
		if !ok {
			return false
		}
		return d.slicesEqual(existingVal, desiredSlice)
	default:
		return existingVal == desired
	}
}
