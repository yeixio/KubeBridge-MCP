package special

import (
	"context"
	"fmt"
	"io"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// LogsOperation handles getting pod logs
type LogsOperation struct {
	client kubernetes.Interface
}

// NewLogsOperation creates a new logs operation handler
func NewLogsOperation(client kubernetes.Interface) *LogsOperation {
	return &LogsOperation{
		client: client,
	}
}

// GetLogs gets logs from a pod container
func (o *LogsOperation) GetLogs(ctx context.Context, namespace, name, container string, tailLines int64) ([]string, error) {
	// Validate tailLines
	if tailLines <= 0 {
		tailLines = 100
	}
	if tailLines > 1000 {
		tailLines = 1000
	}

	opts := &corev1.PodLogOptions{
		TailLines: &tailLines,
	}

	if container != "" {
		opts.Container = container
	}

	req := o.client.CoreV1().Pods(namespace).GetLogs(name, opts)
	stream, err := req.Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get log stream: %w", err)
	}
	defer stream.Close()

	// Read all logs
	logBytes, err := io.ReadAll(stream)
	if err != nil {
		return nil, fmt.Errorf("failed to read logs: %w", err)
	}

	// Split into lines
	logStr := string(logBytes)
	lines := []string{}
	for _, line := range splitLines(logStr) {
		if line != "" {
			lines = append(lines, line)
		}
	}

	return lines, nil
}

// splitLines splits a string into lines
func splitLines(s string) []string {
	var lines []string
	var current []rune
	for _, r := range s {
		if r == '\n' {
			lines = append(lines, string(current))
			current = nil
		} else {
			current = append(current, r)
		}
	}
	if len(current) > 0 {
		lines = append(lines, string(current))
	}
	return lines
}

