package kubernetes

import (
	"context"
	"fmt"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Cluster represents a Kubernetes cluster connection
// Designed to be extensible for multi-cluster support
type Cluster struct {
	name       string
	config     *rest.Config
	typed      kubernetes.Interface
	dynamic    dynamic.Interface
	discovery  DiscoveryInterface
}

// NewCluster creates a new cluster instance
func NewCluster(name string, config *rest.Config) (*Cluster, error) {
	typedClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create typed client: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	discoveryClient, err := NewDiscoveryClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery client: %w", err)
	}

	return &Cluster{
		name:      name,
		config:    config,
		typed:     typedClient,
		dynamic:   dynamicClient,
		discovery: discoveryClient,
	}, nil
}

// Name returns the cluster name
func (c *Cluster) Name() string {
	return c.name
}

// Typed returns the typed Kubernetes client
func (c *Cluster) Typed() kubernetes.Interface {
	return c.typed
}

// Dynamic returns the dynamic Kubernetes client
func (c *Cluster) Dynamic() dynamic.Interface {
	return c.dynamic
}

// Discovery returns the discovery client
func (c *Cluster) Discovery() DiscoveryInterface {
	return c.discovery
}

// Health checks if the cluster is accessible
func (c *Cluster) Health(ctx context.Context) error {
	_, err := c.typed.Discovery().ServerVersion()
	return err
}


