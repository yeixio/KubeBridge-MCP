package server

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ToolMetadata stores metadata about a tool for searching
type ToolMetadata struct {
	Name        string
	Description string
	InputSchema map[string]interface{}
	Handler     interface{} // Store the handler for later execution
}

// Router manages tool routing for the 40-tool limit workaround
type Router struct {
	tools map[string]*ToolMetadata // All available tools
}

// NewRouter creates a new tool router
func NewRouter() *Router {
	return &Router{
		tools: make(map[string]*ToolMetadata),
	}
}

// RegisterTool registers a tool in the router
func (r *Router) RegisterTool(name, description string, inputSchema map[string]interface{}, handler interface{}) {
	r.tools[name] = &ToolMetadata{
		Name:        name,
		Description: description,
		InputSchema: inputSchema,
		Handler:     handler,
	}
}

// RegisterToolWithSchema registers a tool with automatic schema extraction
func (r *Router) RegisterToolWithSchema(name, description string, handler interface{}) {
	// Extract input schema from handler type using reflection
	inputSchema := extractInputSchema(handler)
	r.RegisterTool(name, description, inputSchema, handler)
}

// extractInputSchema extracts the input schema from a handler type
func extractInputSchema(handler interface{}) map[string]interface{} {
	handlerType := reflect.TypeOf(handler)
	if handlerType.Kind() != reflect.Func || handlerType.NumIn() < 3 {
		return map[string]interface{}{"type": "object"}
	}

	// Get the input type (3rd parameter)
	inputType := handlerType.In(2)
	return structToJSONSchema(inputType)
}

// structToJSONSchema converts a Go struct type to a JSON schema
func structToJSONSchema(structType reflect.Type) map[string]interface{} {
	schema := map[string]interface{}{
		"type":       "object",
		"properties": make(map[string]interface{}),
	}

	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}

	if structType.Kind() != reflect.Struct {
		return schema
	}

	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		return schema
	}

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		jsonName := strings.Split(jsonTag, ",")[0]
		if jsonName == "" {
			jsonName = strings.ToLower(field.Name)
		}

		fieldSchema := typeToJSONSchema(field.Type)
		properties[jsonName] = fieldSchema
	}

	return schema
}

// typeToJSONSchema converts a Go type to a JSON schema type
func typeToJSONSchema(goType reflect.Type) map[string]interface{} {
	schema := make(map[string]interface{})

	switch goType.Kind() {
	case reflect.String:
		schema["type"] = "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		schema["type"] = "integer"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		schema["type"] = "integer"
	case reflect.Float32, reflect.Float64:
		schema["type"] = "number"
	case reflect.Bool:
		schema["type"] = "boolean"
	case reflect.Slice, reflect.Array:
		schema["type"] = "array"
		schema["items"] = typeToJSONSchema(goType.Elem())
	case reflect.Map:
		schema["type"] = "object"
		schema["additionalProperties"] = typeToJSONSchema(goType.Elem())
	case reflect.Ptr:
		return typeToJSONSchema(goType.Elem())
	case reflect.Struct:
		return structToJSONSchema(goType)
	default:
		schema["type"] = "object"
	}

	return schema
}

