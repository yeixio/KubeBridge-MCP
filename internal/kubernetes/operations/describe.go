package operations

import (
	"context"
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// DescribeOperation handles describing Kubernetes resources
type DescribeOperation struct {
	client dynamic.Interface
}

// NewDescribeOperation creates a new describe operation handler
func NewDescribeOperation(client dynamic.Interface) *DescribeOperation {
	return &DescribeOperation{
		client: client,
	}
}

// Describe describes a resource in a human-readable format
// For MVP, we return formatted JSON. In the future, we can enhance this
// to use kubectl's describe logic for better formatting
func (o *DescribeOperation) Describe(ctx context.Context, gvr schema.GroupVersionResource, namespace, name string) (string, error) {
	getOp := NewGetOperation(o.client)
	obj, err := getOp.Get(ctx, gvr, namespace, name)
	if err != nil {
		return "", err
	}

	// Format as pretty JSON
	jsonBytes, err := json.MarshalIndent(obj.Object, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal resource: %w", err)
	}

	return string(jsonBytes), nil
}
