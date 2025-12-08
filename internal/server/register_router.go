package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yeixio/KubeBridge-MCP/internal/kubernetes"
	"github.com/yeixio/KubeBridge-MCP/internal/kubernetes/operations"
	"github.com/yeixio/KubeBridge-MCP/internal/kubernetes/special"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

// registerListToolInRouter registers a list tool in router only
func (s *Server) registerListToolInRouter(resource *kubernetes.APIResource, toolResourceName string) {
	type ListArgs struct {
		Namespace string `json:"namespace"`
	}
	type ListResult struct {
		Items []map[string]interface{} `json:"items"`
	}

	var handler mcp.ToolHandlerFor[ListArgs, ListResult] = func(ctx context.Context, req *mcp.CallToolRequest, args ListArgs) (*mcp.CallToolResult, ListResult, error) {
		gvr := schema.GroupVersionResource{
			Group:    resource.Group,
			Version:  resource.Version,
			Resource: resource.Resource,
		}

		listOp := operations.NewListOperation(s.clientManager.GetCluster().Dynamic())
		items, err := listOp.List(ctx, gvr, args.Namespace)
		if err != nil {
			return nil, ListResult{}, kubernetes.SanitizeError(err)
		}

		result := make([]map[string]interface{}, len(items))
		names := make([]string, 0, len(items))
		for i, item := range items {
			result[i] = item.Object
			// Extract name from metadata
			if metadata, ok := item.Object["metadata"].(map[string]interface{}); ok {
				if name, ok := metadata["name"].(string); ok {
					names = append(names, name)
				}
			}
		}

		// Build text summary with names
		text := fmt.Sprintf("Found %d %s", len(result), toolResourceName)
		if len(names) > 0 {
			if len(names) <= 10 {
				text += fmt.Sprintf(": %s", strings.Join(names, ", "))
			} else {
				text += fmt.Sprintf(": %s ... and %d more", strings.Join(names[:10], ", "), len(names)-10)
			}
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: text},
			},
		}, ListResult{Items: result}, nil
	}

	toolName := fmt.Sprintf("k8s_list_%s", toolResourceName)
	s.router.RegisterToolWithSchema(toolName, fmt.Sprintf("List %s resources in a namespace", toolResourceName), handler)
}

// registerGetToolInRouter registers a get tool in router only
func (s *Server) registerGetToolInRouter(resource *kubernetes.APIResource, toolResourceName string) {
	type GetArgs struct {
		Namespace string `json:"namespace"`
		Name      string `json:"name"`
	}
	type GetResult struct {
		Resource map[string]interface{} `json:"resource"`
	}

	var handler mcp.ToolHandlerFor[GetArgs, GetResult] = func(ctx context.Context, req *mcp.CallToolRequest, args GetArgs) (*mcp.CallToolResult, GetResult, error) {
		if args.Name == "" {
			return nil, GetResult{}, fmt.Errorf("name is required")
		}

		gvr := schema.GroupVersionResource{
			Group:    resource.Group,
			Version:  resource.Version,
			Resource: resource.Resource,
		}

		getOp := operations.NewGetOperation(s.clientManager.GetCluster().Dynamic())
		obj, err := getOp.Get(ctx, gvr, args.Namespace, args.Name)
		if err != nil {
			return nil, GetResult{}, kubernetes.SanitizeError(err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Retrieved %s %s", toolResourceName, args.Name)},
			},
		}, GetResult{Resource: obj.Object}, nil
	}

	toolName := fmt.Sprintf("k8s_get_%s", toolResourceName)
	s.router.RegisterToolWithSchema(toolName, fmt.Sprintf("Get a %s resource by name", toolResourceName), handler)
}