// SearchTools searches for tools matching the query
func (r *Router) SearchTools(query string, limit int) []*ToolMetadata {
	if limit <= 0 {
		limit = 10
	}

	query = strings.ToLower(query)
	queryTerms := strings.Fields(query)

	var matches []*ToolMetadata
	scores := make(map[string]int)

	// Score each tool based on query match
	for _, tool := range r.tools {
		score := 0
		nameLower := strings.ToLower(tool.Name)
		descLower := strings.ToLower(tool.Description)

		for _, term := range queryTerms {
			// Exact name match gets highest score
			if nameLower == term {
				score += 100
			} else if strings.Contains(nameLower, term) {
				score += 50
			}

			// Description match
			if strings.Contains(descLower, term) {
				score += 10
			}
		}

		if score > 0 {
			scores[tool.Name] = score
			matches = append(matches, tool)
		}
	}

	// Sort by score (simple bubble sort for small lists)
	for i := 0; i < len(matches)-1; i++ {
		for j := i + 1; j < len(matches); j++ {
			if scores[matches[i].Name] < scores[matches[j].Name] {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	// Limit results
	if len(matches) > limit {
		matches = matches[:limit]
	}

	return matches
}

// GetTool returns a tool by name
func (r *Router) GetTool(name string) (*ToolMetadata, bool) {
	tool, exists := r.tools[name]
	return tool, exists
}

// GetAllTools returns all registered tools
func (r *Router) GetAllTools() []*ToolMetadata {
	tools := make([]*ToolMetadata, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// registerRouterTools registers the router tools (retrieve_tools and call_tool)
func (s *Server) registerRouterTools(router *Router) {
	// Register retrieve_tools
	type RetrieveToolsArgs struct {
		Query string `json:"query"`
		Limit int    `json:"limit"`
	}
	type RetrieveToolsResult struct {
		Tools []map[string]interface{} `json:"tools"`
	}

	var retrieveHandler mcp.ToolHandlerFor[RetrieveToolsArgs, RetrieveToolsResult] = func(
		ctx context.Context,
		req *mcp.CallToolRequest,
		args RetrieveToolsArgs,
	) (*mcp.CallToolResult, RetrieveToolsResult, error) {
		limit := args.Limit
		if limit <= 0 {
			limit = 10
		}

		matches := router.SearchTools(args.Query, limit)

		tools := make([]map[string]interface{}, len(matches))
		for i, tool := range matches {
			tools[i] = map[string]interface{}{
				"name":        tool.Name,
				"description": tool.Description,
			}
		}

		text := fmt.Sprintf("Found %d matching tools", len(tools))
		if args.Query != "" {
			text = fmt.Sprintf("Found %d tools matching '%s'", len(tools), args.Query)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: text},
			},
		}, RetrieveToolsResult{Tools: tools}, nil
	}

	addToolAndTrack(s, &mcp.Tool{
		Name:        "k8s_retrieve_tools",
		Description: "Search and retrieve available Kubernetes tools. Use this to discover tools before calling them.",
	}, retrieveHandler)

	// Register call_tool
	type CallToolArgs struct {
		ToolName string                 `json:"toolName"`
		Args     map[string]interface{} `json:"args"`
	}
	type CallToolResult struct {
		Result map[string]interface{} `json:"result"`
	}

	var callHandler mcp.ToolHandlerFor[CallToolArgs, CallToolResult] = func(
		ctx context.Context,
		req *mcp.CallToolRequest,
		args CallToolArgs,
	) (*mcp.CallToolResult, CallToolResult, error) {
		if args.ToolName == "" {
			return nil, CallToolResult{}, fmt.Errorf("toolName is required")
		}

		tool, exists := router.GetTool(args.ToolName)
		if !exists {
			return nil, CallToolResult{}, fmt.Errorf("tool '%s' not found. Use k8s_retrieve_tools to search for available tools", args.ToolName)
		}

		// No need to marshal - invokeToolHandler will handle it

		// Invoke the stored handler using reflection
		// The handler is stored as interface{}, so we need to use reflection to call it
		invocationResult, callErr := invokeToolHandler(ctx, tool.Handler, args.Args, args.ToolName)
		if callErr != nil {
			return nil, CallToolResult{}, fmt.Errorf("failed to invoke tool handler: %w", callErr)
		}

		// Extract CallToolResult and output
		var callToolResult *mcp.CallToolResult
		var outputMap map[string]interface{}

		if invocationResult != nil {
			callToolResult = invocationResult.CallToolResult

			// Convert output to map
			if invocationResult.Output != nil {
				outputJSON, err := json.Marshal(invocationResult.Output)
				if err == nil {
					if err := json.Unmarshal(outputJSON, &outputMap); err != nil {
						outputMap = make(map[string]interface{})
					}
				}
			}
		}

		// Use the CallToolResult from the handler, or create a default one
		if callToolResult == nil {
			callToolResult = &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Tool '%s' executed successfully", args.ToolName)},
				},
			}
		}

		return callToolResult, CallToolResult{Result: outputMap}, nil
	}

	addToolAndTrack(s, &mcp.Tool{
		Name:        "k8s_call_tool",
		Description: "Call any Kubernetes tool by name. First use k8s_retrieve_tools to find available tools.",
	}, callHandler)
}
