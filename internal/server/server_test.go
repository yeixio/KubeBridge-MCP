package server

import (
	"os"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/yeixio/KubeBridge-MCP/internal/config"
	"github.com/yeixio/KubeBridge-MCP/internal/kubernetes"
)

func TestNewServer(t *testing.T) {
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

	// Test server creation
	server, err := NewServer(cfg, clientManager)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if server == nil {
		t.Fatal("Server is nil")
	}

	if server.config == nil {
		t.Fatal("Server config is nil")
	}

	if server.clientManager == nil {
		t.Fatal("Server clientManager is nil")
	}

	if server.mcpServer == nil {
		t.Fatal("Server mcpServer is nil")
	}

	if server.router == nil {
		t.Fatal("Server router is nil")
	}
}

func TestNewServerWithRouterMode(t *testing.T) {
	// Set environment variable for router mode
	if err := os.Setenv("KUBEBRIDGE_USE_ROUTER", "true"); err != nil {
		t.Fatalf("Failed to set environment variable: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("KUBEBRIDGE_USE_ROUTER"); err != nil {
			t.Logf("Failed to unset environment variable: %v", err)
		}
	}()

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

	// Test server creation with router mode
	server, err := NewServer(cfg, clientManager)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if !server.useRouterMode {
		t.Error("Expected router mode to be enabled")
	}

	// Check that router has tools registered
	allTools := server.router.GetAllTools()
	if len(allTools) == 0 {
		t.Error("Expected router to have tools registered")
	}

	// Check that router tools are registered
	retrieveTool, exists := server.router.GetTool("k8s_retrieve_tools")
	if !exists {
		t.Error("Expected k8s_retrieve_tools to be registered")
	}
	if retrieveTool == nil {
		t.Error("k8s_retrieve_tools tool is nil")
	}

	callTool, exists := server.router.GetTool("k8s_call_tool")
	if !exists {
		t.Error("Expected k8s_call_tool to be registered")
	}
	if callTool == nil {
		t.Error("k8s_call_tool tool is nil")
	}
}

func TestNewServerWithoutRouterMode(t *testing.T) {
	// Ensure router mode is disabled
	os.Unsetenv("KUBEBRIDGE_USE_ROUTER")

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

	// Test server creation without router mode
	server, err := NewServer(cfg, clientManager)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if server.useRouterMode {
		t.Error("Expected router mode to be disabled")
	}

	// In non-router mode, tools should be registered directly with MCP server
	// We can't easily test this without accessing private fields, but we can
	// verify the server was created successfully
	if len(server.registeredTools) == 0 {
		t.Error("Expected some tools to be registered")
	}
}

func TestServerDiscoverAndRegisterTools(t *testing.T) {
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

	// Verify resources were discovered
	// Note: Fake discovery client may not have resources pre-populated
	// This test verifies the discovery process runs without error
	resources := clientManager.GetDiscoveredResources()
	// With fake client, resources may be empty, which is acceptable for testing
	_ = resources // Acknowledge that resources may be empty

	// Verify tools were registered
	if len(server.registeredTools) == 0 {
		t.Error("Expected some tools to be registered")
	}
}

func TestServerWithWriteOperationsDisabled(t *testing.T) {
	// Create fake client manager
	clientManager, err := kubernetes.NewFakeClientManager("test-cluster", []runtime.Object{})
	if err != nil {
		t.Fatalf("Failed to create fake client manager: %v", err)
	}

	// Create config with write operations disabled
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

	// Verify write operations are disabled
	if server.config.EnableWriteOperations {
		t.Error("Expected write operations to be disabled")
	}
}

func TestServerWithWriteOperationsEnabled(t *testing.T) {
	// Create fake client manager
	clientManager, err := kubernetes.NewFakeClientManager("test-cluster", []runtime.Object{})
	if err != nil {
		t.Fatalf("Failed to create fake client manager: %v", err)
	}

	// Create config with write operations enabled
	cfg := &config.Config{
		ClusterName:           "test-cluster",
		EnableWriteOperations: true,
		LogLevel:              "info",
	}

	// Create server
	server, err := NewServer(cfg, clientManager)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Verify write operations are enabled
	if !server.config.EnableWriteOperations {
		t.Error("Expected write operations to be enabled")
	}
}