// registerDescribeToolInRouter registers a describe tool in router only
func (s *Server) registerDescribeToolInRouter(resource *kubernetes.APIResource, toolResourceName string) {
	type DescribeArgs struct {
		Namespace string `json:"namespace"`
		Name      string `json:"name"`
	}
	type DescribeResult struct {
		Description string `json:"description"`
	}

	var handler mcp.ToolHandlerFor[DescribeArgs, DescribeResult] = func(ctx context.Context, req *mcp.CallToolRequest, args DescribeArgs) (*mcp.CallToolResult, DescribeResult, error) {
		if args.Name == "" {
			return nil, DescribeResult{}, fmt.Errorf("name is required")
		}

		gvr := schema.GroupVersionResource{
			Group:    resource.Group,
			Version:  resource.Version,
			Resource: resource.Resource,
		}

		describeOp := operations.NewDescribeOperation(s.clientManager.GetCluster().Dynamic())
		description, err := describeOp.Describe(ctx, gvr, args.Namespace, args.Name)
		if err != nil {
			return nil, DescribeResult{}, kubernetes.SanitizeError(err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: description},
			},
		}, DescribeResult{Description: description}, nil
	}

	toolName := fmt.Sprintf("k8s_describe_%s", toolResourceName)
	s.router.RegisterToolWithSchema(toolName, fmt.Sprintf("Get a human-readable description of a %s resource", toolResourceName), handler)
}

// registerCreateToolInRouter registers a create tool in router only
func (s *Server) registerCreateToolInRouter(resource *kubernetes.APIResource, toolResourceName string) {
	type CreateArgs struct {
		Namespace string                 `json:"namespace"`
		Manifest  map[string]interface{} `json:"manifest"`
	}
	type CreateResult struct {
		Resource map[string]interface{} `json:"resource"`
	}

	var handler mcp.ToolHandlerFor[CreateArgs, CreateResult] = func(ctx context.Context, req *mcp.CallToolRequest, args CreateArgs) (*mcp.CallToolResult, CreateResult, error) {
		if args.Manifest == nil {
			return nil, CreateResult{}, fmt.Errorf("manifest is required")
		}

		obj := &unstructured.Unstructured{Object: args.Manifest}
		gvr := schema.GroupVersionResource{
			Group:    resource.Group,
			Version:  resource.Version,
			Resource: resource.Resource,
		}

		createOp := operations.NewCreateOperation(s.clientManager.GetCluster().Dynamic())
		created, err := createOp.Create(ctx, gvr, args.Namespace, obj)
		if err != nil {
			return nil, CreateResult{}, kubernetes.SanitizeError(err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Created %s %s", toolResourceName, created.GetName())},
			},
		}, CreateResult{Resource: created.Object}, nil
	}

	toolName := fmt.Sprintf("k8s_create_%s", toolResourceName)
	s.router.RegisterToolWithSchema(toolName, fmt.Sprintf("Create a new %s resource", toolResourceName), handler)
}

// registerUpdateToolInRouter registers an update tool in router only
func (s *Server) registerUpdateToolInRouter(resource *kubernetes.APIResource, toolResourceName string) {
	type UpdateArgs struct {
		Namespace string                 `json:"namespace"`
		Manifest  map[string]interface{} `json:"manifest"`
	}
	type UpdateResult struct {
		Resource map[string]interface{} `json:"resource"`
	}

	var handler mcp.ToolHandlerFor[UpdateArgs, UpdateResult] = func(ctx context.Context, req *mcp.CallToolRequest, args UpdateArgs) (*mcp.CallToolResult, UpdateResult, error) {
		if args.Manifest == nil {
			return nil, UpdateResult{}, fmt.Errorf("manifest is required")
		}

		obj := &unstructured.Unstructured{Object: args.Manifest}
		gvr := schema.GroupVersionResource{
			Group:    resource.Group,
			Version:  resource.Version,
			Resource: resource.Resource,
		}

		updateOp := operations.NewUpdateOperation(s.clientManager.GetCluster().Dynamic())
		updated, err := updateOp.Update(ctx, gvr, args.Namespace, obj)
		if err != nil {
			return nil, UpdateResult{}, kubernetes.SanitizeError(err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Updated %s %s", toolResourceName, updated.GetName())},
			},
		}, UpdateResult{Resource: updated.Object}, nil
	}

	toolName := fmt.Sprintf("k8s_update_%s", toolResourceName)
	s.router.RegisterToolWithSchema(toolName, fmt.Sprintf("Update an existing %s resource", toolResourceName), handler)
}

