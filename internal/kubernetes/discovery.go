package kubernetes

import (
	"context"
	"fmt"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
)

// DiscoveryInterface abstracts the discovery client
type DiscoveryInterface interface {
	ServerPreferredResources() ([]*metav1.APIResourceList, error)
	ServerGroups() (*metav1.APIGroupList, error)
}

// DiscoveryClient wraps the Kubernetes discovery client
type DiscoveryClient struct {
	client discovery.DiscoveryInterface
	mu     sync.RWMutex
	cache  *ResourceCache
}

// ResourceCache caches discovered resources
type ResourceCache struct {
	resources map[string]*APIResource
	groups    []metav1.APIGroup
}

// APIResource represents a Kubernetes API resource
type APIResource struct {
	Group        string
	Version      string
	Kind         string
	Resource     string
	Namespaced   bool
	Verbs        []string
	ShortNames   []string
	SingularName string
}

// NewDiscoveryClient creates a new discovery client
func NewDiscoveryClient(config *rest.Config) (*DiscoveryClient, error) {
	client, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery client: %w", err)
	}

	return &DiscoveryClient{
		client: client,
		cache: &ResourceCache{
			resources: make(map[string]*APIResource),
		},
	}, nil
}

// DiscoverResources discovers all available API resources
func (d *DiscoveryClient) DiscoverResources(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Get server groups
	groups, err := d.client.ServerGroups()
	if err != nil {
		return fmt.Errorf("failed to get server groups: %w", err)
	}
	d.cache.groups = groups.Groups

	// Get preferred resources
	resourceLists, err := d.client.ServerPreferredResources()
	if err != nil {
		return fmt.Errorf("failed to get server resources: %w", err)
	}

	// Clear and rebuild cache
	d.cache.resources = make(map[string]*APIResource)

	for _, resourceList := range resourceLists {
		gv, err := schema.ParseGroupVersion(resourceList.GroupVersion)
		if err != nil {
			continue
		}

		for _, resource := range resourceList.APIResources {
			// Skip subresources (e.g., pods/status, pods/log)
			if len(resource.Name) > 0 && resource.Name[len(resource.Name)-1] == '/' {
				continue
			}

			key := resourceKey(gv.Group, gv.Version, resource.Kind)
			d.cache.resources[key] = &APIResource{
				Group:        gv.Group,
				Version:      gv.Version,
				Kind:         resource.Kind,
				Resource:     resource.Name,
				Namespaced:   resource.Namespaced,
				Verbs:        resource.Verbs,
				ShortNames:   resource.ShortNames,
				SingularName: resource.SingularName,
			}
		}
	}

	return nil
}

// GetResources returns all cached resources
func (d *DiscoveryClient) GetResources() []*APIResource {
	d.mu.RLock()
	defer d.mu.RUnlock()

	resources := make([]*APIResource, 0, len(d.cache.resources))
	for _, resource := range d.cache.resources {
		resources = append(resources, resource)
	}
	return resources
}

// GetResource returns a specific resource by group, version, and kind
func (d *DiscoveryClient) GetResource(group, version, kind string) (*APIResource, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	key := resourceKey(group, version, kind)
	resource, exists := d.cache.resources[key]
	return resource, exists
}

// ServerPreferredResources returns the server preferred resources
func (d *DiscoveryClient) ServerPreferredResources() ([]*metav1.APIResourceList, error) {
	return d.client.ServerPreferredResources()
}

// ServerGroups returns the server groups
func (d *DiscoveryClient) ServerGroups() (*metav1.APIGroupList, error) {
	return d.client.ServerGroups()
}

// resourceKey creates a unique key for a resource
func resourceKey(group, version, kind string) string {
	if group == "" {
		return fmt.Sprintf("core/%s/%s", version, kind)
	}
	return fmt.Sprintf("%s/%s/%s", group, version, kind)
}


