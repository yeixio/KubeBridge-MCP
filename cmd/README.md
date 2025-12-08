# Command Line Applications

This directory contains the main applications for this project.

## Structure

Each subdirectory under `/cmd` represents a separate executable application:

- `kubebridge-mcp/` - Main MCP server application for Kubernetes integration

## Building

To build a specific application:

```bash
go build -o build/kubebridge-mcp ./cmd/kubebridge-mcp
```

Or use the Makefile:

```bash
make build
```
