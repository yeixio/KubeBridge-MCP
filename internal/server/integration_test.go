package server

import (
	"context"
	"os"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/yeixio/KubeBridge-MCP/internal/config"
	"github.com/yeixio/KubeBridge-MCP/internal/kubernetes"
)

// TestIntegration_RouterMode tests the full flow of router mode
func TestIntegration_RouterMode(t *testing.T) {
	// Set router mode
	os.Setenv("KUBEBRIDGE_USE_ROUTER", "true")
	defer os.Unsetenv("KUBEBRIDGE_USE_ROUTER")

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

	// Create fake client manager
	clientManager, err := kubernetes.NewFakeClientManager("test-cluster", []runtime.Object{pod})
	if err != nil {
		t.Fatalf("Failed to create fake client manager: %v", err)
	}

	// Create config
	cfg := &config.Config{
		ClusterName:           "test-cluster",
		EnableWriteOperations: false,
		LogLevel:              "info",
	}

	// Create server
	server, err := NewServer(cfg, clientManager)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Test router tool retrieval
	ctx := context.Background()

	// Test k8s_retrieve_tools using invokeToolHandler
	retrieveTool, exists := server.router.GetTool("k8s_retrieve_tools")
	if !exists {
		t.Fatal("k8s_retrieve_tools not found in router")
	}

	// Invoke using the router's invoke mechanism
	args := map[string]interface{}{
		"query": "pod",
		"limit": 10,
	}

	result, err := invokeToolHandler(ctx, retrieveTool.Handler, args, "k8s_retrieve_tools")
	if err != nil {
		t.Fatalf("Failed to invoke retrieve_tools: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.CallToolResult == nil {
		t.Fatal("Expected CallToolResult, got nil")
	}
}

// TestIntegration_ToolInvocation tests invoking a tool through the router
func TestIntegration_ToolInvocation(t *testing.T) {
	// Set router mode
	os.Setenv("KUBEBRIDGE_USE_ROUTER", "true")
	defer os.Unsetenv("KUBEBRIDGE_USE_ROUTER")

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

	// Create fake client manager
	clientManager, err := kubernetes.NewFakeClientManager("test-cluster", []runtime.Object{pod})
	if err != nil {
		t.Fatalf("Failed to create fake client manager: %v", err)
	}

	// Create config
	cfg := &config.Config{
		ClusterName:           "test-cluster",
		EnableWriteOperations: false,
		LogLevel:              "info",
	}

	// Create server
	server, err := NewServer(cfg, clientManager)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Test k8s_call_tool using invokeToolHandler
	// Note: Since no resources are discovered in fake client, k8s_list_pods won't exist
	// This test verifies the router mechanism works, even if the tool doesn't exist
	callTool, exists := server.router.GetTool("k8s_call_tool")
	if !exists {
		t.Fatal("k8s_call_tool not found in router")
	}

	// Invoke using the router's invoke mechanism
	// This will fail because k8s_list_pods isn't registered (no resources discovered)
	ctx := context.Background()
	args := map[string]interface{}{
		"toolName": "k8s_list_pods",
		"args": map[string]interface{}{
			"namespace": "default",
		},
	}

	result, err := invokeToolHandler(ctx, callTool.Handler, args, "k8s_call_tool")
	// We expect an error because the tool doesn't exist (no resources discovered)
	if err == nil {
		t.Error("Expected error for non-existent tool, got nil")
	}
	_ = result // Acknowledge result may be nil on error
}

// TestIntegration_ServerInitialization tests that server initializes correctly
func TestIntegration_ServerInitialization(t *testing.T) {
	// Create fake client manager
	clientManager, err := kubernetes.NewFakeClientManager("test-cluster", []runtime.Object{})
	if err != nil {
		t.Fatalf("Failed to create fake client manager: %v", err)
	}

	// Create config
	cfg := &config.Config{
		ClusterName:           "test-cluster",
		EnableWriteOperations: false,
		LogLevel:              "info",
	}

	// Create server
	server, err := NewServer(cfg, clientManager)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Verify server components
	if server.config == nil {
		t.Error("Server config is nil")
	}

	if server.clientManager == nil {
		t.Error("Server clientManager is nil")
	}

	if server.mcpServer == nil {
		t.Error("Server mcpServer is nil")
	}

	if server.router == nil {
		t.Error("Server router is nil")
	}

	// Verify tools are registered
	if len(server.registeredTools) == 0 {
		t.Error("Expected tools to be registered")
	}
}
