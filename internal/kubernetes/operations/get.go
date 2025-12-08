package operations

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// GetOperation handles getting Kubernetes resources
type GetOperation struct {
	client dynamic.Interface
}

// NewGetOperation creates a new get operation handler
func NewGetOperation(client dynamic.Interface) *GetOperation {
	return &GetOperation{
		client: client,
	}
}

// Get gets a specific resource
func (o *GetOperation) Get(ctx context.Context, gvr schema.GroupVersionResource, namespace, name string) (*unstructured.Unstructured, error) {
	var obj *unstructured.Unstructured
	var err error

	if namespace == "" {
		// Cluster-scoped resource
		obj, err = o.client.Resource(gvr).Get(ctx, name, metav1.GetOptions{})
	} else {
		// Namespaced resource
		obj, err = o.client.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get resource: %w", err)
	}

	return obj, nil
}


