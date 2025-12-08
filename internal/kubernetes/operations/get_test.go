package operations

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
)

func TestGetOperation_Get(t *testing.T) {
	ctx := context.Background()

	// Create test pod
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "container1", Image: "nginx:latest"},
			},
		},
	}

	// Create fake dynamic client
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("Failed to add corev1 to scheme: %v", err)
	}
	client := fake.NewSimpleDynamicClient(scheme, pod)

	// Create get operation
	op := NewGetOperation(client)

	// Test getting pod
	gvr := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}

	item, err := op.Get(ctx, gvr, "default", "test-pod")
	if err != nil {
		t.Fatalf("Failed to get pod: %v", err)
	}

	if item == nil {
		t.Fatal("Expected pod, got nil")
	}

	if item.GetName() != "test-pod" {
		t.Errorf("Expected pod name 'test-pod', got '%s'", item.GetName())
	}

	if item.GetNamespace() != "default" {
		t.Errorf("Expected namespace 'default', got '%s'", item.GetNamespace())
	}
}

func TestGetOperation_GetNotFound(t *testing.T) {
	ctx := context.Background()

	// Create fake dynamic client with no objects
	scheme := runtime.NewScheme()
	corev1.AddToScheme(scheme)
	client := fake.NewSimpleDynamicClient(scheme)

	// Create get operation
	op := NewGetOperation(client)

	// Test getting non-existent pod
	gvr := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}

	_, err := op.Get(ctx, gvr, "default", "nonexistent-pod")
	if err == nil {
		t.Fatal("Expected error for non-existent pod")
	}
}

func TestGetOperation_GetClusterScoped(t *testing.T) {
	ctx := context.Background()

	// Create test namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-ns",
		},
	}

	// Create fake dynamic client
	scheme := runtime.NewScheme()
	corev1.AddToScheme(scheme)
	client := fake.NewSimpleDynamicClient(scheme, ns)

	// Create get operation
	op := NewGetOperation(client)

	// Test getting namespace (cluster-scoped)
	gvr := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "namespaces",
	}

	item, err := op.Get(ctx, gvr, "", "test-ns")
	if err != nil {
		t.Fatalf("Failed to get namespace: %v", err)
	}

	if item == nil {
		t.Fatal("Expected namespace, got nil")
	}

	if item.GetName() != "test-ns" {
		t.Errorf("Expected namespace name 'test-ns', got '%s'", item.GetName())
	}
}
