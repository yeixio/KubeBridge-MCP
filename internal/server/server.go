package server

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yeixio/KubeBridge-MCP/internal/config"
	"github.com/yeixio/KubeBridge-MCP/internal/kubernetes"
)

// getEnvBool is a helper to get boolean env vars (duplicated from config for now)
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}

// Server represents the MCP server for Kubernetes
type Server struct {
	config          *config.Config
	clientManager   *kubernetes.ClientManager
	mcpServer       *mcp.Server
	registeredTools []string // Track registered tool names for debugging
	router          *Router  // Router for dynamic tool access (workaround for 40-tool limit)
	useRouterMode   bool     // If true, only register router tools, not individual tools
}

// NewServer creates a new instance of the MCP server
func NewServer(cfg *config.Config, clientManager *kubernetes.ClientManager) (*Server, error) {
	// Create MCP server
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "kubebridge-mcp",
		Version: "0.1.0",
	}, nil)

	s := &Server{
		config:          cfg,
		clientManager:   clientManager,
		mcpServer:       mcpServer,
		registeredTools: make([]string, 0),
		router:          NewRouter(),
		useRouterMode:   getEnvBool("KUBEBRIDGE_USE_ROUTER", false), // Enable router mode via env var
	}

	// Discover resources and generate tools
	if err := s.discoverAndRegisterTools(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to discover resources and register tools: %w", err)
	}

	// If router mode is enabled, register router tools
	if s.useRouterMode {
		s.registerRouterTools(s.router)
		fmt.Fprintf(os.Stderr, "Router mode enabled: Only 2 router tools registered with MCP server\n")
	}

	return s, nil
}

// discoverAndRegisterTools discovers Kubernetes resources and registers MCP tools
func (s *Server) discoverAndRegisterTools(ctx context.Context) error {
	// Discover resources
	if err := s.clientManager.DiscoverResources(ctx); err != nil {
		return fmt.Errorf("failed to discover resources: %w", err)
	}

	resources := s.clientManager.GetDiscoveredResources()
	fmt.Fprintf(os.Stderr, "Discovered %d Kubernetes resources\n", len(resources))

	// If router mode is enabled, register all tools in router but only expose router tools to MCP
	if s.useRouterMode {
		// Register all tools in router (they'll be accessible via k8s_call_tool)
		s.registerAllToolsInRouter(resources)
		// Router tools will be registered at the end
		return nil
	}

	// Normal mode: Register special tools first (logs, events) - these are critical and should be in first 40
	s.registerSpecialTools()
	fmt.Fprintf(os.Stderr, "Registered special tools (logs, events): %d tools\n", len(s.registeredTools))

	// Core resources that should always have tools (prioritized for Cursor's 40-tool limit)
	coreResources := map[string]bool{
		"pods": true, "pod": true,
		"namespaces": true, "namespace": true,
		"deployments": true, "deployment": true,
		"services": true, "service": true,
		"configmaps": true, "configmap": true,
		"secrets": true, "secret": true,
		"nodes": true, "node": true,
		"events": true, "event": true,
		"replicasets": true, "replicaset": true,
		"daemonsets": true, "daemonset": true,
		"statefulsets": true, "statefulset": true,
		"jobs": true, "job": true,
		"cronjobs": true, "cronjob": true,
		"persistentvolumes": true, "persistentvolume": true,
		"persistentvolumeclaims": true, "persistentvolumeclaim": true,
		"ingresses": true, "ingress": true,
	}

	// Separate resources into core and others
	var coreResourceList []*kubernetes.APIResource
	var otherResourceList []*kubernetes.APIResource

	for _, resource := range resources {
		// Check if resource name or singular name is in core list
		isCore := coreResources[resource.Resource] || coreResources[resource.SingularName] ||
			coreResources[resource.Kind] || coreResources[strings.ToLower(resource.Kind)]

		if isCore {
			coreResourceList = append(coreResourceList, resource)
		} else {
			otherResourceList = append(otherResourceList, resource)
		}
	}

	fmt.Fprintf(os.Stderr, "Prioritizing %d core resources (out of %d total)\n", len(coreResourceList), len(resources))

	// Register tools for core resources first (prioritize list/get over describe)
	toolCount := len(s.registeredTools)
	for _, resource := range coreResourceList {
		count := s.registerResourceToolsPrioritized(resource)
		toolCount += count
	}

	// Only register other resources if we haven't hit the limit
	// Cursor has a 40-tool limit, so we prioritize core resources
	if toolCount < 35 { // Leave some room
		for _, resource := range otherResourceList {
			if toolCount >= 35 {
				break
			}
			count := s.registerResourceToolsPrioritized(resource)
			toolCount += count
		}
	}

	fmt.Fprintf(os.Stderr, "Registered %d MCP tools total (Cursor limit: 40)\n", toolCount)

	// Log all registered tools in debug mode
	if s.config.LogLevel == "debug" {
		fmt.Fprintf(os.Stderr, "Registered tools:\n")
		for _, toolName := range s.registeredTools {
			fmt.Fprintf(os.Stderr, "  - %s\n", toolName)
		}
	}

	return nil
}

// registerAllToolsInRouter registers all tools in the router (for router mode)
func (s *Server) registerAllToolsInRouter(resources []*kubernetes.APIResource) {
	// Register special tools in router
	s.registerPodLogsToolInRouter()
	s.registerEventsToolInRouter()

	// Register all resource tools in router
	for _, resource := range resources {
		s.registerResourceToolsInRouter(resource)
	}

	fmt.Fprintf(os.Stderr, "Registered %d tools in router (accessible via k8s_call_tool)\n", len(s.router.tools))
}

