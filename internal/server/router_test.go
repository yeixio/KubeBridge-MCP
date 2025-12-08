package server

import (
	"context"
	"reflect"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestNewRouter(t *testing.T) {
	router := NewRouter()
	if router == nil {
		t.Fatal("NewRouter() returned nil")
	}
	if router.tools == nil {
		t.Fatal("Router.tools is nil")
	}
	if len(router.tools) != 0 {
		t.Fatalf("Expected empty router, got %d tools", len(router.tools))
	}
}

func TestRegisterTool(t *testing.T) {
	router := NewRouter()

	// Create a simple handler
	type TestArgs struct {
		Name string `json:"name"`
	}
	type TestResult struct {
		Message string `json:"message"`
	}

	var handler mcp.ToolHandlerFor[TestArgs, TestResult] = func(
		ctx context.Context,
		req *mcp.CallToolRequest,
		args TestArgs,
	) (*mcp.CallToolResult, TestResult, error) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Test"},
			},
		}, TestResult{Message: "success"}, nil
	}

	router.RegisterToolWithSchema("test_tool", "A test tool", handler)

	tool, exists := router.GetTool("test_tool")
	if !exists {
		t.Fatal("Tool was not registered")
	}
	if tool.Name != "test_tool" {
		t.Errorf("Expected name 'test_tool', got '%s'", tool.Name)
	}
	if tool.Description != "A test tool" {
		t.Errorf("Expected description 'A test tool', got '%s'", tool.Description)
	}
	if tool.Handler == nil {
		t.Fatal("Handler is nil")
	}
}

func TestGetTool(t *testing.T) {
	router := NewRouter()

	// Test non-existent tool
	_, exists := router.GetTool("nonexistent")
	if exists {
		t.Error("Expected nonexistent tool to not exist")
	}

	// Register a tool
	type TestArgs struct {
		Value string `json:"value"`
	}
	type TestResult struct {
		Result string `json:"result"`
	}

	var handler mcp.ToolHandlerFor[TestArgs, TestResult] = func(
		ctx context.Context,
		req *mcp.CallToolRequest,
		args TestArgs,
	) (*mcp.CallToolResult, TestResult, error) {
		return nil, TestResult{}, nil
	}

	router.RegisterToolWithSchema("existing_tool", "An existing tool", handler)

	// Test existing tool
	tool, exists := router.GetTool("existing_tool")
	if !exists {
		t.Fatal("Expected tool to exist")
	}
	if tool.Name != "existing_tool" {
		t.Errorf("Expected name 'existing_tool', got '%s'", tool.Name)
	}
}

func TestSearchTools(t *testing.T) {
	router := NewRouter()

	// Register multiple tools
	type TestArgs struct {
		Input string `json:"input"`
	}
	type TestResult struct {
		Output string `json:"output"`
	}

	var handler mcp.ToolHandlerFor[TestArgs, TestResult] = func(
		ctx context.Context,
		req *mcp.CallToolRequest,
		args TestArgs,
	) (*mcp.CallToolResult, TestResult, error) {
		return nil, TestResult{}, nil
	}

	router.RegisterToolWithSchema("k8s_list_pods", "List pods in a namespace", handler)
	router.RegisterToolWithSchema("k8s_get_pod", "Get a specific pod", handler)
	router.RegisterToolWithSchema("k8s_list_services", "List services", handler)
	router.RegisterToolWithSchema("k8s_get_deployment", "Get a deployment", handler)

	// Test exact match
	results := router.SearchTools("k8s_list_pods", 10)
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0].Name != "k8s_list_pods" {
		t.Errorf("Expected 'k8s_list_pods', got '%s'", results[0].Name)
	}

	// Test partial match
	results = router.SearchTools("pod", 10)
	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	// Test description match
	results = router.SearchTools("namespace", 10)
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0].Name != "k8s_list_pods" {
		t.Errorf("Expected 'k8s_list_pods', got '%s'", results[0].Name)
	}

	// Test limit
	results = router.SearchTools("k8s", 2)
	if len(results) != 2 {
		t.Fatalf("Expected 2 results (limited), got %d", len(results))
	}

	// Test empty query (should return empty - requires a match)
	results = router.SearchTools("", 10)
	if len(results) != 0 {
		t.Fatalf("Expected 0 results for empty query (requires match), got %d", len(results))
	}

	// Test no matches
	results = router.SearchTools("nonexistent", 10)
	if len(results) != 0 {
		t.Fatalf("Expected 0 results, got %d", len(results))
	}
}

