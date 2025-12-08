# KubeBridge-MCP

[![CI](https://github.com/yeixio/KubeBridge-MCP/actions/workflows/ci.yml/badge.svg)](https://github.com/yeixio/KubeBridge-MCP/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/yeixio/KubeBridge-MCP)](https://goreportcard.com/report/github.com/yeixio/KubeBridge-MCP)
[![License](https://img.shields.io/github/license/yeixio/KubeBridge-MCP)](LICENSE)

A Model Context Protocol (MCP) server that lets ChatGPT talk to your Kubernetes cluster safely and sanely.

## Features

- 🔐 Secure Kubernetes cluster integration
- 🤖 Model Context Protocol (MCP) support
- 🚀 High-performance Go implementation
- 📦 Easy deployment with Docker

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/yeixio/KubeBridge-MCP.git
cd KubeBridge-MCP

# Build the binary
make build

# Install to $GOPATH/bin
make install
```

### Using Docker

```bash
docker build -t kubebridge-mcp .
docker run -p 8080:8080 kubebridge-mcp
```

## Usage

```bash
# Run the server
./build/kubebridge-mcp

# Or use make
make run
```

## Development

### Prerequisites

- Go 1.21 or later
- Make
- Docker (optional)

### Building

```bash
# Install dependencies
make mod-download

# Format code
make fmt

# Run linters
make lint

# Run tests
make test

# Run all checks
make check

# Build binary
make build
```

### Project Structure

```
KubeBridge-MCP/
├── cmd/              # Main applications
│   └── kubebridge-mcp/
├── internal/         # Private application code
│   ├── server/
│   └── kubernetes/
├── pkg/              # Public library code
│   └── mcp/
├── api/              # API definitions
├── configs/          # Configuration files
├── scripts/          # Build and utility scripts
├── test/             # Additional test data
├── docs/             # Documentation
└── examples/         # Example usage
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

This project is licensed under the terms specified in the [LICENSE](LICENSE) file.

## Acknowledgments

Built with ❤️ using Go and following Go project layout standards.