// registerResourceToolsInRouter registers tools in router only (not with MCP server)
func (s *Server) registerResourceToolsInRouter(resource *kubernetes.APIResource) {
	toolResourceName := resource.Resource
	if resource.SingularName != "" {
		toolResourceName = resource.SingularName
	}

	// Register list tool
	if hasVerb(resource.Verbs, "list") {
		s.registerListToolInRouter(resource, toolResourceName)
	}

	// Register get tool
	if hasVerb(resource.Verbs, "get") {
		s.registerGetToolInRouter(resource, toolResourceName)
		s.registerDescribeToolInRouter(resource, toolResourceName)
	}

	// Register write tools if enabled
	if s.config.EnableWriteOperations {
		if hasVerb(resource.Verbs, "create") {
			s.registerCreateToolInRouter(resource, toolResourceName)
			s.registerApplyToolInRouter(resource, toolResourceName)
		}
		if hasVerb(resource.Verbs, "update") {
			s.registerUpdateToolInRouter(resource, toolResourceName)
		}
		if hasVerb(resource.Verbs, "patch") {
			s.registerPatchToolInRouter(resource, toolResourceName)
		}
		if hasVerb(resource.Verbs, "delete") {
			s.registerDeleteToolInRouter(resource, toolResourceName)
		}
	}
}

// registerResourceToolsPrioritized registers MCP tools for a Kubernetes resource
// Prioritizes list/get tools over describe to fit within Cursor's 40-tool limit
// Returns the number of tools registered
func (s *Server) registerResourceToolsPrioritized(resource *kubernetes.APIResource) int {
	// Generate resource name for tool (e.g., "deployments" -> "deployment")
	toolResourceName := resource.Resource
	if resource.SingularName != "" {
		toolResourceName = resource.SingularName
	}

	count := 0

	// Register list and get tools first (most useful, come before describe alphabetically)
	if hasVerb(resource.Verbs, "list") {
		s.registerListTool(resource, toolResourceName)
		count++
	}

	if hasVerb(resource.Verbs, "get") {
		s.registerGetTool(resource, toolResourceName)
		count++
		// Skip describe for now to prioritize list/get tools
		// Describe can be added later if we have room
		// s.registerDescribeTool(resource, toolResourceName)
		// count++
	}

	// Register write tools (only if write operations are enabled)
	if s.config.EnableWriteOperations {
		if hasVerb(resource.Verbs, "create") {
			s.registerCreateTool(resource, toolResourceName)
			count++
			s.registerApplyTool(resource, toolResourceName)
			count++
		}

		if hasVerb(resource.Verbs, "update") {
			s.registerUpdateTool(resource, toolResourceName)
			count++
		}

		if hasVerb(resource.Verbs, "patch") {
			s.registerPatchTool(resource, toolResourceName)
			count++
		}

		if hasVerb(resource.Verbs, "delete") {
			s.registerDeleteTool(resource, toolResourceName)
			count++
		}
	}

	return count
}

// registerResourceTools registers MCP tools for a Kubernetes resource (legacy, full registration)
// Returns the number of tools registered
// This function is currently unused but kept for reference
//
//nolint:unused
func (s *Server) registerResourceTools(resource *kubernetes.APIResource) int {
	// Generate resource name for tool (e.g., "deployments" -> "deployment")
	toolResourceName := resource.Resource
	if resource.SingularName != "" {
		toolResourceName = resource.SingularName
	}

	count := 0

	// Register read-only tools (always available)
	if hasVerb(resource.Verbs, "list") {
		s.registerListTool(resource, toolResourceName)
		count++
	}

	if hasVerb(resource.Verbs, "get") {
		s.registerGetTool(resource, toolResourceName)
		count++
		s.registerDescribeTool(resource, toolResourceName)
		count++
	}

	// Register write tools (only if write operations are enabled)
	if s.config.EnableWriteOperations {
		if hasVerb(resource.Verbs, "create") {
			s.registerCreateTool(resource, toolResourceName)
			count++
			s.registerApplyTool(resource, toolResourceName)
			count++
		}

		if hasVerb(resource.Verbs, "update") {
			s.registerUpdateTool(resource, toolResourceName)
			count++
		}

		if hasVerb(resource.Verbs, "patch") {
			s.registerPatchTool(resource, toolResourceName)
			count++
		}

		if hasVerb(resource.Verbs, "delete") {
			s.registerDeleteTool(resource, toolResourceName)
			count++
		}
	}

	return count
}

// registerSpecialTools registers special tools (logs, events)
func (s *Server) registerSpecialTools() {
	// Pod logs tool - uses typed client for pod logs
	s.registerPodLogsTool()

	// Events tool - uses typed client for events
	s.registerEventsTool()
}

// hasVerb checks if a verb is in the verbs list
func hasVerb(verbs []string, verb string) bool {
	for _, v := range verbs {
		if v == verb {
			return true
		}
	}
	return false
}

// addToolAndTrack registers a tool with the MCP server and tracks it for debugging
func addToolAndTrack[In, Out any](s *Server, tool *mcp.Tool, handler mcp.ToolHandlerFor[In, Out]) {
	// Always register with MCP server (needed for router mode)
	mcp.AddTool(s.mcpServer, tool, handler)
	s.registeredTools = append(s.registeredTools, tool.Name)

	// Also register in router for dynamic access (if router mode is enabled or for future use)
	if s.router != nil {
		// Extract input schema from the handler type
		s.router.RegisterToolWithSchema(tool.Name, tool.Description, handler)
	}
}

// Start initializes and starts the MCP server
func (s *Server) Start(ctx context.Context) error {
	// Use stdio transport with logging
	transport := &mcp.LoggingTransport{
		Transport: &mcp.StdioTransport{},
		Writer:    os.Stderr,
	}

	// Run the server
	return s.mcpServer.Run(ctx, transport)
}
