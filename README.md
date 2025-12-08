# KubeBridge-MCP

[![Go Report Card](https://goreportcard.com/badge/github.com/yeixio/KubeBridge-MCP)](https://goreportcard.com/report/github.com/yeixio/KubeBridge-MCP)
[![License](https://img.shields.io/github/license/yeixio/KubeBridge-MCP)](LICENSE)

> A Model Context Protocol (MCP) server that enables AI assistants (ChatGPT, Claude, Gemini, etc.) to safely interact with your Kubernetes cluster through a standardized tool interface.

## Overview

KubeBridge-MCP is an MCP server that dynamically discovers Kubernetes resources and automatically generates MCP tools for them. It works with **any** Kubernetes resource, including Custom Resource Definitions (CRDs), without requiring code changes.

### Key Features

- 🔍 **Dynamic Resource Discovery** - Automatically discovers all Kubernetes resources at startup
- 🛠️ **Auto-Generated Tools** - Creates MCP tools for every discovered resource (pods, deployments, services, CRDs, etc.)
- 🔐 **Read-Only by Default** - Safe default behavior; write operations require explicit opt-in
- ✍️ **Optional Write Operations** - Enable create, update, patch, delete, and apply operations when needed
- 🤖 **Universal MCP Compatibility** - Works with ChatGPT, Claude, Gemini, and any MCP-compatible client
- 🏗️ **Extensible Architecture** - Designed for easy multi-cluster support (single cluster in MVP)

## Architecture

KubeBridge-MCP uses a **dynamic resource discovery and tool generation** approach:

1. **Discovery Phase**: On startup, queries the Kubernetes Discovery API to find all available resources
2. **Tool Generation**: Automatically generates MCP tools for each discovered resource:
   - Read tools: `k8s_list_{resource}`, `k8s_get_{resource}`, `k8s_describe_{resource}`
   - Write tools (when enabled): `k8s_create_{resource}`, `k8s_update_{resource}`, `k8s_patch_{resource}`, `k8s_delete_{resource}`, `k8s_apply_{resource}`
3. **Router Mode** (Recommended): To work around MCP client tool limits (e.g., Cursor's 40-tool limit), KubeBridge-MCP includes a native Go router that exposes only 2 tools (`k8s_retrieve_tools` and `k8s_call_tool`) while providing access to all 300+ underlying Kubernetes tools
4. **Operation Execution**: Uses Kubernetes dynamic client for generic operations on any resource type

This means:
- ✅ Works with **any** Kubernetes resource, including CRDs
- ✅ No code changes needed when new resources are added to your cluster
- ✅ Automatically respects RBAC (only resources you can access are discovered)
- ✅ Router mode bypasses tool limits while maintaining full functionality

## Installation

### Prerequisites

- Go 1.24+ (or use pre-built binaries)
- A Kubernetes cluster and kubeconfig (or in-cluster service account)
- An MCP-compatible client (e.g., ChatGPT with MCP, Claude Desktop, etc.)

### From Source

```bash
# Clone the repository
git clone https://github.com/yeixio/KubeBridge-MCP.git
cd KubeBridge-MCP

# Build the binary (outputs to ./bin/kubebridge-mcp)
make build

# Or build directly with go
go build -o bin/kubebridge-mcp ./cmd/kubebridge-mcp
chmod +x bin/kubebridge-mcp
```

### Using Make

```bash
# Install dependencies
make mod-download

# Build binary (creates ./bin/kubebridge-mcp with executable permissions)
make build

# Install to $GOPATH/bin
make install

# Run the server
make run

# Clean build artifacts
make clean
```

## Usage

### Basic Usage (Read-Only)

```bash
# Run with default kubeconfig (~/.kube/config or $KUBECONFIG)
./bin/kubebridge-mcp

# Specify a kubeconfig file
./bin/kubebridge-mcp --kubeconfig /path/to/kubeconfig

# Specify cluster name (for display purposes)
./bin/kubebridge-mcp --cluster-name my-dev-cluster
```

### Enable Write Operations

⚠️ **Warning**: Enabling write operations allows the MCP client to create, update, and delete Kubernetes resources. Use with caution!

```bash
# Enable write operations via CLI flag
./bin/kubebridge-mcp --enable-write-operations

# Or via environment variable
export KUBEBRIDGE_ENABLE_WRITE=true
./bin/kubebridge-mcp
```

### Command-Line Options

```
--cluster-name string          Kubernetes cluster name (optional, for display)
--kubeconfig string            Path to kubeconfig file (optional, defaults to $KUBECONFIG or ~/.kube/config)
--enable-write-operations      Enable write operations (default: false)
--log-level string             Log level: debug, info, warn, error (default: "info")
```

### Environment Variables

- `KUBECONFIG` - Path to kubeconfig file
- `KUBEBRIDGE_CLUSTER_NAME` - Cluster name
- `KUBEBRIDGE_ENABLE_WRITE` - Enable write operations (true/false)
- `KUBEBRIDGE_LOG_LEVEL` - Log level
- `KUBEBRIDGE_USE_ROUTER` - Enable router mode (true/false, recommended: true)

## MCP Client Configuration

### ChatGPT (OpenAI) / Cursor

Add to your MCP configuration file. For Cursor, this is typically `~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "kubebridge": {
      "command": "/absolute/path/to/KubeBridge-MCP/bin/kubebridge-mcp",
      "args": ["--cluster-name", "my-cluster"],
      "env": {
        "KUBEBRIDGE_USE_ROUTER": "true"
      }
    }
  }
}
```

**Note**: 
- Use the absolute path to the binary in the `bin/` directory after building with `make build`
- Router mode (`KUBEBRIDGE_USE_ROUTER: "true"`) is recommended to work around MCP client tool limits (e.g., Cursor's 40-tool limit)

### Claude Desktop

Add to your Claude Desktop configuration (typically `~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "kubebridge": {
      "command": "/absolute/path/to/KubeBridge-MCP/bin/kubebridge-mcp",
      "args": ["--cluster-name", "my-cluster"],
      "env": {
        "KUBEBRIDGE_USE_ROUTER": "true"
      }
    }
  }
}
```

**Note**: 
- Use the absolute path to the binary in the `bin/` directory after building with `make build`
- Router mode (`KUBEBRIDGE_USE_ROUTER: "true"`) is recommended to work around MCP client tool limits

## Available Tools

### Router Mode (Recommended)

When `KUBEBRIDGE_USE_ROUTER=true` is set, KubeBridge-MCP exposes only 2 router tools that provide access to all 300+ Kubernetes tools:

- **`k8s_retrieve_tools`** - Search and discover available Kubernetes tools
  - Parameters: `query` (optional, search term), `limit` (optional, default: 10)
  - Returns: List of matching tools with names and descriptions
  - Example: Search for "pod" tools to find `k8s_list_pods`, `k8s_get_pod`, etc.

- **`k8s_call_tool`** - Call any Kubernetes tool by name
  - Parameters: `toolName` (required), `args` (required, tool-specific arguments)
  - Returns: Tool execution result
  - Example: `{"toolName": "k8s_list_pods", "args": {"namespace": "default"}}`

**Why Router Mode?**
- Works around MCP client tool limits (e.g., Cursor's 40-tool limit)
- Provides access to all discovered Kubernetes tools
- Maintains full functionality while exposing minimal tools to the client

### Direct Tool Mode (Legacy)

When router mode is disabled, tools are registered directly with the MCP server. This mode is limited by MCP client tool limits.

#### Read-Only Tools (Always Available)

For each discovered Kubernetes resource, the following tools are automatically generated:

- **`k8s_list_{resource}`** - List resources in a namespace
  - Parameters: `namespace` (optional)
  - Returns: Array of resource objects

- **`k8s_get_{resource}`** - Get a specific resource by name
  - Parameters: `namespace`, `name`
  - Returns: Resource manifest (JSON)

- **`k8s_describe_{resource}`** - Get human-readable description
  - Parameters: `namespace`, `name`
  - Returns: Formatted description

#### Special Tools

- **`k8s_get_pod_logs`** - Get logs from a pod container (includes error analysis)
  - Parameters: `namespace`, `name`, `container` (optional), `tailLines` (optional, default: 100, max: 1000)
  - Returns: Array of log lines with error summary in text output

- **`k8s_get_events`** - Get events for Kubernetes resources
  - Parameters: `namespace`, `kind` (optional), `name` (optional)
  - Returns: Object with `events` field containing array of events sorted by time

### Write Tools (Only when `--enable-write-operations` is set)

- **`k8s_create_{resource}`** - Create a new resource
  - Parameters: `namespace`, `manifest` (JSON)
  - Returns: Created resource

- **`k8s_update_{resource}`** - Update an existing resource
  - Parameters: `namespace`, `manifest` (JSON)
  - Returns: Updated resource

- **`k8s_patch_{resource}`** - Patch a resource
  - Parameters: `namespace`, `name`, `patch` (JSON patch), `patchType` (optional, default: "merge")
  - Returns: Patched resource

- **`k8s_delete_{resource}`** - Delete a resource
  - Parameters: `namespace`, `name`
  - Returns: Success status

- **`k8s_apply_{resource}`** - Apply a resource (server-side apply)
  - Parameters: `namespace`, `manifest` (JSON), `fieldManager` (optional)
  - Returns: Applied resource

### Example Tool Names

If your cluster has these resources, you'll automatically get tools like:
- `k8s_list_pods`, `k8s_get_pod`, `k8s_describe_pod`
- `k8s_list_deployments`, `k8s_get_deployment`, `k8s_create_deployment` (if write enabled)
- `k8s_list_services`, `k8s_get_service`, `k8s_update_service` (if write enabled)
- `k8s_list_configmaps`, `k8s_get_configmap`, etc.
- Works for **any** resource, including Custom Resource Definitions!

## Security Considerations

### RBAC

KubeBridge-MCP respects Kubernetes RBAC:

- If you don't have permission to list a resource type, no tools will be generated for it
- If you don't have permission to perform an operation, it will fail with a clear error
- The service account or user credentials determine what tools are available

### Read-Only by Default

- Write operations are **disabled by default**
- Must explicitly enable with `--enable-write-operations` flag
- All write operations are logged (audit trail)

### Error Handling

- Errors are sanitized to avoid exposing sensitive information
- Detailed errors are logged server-side only
- User-friendly error messages are returned to MCP clients

## Development

### Prerequisites

- Go 1.24 or later
- Make (optional, for using Makefile)
- A Kubernetes cluster for testing (kind, minikube, or real cluster)

### Building

```bash
# Install dependencies
go mod download

# Format code
make fmt

# Run linters
make lint

# Run tests
make test

# Build binary
make build
```

### Project Structure

```
KubeBridge-MCP/
├── cmd/
│   └── kubebridge-mcp/      # Main application entry point
├── internal/
│   ├── config/              # Configuration management
│   ├── kubernetes/           # Kubernetes client layer
│   │   ├── operations/      # Generic K8s operations (list, get, create, etc.)
│   │   ├── special/         # Special operations (logs, events)
│   │   └── test_helpers.go  # Test utilities with fake clients
│   └── server/              # MCP server implementation
│       ├── router.go        # Router for tool access (40-tool limit workaround)
│       ├── router_invoke.go  # Reflection-based tool invocation
│       ├── register.go      # Direct tool registration
│       └── register_router.go # Router-based tool registration
├── bin/                     # Build output (gitignored)
├── docs/                     # Documentation
│   ├── MVP_PLAN.md          # MVP planning document
│   ├── ARCHITECTURE.md      # Architecture documentation
│   └── TESTING.md           # Testing guide
└── pkg/                      # Public library code (currently empty)
```

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make coverage

# Run specific test package
go test ./internal/server -v
go test ./internal/kubernetes/operations -v
```

### Test Coverage

The project includes comprehensive test coverage:

- **Router Tests**: Tool registration, search, and invocation
- **Server Tests**: Server initialization, router mode, tool registration
- **Kubernetes Operations Tests**: List, get operations using fake clients
- **Integration Tests**: End-to-end flow testing

See [docs/TESTING.md](docs/TESTING.md) for comprehensive testing instructions with MCP clients.

## Roadmap

- [ ] Multi-cluster support
- [ ] HTTP/SSE transport (in addition to stdio)
- [ ] Resource watching/streaming
- [ ] Advanced filtering (label selectors, field selectors)
- [ ] Dry-run mode for write operations
- [ ] Rate limiting for write operations

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

This project is licensed under the terms specified in the [LICENSE](LICENSE) file.

## Acknowledgments

- Built with [Model Context Protocol Go SDK](https://github.com/modelcontextprotocol/go-sdk)
- Uses [Kubernetes client-go](https://github.com/kubernetes/client-go)
- Built with ❤️ using Go