func TestSearchToolsScoring(t *testing.T) {
	router := NewRouter()

	type TestArgs struct {
		Input string `json:"input"`
	}
	type TestResult struct {
		Output string `json:"output"`
	}

	var handler mcp.ToolHandlerFor[TestArgs, TestResult] = func(
		ctx context.Context,
		req *mcp.CallToolRequest,
		args TestArgs,
	) (*mcp.CallToolResult, TestResult, error) {
		return nil, TestResult{}, nil
	}

	// Register tools with different relevance
	router.RegisterToolWithSchema("k8s_list_pods", "List pods", handler)
	router.RegisterToolWithSchema("k8s_get_pod_logs", "Get pod logs", handler)
	router.RegisterToolWithSchema("k8s_list_services", "List services", handler)

	// Exact match should be first
	results := router.SearchTools("k8s_list_pods", 10)
	if len(results) == 0 {
		t.Fatal("Expected at least one result")
	}
	if results[0].Name != "k8s_list_pods" {
		t.Errorf("Expected 'k8s_list_pods' to be first, got '%s'", results[0].Name)
	}
}

func TestGetAllTools(t *testing.T) {
	router := NewRouter()

	type TestArgs struct {
		Input string `json:"input"`
	}
	type TestResult struct {
		Output string `json:"output"`
	}

	var handler mcp.ToolHandlerFor[TestArgs, TestResult] = func(
		ctx context.Context,
		req *mcp.CallToolRequest,
		args TestArgs,
	) (*mcp.CallToolResult, TestResult, error) {
		return nil, TestResult{}, nil
	}

	// Register multiple tools
	router.RegisterToolWithSchema("tool1", "Tool 1", handler)
	router.RegisterToolWithSchema("tool2", "Tool 2", handler)
	router.RegisterToolWithSchema("tool3", "Tool 3", handler)

	allTools := router.GetAllTools()
	if len(allTools) != 3 {
		t.Fatalf("Expected 3 tools, got %d", len(allTools))
	}

	// Check that all tools are present
	toolNames := make(map[string]bool)
	for _, tool := range allTools {
		toolNames[tool.Name] = true
	}

	expectedNames := []string{"tool1", "tool2", "tool3"}
	for _, name := range expectedNames {
		if !toolNames[name] {
			t.Errorf("Expected tool '%s' not found", name)
		}
	}
}

func TestExtractInputSchema(t *testing.T) {
	type TestArgs struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Count     int    `json:"count"`
		Enabled   bool   `json:"enabled"`
	}
	type TestResult struct {
		Success bool `json:"success"`
	}

	var handler mcp.ToolHandlerFor[TestArgs, TestResult] = func(
		ctx context.Context,
		req *mcp.CallToolRequest,
		args TestArgs,
	) (*mcp.CallToolResult, TestResult, error) {
		return nil, TestResult{}, nil
	}

	schema := extractInputSchema(handler)

	// Check schema structure
	if schema["type"] != "object" {
		t.Errorf("Expected type 'object', got '%v'", schema["type"])
	}

	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected 'properties' to be a map")
	}

	// Check that all fields are present
	expectedFields := []string{"name", "namespace", "count", "enabled"}
	for _, field := range expectedFields {
		if _, exists := properties[field]; !exists {
			t.Errorf("Expected field '%s' in schema", field)
		}
	}

	// Check field types
	if nameSchema, ok := properties["name"].(map[string]interface{}); ok {
		if nameSchema["type"] != "string" {
			t.Errorf("Expected 'name' to be string, got '%v'", nameSchema["type"])
		}
	}

	if countSchema, ok := properties["count"].(map[string]interface{}); ok {
		if countSchema["type"] != "integer" {
			t.Errorf("Expected 'count' to be integer, got '%v'", countSchema["type"])
		}
	}

	if enabledSchema, ok := properties["enabled"].(map[string]interface{}); ok {
		if enabledSchema["type"] != "boolean" {
			t.Errorf("Expected 'enabled' to be boolean, got '%v'", enabledSchema["type"])
		}
	}
}

func TestTypeToJSONSchema(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"string", "", "string"},
		{"int", 0, "integer"},
		{"int64", int64(0), "integer"},
		{"float64", 0.0, "number"},
		{"bool", false, "boolean"},
		{"slice", []string{}, "array"},
		{"map", map[string]interface{}{}, "object"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := typeToJSONSchema(reflect.TypeOf(tt.input))
			if schema["type"] != tt.expected {
				t.Errorf("Expected type '%s', got '%v'", tt.expected, schema["type"])
			}
		})
	}
}

