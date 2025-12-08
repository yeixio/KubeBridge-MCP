package operations

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

// PatchOperation handles patching Kubernetes resources
type PatchOperation struct {
	client dynamic.Interface
}

// NewPatchOperation creates a new patch operation handler
func NewPatchOperation(client dynamic.Interface) *PatchOperation {
	return &PatchOperation{
		client: client,
	}
}

// Patch patches a resource using JSON patch
func (o *PatchOperation) Patch(ctx context.Context, gvr schema.GroupVersionResource, namespace, name string, patchType types.PatchType, patchData []byte) (*unstructured.Unstructured, error) {
	var patched *unstructured.Unstructured
	var err error

	if namespace == "" {
		// Cluster-scoped resource
		patched, err = o.client.Resource(gvr).Patch(ctx, name, patchType, patchData, metav1.PatchOptions{})
	} else {
		// Namespaced resource
		patched, err = o.client.Resource(gvr).Namespace(namespace).Patch(ctx, name, patchType, patchData, metav1.PatchOptions{})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to patch resource: %w", err)
	}

	return patched, nil
}
