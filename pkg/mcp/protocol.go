package mcp

// Protocol defines the Model Context Protocol interface
type Protocol interface {
	// TODO: Define MCP protocol methods
}

// Message represents an MCP message
type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}
