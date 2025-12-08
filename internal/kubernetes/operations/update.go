package operations

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// UpdateOperation handles updating Kubernetes resources
type UpdateOperation struct {
	client dynamic.Interface
}

// NewUpdateOperation creates a new update operation handler
func NewUpdateOperation(client dynamic.Interface) *UpdateOperation {
	return &UpdateOperation{
		client: client,
	}
}

// Update updates an existing resource
func (o *UpdateOperation) Update(ctx context.Context, gvr schema.GroupVersionResource, namespace string, obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	var updated *unstructured.Unstructured
	var err error

	if namespace == "" {
		// Cluster-scoped resource
		updated, err = o.client.Resource(gvr).Update(ctx, obj, metav1.UpdateOptions{})
	} else {
		// Namespaced resource
		updated, err = o.client.Resource(gvr).Namespace(namespace).Update(ctx, obj, metav1.UpdateOptions{})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to update resource: %w", err)
	}

	return updated, nil
}


