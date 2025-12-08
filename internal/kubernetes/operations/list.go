package operations

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// ListOperation handles listing Kubernetes resources
type ListOperation struct {
	client dynamic.Interface
}

// NewListOperation creates a new list operation handler
func NewListOperation(client dynamic.Interface) *ListOperation {
	return &ListOperation{
		client: client,
	}
}

// List lists resources of the specified type
func (o *ListOperation) List(ctx context.Context, gvr schema.GroupVersionResource, namespace string) ([]*unstructured.Unstructured, error) {
	var list *unstructured.UnstructuredList
	var err error

	if namespace == "" {
		// Cluster-scoped resource or list all namespaces
		list, err = o.client.Resource(gvr).List(ctx, metav1.ListOptions{})
	} else {
		// Namespaced resource
		list, err = o.client.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}

	// Convert []unstructured.Unstructured to []*unstructured.Unstructured
	items := make([]*unstructured.Unstructured, len(list.Items))
	for i := range list.Items {
		items[i] = &list.Items[i]
	}

	return items, nil
}

