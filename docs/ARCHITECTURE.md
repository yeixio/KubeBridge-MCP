# KubeBridge-MCP Architecture

## Overview

KubeBridge-MCP uses a **dynamic resource discovery and tool generation** approach to automatically expose all Kubernetes resources as MCP tools. This eliminates the need to hardcode support for specific resource types.

## Key Architectural Decisions

### 1. Dynamic Resource Discovery

Instead of hardcoding tools for specific resources (pods, deployments, etc.), we:

1. **Discover resources at startup** using Kubernetes Discovery API
2. **Generate MCP tools dynamically** for each discovered resource
3. **Use dynamic client** for generic operations on any resource type

**Benefits:**
- Works with ANY Kubernetes resource, including Custom Resource Definitions (CRDs)
- No need to update code when new resources are added to cluster
- Automatically respects RBAC (if user can't discover a resource, no tools are generated)

### 2. Official MCP Go SDK

We use `github.com/modelcontextprotocol/go-sdk` instead of implementing MCP from scratch.

**Benefits:**
- Well-maintained by MCP team and Google
- Handles JSON-RPC 2.0 protocol correctly
- Supports stdio transport (and HTTP/SSE for future)
- Reduces maintenance burden

### 3. Kubernetes client-go

We use `k8s.io/client-go` (the official Kubernetes Go client library).

**Benefits:**
- Most widely used and maintained
- Stays aligned with Kubernetes core
- No need to maintain API compatibility code

### 4. Cluster Abstraction Layer

Code is designed with a `Cluster` abstraction, even though MVP only supports single cluster.

**Benefits:**
- Easy to add multi-cluster support later
- No major refactoring needed
- Clean separation of concerns

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                    MCP Client (ChatGPT, Claude, etc.)        │
└───────────────────────┬─────────────────────────────────────┘
                        │ JSON-RPC 2.0 over stdio
                        │
┌───────────────────────▼─────────────────────────────────────┐
│              MCP Server (go-sdk)                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │         Tool Registry (Dynamic)                      │  │
│  │  - k8s_list_pods                                     │  │
│  │  - k8s_get_pod                                       │  │
│  │  - k8s_create_deployment (if write enabled)         │  │
│  │  - k8s_get_pod_logs (special)                       │  │
│  │  - ... (auto-generated for all resources)            │  │
│  └───────────────────┬──────────────────────────────────┘  │
│                      │                                       │
│  ┌───────────────────▼──────────────────────────────────┐  │
│  │         Tool Handlers                                 │  │
│  │  - Generic operations (list, get, create, etc.)      │  │
│  │  - Special operations (logs, events)                  │  │
│  └───────────────────┬──────────────────────────────────┘  │
└───────────────────────┼─────────────────────────────────────┘
                        │
┌───────────────────────▼─────────────────────────────────────┐
│         Kubernetes Client Layer                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Cluster Manager                                      │  │
│  │  - Discovery Client (find resources)                 │  │
│  │  - Dynamic Client (generic operations)                │  │
│  │  - Typed Clients (logs, events)                      │  │
│  └───────────────────┬──────────────────────────────────┘  │
└───────────────────────┼─────────────────────────────────────┘
                        │
┌───────────────────────▼─────────────────────────────────────┐
│              Kubernetes API Server                            │
│         (via kubeconfig or in-cluster SA)                    │
└─────────────────────────────────────────────────────────────┘
```

## Component Details

### MCP Server (`internal/server/`)

- **server.go**: Main MCP server using go-sdk, handles router mode vs direct mode
- **router.go**: Native Go router for tool access (workaround for 40-tool limit)
- **router_invoke.go**: Reflection-based tool invocation for router mode
- **register.go**: Direct tool registration (legacy mode)
- **register_router.go**: Router-based tool registration (router mode)

### Kubernetes Client (`internal/kubernetes/`)

- **client.go**: Client manager (creates and manages all client-go clients)
- **cluster.go**: Cluster abstraction (single cluster for MVP)
- **discovery.go**: Resource discovery and caching
- **operations/**: Generic operation handlers (list, get, create, update, patch, delete, apply)
- **special/**: Special operations (logs, events, describe)

### Resource Discovery Flow

1. **Startup**: Server initializes Kubernetes clients
2. **Discovery**: Query Discovery API for all available resources
3. **Cache**: Store resource metadata (API version, kind, namespaced, verbs)
4. **Tool Generation**: For each resource, generate MCP tools:
   - Read tools: `k8s_list_{resource}`, `k8s_get_{resource}`, `k8s_describe_{resource}`
   - Write tools (if enabled): `k8s_create_{resource}`, `k8s_update_{resource}`, etc.
5. **Registration**: 
   - **Router Mode** (if `KUBEBRIDGE_USE_ROUTER=true`): Register all tools in internal router, expose only `k8s_retrieve_tools` and `k8s_call_tool` to MCP server
   - **Direct Mode** (default): Register tools directly with MCP server (limited by client tool limits)

### Operation Execution Flow

#### Router Mode (Recommended)

1. **MCP Request**: Client calls `k8s_retrieve_tools` or `k8s_call_tool`
2. **Router Handler**: 
   - `k8s_retrieve_tools`: Search internal router for matching tools
   - `k8s_call_tool`: Look up tool in router and invoke via reflection
3. **Tool Invocation**: Use reflection to call the actual tool handler
4. **Parameter Validation**: Validate input parameters
5. **Kubernetes Operation**: Execute operation using dynamic or typed client
6. **Response Formatting**: Format result as JSON
7. **MCP Response**: Return to client

#### Direct Mode (Legacy)

1. **MCP Request**: Client calls tool directly (e.g., `k8s_list_pods`)
2. **Tool Handler**: Route to appropriate handler
3. **Parameter Validation**: Validate input parameters
4. **Kubernetes Operation**: Execute operation using dynamic or typed client
5. **Response Formatting**: Format result as JSON
6. **MCP Response**: Return to client

## Security Architecture

### RBAC Integration

- Kubernetes RBAC naturally limits what tools can do
- If user doesn't have permission, operation fails with clear error
- Discovery API respects RBAC - only resources user can list are discovered
- No tools are generated for resources user can't access

### Input Validation

- All parameters validated before sending to Kubernetes API
- Namespace and resource names validated against K8s naming rules
- JSON manifests validated before create/update operations
- Log limits enforced (max 1000 lines)

### Error Handling

- Errors sanitized (no sensitive info exposed)
- Detailed errors logged server-side only
- User-friendly error messages returned to MCP client
- RBAC errors distinguished from other errors

### Write Operations

- **Default: OFF** - Must explicitly enable
- All write operations logged (audit trail)
- Kubernetes admission controllers provide additional protection
- RBAC enforces permissions

## Extensibility

### Adding Multi-Cluster Support

The cluster abstraction makes this straightforward:

1. Create `MultiClusterManager` that manages multiple `Cluster` instances
2. Add cluster selection to tool parameters (or context)
3. Route operations to appropriate cluster
4. No changes needed to operation handlers

### Adding New Special Operations

1. Add handler in `internal/kubernetes/special/`
2. Register tool in `internal/server/tools.go`
3. Use typed client for resource-specific operations

### Adding HTTP/SSE Transport

MCP Go SDK supports this - just need to:
1. Configure server to use HTTP/SSE transport instead of stdio
2. Add authentication/authorization layer
3. No changes to tool handlers needed

## Performance Considerations

### Resource Discovery Caching

- Discovery happens once at startup
- Resource metadata cached in memory
- Can be refreshed on-demand (future enhancement)

### Dynamic Client Efficiency

- Dynamic client is efficient for generic operations
- Typed clients used only when needed (logs, events)
- No unnecessary client creation

### Tool Generation

- Tools generated once at startup
- No runtime overhead for tool generation
- Tool definitions cached

## Router Mode Architecture

### Problem: MCP Client Tool Limits

Some MCP clients (e.g., Cursor) have a limit on the number of tools they can expose (40 tools). With 300+ Kubernetes resources, this limit is quickly exceeded.

### Solution: Native Go Router

KubeBridge-MCP implements a native Go router that:

1. **Registers all tools internally** (300+ Kubernetes tools)
2. **Exposes only 2 router tools** to the MCP client:
   - `k8s_retrieve_tools`: Search and discover available tools
   - `k8s_call_tool`: Call any tool by name with arguments
3. **Uses reflection** to dynamically invoke tool handlers stored in the router

### Benefits

- ✅ Bypasses MCP client tool limits
- ✅ Provides access to all discovered Kubernetes tools
- ✅ No functionality loss
- ✅ Clean, maintainable Go implementation

### Router Implementation

- **router.go**: Router data structure, tool registration, search functionality
- **router_invoke.go**: Reflection-based handler invocation
- **register_router.go**: Register all Kubernetes tools in router
- **server.go**: Conditionally use router mode based on `KUBEBRIDGE_USE_ROUTER` env var

## Testing Strategy

### Unit Tests

- Router tests: Tool registration, search, invocation
- Server tests: Initialization, router mode, tool registration
- Kubernetes operations tests: List, get operations using fake clients
- Reflection-based invocation tests

### Integration Tests

- End-to-end router mode flow
- Tool discovery and invocation through router
- Server initialization with router enabled/disabled

### MCP Protocol Tests

- Test MCP protocol compliance
- Test with real MCP clients (Cursor, ChatGPT, Claude)
- Test error responses
- Test tool discovery via router


