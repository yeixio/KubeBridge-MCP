# KubeBridge-MCP MVP Plan

## Overview

KubeBridge-MCP is a Model Context Protocol (MCP) server that enables AI assistants (ChatGPT, Claude, and other MCP-compatible clients) to safely interact with Kubernetes clusters through a standardized tool interface.

## Core Requirements

### ✅ Confirmed Requirements
1. **Use `k8s.io/client-go`** - The official, widely-used Kubernetes Go client library (what kubectl uses internally)
2. **Golang implementation** - Go 1.23+
3. **MCP-compatible with any AI chat** - ChatGPT, Claude, Gemini, and other MCP clients
4. **Read-only by default** - Safe default behavior
5. **Optional write controls** - User can enable write operations via `--enable-write-operations` flag or `KUBEBRIDGE_ENABLE_WRITE` env var
6. **Automatic resource exposure** - Dynamically expose all Kubernetes resources that client-go can interact with
7. **Single cluster (extensible)** - MVP supports single cluster, but architecture designed for easy multi-cluster addition

### ✅ Decisions Made

1. **Kubernetes Client Library**: `k8s.io/client-go` ✅
   - Most widely used and maintained
   - Stays aligned with Kubernetes core without maintaining API code

2. **MCP Protocol Implementation**: Use official MCP Go SDK ✅
   - Library: `github.com/modelcontextprotocol/go-sdk`
   - Maintained by MCP team and Google
   - Well-tested and up-to-date
   - Supports stdio transport (and HTTP/SSE for future)

3. **Write Operations**: Support ALL operations client-go can perform ✅
   - **Read operations** (always available): `list`, `get`, `describe`, `logs`, `events`
   - **Write operations** (when enabled): `create`, `update`, `patch`, `delete`, `apply`
   - Automatically generated for any discovered Kubernetes resource
   - No hardcoded resource types - uses dynamic client + discovery API

4. **Configuration Method**: CLI flag + environment variable ✅
   - `--enable-write-operations` (default: false)
   - `KUBEBRIDGE_ENABLE_WRITE=true` (env var override)

5. **Multi-cluster Support**: Single cluster for MVP, extensible architecture ✅
   - Code designed with cluster abstraction layer
   - Easy to add multi-cluster support later without refactoring

6. **Transport Protocol**: stdio transport ✅
   - Recommended approach for local MCP servers
   - MCP Go SDK handles this automatically

## MVP Scope

### Phase 1: Core Infrastructure (Week 1)

#### 1.1 MCP Protocol Implementation
- [ ] Integrate `github.com/modelcontextprotocol/go-sdk`
- [ ] Set up stdio transport handler
- [ ] Implement MCP server initialization
- [ ] Tool registration and discovery system
- [ ] Error handling and response formatting

#### 1.2 Kubernetes Client Setup
- [ ] Initialize client-go with kubeconfig or in-cluster config
- [ ] Support both local kubeconfig and in-cluster service account
- [ ] Set up dynamic client (`client-go/dynamic`) for generic operations
- [ ] Set up discovery client (`client-go/discovery`) for resource discovery
- [ ] Set up typed clients for special operations (logs, events)
- [ ] Connection validation and health checks
- [ ] RBAC-aware error handling
- [ ] Cluster abstraction layer (single cluster, but extensible)

#### 1.3 Resource Discovery & Tool Generation
- [ ] Implement resource discovery using discovery API
- [ ] Cache discovered API resources (groups, versions, kinds)
- [ ] Tool generator that creates MCP tools for each discovered resource
- [ ] Generate tools for standard operations:
  - `k8s_list_{resource}` - List resources (read-only, always available)
  - `k8s_get_{resource}` - Get resource details (read-only, always available)
  - `k8s_describe_{resource}` - Human-readable description (read-only, always available)
  - `k8s_create_{resource}` - Create resource (write, requires flag)
  - `k8s_update_{resource}` - Update resource (write, requires flag)
  - `k8s_patch_{resource}` - Patch resource (write, requires flag)
  - `k8s_delete_{resource}` - Delete resource (write, requires flag)
  - `k8s_apply_{resource}` - Apply resource (write, requires flag)
- [ ] Special tools for pods (logs) and events (not generic resources)

#### 1.4 Configuration Management
- [ ] Command-line argument parsing (cobra recommended)
- [ ] Configuration struct with:
  - `--cluster-name` (optional, for display)
  - `--kubeconfig` (optional, defaults to `$KUBECONFIG` or `~/.kube/config`)
  - `--enable-write-operations` (default: false)
  - `--log-level` (default: info)
