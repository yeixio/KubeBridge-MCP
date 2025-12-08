package mcp

import (
	"encoding/json"
	"testing"
)

func TestMessageJSON(t *testing.T) {
	msg := Message{
		Type:    "test",
		Payload: "test payload",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	var decoded Message
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}

	if decoded.Type != msg.Type {
		t.Errorf("Expected type %s, got %s", msg.Type, decoded.Type)
	}
}
