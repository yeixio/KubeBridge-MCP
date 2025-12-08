package operations

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// DeleteOperation handles deleting Kubernetes resources
type DeleteOperation struct {
	client dynamic.Interface
}

// NewDeleteOperation creates a new delete operation handler
func NewDeleteOperation(client dynamic.Interface) *DeleteOperation {
	return &DeleteOperation{
		client: client,
	}
}

// Delete deletes a resource
func (o *DeleteOperation) Delete(ctx context.Context, gvr schema.GroupVersionResource, namespace, name string) error {
	var err error

	if namespace == "" {
		// Cluster-scoped resource
		err = o.client.Resource(gvr).Delete(ctx, name, metav1.DeleteOptions{})
	} else {
		// Namespaced resource
		err = o.client.Resource(gvr).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	}

	if err != nil {
		return fmt.Errorf("failed to delete resource: %w", err)
	}

	return nil
}
