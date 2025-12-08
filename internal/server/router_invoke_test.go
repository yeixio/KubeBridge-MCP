package server

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestInvokeToolHandler(t *testing.T) {
	ctx := context.Background()

	// Create a simple handler
	type TestArgs struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Count     int    `json:"count"`
	}
	type TestResult struct {
		Message string `json:"message"`
		Count   int    `json:"count"`
	}

	var handler mcp.ToolHandlerFor[TestArgs, TestResult] = func(
		ctx context.Context,
		req *mcp.CallToolRequest,
		args TestArgs,
	) (*mcp.CallToolResult, TestResult, error) {
		return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Test executed"},
				},
			}, TestResult{
				Message: "success",
				Count:   args.Count,
			}, nil
	}

	// Test successful invocation
	args := map[string]interface{}{
		"name":      "test-pod",
		"namespace": "default",
		"count":     42,
	}

	result, err := invokeToolHandler(ctx, handler, args, "test_tool")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.CallToolResult == nil {
		t.Fatal("Expected CallToolResult, got nil")
	}

	if len(result.CallToolResult.Content) == 0 {
		t.Fatal("Expected content in CallToolResult")
	}

	output, ok := result.Output.(TestResult)
	if !ok {
		t.Fatalf("Expected TestResult output, got %T", result.Output)
	}

	if output.Message != "success" {
		t.Errorf("Expected message 'success', got '%s'", output.Message)
	}

	if output.Count != 42 {
		t.Errorf("Expected count 42, got %d", output.Count)
	}
}

func TestInvokeToolHandlerWithError(t *testing.T) {
	ctx := context.Background()

	type TestArgs struct {
		ShouldError bool `json:"shouldError"`
	}
	type TestResult struct {
		Message string `json:"message"`
	}

	var handler mcp.ToolHandlerFor[TestArgs, TestResult] = func(
		ctx context.Context,
		req *mcp.CallToolRequest,
		args TestArgs,
	) (*mcp.CallToolResult, TestResult, error) {
		if args.ShouldError {
			return nil, TestResult{}, fmt.Errorf("test error")
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Success"},
			},
		}, TestResult{Message: "ok"}, nil
	}

	// Test error case
	args := map[string]interface{}{
		"shouldError": true,
	}

	result, err := invokeToolHandler(ctx, handler, args, "test_tool")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if result != nil {
		t.Error("Expected nil result on error")
	}
}

func TestInvokeToolHandlerNilHandler(t *testing.T) {
	ctx := context.Background()

	_, err := invokeToolHandler(ctx, nil, map[string]interface{}{}, "test")
	if err == nil {
		t.Fatal("Expected error for nil handler")
	}
}

func TestFillStructFromMap(t *testing.T) {
	type TestStruct struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Count     int    `json:"count"`
		Enabled   bool   `json:"enabled"`
		Omit      string `json:"-"`
		NoTag     string
	}

	structValue := reflect.New(reflect.TypeOf(TestStruct{})).Elem()

	m := map[string]interface{}{
		"name":      "test",
		"namespace": "default",
		"count":     42,
		"enabled":   true,
		"omit":      "should be ignored",
		"notag":     "should be ignored",
	}

	err := fillStructFromMap(structValue, m)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check values
	if structValue.FieldByName("Name").String() != "test" {
		t.Errorf("Expected Name 'test', got '%s'", structValue.FieldByName("Name").String())
	}

	if structValue.FieldByName("Namespace").String() != "default" {
		t.Errorf("Expected Namespace 'default', got '%s'", structValue.FieldByName("Namespace").String())
	}

	if structValue.FieldByName("Count").Int() != 42 {
		t.Errorf("Expected Count 42, got %d", structValue.FieldByName("Count").Int())
	}

	if !structValue.FieldByName("Enabled").Bool() {
		t.Error("Expected Enabled true, got false")
	}
}

func TestFillStructFromMapWithPointer(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name"`
	}

	structPtr := reflect.New(reflect.TypeOf(TestStruct{}))

	m := map[string]interface{}{
		"name": "test",
	}

	err := fillStructFromMap(structPtr, m)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	structValue := structPtr.Elem()
	if structValue.FieldByName("Name").String() != "test" {
		t.Errorf("Expected Name 'test', got '%s'", structValue.FieldByName("Name").String())
	}
}
