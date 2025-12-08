package operations

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// ApplyOperation handles applying Kubernetes resources (server-side apply)
type ApplyOperation struct {
	client dynamic.Interface
}

// NewApplyOperation creates a new apply operation handler
func NewApplyOperation(client dynamic.Interface) *ApplyOperation {
	return &ApplyOperation{
		client: client,
	}
}

// Apply applies a resource using server-side apply
func (o *ApplyOperation) Apply(ctx context.Context, gvr schema.GroupVersionResource, namespace string, obj *unstructured.Unstructured, fieldManager string) (*unstructured.Unstructured, error) {
	if fieldManager == "" {
		fieldManager = "kubebridge-mcp"
	}

	var applied *unstructured.Unstructured
	var err error

	applyOptions := metav1.ApplyOptions{
		FieldManager: fieldManager,
	}

	if namespace == "" {
		// Cluster-scoped resource
		applied, err = o.client.Resource(gvr).Apply(ctx, obj.GetName(), obj, applyOptions)
	} else {
		// Namespaced resource - ensure namespace is set
		if obj.GetNamespace() == "" {
			obj.SetNamespace(namespace)
		}
		applied, err = o.client.Resource(gvr).Namespace(namespace).Apply(ctx, obj.GetName(), obj, applyOptions)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to apply resource: %w", err)
	}

	return applied, nil
}
