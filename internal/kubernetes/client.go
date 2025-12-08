package kubernetes

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ClientManager manages Kubernetes client connections
type ClientManager struct {
	cluster *Cluster
}

// NewClientManager creates a new client manager
func NewClientManager(kubeconfigPath string, clusterName string) (*ClientManager, error) {
	config, err := loadKubeconfig(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	cluster, err := NewCluster(clusterName, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create cluster: %w", err)
	}

	return &ClientManager{
		cluster: cluster,
	}, nil
}

// NewInClusterClientManager creates a client manager using in-cluster config
func NewInClusterClientManager(clusterName string) (*ClientManager, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load in-cluster config: %w", err)
	}

	cluster, err := NewCluster(clusterName, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create cluster: %w", err)
	}

	return &ClientManager{
		cluster: cluster,
	}, nil
}

// GetCluster returns the cluster instance
func (cm *ClientManager) GetCluster() *Cluster {
	return cm.cluster
}

// Health checks cluster connectivity
func (cm *ClientManager) Health(ctx context.Context) error {
	return cm.cluster.Health(ctx)
}

// DiscoverResources discovers all available resources
func (cm *ClientManager) DiscoverResources(ctx context.Context) error {
	discoveryClient, ok := cm.cluster.Discovery().(*DiscoveryClient)
	if !ok {
		return fmt.Errorf("discovery client is not a DiscoveryClient")
	}
	return discoveryClient.DiscoverResources(ctx)
}

// GetDiscoveredResources returns all discovered resources
func (cm *ClientManager) GetDiscoveredResources() []*APIResource {
	discoveryClient, ok := cm.cluster.Discovery().(*DiscoveryClient)
	if !ok {
		return nil
	}
	return discoveryClient.GetResources()
}

// loadKubeconfig loads kubeconfig from file or uses default location
func loadKubeconfig(kubeconfigPath string) (*rest.Config, error) {
	// Try in-cluster config first
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}

	// Fall back to kubeconfig file
	if kubeconfigPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		kubeconfigPath = filepath.Join(home, ".kube", "config")
	}

	// Check if file exists
	if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("kubeconfig file not found: %s", kubeconfigPath)
	}

	config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build config from kubeconfig: %w", err)
	}

	return config, nil
}

// GetTypedClient returns the typed Kubernetes client (for compatibility)
func (cm *ClientManager) GetTypedClient() kubernetes.Interface {
	return cm.cluster.Typed()
}