- [ ] Environment variable support (`KUBEBRIDGE_ENABLE_WRITE`, etc.)

### Phase 2: Core Operations (Week 1-2)

#### 2.1 Generic Resource Operations (Auto-generated)
- [ ] `k8s_list_{resource}` - List any Kubernetes resource
  - Parameters: `apiVersion`, `kind`, `namespace` (optional)
  - Returns: Array of resource objects
  - **Auto-generated** for all discovered resources

- [ ] `k8s_get_{resource}` - Get any Kubernetes resource
  - Parameters: `apiVersion`, `kind`, `namespace`, `name`
  - Returns: Resource manifest (JSON)
  - **Auto-generated** for all discovered resources

- [ ] `k8s_describe_{resource}` - Describe any Kubernetes resource
  - Parameters: `apiVersion`, `kind`, `namespace`, `name`
  - Returns: Human-readable description
  - **Auto-generated** for all discovered resources

#### 2.2 Special Operations (Hand-coded)
- [ ] `k8s_get_pod_logs` - Get pod logs
  - Parameters: `namespace`, `name`, `tailLines` (optional, default: 100), `container` (optional)
  - Returns: Log lines as array
  - **Safety**: Limit `tailLines` to max 1000

- [ ] `k8s_get_events` - Get events for resources
  - Parameters: `namespace`, `kind` (optional), `name` (optional)
  - Returns: Array of events sorted by time

### Phase 3: Write Operations (Week 2)

**Only available when `--enable-write-operations` is set**

#### 3.1 Generic Write Operations (Auto-generated)
- [ ] `k8s_create_{resource}` - Create any Kubernetes resource
  - Parameters: `apiVersion`, `kind`, `namespace` (optional), `manifest` (JSON)
  - Returns: Created resource
  - **Auto-generated** for all discovered resources

- [ ] `k8s_update_{resource}` - Update any Kubernetes resource
  - Parameters: `apiVersion`, `kind`, `namespace`, `name`, `manifest` (JSON)
  - Returns: Updated resource
  - **Auto-generated** for all discovered resources

- [ ] `k8s_patch_{resource}` - Patch any Kubernetes resource
  - Parameters: `apiVersion`, `kind`, `namespace`, `name`, `patch` (JSON patch or strategic merge)
  - Returns: Patched resource
  - **Auto-generated** for all discovered resources

- [ ] `k8s_delete_{resource}` - Delete any Kubernetes resource
  - Parameters: `apiVersion`, `kind`, `namespace`, `name`
  - Returns: Success/failure status
  - **Auto-generated** for all discovered resources
  - **Safety**: Consider dry-run mode or confirmation

- [ ] `k8s_apply_{resource}` - Apply any Kubernetes resource (server-side apply)
  - Parameters: `apiVersion`, `kind`, `namespace` (optional), `manifest` (JSON)
  - Returns: Applied resource
  - **Auto-generated** for all discovered resources

### Phase 4: Testing & Documentation (Week 2-3)

#### 4.1 Testing
- [ ] Unit tests for Kubernetes client operations
- [ ] Integration tests with kind/minikube cluster
- [ ] MCP protocol compliance tests
- [ ] Error handling tests (RBAC failures, network errors, etc.)

#### 4.2 Documentation
- [ ] Updated README with MVP features
- [ ] MCP client configuration examples (ChatGPT, Claude, etc.)
- [ ] Tool reference documentation
- [ ] Security best practices guide
- [ ] Troubleshooting guide

## Technical Architecture

### Project Structure
```
KubeBridge-MCP/
├── cmd/
│   └── kubebridge-mcp/
│       └── main.go              # Entry point, CLI parsing, server startup
├── internal/
│   ├── server/
│   │   ├── server.go            # MCP server implementation (using go-sdk)
│   │   ├── tools.go             # Tool registry and dynamic tool generation
│   │   └── handlers.go          # Tool execution handlers
│   └── kubernetes/
│       ├── client.go             # Client manager (typed, dynamic, discovery)
│       ├── cluster.go            # Cluster abstraction (single cluster, extensible)
│       ├── discovery.go          # Resource discovery and caching
│       ├── operations.go         # Generic K8s operations (list, get, create, etc.)
│       ├── special.go            # Special operations (logs, events, describe)
│       └── errors.go             # K8s error handling and sanitization
├── pkg/
│   └── mcp/
│       └── tools.go              # Tool definition types and helpers
├── configs/
│   └── examples/                 # Example MCP client configs
└── docs/
    ├── MVP_PLAN.md              # This file
    └── TOOLS.md                  # Tool reference (auto-generated docs)
```