// registerPatchToolInRouter registers a patch tool in router only
func (s *Server) registerPatchToolInRouter(resource *kubernetes.APIResource, toolResourceName string) {
	type PatchArgs struct {
		Namespace string          `json:"namespace"`
		Name      string          `json:"name"`
		Patch     json.RawMessage `json:"patch"`
		PatchType string          `json:"patchType"`
	}
	type PatchResult struct {
		Resource map[string]interface{} `json:"resource"`
	}

	var handler mcp.ToolHandlerFor[PatchArgs, PatchResult] = func(ctx context.Context, req *mcp.CallToolRequest, args PatchArgs) (*mcp.CallToolResult, PatchResult, error) {
		if args.Name == "" {
			return nil, PatchResult{}, fmt.Errorf("name is required")
		}
		if args.Patch == nil {
			return nil, PatchResult{}, fmt.Errorf("patch is required")
		}

		patchType := types.PatchType(args.PatchType)
		if patchType == "" {
			patchType = types.MergePatchType
		}

		gvr := schema.GroupVersionResource{
			Group:    resource.Group,
			Version:  resource.Version,
			Resource: resource.Resource,
		}

		patchOp := operations.NewPatchOperation(s.clientManager.GetCluster().Dynamic())
		patched, err := patchOp.Patch(ctx, gvr, args.Namespace, args.Name, patchType, args.Patch)
		if err != nil {
			return nil, PatchResult{}, kubernetes.SanitizeError(err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Patched %s %s", toolResourceName, args.Name)},
			},
		}, PatchResult{Resource: patched.Object}, nil
	}

	toolName := fmt.Sprintf("k8s_patch_%s", toolResourceName)
	s.router.RegisterToolWithSchema(toolName, fmt.Sprintf("Patch a %s resource using JSON patch or strategic merge", toolResourceName), handler)
}

// registerDeleteToolInRouter registers a delete tool in router only
func (s *Server) registerDeleteToolInRouter(resource *kubernetes.APIResource, toolResourceName string) {
	type DeleteArgs struct {
		Namespace string `json:"namespace"`
		Name      string `json:"name"`
	}
	type DeleteResult struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	var handler mcp.ToolHandlerFor[DeleteArgs, DeleteResult] = func(ctx context.Context, req *mcp.CallToolRequest, args DeleteArgs) (*mcp.CallToolResult, DeleteResult, error) {
		if args.Name == "" {
			return nil, DeleteResult{}, fmt.Errorf("name is required")
		}

		gvr := schema.GroupVersionResource{
			Group:    resource.Group,
			Version:  resource.Version,
			Resource: resource.Resource,
		}

		deleteOp := operations.NewDeleteOperation(s.clientManager.GetCluster().Dynamic())
		if err := deleteOp.Delete(ctx, gvr, args.Namespace, args.Name); err != nil {
			return nil, DeleteResult{}, kubernetes.SanitizeError(err)
		}

		msg := fmt.Sprintf("%s %s deleted successfully", toolResourceName, args.Name)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: msg},
			},
		}, DeleteResult{Success: true, Message: msg}, nil
	}

	toolName := fmt.Sprintf("k8s_delete_%s", toolResourceName)
	s.router.RegisterToolWithSchema(toolName, fmt.Sprintf("Delete a %s resource", toolResourceName), handler)
}

// registerApplyToolInRouter registers an apply tool in router only
func (s *Server) registerApplyToolInRouter(resource *kubernetes.APIResource, toolResourceName string) {
	type ApplyArgs struct {
		Namespace    string                 `json:"namespace"`
		Manifest     map[string]interface{} `json:"manifest"`
		FieldManager string                 `json:"fieldManager"`
	}
	type ApplyResult struct {
		Resource map[string]interface{} `json:"resource"`
	}

	var handler mcp.ToolHandlerFor[ApplyArgs, ApplyResult] = func(ctx context.Context, req *mcp.CallToolRequest, args ApplyArgs) (*mcp.CallToolResult, ApplyResult, error) {
		if args.Manifest == nil {
			return nil, ApplyResult{}, fmt.Errorf("manifest is required")
		}

		obj := &unstructured.Unstructured{Object: args.Manifest}
		fieldManager := args.FieldManager
		if fieldManager == "" {
			fieldManager = "kubebridge-mcp"
		}

		gvr := schema.GroupVersionResource{
			Group:    resource.Group,
			Version:  resource.Version,
			Resource: resource.Resource,
		}

		applyOp := operations.NewApplyOperation(s.clientManager.GetCluster().Dynamic())
		applied, err := applyOp.Apply(ctx, gvr, args.Namespace, obj, fieldManager)
		if err != nil {
			return nil, ApplyResult{}, kubernetes.SanitizeError(err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Applied %s %s", toolResourceName, applied.GetName())},
			},
		}, ApplyResult{Resource: applied.Object}, nil
	}

	toolName := fmt.Sprintf("k8s_apply_%s", toolResourceName)
	s.router.RegisterToolWithSchema(toolName, fmt.Sprintf("Apply a %s resource using server-side apply", toolResourceName), handler)
}

