package kubernetes

import (
	"log/slog"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// Deployer handles actual Kubernetes applications
type Deployer struct {
	dynamicClient dynamic.Interface
	clientset     kubernetes.Interface
	logger        *slog.Logger
}

// DeployResult represents the result of a single Kubernetes resource application
type DeployResult struct {
	Resource  *unstructured.Unstructured
	Operation string // "created", "applied", "unchanged"
	Status    string // "Progressing", "Available", "Ready", "Complete", "Failed", "ReplicaFailure", "timeout"
	Error     error
	Timestamp time.Time
}

// DeleteResult represents the result of a delete operation
type DeleteResult struct {
	StackName    string
	ResourceType string
	ResourceName string
	Namespace    string
	Error        error
}
