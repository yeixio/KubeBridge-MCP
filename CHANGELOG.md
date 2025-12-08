# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Router mode to work around MCP client tool limits (40-tool limit)
- Comprehensive test suite with 28 tests
- Test helpers with fake Kubernetes clients
- Error analysis in pod logs tool
- Documentation for router mode and architecture

### Changed
- Build output moved to `./bin` directory
- Updated Go version requirement to 1.24
- Improved list tool to include resource names in summary

### Fixed
- Fixed linter errors in error handling
- Fixed Dockerfile build path
- Fixed CI workflow Go version

## [0.1.0] - 2025-12-08

### Added
- Initial release
- Dynamic Kubernetes resource discovery
- Auto-generated MCP tools for all discovered resources
- Read-only operations (list, get, describe)
- Optional write operations (create, update, patch, delete, apply)
- Special operations (pod logs, events)
- RBAC integration
- Error sanitization
- Comprehensive documentation

[Unreleased]: https://github.com/yeixio/KubeBridge-MCP/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/yeixio/KubeBridge-MCP/releases/tag/v0.1.0