// registerPodLogsToolInRouter registers pod logs tool in router only
func (s *Server) registerPodLogsToolInRouter() {
	type PodLogsArgs struct {
		Namespace string `json:"namespace"`
		Name      string `json:"name"`
		Container string `json:"container"`
		TailLines int64  `json:"tailLines"`
	}
	type PodLogsResult struct {
		Logs []string `json:"logs"`
	}

	var handler mcp.ToolHandlerFor[PodLogsArgs, PodLogsResult] = func(ctx context.Context, req *mcp.CallToolRequest, args PodLogsArgs) (*mcp.CallToolResult, PodLogsResult, error) {
		if args.Namespace == "" || args.Name == "" {
			return nil, PodLogsResult{}, fmt.Errorf("namespace and name are required")
		}

		logsOp := special.NewLogsOperation(s.clientManager.GetCluster().Typed())
		logs, err := logsOp.GetLogs(ctx, args.Namespace, args.Name, args.Container, args.TailLines)
		if err != nil {
			return nil, PodLogsResult{}, kubernetes.SanitizeError(err)
		}

		// Analyze logs for errors
		errorLines := make([]string, 0)
		errorKeywords := []string{"error", "ERROR", "failed", "FAILED", "fatal", "FATAL", "panic", "PANIC", "exception", "EXCEPTION"}
		for _, logLine := range logs {
			lowerLog := strings.ToLower(logLine)
			for _, keyword := range errorKeywords {
				if strings.Contains(lowerLog, strings.ToLower(keyword)) {
					errorLines = append(errorLines, logLine)
					break // Only count each line once
				}
			}
		}

		// Build summary text
		text := fmt.Sprintf("Retrieved %d log lines from pod %s", len(logs), args.Name)
		if len(errorLines) > 0 {
			text += fmt.Sprintf("\n\nFound %d error/warning lines:", len(errorLines))
			// Show up to 5 error examples
			maxErrors := 5
			if len(errorLines) < maxErrors {
				maxErrors = len(errorLines)
			}
			for i := 0; i < maxErrors; i++ {
				// Truncate long lines
				errorText := errorLines[i]
				if len(errorText) > 200 {
					errorText = errorText[:200] + "..."
				}
				text += fmt.Sprintf("\n  - %s", errorText)
			}
			if len(errorLines) > maxErrors {
				text += fmt.Sprintf("\n  ... and %d more error lines", len(errorLines)-maxErrors)
			}
		} else {
			text += "\n\nNo errors or warnings found in logs."
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: text},
			},
		}, PodLogsResult{Logs: logs}, nil
	}

	s.router.RegisterToolWithSchema("k8s_get_pod_logs", "Get logs from a pod container", handler)
}

// registerEventsToolInRouter registers events tool in router only
func (s *Server) registerEventsToolInRouter() {
	type EventsArgs struct {
		Namespace string `json:"namespace"`
		Kind      string `json:"kind"`
		Name      string `json:"name"`
	}
	type EventsResult struct {
		Events []special.Event `json:"events"`
	}

	var handler mcp.ToolHandlerFor[EventsArgs, EventsResult] = func(ctx context.Context, req *mcp.CallToolRequest, args EventsArgs) (*mcp.CallToolResult, EventsResult, error) {
		if args.Namespace == "" {
			return nil, EventsResult{}, fmt.Errorf("namespace is required")
		}

		eventsOp := special.NewEventsOperation(s.clientManager.GetCluster().Typed())
		events, err := eventsOp.GetEvents(ctx, args.Namespace, args.Kind, args.Name)
		if err != nil {
			return nil, EventsResult{}, kubernetes.SanitizeError(err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Retrieved %d events", len(events))},
			},
		}, EventsResult{Events: events}, nil
	}

	s.router.RegisterToolWithSchema("k8s_get_events", "Get events for Kubernetes resources", handler)
}
