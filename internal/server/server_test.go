package server

import "testing"

func TestNewServer(t *testing.T) {
	srv := NewServer()
	if srv == nil {
		t.Fatal("NewServer returned nil")
	}
}

func TestServerStart(t *testing.T) {
	srv := NewServer()
	err := srv.Start()
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
}
