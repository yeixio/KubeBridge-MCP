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

func TestListOperation_List(t *testing.T) {
	ctx := context.Background()

	// Create test pods
	pod1 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "container1", Image: "nginx:latest"},
			},
		},
	}

	pod2 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod2",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "container2", Image: "nginx:latest"},
			},
		},
	}

	// Create fake dynamic client
	scheme := runtime.NewScheme()
	corev1.AddToScheme(scheme)
	client := fake.NewSimpleDynamicClient(scheme, pod1, pod2)

	// Create list operation
	op := NewListOperation(client)

	// Test listing pods in default namespace
	gvr := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}

	items, err := op.List(ctx, gvr, "default")
	if err != nil {
		t.Fatalf("Failed to list pods: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("Expected 2 pods, got %d", len(items))
	}

	// Verify pod names
	names := make(map[string]bool)
	for _, item := range items {
		names[item.GetName()] = true
	}

	if !names["pod1"] {
		t.Error("Expected pod1 in results")
	}
	if !names["pod2"] {
		t.Error("Expected pod2 in results")
	}
}

func TestListOperation_ListEmpty(t *testing.T) {
	ctx := context.Background()

	// Create fake dynamic client with no objects
	scheme := runtime.NewScheme()
	corev1.AddToScheme(scheme)
	client := fake.NewSimpleDynamicClient(scheme)

	// Create list operation
	op := NewListOperation(client)

	// Test listing pods in default namespace
	gvr := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}

	items, err := op.List(ctx, gvr, "default")
	if err != nil {
		t.Fatalf("Failed to list pods: %v", err)
	}

	if len(items) != 0 {
		t.Fatalf("Expected 0 pods, got %d", len(items))
	}
}

func TestListOperation_ListClusterScoped(t *testing.T) {
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

	// Create list operation
	op := NewListOperation(client)

	// Test listing namespaces (cluster-scoped)
	gvr := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "namespaces",
	}

	items, err := op.List(ctx, gvr, "")
	if err != nil {
		t.Fatalf("Failed to list namespaces: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("Expected 1 namespace, got %d", len(items))
	}

	if items[0].GetName() != "test-ns" {
		t.Errorf("Expected namespace 'test-ns', got '%s'", items[0].GetName())
	}
}