### Dependencies (go.mod)
```go
require (
    // MCP Protocol SDK
    github.com/modelcontextprotocol/go-sdk v1.0.0  // Official MCP Go SDK
    
    // Kubernetes client libraries
    k8s.io/client-go v0.30.0        // Kubernetes client (typed + dynamic)
    k8s.io/api v0.30.0              // Kubernetes API types
    k8s.io/apimachinery v0.30.0     // Kubernetes object types
    
    // CLI framework
    github.com/spf13/cobra v1.8.0   // CLI framework
    
    // Utilities
    github.com/spf13/viper v1.18.0  // Config management (optional)
)
```

### Architecture: Dynamic Resource Discovery

The key innovation is **automatic tool generation** based on Kubernetes resource discovery:

1. **Discovery Phase** (on startup):
   - Use `client-go/discovery` to query the Kubernetes API server
   - Discover all available API groups, versions, and resources
   - Cache resource metadata (names, namespaced vs cluster-scoped, verbs supported)

2. **Tool Generation** (on startup):
   - For each discovered resource, generate MCP tools:
     - Read tools: `k8s_list_{resource}`, `k8s_get_{resource}`, `k8s_describe_{resource}`
     - Write tools (if enabled): `k8s_create_{resource}`, `k8s_update_{resource}`, etc.
   - Register tools with MCP server

3. **Operation Execution** (runtime):
   - Use `client-go/dynamic` client for generic operations
   - Use typed clients for special cases (logs, events)
   - Handle RBAC errors gracefully

**Example**: If cluster has `Deployments`, `Services`, `ConfigMaps`, `CustomResources`, etc., tools are automatically generated:
- `k8s_list_deployments`, `k8s_get_deployment`, `k8s_create_deployment`, etc.
- `k8s_list_services`, `k8s_get_service`, `k8s_create_service`, etc.
- `k8s_list_configmaps`, `k8s_get_configmap`, `k8s_create_configmap`, etc.
- Works for **any** resource, including CRDs!

### MCP Tool Definition Examples

#### Auto-generated tool (example for Deployments):
```json
{
  "name": "k8s_list_deployments",
  "description": "List deployments in a namespace. Auto-generated for apps/v1/Deployment resource.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "namespace": {
        "type": "string",
        "description": "Kubernetes namespace (optional, defaults to all namespaces)"
      }
    }
  }
}
```

#### Write tool (only when enabled):
```json
{
  "name": "k8s_create_deployment",
  "description": "Create a deployment. Requires --enable-write-operations flag.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "namespace": {
        "type": "string",
        "description": "Kubernetes namespace"
      },
      "manifest": {
        "type": "object",
        "description": "Deployment manifest (JSON)"
      }
    },
    "required": ["namespace", "manifest"]
  }
}
```

#### Special tool (hand-coded):
```json
{
  "name": "k8s_get_pod_logs",
  "description": "Get logs from a pod container.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "namespace": {
        "type": "string",
        "description": "Kubernetes namespace"
      },
      "name": {
        "type": "string",
        "description": "Pod name"
      },
      "container": {
        "type": "string",
        "description": "Container name (optional, defaults to first container)"
      },
      "tailLines": {
        "type": "integer",
        "description": "Number of lines to tail (optional, default: 100, max: 1000)"
      }
    },
    "required": ["namespace", "name"]
  }
}
```

## Security Considerations

1. **RBAC**: The service account/user must have appropriate RBAC permissions
   - Read-only mode: `get`, `list`, `watch` on resources (enforced by Kubernetes)
   - Write mode: Additional `create`, `update`, `patch`, `delete` permissions (enforced by Kubernetes)
   - **Key Point**: Kubernetes RBAC naturally limits what tools can do - if user doesn't have permission, operation fails

2. **Input Validation**: All tool parameters must be validated
   - Namespace names: validated against K8s naming rules (RFC 1123)
   - Resource names: validated against K8s naming rules
   - API versions and kinds: validated against discovered resources
   - Log limits: enforced maximums (1000 lines)
   - JSON manifests: validated before sending to API server

3. **Error Handling**: Never expose sensitive information in errors
   - Sanitize error messages (remove tokens, credentials, etc.)
   - Log detailed errors server-side only
   - Return user-friendly error messages to MCP client
   - Distinguish between RBAC errors, not found, validation errors, etc.

