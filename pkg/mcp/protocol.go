// Package mcp provides Model Context Protocol definitions and utilities.
package mcp

// Protocol defines the Model Context Protocol interface
type Protocol interface {
	// TODO: Define MCP protocol methods
}

// Message represents an MCP message
type Message struct {
	Payload interface{} `json:"payload"`
	Type    string      `json:"type"`
}
