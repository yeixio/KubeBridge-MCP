package special

import (
	"context"
	"fmt"
	"sort"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Event represents a Kubernetes event
type Event struct {
	Type      string    `json:"type"`
	Reason    string    `json:"reason"`
	Message   string    `json:"message"`
	Source    string    `json:"source"`
	FirstSeen time.Time `json:"firstSeen"`
	LastSeen  time.Time `json:"lastSeen"`
	Count     int32     `json:"count"`
}

// EventsOperation handles getting Kubernetes events
type EventsOperation struct {
	client kubernetes.Interface
}

// NewEventsOperation creates a new events operation handler
func NewEventsOperation(client kubernetes.Interface) *EventsOperation {
	return &EventsOperation{
		client: client,
	}
}

// GetEvents gets events for a resource
func (o *EventsOperation) GetEvents(ctx context.Context, namespace, kind, name string) ([]Event, error) {
	// Get all events in the namespace
	eventList, err := o.client.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	events := []Event{}

	// Filter events by kind and name if provided
	for _, event := range eventList.Items {
		if kind != "" && event.InvolvedObject.Kind != kind {
			continue
		}
		if name != "" && event.InvolvedObject.Name != name {
			continue
		}

		events = append(events, Event{
			Type:      event.Type,
			Reason:    event.Reason,
			Message:   event.Message,
			Source:    fmt.Sprintf("%s/%s", event.Source.Component, event.Source.Host),
			FirstSeen: event.FirstTimestamp.Time,
			LastSeen:  event.LastTimestamp.Time,
			Count:     event.Count,
		})
	}

	// Sort by last seen time (most recent first)
	sort.Slice(events, func(i, j int) bool {
		return events[i].LastSeen.After(events[j].LastSeen)
	})

	return events, nil
}