4. **Write Operations**: When enabled, add safeguards
   - **Default: OFF** - Must explicitly enable with `--enable-write-operations`
   - Log all write operations (audit trail)
   - Consider dry-run mode for destructive operations (future enhancement)
   - Consider rate limiting (future enhancement)
   - **Note**: Kubernetes admission controllers and RBAC provide additional protection

5. **Resource Discovery**: Only expose resources the user has permission to discover
   - Discovery API respects RBAC
   - If user can't list a resource type, tools won't be generated for it
   - This is a natural security boundary

## Success Criteria

MVP is complete when:
- ✅ MCP server starts and connects to Kubernetes cluster
- ✅ Resource discovery works and finds available resources
- ✅ Tools are automatically generated for discovered resources
- ✅ Read-only tools work for all discovered resources (list, get, describe)
- ✅ Special tools work (pod logs, events)
- ✅ Write tools work when enabled (create, update, patch, delete, apply)
- ✅ MCP protocol is properly implemented (compatible with ChatGPT, Claude, etc.)
- ✅ Proper error handling for RBAC failures, not found, validation errors
- ✅ Configuration works (CLI flags, env vars)
- ✅ Documentation is complete (README, tool reference, examples)
- ✅ Basic tests pass (unit tests, integration tests with kind/minikube)

## Out of Scope for MVP

- Multi-cluster support (architecture supports it, but not implemented)
- HTTP/SSE transport (stdio only - MCP SDK supports HTTP/SSE, but not needed for MVP)
- Resource watching/streaming (future: real-time updates)
- Metrics and monitoring endpoints
- Web UI or dashboard
- Authentication/authorization beyond RBAC (Kubernetes handles this)
- Dry-run mode for write operations (future enhancement)
- Rate limiting (future enhancement)
- Resource filtering/search beyond namespace (future: label selectors, field selectors)

## Implementation Strategy

### Key Design Decisions

1. **Dynamic Tool Generation**: Instead of hardcoding tools for specific resources, we:
   - Discover available resources at startup
   - Generate MCP tools dynamically
   - Use dynamic client for operations
   - Result: Works with ANY Kubernetes resource, including CRDs!

2. **Cluster Abstraction**: Design for future multi-cluster support:
   - `Cluster` interface/struct that wraps client-go clients
   - Single cluster implementation for MVP
   - Easy to add multi-cluster manager later

3. **Operation Abstraction**: Generic operation handlers:
   - `ListOperation`, `GetOperation`, `CreateOperation`, etc.
   - Work with any resource type via dynamic client
   - Special operations (logs, events) use typed clients

### Code Organization

```
internal/kubernetes/
├── client.go          # Client manager (creates typed, dynamic, discovery clients)
├── cluster.go         # Cluster abstraction (single cluster for MVP)
├── discovery.go       # Resource discovery and caching
├── operations/
│   ├── list.go        # Generic list operation
│   ├── get.go         # Generic get operation
│   ├── describe.go    # Generic describe operation
│   ├── create.go      # Generic create operation (write)
│   ├── update.go      # Generic update operation (write)
│   ├── patch.go       # Generic patch operation (write)
│   ├── delete.go      # Generic delete operation (write)
│   └── apply.go       # Generic apply operation (write)
└── special/
    ├── logs.go        # Pod logs (uses typed client)
    └── events.go      # Events (uses typed client)
```

## Next Steps

1. ✅ **MVP plan approved** - All questions answered
2. **Set up dependencies**:
   - Add `github.com/modelcontextprotocol/go-sdk` to go.mod
   - Add `k8s.io/client-go` and related packages
   - Add `github.com/spf13/cobra` for CLI
3. **Set up development environment**:
   - Install kind or minikube for local testing
   - Create test cluster with sample resources
4. **Implement Phase 1** (Core Infrastructure):
   - MCP server setup with go-sdk
   - Kubernetes client initialization
   - Resource discovery
   - Tool generation system
5. **Implement Phase 2** (Core Operations):
   - Generic operation handlers
   - Special operations (logs, events)
6. **Implement Phase 3** (Write Operations):
   - Write operation handlers (when enabled)
7. **Testing & Documentation**:
   - Unit tests
   - Integration tests
   - Documentation

## Timeline Estimate

- **Week 1**: 
  - Core infrastructure (MCP server, K8s client, discovery)
  - Tool generation system
  - Read operations (list, get, describe)
- **Week 2**: 
  - Special operations (logs, events)
  - Write operations (create, update, patch, delete, apply)
  - Testing
- **Week 3**: 
  - Documentation
  - Polish
  - MVP release

**Total: 2-3 weeks for MVP**

