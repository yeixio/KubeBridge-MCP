# Contributing to KubeBridge-MCP

Thank you for your interest in contributing to KubeBridge-MCP! This document provides guidelines and instructions for contributing.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/KubeBridge-MCP.git`
3. Add upstream remote: `git remote add upstream https://github.com/yeixio/KubeBridge-MCP.git`
4. Create a new branch: `git checkout -b feature/your-feature-name`

## Development Setup

### Prerequisites

- Go 1.21 or later
- Make
- Docker (optional, for containerized development)

### Building the Project

```bash
# Install dependencies
make mod-download

# Build the binary
make build

# Run tests
make test

# Run all checks (format, vet, lint, test)
make check
```

## Code Standards

### Go Code Style

- Follow standard Go formatting guidelines (enforced by `gofmt`)
- Use `go vet` to check for common mistakes
- Follow the [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Write clear, descriptive comments for exported functions and types

### Project Structure

```
KubeBridge-MCP/
├── cmd/              # Main applications
├── internal/         # Private application code
├── pkg/              # Public library code
├── api/              # API definitions
├── configs/          # Configuration files
├── scripts/          # Build and utility scripts
├── test/             # Additional test data
├── docs/             # Documentation
└── examples/         # Example usage
```

### Commit Messages

- Use clear, descriptive commit messages
- Start with a verb in present tense (e.g., "Add", "Fix", "Update")
- Reference issue numbers when applicable (e.g., "Fix #123: Description")

Example:
```
Add Kubernetes pod listing functionality

- Implement pod client in internal/kubernetes
- Add tests for pod operations
- Update server to handle pod list requests
```

## Testing

- Write unit tests for new functionality
- Ensure all tests pass before submitting a PR: `make test`
- Aim for good test coverage: `make coverage`
- Include integration tests where appropriate

## Pull Request Process

1. Update your branch with the latest upstream changes:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. Ensure all tests pass and code is formatted:
   ```bash
   make check
   ```

3. Push your changes to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

4. Create a Pull Request with:
   - Clear description of changes
   - Reference to related issues
   - Screenshots/examples if applicable

5. Address review comments and update your PR as needed

## Code Review Guidelines

- Be respectful and constructive
- Focus on code quality and maintainability
- Ensure changes align with project goals
- Check for security implications

## Questions?

Feel free to open an issue for questions or discussions about contributing.
