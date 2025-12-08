package main

import (
	"fmt"
	"os"

	"github.com/yeixio/KubeBridge-MCP/internal/server"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	srv := server.NewServer()
	return srv.Start()
}
