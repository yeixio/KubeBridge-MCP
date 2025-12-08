package kubernetes

import "testing"

func TestNewClient(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
}
