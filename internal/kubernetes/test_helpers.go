package kubernetes

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	kubernetesfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

// NewFakeClientManager creates a ClientManager with fake Kubernetes clients for testing
func NewFakeClientManager(clusterName string, objects []runtime.Object) (*ClientManager, error) {
	// Create fake typed client
	typedClient := kubernetesfake.NewSimpleClientset(objects...)

	// Create fake dynamic client with proper scheme
	scheme := runtime.NewScheme()
	// Register core v1 types
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add corev1 to scheme: %w", err)
	}
	dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme, objects...)

	// Create a minimal rest config (not used by fake clients)
	config := &rest.Config{
		Host: "https://fake-k8s-server",
	}

	// Create fake discovery client
	discoveryClient, err := NewFakeDiscoveryClient(config, typedClient.Discovery())
	if err != nil {
		return nil, err
	}

	cluster := &Cluster{
		name:      clusterName,
		config:    config,
		typed:     typedClient,
		dynamic:   dynamicClient,
		discovery: discoveryClient,
	}

	return &ClientManager{
		cluster: cluster,
	}, nil
}

// NewFakeDiscoveryClient creates a fake discovery client for testing
func NewFakeDiscoveryClient(config *rest.Config, discoveryInterface discovery.DiscoveryInterface) (*DiscoveryClient, error) {
	// Use the fake discovery interface from the fake client
	discoveryClient := &DiscoveryClient{
		client: discoveryInterface,
		cache: &ResourceCache{
			resources: make(map[string]*APIResource),
		},
	}

	// Pre-populate with common resources for testing
	ctx := context.Background()
	if err := discoveryClient.DiscoverResources(ctx); err != nil {
		// If discovery fails (e.g., with fake client), populate manually
		discoveryClient.cache.resources = map[string]*APIResource{
			"core/v1/Pod": {
				Group:        "",
				Version:      "v1",
				Kind:         "Pod",
				Resource:     "pods",
				Namespaced:   true,
				Verbs:        []string{"get", "list", "watch", "create", "update", "patch", "delete"},
				ShortNames:   []string{"po"},
				SingularName: "pod",
			},
			"core/v1/Service": {
				Group:        "",
				Version:      "v1",
				Kind:         "Service",
				Resource:     "services",
				Namespaced:   true,
				Verbs:        []string{"get", "list", "watch", "create", "update", "patch", "delete"},
				ShortNames:   []string{"svc"},
				SingularName: "service",
			},
			"core/v1/Namespace": {
				Group:        "",
				Version:      "v1",
				Kind:         "Namespace",
				Resource:     "namespaces",
				Namespaced:   false,
				Verbs:        []string{"get", "list", "watch", "create", "update", "patch", "delete"},
				ShortNames:   []string{"ns"},
				SingularName: "namespace",
			},
		}
	}

	return discoveryClient, nil
}
