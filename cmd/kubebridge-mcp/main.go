package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/yeixio/KubeBridge-MCP/internal/config"
	"github.com/yeixio/KubeBridge-MCP/internal/kubernetes"
	"github.com/yeixio/KubeBridge-MCP/internal/server"
)

var (
	clusterName          string
	kubeconfig          string
	enableWriteOps      bool
	logLevel            string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "kubebridge-mcp",
		Short: "KubeBridge-MCP - MCP server for Kubernetes",
		Long:  "A Model Context Protocol (MCP) server that enables AI assistants to safely interact with Kubernetes clusters.",
		RunE:  run,
	}

	rootCmd.Flags().StringVar(&clusterName, "cluster-name", "", "Kubernetes cluster name (optional, for display)")
	rootCmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file (optional, defaults to $KUBECONFIG or ~/.kube/config)")
	rootCmd.Flags().BoolVar(&enableWriteOps, "enable-write-operations", false, "Enable write operations (create, update, patch, delete, apply)")
	rootCmd.Flags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg := config.LoadConfig()
	
	// Override with CLI flags if provided
	if clusterName != "" {
		cfg.ClusterName = clusterName
	}
	if kubeconfig != "" {
		cfg.Kubeconfig = kubeconfig
	}
	if cmd.Flags().Changed("enable-write-operations") {
		cfg.EnableWriteOperations = enableWriteOps
	}
	if logLevel != "" {
		cfg.LogLevel = logLevel
	}

	// Set default cluster name if not provided
	if cfg.ClusterName == "" {
		cfg.ClusterName = "default"
	}

	// Initialize Kubernetes client manager
	var clientManager *kubernetes.ClientManager
	var err error

	// Try in-cluster config first, then fall back to kubeconfig
	clientManager, err = kubernetes.NewInClusterClientManager(cfg.ClusterName)
	if err != nil {
		// Fall back to kubeconfig
		clientManager, err = kubernetes.NewClientManager(cfg.Kubeconfig, cfg.ClusterName)
		if err != nil {
			return fmt.Errorf("failed to initialize Kubernetes client: %w", err)
		}
	}

	// Health check
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := clientManager.Health(ctx); err != nil {
		return fmt.Errorf("failed to connect to Kubernetes cluster: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Connected to Kubernetes cluster: %s\n", cfg.ClusterName)

	// Create MCP server
	mcpServer, err := server.NewServer(cfg, clientManager)
	if err != nil {
		return fmt.Errorf("failed to create MCP server: %w", err)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Fprintf(os.Stderr, "\nShutting down...\n")
		cancel()
	}()

	// Start MCP server
	fmt.Fprintf(os.Stderr, "Starting KubeBridge-MCP server...\n")
	if cfg.EnableWriteOperations {
		fmt.Fprintf(os.Stderr, "WARNING: Write operations are ENABLED\n")
	} else {
		fmt.Fprintf(os.Stderr, "Write operations are disabled (use --enable-write-operations to enable)\n")
	}

	return mcpServer.Start(ctx)
}
