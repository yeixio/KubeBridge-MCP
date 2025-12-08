package operations

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// CreateOperation handles creating Kubernetes resources
type CreateOperation struct {
	client dynamic.Interface
}

// NewCreateOperation creates a new create operation handler
func NewCreateOperation(client dynamic.Interface) *CreateOperation {
	return &CreateOperation{
		client: client,
	}
}

// Create creates a new resource
func (o *CreateOperation) Create(ctx context.Context, gvr schema.GroupVersionResource, namespace string, obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	var created *unstructured.Unstructured
	var err error

	if namespace == "" {
		// Cluster-scoped resource
		created, err = o.client.Resource(gvr).Create(ctx, obj, metav1.CreateOptions{})
	} else {
		// Namespaced resource - ensure namespace is set
		if obj.GetNamespace() == "" {
			obj.SetNamespace(namespace)
		}
		created, err = o.client.Resource(gvr).Namespace(namespace).Create(ctx, obj, metav1.CreateOptions{})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	return created, nil
}
